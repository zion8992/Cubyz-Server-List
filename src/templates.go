package main

import (
	"html/template"
	"path/filepath"
	"io/fs"
	"net/http"
	"fmt"
)

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	// Root-level page templates
	rootPages, err := filepath.Glob("./templates/*.html")
	if err != nil {
		return nil, err
	}

	// Auth subfolder templates
	authPages, err := filepath.Glob("./templates/auth/*.html")
	if err != nil {
		return nil, err
	}

	pages := append(rootPages, authPages...)

	for _, page := range pages {
		name := filepath.Base(page)

		// Skip base.html, it's a layout, not a page
		if name == "base.html" {
			continue
		}

		files := []string{
			"./templates/base.html",
			page,
		}

		ts, err := template.ParseFiles(files...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}

func newEmbedTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	rootPages, err := fs.Glob(templatesFS, "templates/*.html")
	if err != nil {
		return nil, err
	}

	authPages, err := fs.Glob(templatesFS, "templates/auth/*.html")
	if err != nil {
		return nil, err
	}

	pages := append(rootPages, authPages...)

	for _, page := range pages {
		name := filepath.Base(page)

		if name == "base.html" {
			continue
		}

		files := []string{
			"templates/base.html",
			page,
		}

		ts, err := template.ParseFS(templatesFS, files...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}

func (a *App) render(w http.ResponseWriter, r *http.Request, status int, page string, data any) {
	ts, ok := a.TemplateCache[page]
	if !ok {
		a.Error(w, r, fmt.Sprintf("the template %s does not exist", page))
		return
	}

	w.WriteHeader(status)

	if err := ts.Execute(w, data); err != nil {
		a.Error(w, r, "template execution failed: "+err.Error())
	}
}
