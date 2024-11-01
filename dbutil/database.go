package dbutil

import "database/sql"

type Database interface {
	Init()
	GetAccount(id int) (*Account, error)
	GetAccounts() []Account
	MockData()
	CreateAccount(account *Account) error
	Transfer(fromAccountId, toAccountId int, amount float64) error
	UpdateAccountBalance(tx *sql.Tx, account *Account) error
}
