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
	return c.Render(http.StatusOK, "account", map[string]interface{}{
		"Account": account,
	})
}

func allAccountsHandler(db dbutil.Database, c echo.Context) error {
	return c.Render(http.StatusOK, "all-accounts", map[string]interface{}{
		"Accounts": db.GetAccounts(),
	})
}

func createAccountHandler(db dbutil.Database, c echo.Context) error {
	if c.Request().Method == http.MethodGet {
		return c.Render(http.StatusOK, "create-account", nil)
	}

	if c.Request().Method == http.MethodPost {
		firstName := c.FormValue("first_name")
		lastName := c.FormValue("last_name")
		email := c.FormValue("email")
		phoneNumberStr := c.FormValue("phone_number")
		password := c.FormValue("password")

		// Validate required fields
		if firstName == "" || lastName == "" || email == "" || password == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Please fill in all required fields."})
		}

		// Validate phone number
		number, err := strconv.Atoi(phoneNumberStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid phone number. Only digits are allowed."})
		}

		// Format the first name and hash the password
		firstName = strings.Replace(firstName, string(firstName[0]), strings.ToUpper(string(firstName[0])), 1)
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error hashing password."})
		}

		// Create the new account in the database
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
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error creating account. Please try again."})
		}

		// Automatically log in the user by creating a session
		sess, _ := session.Get("session", c)
		sess.Values["userID"] = newAccount.Id
		sess.Save(c.Request(), c.Response())

		// Return success response
		return c.JSON(http.StatusOK, map[string]string{"status": "success"})
	}

	return nil
}

func loginHandler(db dbutil.Database, c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		email := c.FormValue("email")
		password := c.FormValue("password")

		// Check for empty fields
		if email == "" || password == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Please enter both email and password."})
		}

		// Fetch account by email
		account, err := db.GetAccountByEmail(email)
		if err != nil {
			// Check if the error is "no rows found," meaning the account doesn't exist
			if err == sql.ErrNoRows {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid email or password."})
			}
			// Unexpected database error
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "An error occurred. Please try again."})
		}

		// Check password
		err = bcrypt.CompareHashAndPassword([]byte(account.Encrypted_password), []byte(password))
		if err != nil {
			// Incorrect password
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid email or password."})
		}

		// Successful login: create session
		sess, _ := session.Get("session", c)
		sess.Values["userID"] = account.Id
		sess.Save(c.Request(), c.Response())

		// Return success response
		return c.JSON(http.StatusOK, map[string]string{"status": "success"})
	}

	return c.Render(http.StatusOK, "login", nil)
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
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "not_logged_in"})
	}

	if c.Request().Method == http.MethodGet {
		accounts := db.GetAccounts()
		return c.Render(http.StatusOK, "delete-account", map[string]interface{}{
			"Accounts": accounts,
		})
	}

	if c.Request().Method == http.MethodPost {
		accountIDStr := c.FormValue("account_id")
		accountID, err := strconv.Atoi(accountIDStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid_account_id"})
		}

		// Fetch the account details
		account, err := db.GetAccount(accountID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "fetch_error"})
		}

		// Check if the account is restricted and not owned by the user
		if accountID <= 5 && account.Id != userID.(int) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "unauthorized"})
		}

		// Delete the account
		err = db.DeleteAccount(accountID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "delete_error"})
		}

		// If the user deleted their own account, clear the session and log them out
		if accountID == userID.(int) {
			sess.Options.MaxAge = -1 // Expire the session cookie
			sess.Values["userID"] = nil
			sess.Save(c.Request(), c.Response()) // Ensure session is saved as expired
			return c.JSON(http.StatusOK, map[string]string{"status": "logged_out"})
		}

		// Otherwise, deletion was successful but no logout needed
		return c.JSON(http.StatusOK, map[string]string{"status": "success"})
	}

	return nil
}

