package transaction

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/JosephJoshua/rvm/backend/internal/httputils"
	"github.com/JosephJoshua/rvm/backend/internal/logging"
	"github.com/JosephJoshua/rvm/backend/internal/transaction/domain"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v2"
)

type HTTPHandler struct {
	http.Handler
	s *Service
}

// NewHTTPHandler creates a new transaction HTTP handler.
//   - POST /transactions - starts a new transaction and returns the transaction code.
//   - POST /transactions/{transactionID}/items - adds an item to the transaction.
//     item_id is a form value or query parameter.
//   - DELETE /transactions/{transactionID} - ends the transaction and assigns the user to the transaction.
//     user_id is a form value or query parameter.
func NewHTTPHandler(s *Service) *HTTPHandler {
	handler := &HTTPHandler{s: s}

	r := chi.NewRouter()

	r.Post("/", httputils.HandlerFunc(handler.startTransaction))
	r.Post("/{transactionID}/items", httputils.HandlerFunc(handler.addItemToTransaction))
	r.Post("/{transactionID}/end", httputils.HandlerFunc(handler.endTransactionAndAssignUser))

	handler.Handler = r
	return handler
}

func (h *HTTPHandler) startTransaction(w httputils.ResponseWriter, r *http.Request) {
	oplog := httplog.LogEntry(r.Context())

	code, err := h.s.StartTransaction()
	if err != nil {
		oplog.Error("failed to start transaction", logging.ErrAttr(err))
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusCreated)
	w.TryWrite(&oplog, []byte(code))
}

func (h *HTTPHandler) addItemToTransaction(w httputils.ResponseWriter, r *http.Request) {
	oplog := httplog.LogEntry(r.Context())

	transactionIDStr := chi.URLParam(r, "transactionID")
	itemIDStr := r.FormValue("item_id")

	if itemIDStr == "" {
		oplog.Error("item_id is empty")

		w.WriteHeader(http.StatusBadRequest)
		w.TryWrite(&oplog, []byte("item_id is required"))
	}

	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		oplog.Error("failed to convert item_id to int", logging.ErrAttr(err))

		w.WriteHeader(http.StatusBadRequest)
		w.TryWrite(&oplog, []byte("item_id has to be an integer"))

		return
	}

	transactionID, err := domain.NewTransactionID(transactionIDStr)
	if err != nil {
		oplog.Error("failed to create transaction id", logging.ErrAttr(err))
		w.WriteHeader(http.StatusNotFound)

		return
	}

	c, err := h.s.AddItemToTransaction(transactionID, itemID)
	if err != nil {
		if errors.Is(err, ErrTransactionDoesNotExist) {
			oplog.Error("transaction not found", slog.String("transaction_id", transactionID.String()))
			w.WriteHeader(http.StatusNotFound)

			return
		}

		if errors.Is(err, ErrItemDoesNotExist) {
			oplog.Error("item not found", slog.Int("item_id", itemID))

			w.WriteHeader(http.StatusBadRequest)
			w.TryWrite(&oplog, []byte("item not found"))

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

	w.TryWrite(&oplog, []byte(strconv.Itoa(c)))
}

func (h *HTTPHandler) endTransactionAndAssignUser(w httputils.ResponseWriter, r *http.Request) {
	oplog := httplog.LogEntry(r.Context())

	transactionIDStr := chi.URLParam(r, "transactionID")
	userID := r.FormValue("user_id")

	if userID == "" {
		oplog.Error("user_id is empty")

		w.WriteHeader(http.StatusBadRequest)
		w.TryWrite(&oplog, []byte("user_id is required"))

		return
	}

	transactionID, err := domain.NewTransactionID(transactionIDStr)
	if err != nil {
		oplog.Error("failed to create transaction id", logging.ErrAttr(err))
		w.WriteHeader(http.StatusNotFound)

		return
	}

	c, err := h.s.EndTransactionAndAssignUser(transactionID, userID)
	if err != nil {
		if errors.Is(err, ErrTransactionAlreadyAssigned) {
			oplog.Error("transaction is already assigned", slog.String("transaction_id", transactionIDStr))
			w.WriteHeader(http.StatusConflict)

			return
		}

		if errors.Is(err, ErrTransactionDoesNotExist) {
			oplog.Error("transaction not found", slog.String("transaction_id", transactionID.String()))
			w.WriteHeader(http.StatusNotFound)

			return
		}

		if errors.Is(err, ErrUserDoesNotExist) {
			oplog.Error("user not found", slog.String("user_id", userID))

			w.WriteHeader(http.StatusBadRequest)
			w.TryWrite(&oplog, []byte("user not found"))

			return
		}

		oplog.Error(
			"failed to end transaction",
			logging.ErrAttr(err),
			slog.String("transaction_id", transactionID.String()),
			slog.String("user_id", userID),
		)

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.TryWrite(&oplog, []byte(strconv.Itoa(c)))
}
