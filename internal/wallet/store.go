package wallet

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rishabhkum-Dev-1799/go-payments/internal/wallet/walletdb"
)

// creating the wrapper types so we dont need to directly import from the walletdb ...r
type (
	Account        = walletdb.Account
	AccountBalance = walletdb.AccountBalance
	Entry          = walletdb.Entry
	NewAccount     = walletdb.CreateAccountParams
	Direction      = walletdb.EntryDirection
)

const (
	Debit  = walletdb.EntryDirectionDebit
	Credit = walletdb.EntryDirectionCredit
)

// Defining the common error variables here for the modular uses  ...r
var ErrAccountNotFound = errors.New("wallet: account not found")

// store is the repository layer which calls the sqlc queries functions ...
type Store struct {
	pool *pgxpool.Pool
	q    *walletdb.Queries
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool, q: walletdb.New(pool)}
}

// defining the methods to interact with the entry and account databases ...r
func (s *Store) CreateAccount(ctx context.Context, a NewAccount) (Account, error) {
	acc, err := s.q.CreateAccount(ctx, a)
	if err != nil {
		return Account{}, fmt.Errorf("wallet:create account : %w", err)
	}
	return acc, nil
}

// Account returns the account based on the given id ...r
func (s *Store) Account(ctx context.Context, id uuid.UUID) (Account, error) {
	acc, err := s.q.GetAccount(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Account{}, ErrAccountNotFound
	}
	if err != nil {
		return Account{}, fmt.Errorf("wallet:get account %s:%w", id, err)
	}
	return acc, nil
}

// Balance is the function which will return the account balance which is computed from the account_balances_view which is computed from the entries
func (s *Store) Balance(ctx context.Context, id uuid.UUID) (AccountBalance, error) {
	accBal, err := s.q.GetAccountBalance(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return AccountBalance{}, ErrAccountNotFound
	}
	if err != nil {
		return AccountBalance{}, fmt.Errorf("wallet:account balance %s:%w", id, err)
	}
	return accBal, nil
}

// Entries return the page of the entries newest first pass the before = 0 if you want the first page and after that send the last entry id of the page
func (s *Store) Entries(ctx context.Context, id uuid.UUID, before int64, limit int32) ([]Entry, error) {
	if before <= 0 {
		before = math.MaxInt64
	}
	entries, err := s.q.ListEntriesByAccount(ctx, walletdb.ListEntriesByAccountParams{
		AccountID: id,
		BeforeID:  before,
		PageSize:  limit,
	})
	if err != nil {
		return []Entry{}, fmt.Errorf("wallet: get entries %s:%w", id, err)
	}
	return entries, nil
}
