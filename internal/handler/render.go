package handler

import (
	"html/template"
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

// renderPage renders a page inside the base layout.
func renderPage(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := templates.ExecuteTemplate(w, "layouts/base.html", data); err != nil {
		log.Printf("render %s: %v", name, err)
	}
}

// renderPartial renders a standalone partial template.
func renderPartial(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := templates.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("render %s: %v", name, err)
	}
}

func httpError(w http.ResponseWriter, msg string, code int) {
	http.Error(w, msg, code)
}
