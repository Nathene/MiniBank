package server

import (
	"database/sql"
	"fmt"
	"log"
	"minibank/dbutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

func accountHandler(db dbutil.Database, c echo.Context) error {

	// Get the user ID from the session
	sess, _ := session.Get("session", c)
	userID, ok := sess.Values["userID"]
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login") // Redirect to login if not logged in
	}

	// Fetch the account details from the database
	account, err := db.GetAccount(userID.(int))
	if err != nil {
		log.Println("Error fetching account details:", err) // Log the error for debugging
		return c.Redirect(http.StatusSeeOther, "/login")
	}
	if c.Request().Method == http.MethodPost {
		tx, err := db.Begin() // Assuming your dbutil.Database has a Begin() method
		if err != nil {
			return c.String(http.StatusInternalServerError, "Error starting transaction")
		}
		defer func() {
			if err != nil {
				tx.Rollback()
			} else {
				err = tx.Commit()
			}
		}()

		// Apply the stimulus
		err = db.Stimulus(tx, account)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		return c.Redirect(http.StatusSeeOther, "/")
	}

	// Render the account.html template with the account data
	return c.Render(http.StatusOK, "account.html", map[string]interface{}{
		"Account": account,
	})
}

func allAccountsHandler(db dbutil.Database, c echo.Context) error {
	return c.Render(http.StatusOK, "all-accounts.html", map[string]interface{}{
		"Accounts": db.GetAccounts(),
	})
}

func createAccountHandler(db dbutil.Database, c echo.Context) error {
	if c.Request().Method == http.MethodGet {
		return c.Render(http.StatusOK, "create-account.html", nil)
	}

	if c.Request().Method == http.MethodPost {
		firstName := c.FormValue("first_name")
		lastName := c.FormValue("last_name")
		email := c.FormValue("email")
		phoneNumberStr := c.FormValue("phone_number")
		password := c.FormValue("password")

		if firstName == "" || lastName == "" || email == "" || password == "" {
			return c.String(http.StatusBadRequest, "Please fill in all required fields")
		}
		firstName = strings.Replace(firstName, string(firstName[0]), strings.ToUpper(string(firstName[0])), 1)

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Error hashing password")
		}
		number, err := strconv.Atoi(phoneNumberStr)
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid phone number")
		}

		newAccount := &dbutil.Account{
			First_name:         firstName,
			Last_name:          lastName,
			Email:              email,
			Phone_number:       number,
			Encrypted_password: string(hashedPassword),
			Balance:            0,
			Created_at:         time.Now(),
			Updated_at:         time.Now(),
		}
		err = db.CreateAccount(newAccount)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Error creating account")
		}

		return c.Redirect(http.StatusSeeOther, "/")
	}
	return nil
}

func loginHandler(db dbutil.Database, c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		email := c.FormValue("email")
		password := c.FormValue("password")

		if email == "" || password == "" {
			return c.String(http.StatusBadRequest, "Please enter email and password")
		}

		account, err := db.GetAccountByEmail(email)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.Render(http.StatusUnauthorized, "login.html", "invalid login details")
			}
			return c.String(http.StatusInternalServerError, "Error fetching account")
		}

		err = bcrypt.CompareHashAndPassword([]byte(account.Encrypted_password), []byte(password))
		if err != nil {
			return c.Render(http.StatusUnauthorized, "login.html", map[string]interface{}{
				"Error": "Invalid email or password", // Correct error message key
			})
		}

		sess, _ := session.Get("session", c)
		sess.Values["userID"] = account.Id
		sess.Save(c.Request(), c.Response())

		return c.Redirect(http.StatusSeeOther, "/")
	}
	return c.Render(http.StatusOK, "login.html", nil)
}

func logoutHandler(c echo.Context) error {
	sess, _ := session.Get("session", c)
	sess.Options.MaxAge = -1 // Set the MaxAge option to -1 to expire the cookie immediately
	sess.Values["userID"] = nil
	sess.Save(c.Request(), c.Response())
	return c.Redirect(http.StatusSeeOther, "/")
}

