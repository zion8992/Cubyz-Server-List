package main

import (
	// === GENERAL ===
	"log/slog"
	"net/http"
	"os"

	// === HTTP ===
	"github.com/justinas/alice"
)

type App struct {
	logger *slog.Logger
}

func main() {
	app := &App{
		logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}

	var port string = ":8000"

	chain := alice.New(app.RequestLogger)

	mux := http.NewServeMux()
	mux.Handle("GET /", chain.ThenFunc(app.hello))

	app.logger.Info("starting server", "addr", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		app.logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}
