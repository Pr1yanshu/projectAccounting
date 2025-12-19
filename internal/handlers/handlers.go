package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"accounting/internal/db"
	"accounting/internal/models"
	"accounting/internal/service"

	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
)

type Server struct {
	svc *service.Service
}

func NewServer(svc *service.Service) *Server { return &Server{svc: svc} }

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func (s *Server) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req models.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// validate initial balance is a valid non-negative decimal
	bal, err := decimal.NewFromString(req.InitialBalance)
	if err != nil {
		http.Error(w, "invalid balance", http.StatusBadRequest)
		return
	}
	if bal.IsNegative() {
		http.Error(w, "negative balance not allowed", http.StatusBadRequest)
		return
	}

	if err := s.svc.CreateAccount(r.Context(), req); err != nil {
		if err == service.ErrInvalidBalance {
			http.Error(w, "invalid balance", http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "account created successfully"})
}

func (s *Server) GetAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	bal, err := s.svc.GetAccount(r.Context(), id)
	if err != nil {
		http.Error(w, "account not found", http.StatusNotFound)
		return
	}

	resp := map[string]string{
		"account_id": strconv.FormatInt(id, 10),
		"balance":    bal.String(),
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) GetAllAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := s.svc.GetAllAccounts(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, accounts)
}

func (s *Server) GetTransactions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	txs, err := s.svc.GetTransactions(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, txs)
}

func (s *Server) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := s.svc.CreateTransaction(r.Context(), req); err != nil {
		switch err {
		case service.ErrInvalidAmount:
			http.Error(w, "invalid amount", http.StatusBadRequest)
		case db.ErrAccountNotFound:
			http.Error(w, "account not found", http.StatusNotFound)
		case db.ErrInsufficientFunds:
			http.Error(w, "insufficient funds", http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "transaction completed successfully"})
}
