package main

import (
	// ==== GENERAL ====
	"log/slog"
	"time"

	// ==== PAGE SLUGS ====
	"regexp"
	"strings"

	// ==== HTML ====
	tmpl "html/template"
)

type App struct {
	Log *slog.Logger
	Now string
}

type Server struct {
	// === Public Fields ===
	Name        string
	IP          string
	DateCreated time.Time
}

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

// Slug returns a filesystem/URL-safe identifier for the server.
func (s Server) Slug() string {
	slug := slugRe.ReplaceAllString(strings.ToLower(s.Name), "-")
	return strings.Trim(slug, "-")
}

// ===== PAGES =====

// Page holds fields shared by every rendered page.
type Page struct {
	Style tmpl.CSS
	GenerationDate string
}

type PageList struct {
	Page
	Servers []Server
}

type PageServer struct {
	Page
	Server Server
}