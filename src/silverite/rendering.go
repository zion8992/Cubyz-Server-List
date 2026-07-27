package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"
	"embed"
	"math/rand"
)

//go:embed static/*
var staticFS embed.FS

//go:embed templates/*
var templateFS embed.FS

// Where templates live inside templateFS.
const (
	tmplDir      = "templates"
	tmplBase     = tmplDir + "/base.html"
	tmplError    = "error.html" // page name, must exist as templates/error.html
	tmplPageGlob = tmplDir + "/*.html"
	tmplPartGlob = tmplDir + "/partials/*.html" // optional, ok if empty
)

// BasePageData is what every template is executed with.
// Anything page-specific goes into Data.
type BasePageData struct {
	Root    string // URL prefix, always ends in "/" (used by base.html)
	Title   string
	Path    string // r.URL.Path, handy for marking the active nav item
	Year    int
	Error   string // rendered by error.html
	Flash   string
	Data    any // <- custom page data
}

// ---------------------------------------------------------------------------
// Function cache
// ---------------------------------------------------------------------------

// newFuncMap builds the func cache once at startup.
func newFuncMap(root string) template.FuncMap {
	return template.FuncMap{
		// {{url "static/style.css"}} -> "/prefix/static/style.css"
		"url": func(p string) string {
			return root + strings.TrimPrefix(p, "/")
		},
		"humanDate": func(t time.Time) string {
			if t.IsZero() {
				return ""
			}
			return t.UTC().Format("2006-01-02 15:04")
		},
		"isoDate": func(t time.Time) string {
			if t.IsZero() {
				return ""
			}
			return t.UTC().Format("2006-01-02")
		},
		"yesno": func(b bool) string {
			if b {
				return "Yes"
			}
			return "No"
		},
		"safeHTML": func(s string) template.HTML { return template.HTML(s) },
		"add":      func(a, b int) int { return a + b },
		"join":     strings.Join,
		"lower":    strings.ToLower,
		"title":    strings.Title,
		"motd": func() string {
			motds := make([]string, 0)
			motds = append(motds,
				"Snails",
				"Cubert has one ear",
				"Moffalo is GOATED",
				"This is an motd",
				"For some reason, I put this here",
				"Astr0_Steve was here =)",
				"ikabod?",
			)
			return motds[rand.Intn(len(motds))]
		},
	}
}

// ---------------------------------------------------------------------------
// Template cache
// ---------------------------------------------------------------------------

// newTemplateCache parses base.html + (optional partials) + each page into its
// own *template.Template, keyed by the page's file name ("index.html").
func newTemplateCache(funcs template.FuncMap) (map[string]*template.Template, error) {
	cache := make(map[string]*template.Template)

	pages, err := fs.Glob(templateFS, tmplPageGlob)
	if err != nil {
		return nil, err
	}
	if len(pages) == 0 {
		return nil, fmt.Errorf("no templates matched %q — check your //go:embed line", tmplPageGlob)
	}

	// Partials are optional; a missing directory just yields an empty slice.
	partials, err := fs.Glob(templateFS, tmplPartGlob)
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := path.Base(page)
		if name == path.Base(tmplBase) {
			continue // base isn't a page
		}

		files := make([]string, 0, len(partials)+2)
		files = append(files, tmplBase)
		files = append(files, partials...)
		files = append(files, page)

		ts, err := template.New(name).Funcs(funcs).ParseFS(templateFS, files...)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", page, err)
		}
		cache[name] = ts
	}

	if _, ok := cache[tmplError]; !ok {
		return nil, fmt.Errorf("missing %s/%s", tmplDir, tmplError)
	}
	return cache, nil
}

// ---------------------------------------------------------------------------
// App plumbing
// ---------------------------------------------------------------------------

// InitTemplates fills the func + template caches. Call it once from main.
func (app *App) InitTemplates() error {
	if app.Root == "" {
		app.Root = "/"
	}
	if !strings.HasSuffix(app.Root, "/") {
		app.Root += "/"
	}

	app.funcs = newFuncMap(app.Root)

	cache, err := newTemplateCache(app.funcs)
	if err != nil {
		return err
	}
	app.mu.Lock()
	app.templateCache = cache
	app.mu.Unlock()
	return nil
}

