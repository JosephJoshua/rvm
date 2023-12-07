package apitoken

import (
	"net/http"
	"strings"

	"github.com/JosephJoshua/rvm/backend/internal/logging"
	"github.com/go-chi/httplog/v2"
)

const (
	bearerPrefix    = "Bearer "
	authHeaderParts = 2
)

func ValidTokenMiddleware(s *Service) func(next http.Handler) http.Handler {
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
			ok, err := s.IsValidToken(token)

			if err != nil {
				oplog.Error("failed to validate token", logging.ErrAttr(err))
				w.WriteHeader(http.StatusInternalServerError)

				return
			}

			if !ok {
				oplog.Error("invalid token")
				w.WriteHeader(http.StatusUnauthorized)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
