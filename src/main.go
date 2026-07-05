package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/justinas/alice"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {
	var port = flag.String("port", ":8000", "Port to host the HTTP server")

	app := NewApp()
	mux := http.NewServeMux()
	handler := alice.New(app.RouteLogger).Then(mux)

	staticSubFS, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}
	staticFileServer := http.FileServer(http.FS(staticSubFS))
	mux.Handle("/static/", http.StripPrefix("/static/", staticFileServer))

	mux.HandleFunc("GET /debug/static", func(w http.ResponseWriter, r *http.Request) {
		fs.WalkDir(staticFS, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				fmt.Fprintln(w, "err:", err)
				return nil
			}
			fmt.Fprintln(w, path)
			return nil
		})
	})

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
	mux.HandleFunc("GET /account-delete", app.DeleteAccountGET)
	mux.HandleFunc("POST /account-delete-forever-and-ever", app.DeleteAccountPOST)
	mux.HandleFunc("GET /api/v1/create-token-ui", app.CreateTokenUI)

	// semi-static pages
	mux.HandleFunc("/", app.SlashHandler)

	// servers
	mux.HandleFunc("GET /servers/create", app.ServerCreateGET)
	mux.HandleFunc("POST /servers/create", app.ServerCreatePOST)
	mux.HandleFunc("GET /servers/{id}", app.ServerInfo)
	mux.HandleFunc("GET /servers/edit/{id}", app.ServerEditGET)
	mux.HandleFunc("POST /servers/edit/{id}", app.ServerEditPOST)
	mux.HandleFunc("POST /servers/delete/{id}", app.ServerDeletePOST)

	// api
	mux.HandleFunc("POST /api/v1/sparkUpdate", app.SparkUpdatePOST)
	mux.HandleFunc("POST /api/v1/create-token", app.ApiCreateTokenHandler)
	mux.HandleFunc("POST /api/v1/delete-token", app.ApiDeleteTokenHandler)

	// list
	mux.HandleFunc("GET /list", app.ServerListGET)

	app.Log.Info(fmt.Sprintf("Listening on port %s", *port))
	err = http.ListenAndServe(*port, handler)
	if err != nil {
		app.Log.Error("server failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func NewApp() *App {
	var DBPass = flag.String("dbpass", "H0EeLfLnO,xDEVELOPERSx4c!#%", "Root user's database password for MySQL")
	flag.Parse()

	dsn := fmt.Sprintf("root:%s@tcp(127.0.0.1:3306)/ironite?parseTime=true", *DBPass)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("%s\n", "Failed to connect to database!")
		panic(err)
	}

	if err := db.Ping(); err != nil {
		fmt.Printf("%s\n", "Failed to ping to database!")
		panic(err)
	}

	templateCache, err := newEmbedTemplateCache()
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
