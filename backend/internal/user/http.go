package user

import (
	"net/http"
	"strconv"

	"github.com/JosephJoshua/rvm/backend/internal/auth"
	"github.com/JosephJoshua/rvm/backend/internal/httputils"
	"github.com/JosephJoshua/rvm/backend/internal/logging"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"
)

type HTTPHandler struct {
	http.Handler
	s *Service
}

// NewHTTPHandler creates a new user HTTP handler.
//   - GET /points - returns this user's points.
func NewHTTPHandler(s *Service) *HTTPHandler {
	handler := &HTTPHandler{s: s}

	r := chi.NewRouter()

	r.Get("/points", httputils.HandlerFunc(handler.getPoints))

	handler.Handler = r
	return handler
}

func (h *HTTPHandler) getPoints(w httputils.ResponseWriter, r *http.Request) {
	oplog := httplog.LogEntry(r.Context())

	uid := auth.UIDFromCtx(r.Context())

	p, err := h.s.GetPoints(uid)
	if err != nil {
		oplog.Error("failed to get points", logging.ErrAttr(err))
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.TryWrite(&oplog, []byte(strconv.Itoa(p)))
}