// lookup returns the parsed set for a page, re-parsing everything when
// app.Dev is set so you don't have to restart while editing templates.
func (app *App) lookup(page string) (*template.Template, error) {
	if app.Dev {
		cache, err := newTemplateCache(app.funcs)
		if err != nil {
			return nil, err
		}
		app.mu.Lock()
		app.templateCache = cache
		app.mu.Unlock()
	}

	app.mu.RLock()
	ts, ok := app.templateCache[page]
	app.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("template %q does not exist in cache", page)
	}
	return ts, nil
}

// NewPageData seeds a BasePageData with everything base.html needs.
func (app *App) NewPageData(r *http.Request, title string, data any) *BasePageData {
	p := ""
	if r != nil {
		p = r.URL.Path
	}
	return &BasePageData{
		Root:  app.Root,
		Title: title,
		Path:  p,
		Year:  time.Now().Year(),
		Data:  data,
	}
}

// ---------------------------------------------------------------------------
// The two functions you actually call from handlers
// ---------------------------------------------------------------------------

// Render renders base.html + the named page (e.g. "index.html").
// `data` is anything you like; it lands in {{.Data}}.
func (app *App) Render(w http.ResponseWriter, r *http.Request, status int, page, title string, data any) {
	app.renderPage(w, r, status, page, app.NewPageData(r, title, data))
}

// RenderData is the escape hatch when you want to set Flash/Error/etc. yourself.
func (app *App) RenderData(w http.ResponseWriter, r *http.Request, status int, page string, pd *BasePageData) {
	if pd == nil {
		pd = app.NewPageData(r, "", nil)
	}
	if pd.Root == "" {
		pd.Root = app.Root
	}
	app.renderPage(w, r, status, page, pd)
}

// Error renders base.html + error.html with msg shown to the user.
func (app *App) Error(w http.ResponseWriter, r *http.Request, status int, msg string) {
	if msg == "" {
		msg = http.StatusText(status)
	}
	pd := app.NewPageData(r, fmt.Sprintf("%d %s", status, http.StatusText(status)), nil)
	pd.Error = msg
	app.renderPage(w, r, status, tmplError, pd)
}

// Errorf is the printf flavour.
func (app *App) Errorf(w http.ResponseWriter, r *http.Request, status int, format string, a ...any) {
	app.Error(w, r, status, fmt.Sprintf(format, a...))
}

// ServerError logs the real error and shows a generic message.
func (app *App) ServerError(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("server error", "err", err, "path", r.URL.Path, "method", r.Method)
	app.Error(w, r, http.StatusInternalServerError, "Something went wrong on our end. Please try again later.")
}

// NotFound is usable directly as a http.HandlerFunc / mux NotFoundHandler.
func (app *App) NotFound(w http.ResponseWriter, r *http.Request) {
	app.Error(w, r, http.StatusNotFound, "That page could not be found.")
}

// renderPage does the buffered write so a half-rendered page never reaches
// the client with a 200 already on the wire.
func (app *App) renderPage(w http.ResponseWriter, r *http.Request, status int, page string, pd *BasePageData) {
	ts, err := app.lookup(page)
	if err != nil {
		app.fallbackError(w, r, err)
		return
	}

	buf := new(bytes.Buffer)
	// base.html is the entry point; the page supplies the "title"/"body" blocks.
	if err := ts.ExecuteTemplate(buf, path.Base(tmplBase), pd); err != nil {
		app.fallbackError(w, r, fmt.Errorf("executing %s: %w", page, err))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = buf.WriteTo(w)
}

// fallbackError avoids infinite recursion if error.html itself is broken.
func (app *App) fallbackError(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("render failed", "err", err, "path", r.URL.Path)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// ---------------------------------------------------------------------------
// Static files (bonus, since base.html links {{.Root}}static/…)
// ---------------------------------------------------------------------------

func (app *App) StaticHandler() http.Handler {
    sub, err := fs.Sub(staticFS, "static")
    if err != nil {
        panic(err) // embed guarantees this at build time
    }
    prefix := strings.TrimSuffix(app.Root+"static/", "/") // "/static" or "/silverite/static"
    return http.StripPrefix(prefix, http.FileServer(http.FS(sub)))
}


var _ = sync.RWMutex{} // (mutex lives on App, see below)
