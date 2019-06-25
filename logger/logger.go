package logger

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger provide logging functions
type Logger interface {
	Info(a ...interface{})
	InfoF(format string, a ...interface{})
	Error(a ...interface{})
	ErrorF(format string, a ...interface{})
	Warn(a ...interface{})
	WarnF(format string, a ...interface{})
}

// type Closer interface {
// 	Close() error
// }

type logdata struct {
	name string
}

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

// Get returns a new logger
func Get(name string) Logger {
	return &logdata{name}
}

// func File(filepath string) (Closer, error) {
// 	f, err := os.Create(filepath)
// 	if err != nil {
// 		return nil, err
// 	}

// 	log.Logger = log.Output(zerolog.ConsoleWriter{Out: f})
// 	return f, nil
// }

// func LogLevel(level string) {
// 	switch level {
// 	case "panic":
// 		zerolog.SetGlobalLevel(zerolog.PanicLevel)
// 	}
// }

func (l *logdata) Info(a ...interface{}) {
	log.Info().Str("name", l.name).Msg(fmt.Sprint(a...))
}

func (l *logdata) InfoF(format string, a ...interface{}) {
	log.Info().Str("name", l.name).Msg(fmt.Sprintf(format, a...))
}

func (l *logdata) Warn(a ...interface{}) {
	log.Warn().Str("name", l.name).Msg(fmt.Sprint(a...))
}

func (l *logdata) WarnF(format string, a ...interface{}) {
	log.Warn().Str("name", l.name).Msg(fmt.Sprintf(format, a...))
}

func (l *logdata) Error(a ...interface{}) {
	log.Error().Str("name", l.name).Msg(fmt.Sprint(a...))
}

func (l *logdata) ErrorF(format string, a ...interface{}) {
	log.Error().Str("name", l.name).Msg(fmt.Sprintf(format, a...))
}
