package main

import (
	"html/template"
	"strings"
	"time"
)

// Funcs is where you add your own template functions.
func (a *App) Funcs() template.FuncMap {
	return template.FuncMap{
		// Explicit URL building, for places the HTML rewriter can't reach
		// (CSS url(), JS strings, <meta content="…">, srcset, …).
		"url":  a.URL,
		"year": func() int { return time.Now().Year() },
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"motd": randomMOTD,
	}
}
