package config

import (
	"github.com/rs/zerolog"
	"os"
	"time"
)

type TimestampFormatter struct{}

func NewZeroLog() zerolog.Logger {
	logger := zerolog.New(os.Stdout).With().Caller().Logger().Hook(&TimestampFormatter{}).Level(zerolog.DebugLevel)

	return logger
}

func (t *TimestampFormatter) Run(e *zerolog.Event, _ zerolog.Level, _ string) {
	e.Str("time", time.Now().Format("2006-01-02 15:04:05"))
}
