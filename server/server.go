package server

import (
	"fmt"
	"html/template"
	"io"
	"minibank/dbutil/sqlite"
	"net/http"

	"github.com/labstack/echo/v4"
)

// TemplateRenderer is a custom HTML renderer for Echo
type TemplateRegistry struct {
	templates map[string]*template.Template
}

// Render renders a template document
func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	tmpl, ok := t.templates[name]
	if !ok {
		err := fmt.Errorf("template %s not found", name)
		return c.String(http.StatusInternalServerError, err.Error())
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
	templates["create-account.html"] = template.Must(template.ParseFiles("templates/create-account.html"))
	templates["accounts.html"] = template.Must(template.ParseFiles("templates/accounts.html"))
	// Add more templates if needed

	e := echo.New()

	e.Renderer = &TemplateRegistry{
		templates: templates,
	}

	e.GET("/", func(c echo.Context) error {
		return handleHome(&db, c)
	})

	e.GET("/create-account", func(c echo.Context) error {
		return createAccountHandler(&db, c)
	})
	e.POST("/create-account", func(c echo.Context) error {
		return createAccountHandler(&db, c)
	})

	e.Logger.Fatal(e.Start(":3000"))
}
