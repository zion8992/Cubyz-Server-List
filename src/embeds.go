package main

import (
	"embed"
	"html/template"
	"path"
	"io/fs"
)

//go:embed all:templates
var templatesFS embed.FS

//go:embed all:static
var staticFS embed.FS

func newEmbedTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	err := fs.WalkDir(templatesFS, "templates", func(page string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || path.Ext(page) != ".html" {
			return nil
		}

		name := path.Base(page)
		if name == "base.html" {
			return nil
		}

		files := []string{
			"templates/base.html",
			page,
		}

		ts, err := template.ParseFS(templatesFS, files...)
		if err != nil {
			return err
		}

		cache[name] = ts
		return nil
	})
	if err != nil {
		return nil, err
	}

	return cache, nil
}
