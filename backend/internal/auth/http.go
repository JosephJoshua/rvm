package auth

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/JosephJoshua/rvm/backend/internal/httputils"
	"github.com/JosephJoshua/rvm/backend/internal/logging"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"
)

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
