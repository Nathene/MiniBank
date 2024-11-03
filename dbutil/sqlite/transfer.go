package sqlite

import (
	"database/sql"
	"fmt"
	"log"
	"minibank/dbutil"
	"time"
)

func (s *sqlite) Transfer(fromAccountId, toAccountId int, amount float64) (int, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			log.Printf("Rolling back transaction due to error: %v", err)
			tx.Rollback()
		} else {
			err = tx.Commit()
			if err != nil {
				log.Printf("Error committing transaction: %v", err)
			}
		}
	}()

	// TODO: Dont leave the page
	fromAccount, err := s.GetAccount(fromAccountId)
	if err != nil {
		return 0, fmt.Errorf("error getting from account: %w", err)
	}

	// TODO: Dont leave the page
	toAccount, err := s.GetAccount(toAccountId)
	if err != nil {
		return 0, fmt.Errorf("error getting to account: %w", err)
	}
	// TODO: Make an error page for this or just stay on the same page.
	if fromAccount.Balance < amount {
		return 0, fmt.Errorf("insufficient funds in the from account")
	}

	fromAccount.Balance -= amount
	toAccount.Balance += amount

	// TODO: Dont leave the page
	err = s.UpdateAccountBalance(tx, fromAccount)
	if err != nil {
		return 0, fmt.Errorf("error updating from account balance: %w", err)
	}

	// TODO: Dont leave the page
	err = s.UpdateAccountBalance(tx, toAccount)
	if err != nil {
		return 0, fmt.Errorf("error updating to account balance: %w", err)
	}

	// Create a new transaction using NewTransaction, which returns a pointer
	transaction := dbutil.NewTransaction(fromAccountId, toAccountId, amount, "Transfer")

	// Use the pointer when passing to MakeTransaction
	err = s.MakeTransaction(tx, transaction)
	if err != nil {
		return 0, fmt.Errorf("error making transaction: %w", err)
	}

	if transaction.Id == 0 {
		return 0, fmt.Errorf("error: transaction ID is not set")
	}

	return transaction.Id, nil
}

func (s *sqlite) UpdateAccountBalance(tx *sql.Tx, account *dbutil.Account) error {
	stmt, err := tx.Prepare("UPDATE account SET balance = ?, updated_at = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("error preparing update statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(account.Balance, time.Now(), account.Id)
	if err != nil {
		return fmt.Errorf("error updating account balance: %w", err)
	}

	return nil
}

func (s *sqlite) Stimulus(tx *sql.Tx, account *dbutil.Account) error {
	// Update the account balance
	account.Balance += 1000

	transaction := dbutil.NewTransaction(1, account.Id, 1000.0, "Stimulus")

	err := s.MakeTransaction(tx, transaction)
	if err != nil {
		return err
	}

	return s.UpdateAccountBalance(tx, account)
}
