package main

import(
	// ==== GENERAL ====
	"log/slog"
	"time"

	// ==== HTML ====
	tmpl "html/template"
)

type App struct {
	Log *slog.Logger
}

type Server struct {
	Name string
	IP string
	DateCreated time.Time
}

type PageList struct {
	Style tmpl.CSS
	Servers []Server
}