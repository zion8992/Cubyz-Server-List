package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/justinas/alice"
)

type App struct {
	logger *slog.Logger
}

func main() {
	app := &App{
		logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}

	chain := alice.New(app.requestLogger)

	mux := http.NewServeMux()
	mux.Handle("GET /", chain.ThenFunc(app.hello))

	app.logger.Info("starting server", "addr", ":8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		app.logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}
