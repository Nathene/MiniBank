package sqlite

import (
	"database/sql"
	"fmt"
	"minibank/dbutil"
	"time"
)

func (s *sqlite) Transfer(fromAccountId, toAccountId int, amount float64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// Get the accounts involved in the transfer
	fromAccount, err := s.GetAccount(fromAccountId)
	if err != nil {
		return fmt.Errorf("error getting from account: %w", err)
	}

	toAccount, err := s.GetAccount(toAccountId)
	if err != nil {
		return fmt.Errorf("error getting to account: %w", err)
	}

	// Check if the sender has sufficient balance
	if fromAccount.Balance < amount {
		return fmt.Errorf("insufficient funds in the from account")
	}

	// Update the balances
	fromAccount.Balance -= amount
	toAccount.Balance += amount

	// Update the accounts in the database within the transaction
	err = s.UpdateAccountBalance(tx, fromAccount)
	if err != nil {
		return fmt.Errorf("error updating from account balance: %w", err)
	}

	err = s.UpdateAccountBalance(tx, toAccount)
	if err != nil {
		return fmt.Errorf("error updating to account balance: %w", err)
	}

	return nil
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

	return s.UpdateAccountBalance(tx, account)
}
