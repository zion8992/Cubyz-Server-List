package main

import (
	// === GENERAL ===
	"log/slog"
	"net/http"
	"os"
	"flag"
	"html/template"
	"sync"

	// === HTTP ===
	"github.com/justinas/alice"
)

type App struct {
	logger *slog.Logger
	Root string // e.g. "/" or "/silverite/"
	Dev  bool   // reparse templates on every request

	funcs         template.FuncMap
	mu            sync.RWMutex
	templateCache map[string]*template.Template
}

func main() {
	var hotReload = flag.Bool("dev", false, "Enable hot reloading")

	flag.Parse()

	app := &App{
		logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
		Dev: *hotReload,
		Root: "/",
	}

	if err := app.InitTemplates(); err != nil {
		app.logger.Error("Failed to load HTML templates", slog.String("error", err.Error()))
		os.Exit(1) // bail out
	}

	var port string = ":8000"

	chain := alice.New(app.RequestLogger)

	mux := http.NewServeMux()
	mux.Handle("GET /", chain.ThenFunc(app.GETSlash))
	mux.Handle("GET "+app.Root+"static/", chain.Then(app.StaticHandler()))

	app.logger.Info("starting server", "addr", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		app.logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}
