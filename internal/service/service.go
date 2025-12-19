package service

import (
	"context"
	"errors"

	"accounting/internal/db"
	"accounting/internal/models"

	"github.com/shopspring/decimal"
)

var (
	ErrInvalidAmount  = errors.New("invalid amount")
	ErrInvalidBalance = errors.New("invalid balance")
)

type Service struct {
	store db.Store
}

func New(s db.Store) *Service { return &Service{store: s} }

func (s *Service) CreateAccount(ctx context.Context, req models.CreateAccountRequest) error {
	bal, err := decimal.NewFromString(req.InitialBalance)
	if err != nil {
		return ErrInvalidBalance
	}
	return s.store.CreateAccount(ctx, req.AccountID, bal)
}

func (s *Service) GetAccount(ctx context.Context, id int64) (decimal.Decimal, error) {
	return s.store.GetAccount(ctx, id)
}

func (s *Service) GetAllAccounts(ctx context.Context) ([]models.Account, error) {
	return s.store.ListAccounts(ctx)
}

func (s *Service) GetTransactions(ctx context.Context, accountID int64) ([]models.Transaction, error) {
	return s.store.ListTransactions(ctx, accountID)
}

func (s *Service) CreateTransaction(ctx context.Context, req models.CreateTransactionRequest) error {
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return ErrInvalidAmount
	}
	return s.store.Transfer(ctx, req.SourceAccountID, req.DestinationAccountID, amount)
}
