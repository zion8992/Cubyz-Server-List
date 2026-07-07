package main

import (
	"log/slog"
	"net/http"
	"strings"
)

func (app *App) RouteLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.Log.Info("new request", slog.String("url", r.URL.Path), slog.String("method", r.Method))
		next.ServeHTTP(w, r)
	})
}

func (app *App) CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Skip CSRF entirely for API and static routes
		if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/static/") {
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie("csrf_token")
		if err != nil || cookie.Value == "" {
			token, err := app.GenerateCSRFToken()
			if err != nil {
				app.Error(w, r, "Failed to generate CSRF token: "+err.Error())
				return
			}
			app.SetCSRFTokenCookie(w, token)
			r.AddCookie(&http.Cookie{Name: "csrf_token", Value: token})
		}

		if r.Method == "POST" || r.Method == "PUT" || r.Method == "DELETE" || r.Method == "PATCH" {
			if !app.ValidateCSRFToken(r) {
				http.Error(w, "Invalid or missing CSRF token", http.StatusForbidden)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
