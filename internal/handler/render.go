package handler

import (
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
)

var templates *template.Template

// InitTemplates parses all templates from the embedded filesystem.
func InitTemplates(fsys fs.FS, funcMap template.FuncMap) {
	templates = template.Must(
		template.New("").Funcs(funcMap).ParseFS(fsys,
			"layouts/*.html",
			"pages/*.html",
			"partials/*.html",
		),
	)
}

// renderPage renders a page template inside the base layout.
func renderPage(w http.ResponseWriter, page string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := renderWithLayout(w, page, data); err != nil {
		log.Printf("render %s: %v", page, err)
	}
}

// renderPartial renders a standalone partial template.
func renderPartial(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := templates.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("render %s: %v", name, err)
	}
}

// renderWithLayout executes the page inside the base layout.
func renderWithLayout(w io.Writer, page string, data any) error {
	// Clone base layout and add the page content block
	t, err := templates.Lookup("layouts/base.html").Clone()
	if err != nil {
		return err
	}
	// Add the page template as "content"
	pageT := templates.Lookup(page + ".html")
	if pageT == nil {
		pageT = templates.Lookup(page)
	}
	if pageT != nil {
		_, err = t.AddParseTree("content", pageT.Tree)
		if err != nil {
			return err
		}
	}
	return t.Execute(w, data)
}

func httpError(w http.ResponseWriter, msg string, code int) {
	http.Error(w, msg, code)
}
