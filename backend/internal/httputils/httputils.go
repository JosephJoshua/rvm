package httputils

import (
	"log/slog"
	"net/http"
)

func HandlerFunc(fn func(ResponseWriter, *http.Request)) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nw := newResponseWriter(w)
		fn(nw, r)
	})
}

type ResponseWriter struct {
	http.ResponseWriter
}

func newResponseWriter(w http.ResponseWriter) ResponseWriter {
	return ResponseWriter{w}
}

func (w *ResponseWriter) TryWrite(oplog *slog.Logger, toWrite []byte) (bool, int) {
	c, err := w.Write(toWrite)
	if err != nil {
		oplog.Error("failed to write response", slog.String("error", err.Error()))
		return false, c
	}

	return true, 0
}
