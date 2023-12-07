package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JosephJoshua/rvm/backend/internal/apitoken"
	"github.com/JosephJoshua/rvm/backend/internal/auth"
	"github.com/JosephJoshua/rvm/backend/internal/db"
	"github.com/JosephJoshua/rvm/backend/internal/env"
	"github.com/JosephJoshua/rvm/backend/internal/firebase"
	"github.com/JosephJoshua/rvm/backend/internal/logging"
	"github.com/JosephJoshua/rvm/backend/internal/transaction"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v2"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

const ReadHeaderTimeoutSecs = 3
const GracefulTimeoutSecs = 30
const CORSMaxAge = 300

func main() {
	loadDotEnv()

	seedFlag := flag.Bool("seed", false, "seeds the database with initial data")
	flag.Parse()

	slog.Default().Info("initializing db..")

	dbHandle, err := db.NewDB(env.GetDBPath())
	if err != nil {
		slog.Default().Error("failed to initialize db", logging.ErrAttr(err))
		return
	}

	slog.Default().Info("initialized db")
	defer dbHandle.Close()

	slog.Default().Info("migrating db schema..")
	if err = db.MigrateDB(dbHandle); err != nil {
		slog.Default().Error("failed to migrate db", logging.ErrAttr(err))
		return
	}

	slog.Default().Info("migrated db schema")

	if *seedFlag {
		slog.Default().Info("seeding db..")

		if err = db.SeedDB(dbHandle); err != nil {
			slog.Default().Error("failed to seed db", logging.ErrAttr(err))
			return
		} else {
			slog.Default().Info("seeded db with initial data")
		}
	}

	firebaseCreds, err := env.GetFirebaseCredentialsJSON()
	if err != nil {
		slog.Default().Error("failed to get firebase creds", logging.ErrAttr(err))
		return
	}

	firebaseApp, err := firebase.NewApp(context.Background(), firebaseCreds)
	if err != nil {
		slog.Default().Error("failed to initialize firebase app", logging.ErrAttr(err))
		return
	}

	server := &http.Server{
		Handler:           getRouter(dbHandle, firebaseApp),
		Addr:              "0.0.0.0:3123",
		ReadHeaderTimeout: ReadHeaderTimeoutSecs * time.Second,
	}

	slog.Default().Info("running server..", slog.String("addr", server.Addr))
	runServer(server)
}

func runServer(server *http.Server) {
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

		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Default().Error("failed to shutdown server", logging.ErrAttr(err))
		}

		stopServerCtx()
	}()

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Default().Error("failed to start server", logging.ErrAttr(err))
	}

	<-serverCtx.Done()
}

func getRouter(dbHandle *sqlx.DB, firebaseApp *firebase.App) http.Handler {
	logger := logging.NewRequestLogger(env.GetAppEnv())

	r := chi.NewRouter()

	r.Use(middleware.StripSlashes)
	r.Use(httplog.RequestLogger(logger, []string{"/ping"}))
	r.Use(middleware.Heartbeat("/ping"))
	r.Use(middleware.SetHeader("Content-Type", "text/plain"))

	r.Use(cors.Handler(cors.Options{
		// TODO: change this to the actual frontend url
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           CORSMaxAge,
	}))

	authService := auth.NewService(
		auth.NewSQLRepository(dbHandle),
		auth.NewFirebaseAuthProvider(firebaseApp),
	)

	transactionService := transaction.NewService(
		transaction.NewSQLRepository(dbHandle),
		transaction.NewUUIDIDGenerator(),
	)

	apiTokenService := apitoken.NewService(
		apitoken.NewSQLRepository(dbHandle),
	)

	transactionHandler := transaction.NewHTTPHandler(transactionService)

	authHandler := auth.NewHTTPHandler(authService)

	r.Mount("/auth", authHandler)

	r.Group(func(r chi.Router) {
		r.Use(apitoken.ValidTokenMiddleware(apiTokenService))
		r.Mount("/transactions", transactionHandler)
	})

	return r
}

func loadDotEnv() {
	if err := godotenv.Load(); err != nil {
		slog.Default().Warn("failed to load .env file", logging.ErrAttr(err))
	}
}
