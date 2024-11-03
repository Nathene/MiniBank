package dbutil

import "database/sql"

type Database interface {
	Init()

	GetAccount(id int) (*Account, error)
	GetAccounts() []Account
	GetAccountByEmail(email string) (*Account, error)
	GetAccountByPhoneNumber(number int) (*Account, error)
	CreateAccount(account *Account) error
	UpdateAccountBalance(tx *sql.Tx, account *Account) error
	DeleteAccount(id int) error
	Transfer(fromAccountId, toAccountId int, amount float64) (int, error)

	ListTransactionsFromAccount(id int) ([]Transaction, error)
	MakeTransaction(tx *sql.Tx, transaction *Transaction) error
	GetTransaction(transactionID int) (*Transaction, error)

	Stimulus(tx *sql.Tx, account *Account) error
	MockData()
	Begin() (*sql.Tx, error)
}
