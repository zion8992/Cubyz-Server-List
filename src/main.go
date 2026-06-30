package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/justinas/alice"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {
	app := NewApp()
	mux := http.NewServeMux()
	handler := alice.New(app.RouteLogger).Then(mux)

	staticFS := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", staticFS))

	// auth
	mux.HandleFunc("GET /register", app.RegisterGET)
	mux.HandleFunc("POST /register", app.RegisterPOST)

	mux.HandleFunc("GET /login", app.LoginGET)
	mux.HandleFunc("POST /login", app.LoginPOST)

	mux.HandleFunc("GET /logout", app.LogoutHandler)

	// account page
	mux.HandleFunc("GET /account", app.AccountGET)
	mux.HandleFunc("POST /account", app.AccountPOST)
	mux.HandleFunc("POST /account-verify", app.AccountVerifyPOST)

	// semi-static pages
	mux.HandleFunc("/", app.SlashHandler)

	// servers
	mux.HandleFunc("GET /servers/create", app.ServerCreateGET)
	mux.HandleFunc("POST /servers/create", app.ServerCreatePOST)
	mux.HandleFunc("GET /servers/{id}", app.ServerInfo)
	mux.HandleFunc("GET /servers/edit/{id}", app.ServerEditGET)
	mux.HandleFunc("POST /servers/edit/{id}", app.ServerEditPOST)
	mux.HandleFunc("POST /servers/delete/{id}", app.ServerDeletePOST)

	app.Log.Info("Listening on :8000...")
	err := http.ListenAndServe(":8000", handler)
	if err != nil {
		app.Log.Error("server failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func NewApp() *App {

	db, err := sql.Open("mysql", "root:H0EeLfLnO,xDEVELOPERSx4c!#%@tcp(127.0.0.1:3306)/ironite?parseTime=true")
	if err != nil {
		fmt.Printf("%s\n", "Failed to connect to database!")
		panic(err)
	}

	if err := db.Ping(); err != nil {
		fmt.Printf("%s\n", "Failed to ping to database!")
		panic(err)
	}

	templateCache, err := newTemplateCache()
	if err != nil {
		fmt.Printf("%s\n", "Failed to create template cache!")
		panic(err)
	}

	a := &App{
		Log:           slog.New(slog.NewTextHandler(os.Stderr, nil)),
		DB:            db,
		DefaultExpiry: 4 * time.Hour,
		TemplateCache: templateCache,
	}

	return a
}
