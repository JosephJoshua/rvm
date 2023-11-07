package transaction

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/JosephJoshua/rvm/backend/internal/logging"
	"github.com/JosephJoshua/rvm/backend/internal/transaction/domain"
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
	r.Post("/transactions/{transactionID}/items", http.HandlerFunc(handler.addItemToTransaction))

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
	w.WriteHeader(http.StatusCreated)

	if _, err = w.Write([]byte(code)); err != nil {
		oplog.Error("failed to write response", logging.ErrAttr(err))
		w.WriteHeader(http.StatusInternalServerError)

		return
	}
}

func (h *HTTPHandler) addItemToTransaction(w http.ResponseWriter, r *http.Request) {
	oplog := httplog.LogEntry(r.Context())

	transactionIDStr := chi.URLParam(r, "transactionID")
	itemIDStr := r.FormValue("item_id")

	w.Header().Add("Content-Type", "text/plain")

	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		oplog.Error("failed to convert item_id to int", logging.ErrAttr(err))

		w.WriteHeader(http.StatusBadRequest)

		_, err = w.Write([]byte("item_id has to be an integer"))
		if err != nil {
			oplog.Error("failed to write response", logging.ErrAttr(err))
		}

		return
	}

	transactionID, err := domain.NewTransactionID(transactionIDStr)
	if err != nil {
		oplog.Error("failed to create transaction id", logging.ErrAttr(err))
		w.WriteHeader(http.StatusNotFound)

		return
	}

	if err = h.s.AddItemToTransaction(transactionID, itemID); err != nil {
		if errors.Is(err, ErrTransactionDoesNotExist) {
			oplog.Error("transaction not found", slog.String("transaction_id", transactionID.String()))
			w.WriteHeader(http.StatusNotFound)

			return
		}

		if errors.Is(err, ErrItemDoesNotExist) {
			oplog.Error("item not found", slog.Int("item_id", itemID))
			w.WriteHeader(http.StatusNotFound)

			return
		}

		oplog.Error(
			"failed to add item to transaction",
			logging.ErrAttr(err),
			slog.String("transaction_id", transactionID.String()),
			slog.Int("item_id", itemID),
		)

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
