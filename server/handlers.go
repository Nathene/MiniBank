package server

import (
	"minibank/dbutil"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

func handleHome(db dbutil.Database, c echo.Context) error {
	accounts := db.GetAccounts() // Fetch the accounts

	// Render the accounts.html template with the account data
	return c.Render(http.StatusOK, "accounts.html", map[string]interface{}{
		"Accounts": accounts,
	})
}

func createAccountHandler(db dbutil.Database, c echo.Context) error {
	if c.Request().Method == http.MethodGet {
		// Serve the create account form
		return c.Render(http.StatusOK, "create-account.html", nil)
	} else if c.Request().Method == http.MethodPost {
		// Handle the form submission
		firstName := c.FormValue("first_name")
		lastName := c.FormValue("last_name")
		email := c.FormValue("email")
		phoneNumberStr := c.FormValue("phone_number") // You might need to convert this to an integer
		password := c.FormValue("password")
		balanceStr := c.FormValue("balance") // You might need to convert this to a float64

		// Basic validation (you should add more robust validation)
		if firstName == "" || lastName == "" || email == "" || password == "" {
			return c.String(http.StatusBadRequest, "Please fill in all required fields")
		}

		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Error hashing password")
		}
		number, err := strconv.Atoi(phoneNumberStr) // Convert to int
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid phone number")
		}
		balance, err := strconv.ParseFloat(balanceStr, 64)
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid balance")
		}

		// Create the account
		newAccount := &dbutil.Account{
			First_name:         firstName,
			Last_name:          lastName,
			Email:              email,
			Phone_number:       number,
			Encrypted_password: string(hashedPassword),
			Balance:            balance,
			Created_at:         time.Now(),
			Updated_at:         time.Now(),
		}
		err = db.CreateAccount(newAccount)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Error creating account")
		}

		// Redirect to a success page or display a success message
		return c.String(http.StatusOK, "Account created successfully!")
	}
	return nil
}
