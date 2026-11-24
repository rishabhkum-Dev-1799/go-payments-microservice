package wallet

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rishabhkum-Dev-1799/go-payments/internal/wallet/walletdb"
)

type Posting struct {
	AccountID uuid.UUID
	Direction Direction
	Amount    int64
	Currency  string
}

// the struct to represent new transaction
type NewTransaction struct {
	Description string
	Postings    []Posting
}

// transaction is the posted journal entry along with its transactions
type Transaction struct {
	walletdb.Transaction
	Entries []Entry `json:entries`
}

// inTx is the function which will run the transactions of queries in the single pool connection so rolling back will be easy , in this function it will return nil and if some error occurs then the transaction will be roll backed
func (s *Store) inTx(ctx context.Context, fn func(q *walletdb.Queries) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("transaction:begin %w", err)
	}
	// so this defer will run after the function is finished ,but rollback would happen if some error or panic happens this is fail safe and very important thing to be taken care of ...r
	defer tx.Rollback(ctx)

	// we need to run the queries function here and check for the error
	if err := fn(s.q.WithTx(tx)); err != nil {
		return err
	}
	// now after that when the queries are runned lets start the commit deffered constraints like the check balance is trigger here and catch the error in the commit if any ...r
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("transaction:commit %w", err)
	}
	// final return
	return nil
}

// Post transaction records every transaction and entries and checks for the balances for the credit and debit if it aligns then it will store else it will reject the commit
func (s *Store) PostTransaction(ctx context.Context, nt NewTransaction) (Transaction, error) {
	var posted Transaction
	err := s.inTx(ctx, func(q *walletdb.Queries) error {
		// first created the transaction ...r
		journal, err := q.CreateTransaction(ctx, nt.Description)
		if err != nil {
			return fmt.Errorf("insert transaction %w", err)
		}
		posted.Transaction = journal
		// second adding the entries in to the transaction ...r
		for _, p := range nt.Postings {
			entry, err := q.CreateEntry(ctx, walletdb.CreateEntryParams{
				TransactionID: journal.ID,
				AccountID:     p.AccountID,
				Direction:     p.Direction,
				Currency:      p.Currency,
				Amount:        p.Amount,
			})
			if err != nil {
				return fmt.Errorf("entry transaction error :err")
			}
			posted.Entries = append(posted.Entries, entry)
		}
		// at the end return nil
		return nil
	})
	if err != nil {
		return Transaction{}, fmt.Errorf("post transaction: %w", err)
	}
	return posted, nil
}
