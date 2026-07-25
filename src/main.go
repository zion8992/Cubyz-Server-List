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
	handler := alice.New(app.RouteLogger, app.CSRFMiddleware).Then(mux)

	staticSubFS, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}
	staticFileServer := http.FileServer(http.FS(staticSubFS))
	mux.Handle("/static/", http.StripPrefix("/static/", staticFileServer))

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
	mux.HandleFunc("POST /account-changepass", app.AccountChangePassPOST)
	mux.HandleFunc("GET /account-2fa", app.Enable2faGET)
	mux.HandleFunc("POST /account-2fa-enable", app.Enable2faPOST)

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
	mux.HandleFunc("GET /api/v1/listServers", app.ApiListServers)

	// user profiles
	mux.HandleFunc("GET /user/{id}", app.UserInfo)

	// admin panel
	mux.HandleFunc("GET /admin/panel", app.AdminPanel)
	mux.HandleFunc("POST /admin/suspend", app.AdminSuspendUser)

	// list
	mux.HandleFunc("GET /list", app.ServerListGET)

	// theme toggler
	mux.HandleFunc("/theme", app.ToggleTheme)

	app.Log.Info(fmt.Sprintf("Listening on port %s", *port))
	err = http.ListenAndServe(*port, handler)
	if err != nil {
		app.Log.Error("server failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func NewApp() *App {
	var DBPass = flag.String("dbpass", "H0EeLfLnO,xDEVELOPERSx4c!#%", "Root user's database password for MySQL")
	var DBIP = flag.String("dburl", "127.0.0.1:3306", "IP and Port of the database")
	var AppName = flag.String("appname", "Ironite Server List", "Name for the application")

	flag.Parse()

	dsn := fmt.Sprintf("root:%s@tcp(%s)/ironite?parseTime=true", *DBPass, *DBIP)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("Failed to connect to the database: %s\n", err.Error())
		os.Exit(0)
	}

	if err := db.Ping(); err != nil {
		fmt.Printf("Failed to ping database: %s\n", err.Error())
		os.Exit(0)
	}

	templateCache, err := newEmbedTemplateCache()
	if err != nil {
		fmt.Printf("Failed to load templates from exe: %s\n", err.Error())
		os.Exit(0)
	}

	a := &App{
		Log:                slog.New(slog.NewTextHandler(os.Stderr, nil)),
		DB:                 db,
		DefaultTokenExpiry: 4 * time.Hour,
		TemplateCache:      templateCache,
		AppName: *AppName,
	}

	if err := a.LoadBannedWords(); err != nil {
		fmt.Printf("Failed to load banned words file: %s\n", err.Error())
		os.Exit(0)
	}

	return a
}
