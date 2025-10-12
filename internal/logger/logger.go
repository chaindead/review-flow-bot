package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.ErrorStackMarshaler = marshalStack

	writer := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.TimeOnly}

	log.Logger = log.Output(writer).With().Caller().Stack().Logger().Level(zerolog.DebugLevel)
}

//func Setup() error {
//	level, err := zerolog.ParseLevel(*levelStr)
//	if err != nil {
//		return errors.Wrapf(err, "invalid log level(%s)", *levelStr)
//	}
//
//	log.Logger = log.Logger.Level(level)
//
//	return nil
//}
