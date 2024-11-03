package sqlite

import (
	"database/sql"
	"fmt"
	"log"
	"minibank/dbutil"
	"os"
)

func (s *sqlite) GetAccounts() []dbutil.Account {
	sqlScript, err := os.ReadFile("./sql/queryUsers.sql")
	if err != nil {
		panic(err)
	}
	var accounts []dbutil.Account
	rows, err := s.db.Query(string(sqlScript))
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var account dbutil.Account
		var id sql.NullInt64
		err := rows.Scan(
			&id,
			&account.First_name,
			&account.Last_name,
			&account.Email,
			&account.Phone_number,
			&account.Encrypted_password,
			&account.Balance,
			&account.Created_at,
			&account.Updated_at,
		)
		if err != nil {
			return nil
		}
		if id.Valid {
			account.Id = int(id.Int64)
		}
		accounts = append(accounts, account)
	}
	return accounts
}

func (s *sqlite) CreateAccount(account *dbutil.Account) error {
	stmt, err := s.db.Prepare("INSERT INTO account(first_name, last_name, email, phone_number, encrypted_password, balance, created_at, updated_at) values(?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Println("error preparing statement: ", err)
		return fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(account.First_name, account.Last_name, account.Email, account.Phone_number, account.Encrypted_password, account.Balance, account.Created_at, account.Updated_at)
	if err != nil {
		log.Println("error executing statement: ", err)
		return fmt.Errorf("error executing statement: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		log.Println("error getting last inserted id: ", err)
		return fmt.Errorf("error getting last inserted id: %w", err)
	}
	account.Id = int(id)
	return nil
}

func (s *sqlite) GetAccount(id int) (*dbutil.Account, error) {
	// Prepare the SQL statement
	stmt, err := s.db.Prepare("SELECT * FROM account WHERE id = ?")
	if err != nil {
		return nil, fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	var account dbutil.Account
	err = stmt.QueryRow(id).Scan(&account.Id, &account.First_name, &account.Last_name, &account.Email, &account.Phone_number, &account.Encrypted_password, &account.Balance, &account.Created_at, &account.Updated_at)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("account not found: %w", err)
		}
		return nil, fmt.Errorf("error scanning row: %w", err)
	}

	return &account, nil
}

func (s *sqlite) GetAccountByEmail(email string) (*dbutil.Account, error) {
	// Prepare the SQL statement
	stmt, err := s.db.Prepare("SELECT * FROM account WHERE email = ?") // Use the correct table name "account"
	if err != nil {
		return nil, fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	var account dbutil.Account
	err = stmt.QueryRow(email).Scan(&account.Id, &account.First_name, &account.Last_name, &account.Email, &account.Phone_number, &account.Encrypted_password, &account.Balance, &account.Created_at, &account.Updated_at)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("account not found: %w", err) // Or you can return nil, nil
		}
		return nil, fmt.Errorf("error scanning row: %w", err)
	}

	return &account, nil
}

func (s *sqlite) GetAccountByPhoneNumber(number int) (*dbutil.Account, error) {
	// Prepare the SQL statement
	stmt, err := s.db.Prepare("SELECT * FROM account WHERE phone_number = ?") // Use the correct table name "account"
	if err != nil {
		return nil, fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	var account dbutil.Account
	err = stmt.QueryRow(number).Scan(&account.Id, &account.First_name, &account.Last_name, &account.Email, &account.Phone_number, &account.Encrypted_password, &account.Balance, &account.Created_at, &account.Updated_at)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("account not found: %w", err) // Or you can return nil, nil
		}
		return nil, fmt.Errorf("error scanning row: %w", err)
	}

	return &account, nil
}

func (s *sqlite) DeleteAccount(id int) error {
	stmt, err := s.db.Prepare("DELETE FROM account WHERE id = ?") // Use correct table name "account"
	if err != nil {
		return fmt.Errorf("error preparing delete statement: %w", err)
	}
	defer stmt.Close()

	// Execute the statement with the account ID
	result, err := stmt.Exec(id)
	if err != nil {
		return fmt.Errorf("error deleting account: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf(`no account found with ID %d`, id)
	}

	return nil
}
