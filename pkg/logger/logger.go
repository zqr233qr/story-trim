package logger

import (
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/zqr233qr/story-trim/pkg/config"
)

func Init(cfg config.LogConfig) {
	zerolog.TimeFieldFormat = "2006-01-02 15:04:05"

	var out io.Writer = os.Stdout
	if cfg.Format == "console" {
		out = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "2006-01-02 15:04:05",
		}
	}

	log.Logger = zerolog.New(out).With().Timestamp().Logger()

	level, err := zerolog.ParseLevel(strings.ToLower(cfg.Level))
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	log.Info().Str("level", level.String()).Msg("Logger initialized")
}

func Debug() *zerolog.Event {
	return log.Debug()
}

func Info() *zerolog.Event {
	return log.Info()
}

func Warn() *zerolog.Event {
	return log.Warn()
}

func Error() *zerolog.Event {
	return log.Error()
}
