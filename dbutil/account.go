package dbutil

import (
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Account struct {
	Id                 int       `json:"id"`
	First_name         string    `json:"first_name"`
	Last_name          string    `json:"last_name"`
	Email              string    `json:"email"`
	Phone_number       int       `json:"phone_number,omitempty"`
	Encrypted_password string    `json:"encrypted_password"`
	Balance            float64   `json:"balance"`
	Created_at         time.Time `json:"created_at"`
	Updated_at         time.Time `json:"updated_at"`
}

func (a *Account) ChangeName(first_name, last_name string) {
	a.First_name = first_name
	a.Last_name = last_name
}

func (a *Account) ChangeEmail(email string) {
	a.Email = email
}

func (a *Account) ChangePhoneNumber(number int) {
	a.Phone_number = number
}

func (a *Account) ChangePassword(password string) {
	// encrypt password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		log.Println(err)
		return
	}

	a.Encrypted_password = string(hashedPassword)
}

func (a *Account) Transfer(account *Account, amount float64) {
	if (a.Balance - amount) < 0 {
		fmt.Println(`not enough funds in your account.`)
		return
	}
	a.Balance -= amount
	account.Balance += amount
}

func (a *Account) Print() {
	fmt.Println(fmt.Sprintf(`
		Account:
			id: %v,
			first_name: %s,
			last_name: %s,
			email: %s,
			phone_number %d,
			encrypted_password: %s,
			balance: %f,
			created_at: %q,
			updated_at: %q
		`, a.Id, a.First_name, a.Last_name, a.Email, a.Phone_number, a.Encrypted_password, a.Balance, a.Created_at, a.Updated_at))
}
