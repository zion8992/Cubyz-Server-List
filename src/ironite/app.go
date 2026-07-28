package main

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"github.com/zion8992/netgliss"
)

// 'layoutName' is the name of the base template. Anything on this template is rendered every page
// Page then only need to provide "{{define "content"}}...{{end}}"
const layoutName = "_layout.html"

// === Structs ===
type Config struct {
	BaseURL     string
	TemplateDir string
	StaticDir   string
	OutDir      string
}

type App struct {
	log *slog.Logger
	cfg Config

	// baseURL is normalized: no trailing slash ("" for a site at the root).
	baseURL string
}

func NewApp(log *slog.Logger, cfg Config) *App {
	return &App{
		log:     log,
		cfg:     cfg,
		baseURL: strings.TrimSuffix(cfg.BaseURL, "/"),
	}
}

// Page is the data handed to every template.
type Page struct {
	Title   string
	MOTD    string
	Servers []Server
}

// === Static Site Generator ===

func (a *App) Build() error {
	// Delete and re-create the output directory
	if err := os.RemoveAll(a.cfg.OutDir); err != nil {
		return fmt.Errorf("clean output: %w", err)
	}
	if err := os.MkdirAll(a.cfg.OutDir, 0o755); err != nil {
		return fmt.Errorf("create output: %w", err)
	}

	// copy static files directly to output, "no questions asked"
	if err := a.copyStatic(); err != nil {
		return fmt.Errorf("copy static: %w", err)
	}

	data := &Page{
		Title: "Servers",
		MOTD:  randomMOTD(),
		Servers: servers,
	}
	if err := a.renderPages(data); err != nil {
		return fmt.Errorf("render: %w", err)
	}

	a.log.Info("build complete", "out", a.cfg.OutDir, "base_url", a.URL("/"))
	return nil
}

func (a *App) copyStatic() error {
	if _, err := os.Stat(a.cfg.StaticDir); errors.Is(err, fs.ErrNotExist) {
		a.log.Debug("no static dir, skipping", "dir", a.cfg.StaticDir)
		return nil
	}
	a.log.Debug("copying static files", "from", a.cfg.StaticDir, "to", a.cfg.OutDir)
	return os.CopyFS(a.cfg.OutDir + "/static", os.DirFS(a.cfg.StaticDir))
}

func (a *App) renderPages(data *Page) error {
	partials, err := filepath.Glob(filepath.Join(a.cfg.TemplateDir, "_*.html"))
	if err != nil {
		return err
	}
	pages, err := filepath.Glob(filepath.Join(a.cfg.TemplateDir, "*.html"))
	if err != nil {
		return err
	}

	for _, page := range pages {
		if strings.HasPrefix(filepath.Base(page), "_") { // partials start with '_'
			continue
		}
		if err := a.renderPage(page, partials, data); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) renderPage(page string, partials []string, data *Page) error {
	name := filepath.Base(page)

	ts, err := template.New(name).Funcs(a.Funcs()).ParseFiles(append([]string{page}, partials...)...)
	if err != nil {
		return err
	}

	entry := name
	if ts.Lookup(layoutName) != nil {
		entry = layoutName
	}

	var buf bytes.Buffer
	if err := ts.ExecuteTemplate(&buf, entry, data); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}

	out := filepath.Join(a.cfg.OutDir, name)
	if err := os.WriteFile(out, a.RewriteLinks(buf.Bytes()), 0o644); err != nil {
		return err
	}
	a.log.Debug("rendered page", "template", name, "out", out)
	return nil
}

// URL prefixes a root-relative path with the configured base URL.
func (a *App) URL(p string) string {
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return a.baseURL + p
}

var attrRe = regexp.MustCompile(`(?i)(\b(?:href|src|action|poster|formaction)\s*=\s*)(["'])([^"']*)(["'])`)

// RewriteLinks applies the 'a.baseURL' to every hyperlink tag ("<a>" tag) in the outputted HTML files
// anchor/mailto links are left untouched.
func (a *App) RewriteLinks(html []byte) []byte {
	if a.baseURL == "" {
		return html
	}
	return attrRe.ReplaceAllFunc(html, func(m []byte) []byte {
		g := attrRe.FindSubmatch(m)
		val := string(g[3])
		if !strings.HasPrefix(val, "/") || strings.HasPrefix(val, "//") {
			return m
		}
		return []byte(string(g[1]) + string(g[2]) + a.baseURL + val + string(g[4]))
	})
}
