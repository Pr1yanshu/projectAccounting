package db

import (
	"context"
	"errors"
	"time"

	"accounting/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type DB struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *DB { return &DB{pool: pool} }

type Store interface {
	CreateAccount(ctx context.Context, accountID int64, initial decimal.Decimal) error
	GetAccount(ctx context.Context, accountID int64) (decimal.Decimal, error)
	Transfer(ctx context.Context, srcID, dstID int64, amount decimal.Decimal) error
	ListAccounts(ctx context.Context) ([]models.Account, error)
	ListTransactions(ctx context.Context, accountID int64) ([]models.Transaction, error)
}

var _ Store = (*DB)(nil)

var (
	ErrAccountNotFound   = errors.New("account not found")
	ErrInsufficientFunds = errors.New("insufficient funds")
)

func (d *DB) CreateAccount(ctx context.Context, accountID int64, initial decimal.Decimal) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	_, err := d.pool.Exec(ctx, `INSERT INTO accounts(id, balance) VALUES($1, $2)`, accountID, initial.String())
	return err
}

func (d *DB) GetAccount(ctx context.Context, accountID int64) (decimal.Decimal, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var balStr string
	err := d.pool.QueryRow(ctx, `SELECT balance FROM accounts WHERE id=$1`, accountID).Scan(&balStr)
	if err != nil {
		return decimal.Zero, ErrAccountNotFound
	}
	bal, err := decimal.NewFromString(balStr)
	if err != nil {
		return decimal.Zero, err
	}
	return bal, nil
}

// Transfer moves amount from source to destination atomically.
func (d *DB) Transfer(ctx context.Context, srcID, dstID int64, amount decimal.Decimal) error {
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Lock source row
	var srcBalStr string
	if err := tx.QueryRow(ctx, `SELECT balance FROM accounts WHERE id=$1 FOR UPDATE`, srcID).Scan(&srcBalStr); err != nil {
		return ErrAccountNotFound
	}
	srcBal, err := decimal.NewFromString(srcBalStr)
	if err != nil {
		return err
	}

	// Lock destination row
	var dstBalStr string
	if err := tx.QueryRow(ctx, `SELECT balance FROM accounts WHERE id=$1 FOR UPDATE`, dstID).Scan(&dstBalStr); err != nil {
		return ErrAccountNotFound
	}
	dstBal, err := decimal.NewFromString(dstBalStr)
	if err != nil {
		return err
	}

	if srcBal.Cmp(amount) < 0 {
		return ErrInsufficientFunds
	}

	newSrc := srcBal.Sub(amount)
	newDst := dstBal.Add(amount)

	if _, err := tx.Exec(ctx, `UPDATE accounts SET balance=$1 WHERE id=$2`, newSrc.String(), srcID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE accounts SET balance=$1 WHERE id=$2`, newDst.String(), dstID); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `INSERT INTO transactions(source_account_id, destination_account_id, amount) VALUES($1,$2,$3)`, srcID, dstID, amount.String()); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (d *DB) ListAccounts(ctx context.Context) ([]models.Account, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	rows, err := d.pool.Query(ctx, `SELECT id, balance FROM accounts`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Account
	for rows.Next() {
		var id int64
		var balStr string
		if err := rows.Scan(&id, &balStr); err != nil {
			return nil, err
		}
		bal, err := decimal.NewFromString(balStr)
		if err != nil {
			return nil, err
		}
		out = append(out, models.Account{AccountID: id, Balance: bal})
	}
	return out, nil
}

func (d *DB) ListTransactions(ctx context.Context, accountID int64) ([]models.Transaction, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	rows, err := d.pool.Query(ctx, `SELECT id, source_account_id, destination_account_id, amount FROM transactions WHERE source_account_id=$1 OR destination_account_id=$1 ORDER BY id DESC`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Transaction
	for rows.Next() {
		var id int64
		var srcID int64
		var dstID int64
		var amtStr string
		if err := rows.Scan(&id, &srcID, &dstID, &amtStr); err != nil {
			return nil, err
		}
		amt, err := decimal.NewFromString(amtStr)
		if err != nil {
			return nil, err
		}
		out = append(out, models.Transaction{
			ID:                   id,
			SourceAccountID:      srcID,
			DestinationAccountID: dstID,
			Amount:               amt,
		})
	}
	return out, nil
}
