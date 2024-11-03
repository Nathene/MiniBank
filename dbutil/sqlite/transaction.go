package sqlite

import (
	"database/sql"
	"fmt"
	"minibank/dbutil"
	"time"
)

func (s *sqlite) MakeTransaction(tx *sql.Tx, transaction *dbutil.Transaction) error {
	stmt, err := tx.Prepare("INSERT INTO transactions (from_account, to_account, amount, transaction_type, created_at) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("error preparing insert statement: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(transaction.FromAccount, transaction.ToAccount, transaction.Amount, transaction.TransactionType, time.Now())
	if err != nil {
		return fmt.Errorf("error inserting transaction: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %w", err)
	}

	transaction.Id = int(id)
	return nil
}

func (s *sqlite) ListTransactionsFromAccount(accountID int) ([]dbutil.Transaction, error) {
	// Update the SQL to include both from_account and to_account, and use DISTINCT to avoid duplicates
	stmt, err := s.db.Prepare(`
		SELECT DISTINCT id, from_account, to_account, amount, transaction_type, created_at 
		FROM transactions 
		WHERE from_account = ? OR to_account = ?
	`)
	if err != nil {
		return nil, fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(accountID, accountID)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	defer rows.Close()

	var transactions []dbutil.Transaction
	for rows.Next() {
		var transaction dbutil.Transaction
		// Scan the new fields
		err := rows.Scan(
			&transaction.Id,
			&transaction.FromAccount,
			&transaction.ToAccount,
			&transaction.Amount,
			&transaction.TransactionType,
			&transaction.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning transaction: %w", err)
		}

		transactions = append(transactions, transaction)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return transactions, nil
}

func (s *sqlite) GetTransaction(transactionID int) (*dbutil.Transaction, error) {
	var transaction dbutil.Transaction
	// Update the query to fetch the new fields
	query := "SELECT id, from_account, to_account, amount, transaction_type, created_at FROM transactions WHERE id = ?"
	row := s.db.QueryRow(query, transactionID)

	// Scan the new fields
	err := row.Scan(&transaction.Id, &transaction.FromAccount, &transaction.ToAccount, &transaction.Amount, &transaction.TransactionType, &transaction.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transaction not found")
		}
		return nil, fmt.Errorf("error fetching transaction: %w", err)
	}
	return &transaction, nil
}
