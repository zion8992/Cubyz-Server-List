package main

import (
	"log/slog"
	"net/http"
)

func (app *App) RouteLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.Log.Info("new request", slog.String("url", r.URL.Path), slog.String("method", r.Method))
		next.ServeHTTP(w, r)
	})
}
