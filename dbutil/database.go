package dbutil

import "database/sql"

type Database interface {
	Init()
	GetAccount(id int) (*Account, error)
	GetAccounts() []Account
	GetAccountByEmail(email string) (*Account, error)
	GetAccountByPhoneNumber(number int) (*Account, error)
	MockData()
	CreateAccount(account *Account) error
	Transfer(fromAccountId, toAccountId int, amount float64) error
	UpdateAccountBalance(tx *sql.Tx, account *Account) error
	DeleteAccount(id int) error
	Stimulus(tx *sql.Tx, account *Account) error
	Begin() (*sql.Tx, error)
}
