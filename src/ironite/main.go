package main

import (
	// === General ===
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	// === Static & HTML ===
	"embed"
	tmpl "html/template"
)

//go:embed static
var staticFS embed.FS

const (
	outputDir  = "public"
	outputFile = "index.html"
)

// Parsed once at startup; panics immediately if templates are broken.
var listTmpl = tmpl.Must(tmpl.ParseFS(staticFS, "static/base.html", "static/list.html"))
var serverTmpl = tmpl.Must(tmpl.ParseFS(staticFS, "static/base.html", "static/server.html"))

func CreateApp() *App {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	return &App{Log: logger}
}

func main() {
	a := CreateApp()
	if err := a.Run(); err != nil {
		a.Log.Error("run failed", slog.Any("error", err))
		os.Exit(1)
	}
}

func (a *App) Run() error {
	servers, err := a.LoadServers("servers-*.json")
	if err != nil {
		return fmt.Errorf("loading servers: %w", err)
	}
	if len(servers) == 0 {
		a.Log.Warn("no servers found, generating empty list")
	}

	css, err := staticFS.ReadFile("static/style.css")
	if err != nil {
		return fmt.Errorf("reading style.css: %w", err)
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("creating %s/: %w", outputDir, err)
	}

	if err := a.GenerateList(servers, css); err != nil {
		return fmt.Errorf("generating list: %w", err)
	}
	if err := a.GenerateServerPages(servers, css); err != nil {
		return fmt.Errorf("generating server pages: %w", err)
	}
	return nil
}


// LoadServers reads every file matching the glob pattern (e.g. "servers-*.json")
// and merges them into one slice.
func (a *App) LoadServers(pattern string) ([]Server, error) {
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("bad glob pattern %q: %w", pattern, err)
	}

	var servers []Server
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", file, err)
		}

		var batch []Server
		if err := json.Unmarshal(data, &batch); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", file, err)
		}

		servers = append(servers, batch...)
		a.Log.Info("loaded server file", slog.String("file", file), slog.Int("count", len(batch)))
	}

	return servers, nil
}

func (a *App) GenerateList(servers []Server, css []byte) error {
	outPath := filepath.Join(outputDir, outputFile)
	page := PageList{
		Page: Page{
			Style: tmpl.CSS(css),
		},
		Servers: servers,
	}

	err := a.writeAtomic(outPath, func(out *os.File) error {
		return listTmpl.ExecuteTemplate(out, "base.html", page)
	})
	if err != nil {
		return err
	}

	a.Log.Info("generated list", slog.String("path", outPath), slog.Int("servers", len(servers)))
	return nil
}

func (a *App) GenerateServerPages(servers []Server, css []byte) error {
	serversDir := filepath.Join(outputDir, "servers")
	if err := os.MkdirAll(serversDir, 0o755); err != nil {
		return fmt.Errorf("creating %s/: %w", serversDir, err)
	}

	for _, s := range servers {
		outPath := filepath.Join(serversDir, s.Slug()+".html")
		page := PageServer{
			Page: Page{
				Style: tmpl.CSS(css),
			},
			Server: s,
		}

		err := a.writeAtomic(outPath, func(out *os.File) error {
			return serverTmpl.ExecuteTemplate(out, "base.html", page)
		})
		if err != nil {
			return fmt.Errorf("server page %s: %w", s.Name, err)
		}
	}

	a.Log.Info("generated server pages", slog.String("dir", serversDir), slog.Int("count", len(servers)))
	return nil
}


// writeAtomic renders fn into path via a temp file + rename, so readers
// never see a half-written page.
func (a *App) writeAtomic(path string, render func(out *os.File) error) (err error) {
	tmpPath := path + ".tmp"

	out, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("creating %s: %w", tmpPath, err)
	}
	defer func() {
		out.Close() // safety net; no-op error if already closed
		if err != nil {
			os.Remove(tmpPath)
		}
	}()

	if err := render(out); err != nil {
		return err
	}

	if err := out.Close(); err != nil {
		return fmt.Errorf("closing %s: %w", tmpPath, err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("renaming to %s: %w", path, err)
	}
	return nil
}