func paymentHandler(db dbutil.Database, c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c.Response().Header().Set("Content-Type", "application/json")
		recipient := c.FormValue("recipient")
		amountStr := c.FormValue("amount")

		if recipient == "" || amountStr == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"Error": "Please provide recipient and amount"})
		}
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"Error": "Invalid amount"})
		}

		sess, _ := session.Get("session", c)
		userID, ok := sess.Values["userID"]
		if !ok {
			return c.Redirect(http.StatusSeeOther, "/login")
		}

		var recipientAccount *dbutil.Account
		if strings.Contains(recipient, "@") {
			recipientAccount, err = db.GetAccountByEmail(recipient)
		} else {
			phoneNumber, err := strconv.Atoi(recipient)
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]interface{}{"Error": "Invalid recipient phone number"})
			}
			recipientAccount, err = db.GetAccountByPhoneNumber(phoneNumber)
		}
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"Error": "Error finding recipient account"})
		}

		senderAccount, err := db.GetAccount(userID.(int))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"Error": "Error fetching sender account details"})
		}

		if senderAccount.Balance < amount {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"Error": "Insufficient balance"})
		}

		transactionID, err := db.Transfer(userID.(int), recipientAccount.Id, amount)
		if err != nil {
			log.Printf("Error during transfer: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"Error": "Error processing payment"})
		}

		if transactionID == 0 {
			log.Println("Transaction ID is invalid, possibly due to an error in db.Transfer.")
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"Error": "Error finalizing transaction"})
		}

		tx, err := db.Begin()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, "Error starting transaction")
		}

		err = tx.Commit()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, "Error committing transaction")
		}
		// Redirect to the transaction details page
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/single-transaction/%d", transactionID))
	}

	recipient := c.QueryParam("recipient")
	if recipient == "" {
		return c.Render(http.StatusOK, "payment", nil)
	}

	c.Response().Header().Set("Content-Type", "application/json")

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

func transactionsHandler(db dbutil.Database, c echo.Context) error {
	// Step 1: Get the account ID from the URL parameters or session
	accountIDStr := c.QueryParam("account_id")
	if accountIDStr == "" {
		sess, _ := session.Get("session", c)
		userID, ok := sess.Values["userID"]
		if !ok {
			log.Println("User is not logged in. Redirecting to login.")
			return c.Redirect(http.StatusSeeOther, "/login")
		}
		accountIDStr = strconv.Itoa(userID.(int))
	}
	// Step 2: Convert account ID to integer
	accountID, err := strconv.Atoi(accountIDStr)
	if err != nil {
		log.Printf("Invalid account ID: %s. Error: %v", accountIDStr, err)
		return c.String(http.StatusBadRequest, "Invalid account ID")
	}
	// Step 3: Fetch account details
	account, err := db.GetAccount(accountID)
	if err != nil {
		log.Printf("Error fetching account details for ID %d: %v", accountID, err)
		return c.String(http.StatusInternalServerError, "Error fetching account details")
	}
	// Step 4: Fetch transactions
	transactions, err := db.ListTransactionsFromAccount(accountID)
	if err != nil {
		log.Printf("Error fetching transactions for account %d: %v", accountID, err)
		return c.String(http.StatusInternalServerError, "Error fetching transactions")
	}
	// Step 5: Render the template
	return c.Render(http.StatusOK, "transactions", map[string]interface{}{
		"Transactions": transactions,
		"Account":      account,
		"IsLoggedIn":   true,
	})
}

func singleTransactionHandler(db dbutil.Database, c echo.Context) error {
	// Step 1: Get the transaction ID from the URL parameters
	transactionIDStr := c.Param("transaction_id")
	if transactionIDStr == "" {
		log.Println("Transaction ID is required")
		return c.String(http.StatusBadRequest, "Transaction ID is required")
	}

	// Step 2: Convert transaction ID to integer
	transactionID, err := strconv.Atoi(transactionIDStr)
	if err != nil {
		log.Printf("Invalid transaction ID: %s. Error: %v", transactionIDStr, err)
		return c.String(http.StatusBadRequest, "Invalid transaction ID")
	}

	// Step 3: Fetch the transaction details from the database
	transaction, err := db.GetTransaction(transactionID)
	if err != nil {
		log.Printf("Error fetching transaction details for ID %d: %v", transactionID, err)
		return c.String(http.StatusNotFound, "Transaction not found")
	}

	// Step 4: Fetch associated account details using FromAccount and ToAccount
	fromAccount, err := db.GetAccount(transaction.FromAccount)
	if err != nil {
		log.Printf("Error fetching from account details for ID %d: %v", transaction.FromAccount, err)
		return c.String(http.StatusInternalServerError, "Error fetching from account details")
	}

	toAccount, err := db.GetAccount(transaction.ToAccount)
	if err != nil {
		log.Printf("Error fetching to account details for ID %d: %v", transaction.ToAccount, err)
		return c.String(http.StatusInternalServerError, "Error fetching to account details")
	}

	// Step 5: Render the template
	err = c.Render(http.StatusOK, "single-transaction", map[string]interface{}{
		"Transaction": transaction,
		"FromAccount": fromAccount,
		"ToAccount":   toAccount,
		"IsLoggedIn":  true,
	})

	if err != nil {
		log.Printf("Error rendering single-transaction template: %v", err)
		return c.String(http.StatusInternalServerError, "Error rendering template")
	}

	return nil
}
