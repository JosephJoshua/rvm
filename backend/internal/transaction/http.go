package transaction

import (
	"net/http"

	"github.com/JosephJoshua/rvm/backend/internal/logging"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"
)

type HTTPHandler struct {
	http.Handler
	s *Service
}

func NewHTTPHandler(s *Service) *HTTPHandler {
	handler := &HTTPHandler{s: s}

	r := chi.NewRouter()
	r.Post("/transactions", http.HandlerFunc(handler.startTransaction))

	handler.Handler = r
	return handler
}

func (h *HTTPHandler) startTransaction(w http.ResponseWriter, r *http.Request) {
	oplog := httplog.LogEntry(r.Context())

	code, err := h.s.StartTransaction()
	if err != nil {
		oplog.Error("failed to start transaction", logging.ErrAttr(err))
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Add("Content-Type", "text/plain")

	if _, err = w.Write([]byte(code)); err != nil {
		oplog.Error("failed to write response", logging.ErrAttr(err))
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusCreated)
}
