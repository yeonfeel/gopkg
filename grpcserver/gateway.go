package grpcserver

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"path"
	"strings"
	"time"

	mw "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"

	"github.com/yeonfeel/gopkg/er"
)

type gatewayServer struct {
	grpcPort                int
	httpPort                int
	registerService         RegisterService
	registerEndpoints       map[RegisterServiceURLPattern]RegisterServiceFromEndpoint
	gracefulShutdownTimeout time.Duration
	ssl                     bool
	sslCertPath             string
	sslKeyPath              string
}

type RegisterServiceURLPattern string
type RegisterServiceFromEndpoint func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error

// RunGateway create and run the grpc-gateway server
func RunGateway(ctx context.Context, grpcDone chan<- struct{}, httpDone chan<- struct{}, grpcPort, httpPort int,
	ssl bool, sslCertPath, sslKeyPath string, gracefulShutdownTimeout time.Duration,
	registerService RegisterService, registerEndpoints map[RegisterServiceURLPattern]RegisterServiceFromEndpoint,
	interceptors ...grpc.UnaryServerInterceptor,
) error {
	log.Info("Starting a grpc-gateway")

	s := &gatewayServer{
		grpcPort:                grpcPort,
		httpPort:                httpPort,
		registerService:         registerService,
		registerEndpoints:       registerEndpoints,
		gracefulShutdownTimeout: gracefulShutdownTimeout,
		ssl:         ssl,
		sslCertPath: sslCertPath,
		sslKeyPath:  sslKeyPath,
	}

	if err := s.runGRPC(ctx, grpcDone, interceptors...); err != nil {
		return err
	}

	if err := s.runHTTP(ctx, httpDone); err != nil {
		return err
	}

	return nil
}

func (s *gatewayServer) runGRPC(ctx context.Context, done chan<- struct{}, interceptors ...grpc.UnaryServerInterceptor) error {
	log.InfoF("Starting grpc server on port:%d", s.grpcPort)

	opts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(recoveryHandlerFunc),
	}

	interceptorOpts := []grpc.UnaryServerInterceptor{}
	interceptorOpts = append(interceptorOpts, grpc_recovery.UnaryServerInterceptor(opts...))

	for _, interceptor := range interceptors {
		interceptorOpts = append(interceptorOpts, interceptor)
	}

	serverOptions := []grpc.ServerOption{
		grpc.StreamInterceptor(mw.ChainStreamServer(grpc_recovery.StreamServerInterceptor(opts...))),
		grpc.UnaryInterceptor(mw.ChainUnaryServer(interceptorOpts...)),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime: 30 * time.Second,
		}),
	}

	// Create the TLS credentials
	if s.ssl {
		creds, err := credentials.NewServerTLSFromFile(s.sslCertPath, s.sslKeyPath)
		if err != nil {
			return er.Error(err, "failed to load TLS keys")
		}
		serverOptions = append(serverOptions, grpc.Creds(creds))
	}

	server := grpc.NewServer(serverOptions...)

	// register service
	s.registerService(server)

	// start listening for grpc
	listen, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", s.grpcPort))
	if err != nil {
		return er.Error(err, "can't listen a tcp")
	}

	// Start the gRPC server in goroutine
	go func() {
		defer func() {
			close(done)
			log.Warn("grpc server was stopped")
		}()

		if err := server.Serve(listen); err != nil {
			if err != grpc.ErrServerStopped {
				log.Error(er.ErrorF(err, "grpc server failed to listen"))
			}
			return
		}
	}()

	go func() {
		<-ctx.Done()
		if err := ctx.Err(); err != nil {
			if err == context.Canceled {
				log.Info("context was canceled")
			} else {
				log.Error(er.ErrorF(err, "context was error"))
			}
		}

		stopped := make(chan struct{})
		go func() {
			server.GracefulStop()
			close(stopped)
		}()

		select {
		case <-stopped:
		case <-time.After(s.gracefulShutdownTimeout):
			log.WarnF("graceful shutdown was timeout (%s)", s.gracefulShutdownTimeout)
			server.Stop()
		}

		log.Warn("stop the grpc server")
	}()

	return nil
}

func (s *gatewayServer) runHTTP(ctx context.Context, done chan<- struct{}) error {
	// Start the HTTP server for REST
	log.InfoF("Starting HTTP server on port:%d", s.httpPort)

	//gw := runtime.NewServeMux()
	gw := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{EmitDefaults: true}))
	opts := []grpc.DialOption{grpc.WithInsecure()}

	mux := http.NewServeMux()
	mux.HandleFunc("/swagger/", serveSwagger)

	endpoint := fmt.Sprintf("localhost:%d", s.grpcPort)
	for pattern, registerEndpoint := range s.registerEndpoints {
		err := registerEndpoint(ctx, gw, endpoint, opts)
		if err != nil {
			return er.Error(err, "can't register a service to RESTful API")
		}

		mux.Handle(string(pattern), gw)
		log.InfoF("Register service: %s", string(pattern))
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.httpPort),
		Handler: mux,
	}

	// Start the HTTP server in goroutine
	go func() {
		defer func() {
			close(done)
			log.Warn("http server was stopped")
		}()

		if err := server.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Error(er.ErrorF(err, "http server failed to listen"))
			}
			return
		}
	}()

	go func() {
		<-ctx.Done()
		if err := ctx.Err(); err != nil {
			if err == context.Canceled {
				log.Info("context was canceled")
			} else {
				log.Error(er.ErrorF(err, "context was error"))
			}
		}

		tctx, cancel := context.WithTimeout(ctx, s.gracefulShutdownTimeout)
		defer cancel()

		if err := server.Shutdown(tctx); err != nil {
			log.Error(er.Error(err, "failed to shutdown the http server"))
		}
		log.Warn("stop the http server")
	}()

	return nil
}

func serveSwagger(w http.ResponseWriter, r *http.Request) {
	fmt.Println("request", r.URL.Path)
	p := strings.TrimPrefix(r.URL.Path, "/swagger/")
	p = path.Join("third-party/swagger-ui", p)
	http.ServeFile(w, r, p)
}