func deleteAccountHandler(db dbutil.Database, c echo.Context) error {
	sess, _ := session.Get("session", c)
	userID, ok := sess.Values["userID"]
	if !ok {
		return c.Redirect(http.StatusSeeOther, "/login")
	}

	if c.Request().Method == http.MethodGet {
		accounts := db.GetAccounts()
		accountIDStr := c.QueryParam("account_id")
		if accountIDStr == "" {
			return c.Render(http.StatusOK, "delete-account.html", map[string]interface{}{
				"Accounts": accounts,
			})
		}

		accountID, err := strconv.Atoi(accountIDStr)
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid account ID")
		}

		account, err := db.GetAccount(accountID)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Error fetching account details")
		}

		if account.Id != userID.(int) {
			return c.String(http.StatusForbidden, "You are not authorized to delete this account")
		}

		return c.Redirect(http.StatusSeeOther, "/delete-account")

	}

	if c.Request().Method == http.MethodPost {
		accountIDStr := c.FormValue("account_id")
		accountID, err := strconv.Atoi(accountIDStr)
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid account ID")
		}

		account, err := db.GetAccount(accountID)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Error fetching account details")
		}
		if account.Id != userID.(int) {
			return c.String(http.StatusForbidden, "You are not authorized to delete this account")
		}

		err = db.DeleteAccount(accountID)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Error deleting account")
		}

		return c.Redirect(http.StatusSeeOther, "/")
	}

	return nil
}

func paymentHandler(db dbutil.Database, c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c.Response().Header().Set("Content-Type", "application/json")
		// Get recipient and amount from form data
		recipient := c.FormValue("recipient")
		amountStr := c.FormValue("amount")

		// Validate input (add more robust validation as needed)
		if recipient == "" || amountStr == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"Error": "Please provide recipient and amount"})
		}
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"Error": "Invalid amount"})
		}

		// Get the logged-in user's ID
		sess, _ := session.Get("session", c)
		userID, ok := sess.Values["userID"]
		if !ok {
			return c.Redirect(http.StatusSeeOther, "/login")
		}

		// Identify the recipient account (using email or phone number)
		var recipientAccount *dbutil.Account
		if strings.Contains(recipient, "@") { // Check if it's an email
			recipientAccount, err = db.GetAccountByEmail(recipient)
		} else { // Assume it's a phone number
			phoneNumber, err := strconv.Atoi(recipient)
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]interface{}{"Error": "Invalid recipient phone number"})
			}
			recipientAccount, err = db.GetAccountByPhoneNumber(phoneNumber)
		}
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"Error": "Error finding recipient account"})
		}

		// Process the payment (update balances, record transaction, etc.)
		senderAccount, err := db.GetAccount(userID.(int))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"Error": "Error fetching sender account details"})
		}

		if senderAccount.Balance < amount {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"Error": "Insufficient balance"})
		}

		err = db.Transfer(userID.(int), recipientAccount.Id, amount)
		if err != nil {
			log.Println("Error processing payment:", err)
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"Error": fmt.Sprintf("Error processing payment: %v", err)})
		}

		return c.Redirect(http.StatusSeeOther, "/")
	}

	// For GET requests, fetch the account details and return as JSON
	recipient := c.QueryParam("recipient")
	if recipient == "" {
		return c.Render(http.StatusOK, "payment.html", nil)
	}

	c.Response().Header().Set("Content-Type", "application/json")

	// Identify the recipient account (using email or phone number)
	var recipientAccount *dbutil.Account
	var err error
	if strings.Contains(recipient, "@") {
		recipientAccount, err = db.GetAccountByEmail(recipient)
	} else {
		phoneNumber, err := strconv.Atoi(recipient)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"Error": "Invalid recipient phone number"})
		}
		recipientAccount, err = db.GetAccountByPhoneNumber(phoneNumber)
	}

	// Check if recipientAccount is nil before accessing its properties
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]interface{}{"Error": "Account not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"Error": err.Error()})
	}

	if recipientAccount == nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{"Error": "Account not found"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"Account": recipientAccount})
}
