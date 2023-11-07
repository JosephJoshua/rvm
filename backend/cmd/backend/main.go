package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JosephJoshua/rvm/backend/internal/db"
	"github.com/JosephJoshua/rvm/backend/internal/env"
	"github.com/JosephJoshua/rvm/backend/internal/logging"
	"github.com/JosephJoshua/rvm/backend/internal/transaction"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v2"
	"github.com/joho/godotenv"
)

const ReadHeaderTimeoutSecs = 3
const GracefulTimeoutSecs = 30

func main() {
	loadDotEnv()

	dbHandle, err := db.NewDB(env.GetDBPath())
	if err != nil {
		slog.Default().Error("failed to initialize db", logging.ErrAttr(err))
	}

	defer dbHandle.Close()

	if err = db.MigrateDB(dbHandle); err != nil {
		slog.Default().Error("failed to migrate db", logging.ErrAttr(err))
	}

	server := http.Server{
		Handler:           getRouter(dbHandle),
		Addr:              "0.0.0.0:3123",
		ReadHeaderTimeout: ReadHeaderTimeoutSecs * time.Second,
	}

	serverCtx, stopServerCtx := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sig

		shutdownCtx, stopShutdownCtx := context.WithTimeout(serverCtx, GracefulTimeoutSecs*time.Second)
		defer stopShutdownCtx()

		go func() {
			<-shutdownCtx.Done()

			if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
				slog.Default().Error("graceful shutdown timed out. forcing exit..")
			}
		}()

		if err = server.Shutdown(shutdownCtx); err != nil {
			slog.Default().Error("failed to shutdown server", logging.ErrAttr(err))
		}

		stopServerCtx()
	}()

	if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Default().Error("failed to start server", logging.ErrAttr(err))
	}

	<-serverCtx.Done()
}

func getRouter(dbHandle *sql.DB) http.Handler {
	logger := logging.NewRequestLogger(env.GetAppEnv())

	r := chi.NewRouter()
	r.Use(httplog.RequestLogger(logger, []string{"/ping"}))
	r.Use(middleware.Heartbeat("/ping"))

	transactionService := transaction.NewService(
		transaction.NewSQLRepository(dbHandle),
		transaction.NewUUIDCodeGenerator(),
	)

	transactionHandler := transaction.NewHTTPHandler(transactionService)

	r.Mount("/", transactionHandler)

	return r
}

func loadDotEnv() {
	if err := godotenv.Load(); err != nil {
		slog.Default().Warn("failed to load .env file", logging.ErrAttr(err))
	}
}
