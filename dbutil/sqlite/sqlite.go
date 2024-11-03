package sqlite

import (
	"database/sql"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

type sqlite struct {
	db *sql.DB
}

func New() sqlite {
	db, err := sql.Open("sqlite", "file:minibank?cache=shared&mode=rwc")
	if err != nil {
		log.Fatal(err)
	}
	return sqlite{
		db: db,
	}
}

func (s *sqlite) Init() {
	sqlStmt := `
    CREATE TABLE IF NOT EXISTS account (  -- Use the correct table name "account"
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        first_name VARCHAR(50),
        last_name VARCHAR(50),
        email VARCHAR(50) UNIQUE,
        phone_number INTEGER UNIQUE,
        encrypted_password VARCHAR(100),
        balance INTEGER,
        created_at TIMESTAMP,
        updated_at TIMESTAMP
    );

	CREATE TABLE IF NOT EXISTS transactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		from_account INTEGER,
		to_account INTEGER,
		amount REAL,
		transaction_type TEXT,
		created_at DATETIME
	);

    `
	_, err := s.db.Exec(sqlStmt)
	if err != nil {
		log.Fatal("Error creating account table:", err) // Log with context
	}

	err = s.db.Ping()
	if err != nil {
		log.Fatal("Database connection failed:", err)
		return
	}

	log.Println("Database connection successful!")
}

func (s *sqlite) MockData() {
	sqlScript, err := os.ReadFile("./sql/mock_data.sql")
	if err != nil {
		panic(err)
	}
	_, err = s.db.Exec(string(sqlScript))
	if err != nil {
		panic(err)
	}

	log.Println("Data added")
}

func (s *sqlite) Begin() (*sql.Tx, error) {
	return s.db.Begin()
}
