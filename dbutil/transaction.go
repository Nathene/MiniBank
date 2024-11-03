package dbutil

import (
	"time"
)

type Transaction struct {
	Id              int       `json:"id"`
	FromAccount     int       `json:"from_account"`
	ToAccount       int       `json:"to_account"`
	Amount          float64   `json:"amount"`
	TransactionType string    `json:"transaction_type"`
	CreatedAt       time.Time `json:"created_at"`
}

func NewTransaction(fromAccount, toAccount int, amount float64, transactionType string) *Transaction {
	return &Transaction{
		FromAccount:     fromAccount,
		ToAccount:       toAccount,
		Amount:          amount,
		TransactionType: transactionType,
		CreatedAt:       time.Now(),
	}
}
