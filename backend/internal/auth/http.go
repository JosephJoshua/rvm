package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/JosephJoshua/rvm/backend/internal/httputils"
	"github.com/JosephJoshua/rvm/backend/internal/logging"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"
)

const (
	bearerPrefix    = "Bearer "
	authHeaderParts = 2
)

type uidCtxKey struct{}

type HTTPHandler struct {
	http.Handler
	s *Service
}

// NewHTTPHandler creates a new auth HTTP handler.
//   - POST /register - registers a new user from the given Firebase id token; does nothing if the user already exists.
func NewHTTPHandler(s *Service) *HTTPHandler {
	handler := &HTTPHandler{s: s}

	r := chi.NewRouter()
	r.Post("/register", httputils.HandlerFunc(handler.register))

	handler.Handler = r
	return handler
}

func (h *HTTPHandler) register(w httputils.ResponseWriter, r *http.Request) {
	oplog := httplog.LogEntry(r.Context())

	idToken := r.FormValue("id_token")

	if idToken == "" {
		oplog.Error("id_token is empty")

		w.WriteHeader(http.StatusBadRequest)
		w.TryWrite(&oplog, []byte("id_token is required"))
	}

	err := h.s.Register(idToken)
	if err != nil {
		if errors.Is(err, ErrUserAlreadyExists) {
			oplog.Info("user already exists", slog.String("id_token", idToken))

			w.WriteHeader(http.StatusConflict)
			return
		}

		if errors.Is(err, ErrInvalidIDToken) {
			oplog.Error("invalid id token", slog.String("id_token", idToken))

			w.WriteHeader(http.StatusBadRequest)
			w.TryWrite(&oplog, []byte("invalid id_token"))

			return
		}

		if errors.Is(err, ErrGetUserFailed) {
			oplog.Error("failed to get user", slog.String("id_token", idToken))

			w.WriteHeader(http.StatusInternalServerError)
			w.TryWrite(&oplog, []byte("failed to get user"))

			return
		}

		oplog.Error("failed to register user", logging.ErrAttr(err), slog.String("id_token", idToken))

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func LoggedInMiddleware(s *Service) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			oplog := httplog.LogEntry(r.Context())

			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			parts := strings.Split(authHeader, bearerPrefix)

			if len(parts) != authHeaderParts {
				oplog.Error("invalid authorization header format")
				w.WriteHeader(http.StatusUnauthorized)

				return
			}

			token := strings.TrimSpace(parts[1])
			user, err := s.GetUser(token)

			if err != nil {
				if errors.Is(err, ErrInvalidIDToken) {
					oplog.Error("invalid token", slog.String("id_token", token))
					w.WriteHeader(http.StatusUnauthorized)

					return
				}

				oplog.Error("failed to get user", slog.String("id_token", token), logging.ErrAttr(err))
				w.WriteHeader(http.StatusInternalServerError)

				return
			}

			ctx := context.WithValue(r.Context(), uidCtxKey{}, user.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UIDFromCtx(ctx context.Context) string {
	return ctx.Value(uidCtxKey{}).(string)
}
