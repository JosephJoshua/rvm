package logging

import (
	"log/slog"
	"time"

	"github.com/JosephJoshua/rvm/backend/internal/env"
	"github.com/go-chi/httplog/v2"
)

func NewRequestLogger(appEnv env.AppEnv) *httplog.Logger {
	return httplog.NewLogger("rvm-backend", httplog.Options{
		LogLevel:         slog.LevelDebug,
		JSON:             appEnv == env.AppEnvProduction,
		RequestHeaders:   true,
		ResponseHeaders:  true,
		SourceFieldName:  "source",
		MessageFieldName: "message",
		LevelFieldName:   "severity",
		TimeFieldFormat:  time.RFC3339,
		Tags: map[string]string{
			"APP_ENV":  string(appEnv),
			"APP_NAME": "rvm-backend",
			"VERSION":  "v0.1",
		},
	})
}

func ErrAttr(err error) slog.Attr {
	return slog.Any("err", err)
}
