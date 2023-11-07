package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v2"
)

type AppEnv string

const (
	AppEnvProduction  = "production"
	AppEnvDevelopment = "development"
)

const ReadHeaderTimeoutSecs = 3

func getAppEnv() string {
	env := os.Getenv("APP_ENV")
	if env == "" {
		return AppEnvProduction
	}

	return AppEnvDevelopment
}

func main() {
	logger := httplog.NewLogger("rvm-backend", httplog.Options{
		LogLevel:         slog.LevelDebug,
		JSON:             getAppEnv() == AppEnvProduction,
		RequestHeaders:   true,
		ResponseHeaders:  true,
		SourceFieldName:  "source",
		MessageFieldName: "message",
		LevelFieldName:   "severity",
		TimeFieldFormat:  time.RFC3339,
		Tags: map[string]string{
			"APP_ENV":  getAppEnv(),
			"APP_NAME": "rvm-backend",
			"VERSION":  "v0.1",
		},
	})

	r := chi.NewRouter()
	r.Use(httplog.RequestLogger(logger, []string{"/ping"}))
	r.Use(middleware.Heartbeat("/ping"))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		oplog := httplog.LogEntry(r.Context())
		if _, err := w.Write([]byte("Hello World")); err != nil {
			oplog.Error("failed to write response", httplog.ErrAttr(err))
		}
	})

	server := http.Server{
		Handler:           r,
		Addr:              ":3123",
		ReadHeaderTimeout: ReadHeaderTimeoutSecs * time.Second,
	}

	log.Fatal(server.ListenAndServe())
}
