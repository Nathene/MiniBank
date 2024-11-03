package server

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"minibank/dbutil/sqlite"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type TemplateRegistry struct {
	templates map[string]*template.Template
}

// Render renders a template document
func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	// Check if the user is logged in and pass that information to the template
	sess, _ := session.Get("session", c)
	userID, ok := sess.Values["userID"]
	isLoggedIn := ok && userID != nil

	// Add the IsLoggedIn variable to the template data
	if data == nil {
		data = map[string]interface{}{"IsLoggedIn": isLoggedIn}
	} else {
		dataMap, ok := data.(map[string]interface{}) // Type assertion without the 'ok' check
		if !ok {
			// Handle the case where data is not a map[string]interface{}
			log.Println("Template data is not a map[string]interface{}")
			return fmt.Errorf("invalid template data type: %T", data) // Return an error
		}
		dataMap["IsLoggedIn"] = isLoggedIn
	}

	tmpl, ok := t.templates[name]
	if !ok {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("template %s not found", name))
	}
	return tmpl.Execute(w, data)
}

func Run() {
	db := sqlite.New()
	db.Init()
	tmp := db.GetAccounts()
	if len(tmp) == 0 {
		db.MockData()
	}

	templates := make(map[string]*template.Template)
	templates["create-account"] = template.Must(template.ParseFiles("templates/create-account.html"))
	templates["delete-account"] = template.Must(template.ParseFiles("templates/delete-account.html"))
	templates["payment"] = template.Must(template.ParseFiles("templates/payment.html"))
	templates["account"] = template.Must(template.ParseFiles("templates/account.html"))
	templates["all-accounts"] = template.Must(template.ParseFiles("templates/all-accounts.html"))
	templates["login"] = template.Must(template.ParseFiles("templates/login.html"))

	templates["transactions"] = template.Must(template.ParseFiles("templates/transactions.html"))
	templates["single-transaction"] = template.Must(template.ParseFiles("templates/single-transaction.html"))
	// Add more templates if needed

	e := echo.New()

	e.Renderer = &TemplateRegistry{
		templates: templates,
	}

	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))

	e.GET("/", func(c echo.Context) error {
		return accountHandler(&db, c)
	})
	e.POST("/account", func(c echo.Context) error {
		return accountHandler(&db, c)
	})
	e.GET("/payment", func(c echo.Context) error {
		return paymentHandler(&db, c)
	})
	e.POST("/payment", func(c echo.Context) error {
		return paymentHandler(&db, c)
	})
	e.GET("/all-accounts", func(c echo.Context) error {
		return allAccountsHandler(&db, c)
	})
	e.GET("/create-account", func(c echo.Context) error {
		return createAccountHandler(&db, c)
	})
	e.POST("/create-account", func(c echo.Context) error {
		return createAccountHandler(&db, c)
	})

	e.GET("/delete-account", func(c echo.Context) error {
		return deleteAccountHandler(&db, c)
	})
	e.POST("/delete-account", func(c echo.Context) error {
		return deleteAccountHandler(&db, c)
	})

	e.GET("/transactions", func(c echo.Context) error {
		return transactionsHandler(&db, c)
	})

	e.GET("/single-transaction/:transaction_id", func(c echo.Context) error {
		return singleTransactionHandler(&db, c)
	})

	e.GET("/login", func(c echo.Context) error {
		return loginHandler(&db, c)
	})
	e.POST("/login", func(c echo.Context) error {
		return loginHandler(&db, c)
	})

	e.GET("/logout", func(c echo.Context) error {
		return logoutHandler(c)
	})

	e.Logger.Fatal(e.Start(":3000"))
}
