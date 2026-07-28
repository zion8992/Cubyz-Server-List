package main

import (
	// ==== GENERAL ====
	"log/slog"
	"time"

	// ==== PAGE SLUGS ====
	"regexp"
	"strings"
)

type App struct {
	Log *slog.Logger
	Now time.Time // raw time; formatted where needed

	// === Config (from CLI flags) ===
	OutputDir string
	Pattern   string
	BaseURL   string // empty disables sitemap generation
}

type Server struct {
	// === Core Fields ===
	Name        string    `json:"name"`
	IP          string    `json:"ip"`
	DateCreated time.Time `json:"date_created"`

	// === Server-Provided Fields ===
	Description string `json:"description"`
	Version string `json:"version"`
	Gamemodes string `json:"gamemodes"`
	Languages string `json:"languages"`
	RequiresMods bool `json:"requires_mods"`
}

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

// Slug returns a filesystem/URL-safe identifier for the server.
func (s Server) Slug() string {
	slug := slugRe.ReplaceAllString(strings.ToLower(s.Name), "-")
	return strings.Trim(slug, "-")
}

// Created formats DateCreated for display, so templates never print the
// raw time.Time string.
func (s Server) Created() string {
	if s.DateCreated.IsZero() {
		return "unknown"
	}
	return s.DateCreated.Format("January 2, 2006")
}

// ===== PAGES =====

// Page holds fields shared by every rendered page.
type Page struct {
	// Root is the relative path prefix back to the site root
	// ("" for index.html, "../../" for servers/<slug>/index.html).
	Root           string
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
