package logger

import (
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github/zqr233qr/story-trim/pkg/config"
)

func Init(cfg config.LogConfig) {
	// 1. 设置时间格式
	zerolog.TimeFieldFormat = "2006-01-02 15:04:05"

	// 2. 决定输出目的地
	var out io.Writer = os.Stdout
	if cfg.Format == "console" {
		out = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "2006-01-02 15:04:05",
		}
	}

	log.Logger = zerolog.New(out).With().Timestamp().Logger()

	// 3. 设置级别
	level, err := zerolog.ParseLevel(strings.ToLower(cfg.Level))
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	log.Info().Str("level", level.String()).Msg("Logger initialized")
}
