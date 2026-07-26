package main

import (
	// === General ===
	"os"
	"log/slog"
	"encoding/json"

	// === Static & HTML ===
	tmpl "html/template"
	"embed"
)

//go:embed static
var staticFS embed.FS

func CreateApp() *App {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	a := &App {
		Log: logger,
	}

	return a
}

func main() {
	a := CreateApp()

	file := "servers-0.json" // TODO: load every "servers-*.json" file inside App.LoadServers
	servers := a.LoadServers(file)
	a.GenerateList(servers)
}

func (a App) LoadServers(file string) []Server {
	// Open the file for reading
	data, err := os.ReadFile(file)
	if err != nil {
		a.Log.Error("Failed to open file", slog.String("error", err.Error()))
		return nil
	}

	var servers []Server
	err = json.Unmarshal(data, &servers)
	if err != nil {
		a.Log.Error("Failed to unmarshal JSON", slog.String("error", err.Error()))
		return nil
	}

	return servers
}

func (a App) GenerateList(servers []Server) {
	t, err := tmpl.ParseFS(staticFS, "static/base.html", "static/list.html")
	if err != nil {
		a.Log.Error("Failed to parse template", slog.String("error", err.Error()))
		return
	}

	css, err := staticFS.ReadFile("static/style.css")
	if err != nil {
		a.Log.Error("Failed to read style.css", slog.String("error", err.Error()))
		return
	}

	out, err := os.Create("out.html")
	if err != nil {
		a.Log.Error("Failed to create output file", slog.String("file", "out.html"), slog.String("error", err.Error()))
		return
	}
	defer out.Close()

	page := PageList{
		Style:   tmpl.CSS(css),
		Servers: servers,
	}

	if err := t.ExecuteTemplate(out, "base.html", page); err != nil {
		a.Log.Error("Template execution failed", slog.String("error", err.Error()))
		return
	}

	a.Log.Info("Generated list")
}
