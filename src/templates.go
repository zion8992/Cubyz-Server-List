package main

import (
	"html/template"
	"path/filepath"
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
