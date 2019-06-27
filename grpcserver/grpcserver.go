package grpcserver

import (
	"fmt"
	"net"

	mw "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	gr "google.golang.org/grpc"

	"github.com/yeonfeel/gopkg/er"
	"github.com/yeonfeel/gopkg/logger"
)

var log = logger.Get("server")

type grpcServer struct {
	port   int
	listen net.Listener
	server *gr.Server
}

// RegisterService add a service to server
type RegisterService func(*gr.Server)

// Run create and run the grpc server
func Run(errchan chan<- error, port int, registerService RegisterService) {
	var err error

	s := &grpcServer{}
	s.port = port

	address := fmt.Sprintf("0.0.0.0:%d", port)
	s.listen, err = net.Listen("tcp", address)
	if err != nil {
		errchan <- er.Error(err, "Failed to listen")
		return
	}

	log.InfoF("listen address: %s", address)

	opts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(recoveryHandlerFunc),
	}

	s.server = gr.NewServer(
		gr.StreamInterceptor(mw.ChainStreamServer(
			grpc_recovery.StreamServerInterceptor(opts...),
		)),
		gr.UnaryInterceptor(mw.ChainUnaryServer(
			grpc_recovery.UnaryServerInterceptor(opts...),
		)),
	)

	registerService(s.server)

	go s.Run(errchan)
}

func (s *grpcServer) Run(errchan chan<- error) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("the grpc server shuts down with an unexpected errror", r)
		}
	}()

	if err := s.server.Serve(s.listen); err != nil {
		errchan <- er.Error(err, "failed to serve")
	}
}

func recoveryHandlerFunc(p interface{}) (err error) {
	log.Error(er.ErrorS("recoverd error! %s", p))

	return fmt.Errorf(fmt.Sprint(p))
}
