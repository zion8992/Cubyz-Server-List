package main

import (
	// === General ===
	"bytes"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	// === Static & HTML ===
	"embed"
	tmpl "html/template"
)

//go:embed static
var staticFS embed.FS

// Parsed once at startup; panics immediately if templates are broken.
var (
	listTmpl   = mustPage("static/list.html")
	serverTmpl = mustPage("static/server.html")
)

func mustPage(page string) *tmpl.Template {
	return tmpl.Must(tmpl.ParseFS(staticFS, "static/base.html", page))
}

func CreateApp() *App {
	outDir := flag.String("out", "public", "output directory")
	pattern := flag.String("glob", "servers-*.json", "glob pattern for server JSON files")
	baseURL := flag.String("base-url", "", "public base URL (e.g. https://example.com); enables sitemap.xml")
	logLevel := flag.String("log-level", "info", "log level: debug, info, warn, error")
	flag.Parse()

	var level slog.Level
	if err := level.UnmarshalText([]byte(*logLevel)); err != nil {
		level = slog.LevelInfo
	}
	// stderr keeps stdout free for future pipeable output.
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))

	return &App{
		Log:       logger,
		Now:       buildTime(logger),
		OutputDir: *outDir,
		Pattern:   *pattern,
		BaseURL:   strings.TrimRight(*baseURL, "/"),
	}
}

// buildTime honours SOURCE_DATE_EPOCH so builds are reproducible:
// regenerating with unchanged data yields byte-identical output.
func buildTime(log *slog.Logger) time.Time {
	if v := os.Getenv("SOURCE_DATE_EPOCH"); v != "" {
		if sec, err := strconv.ParseInt(v, 10, 64); err == nil {
			return time.Unix(sec, 0).UTC()
		}
		log.Warn("invalid SOURCE_DATE_EPOCH, using wall clock", slog.String("value", v))
	}
	return time.Now()
}

func main() {
	a := CreateApp()
	if err := a.Run(); err != nil {
		a.Log.Error("run failed", slog.Any("error", err))
		os.Exit(1)
	}
}

func (a *App) Run() error {
	servers, err := a.LoadServers(a.Pattern)
	if err != nil {
		return fmt.Errorf("loading servers: %w", err)
	}
	if len(servers) == 0 {
		a.Log.Warn("no servers found, generating empty list")
	}

	// Deterministic output regardless of JSON file/entry order.
	slices.SortFunc(servers, func(x, y Server) int {
		return strings.Compare(strings.ToLower(x.Name), strings.ToLower(y.Name))
	})

	slugs, err := validateSlugs(servers)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(a.OutputDir, 0o755); err != nil {
		return fmt.Errorf("creating %s/: %w", a.OutputDir, err)
	}

	if err := a.CopyStatic(); err != nil {
		return fmt.Errorf("copying static assets: %w", err)
	}
	if err := a.GenerateList(servers); err != nil {
		return fmt.Errorf("generating list: %w", err)
	}
	if err := a.GenerateServerPages(servers, slugs); err != nil {
		return fmt.Errorf("generating server pages: %w", err)
	}
	if err := a.GenerateJSONIndex(servers); err != nil {
		return fmt.Errorf("generating JSON index: %w", err)
	}
	if a.BaseURL != "" {
		if err := a.GenerateSitemap(servers); err != nil {
			return fmt.Errorf("generating sitemap: %w", err)
		}
	}
	return nil
}

// validateSlugs rejects empty slugs and collisions (e.g. "My Server" vs
// "my_server"), which would otherwise silently overwrite pages.
func validateSlugs(servers []Server) (map[string]bool, error) {
	owner := make(map[string]string, len(servers))
	slugs := make(map[string]bool, len(servers))
	for _, s := range servers {
		slug := s.Slug()
		if slug == "" {
			return nil, fmt.Errorf("server %q produces an empty slug", s.Name)
		}
		if prev, ok := owner[slug]; ok {
			return nil, fmt.Errorf("slug collision: %q and %q both map to %q", prev, s.Name, slug)
		}
		owner[slug] = s.Name
		slugs[slug] = true
	}
	return slugs, nil
}

// LoadServers reads every file matching the glob pattern (e.g. "servers-*.json")
// and merges them into one slice. Unknown JSON fields and entries missing
// name/ip are rejected so typos fail loudly instead of vanishing.
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

		dec := json.NewDecoder(bytes.NewReader(data))
		dec.DisallowUnknownFields()
		var batch []Server
		if err := dec.Decode(&batch); err != nil {
			return nil, fmt.Errorf("parsing %s: %w", file, err)
		}
		for i, s := range batch {
			if s.Name == "" || s.IP == "" {
				return nil, fmt.Errorf("%s: entry %d missing name or ip", file, i)
			}
			if s.DateCreated.IsZero() {
				a.Log.Warn("server has no date_created",
					slog.String("file", file), slog.String("server", s.Name))
			}
		}

		servers = append(servers, batch...)
		a.Log.Info("loaded server file", slog.String("file", file), slog.Int("count", len(batch)))
	}

	return servers, nil
}

// CopyStatic mirrors the embedded static/ tree into <out>/static/,
// skipping the *.html templates (they are build inputs, not site assets).
// Files that no longer exist in the embed are removed.
func (a *App) CopyStatic() error {
	dstRoot := filepath.Join(a.OutputDir, "static")
	keep := make(map[string]bool)

	err := fs.WalkDir(staticFS, "static", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(strings.TrimPrefix(p, "static"), "/")
		if d.IsDir() {
			if rel == "" {
				return os.MkdirAll(dstRoot, 0o755)
			}
			keep[rel] = true
			return os.MkdirAll(filepath.Join(dstRoot, filepath.FromSlash(rel)), 0o755)
		}

		keep[rel] = true
		data, err := staticFS.ReadFile(p)
		if err != nil {
			return fmt.Errorf("reading embedded %s: %w", p, err)
		}
		return a.writeAtomic(filepath.Join(dstRoot, filepath.FromSlash(rel)), func(out *os.File) error {
			_, werr := out.Write(data)
			return werr
		})
	})
	if err != nil {
		return err
	}

	// Remove stale assets left over from previous builds.
	var stale []string
	err = filepath.WalkDir(dstRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil || p == dstRoot {
			return err
		}
		rel := filepath.ToSlash(strings.TrimPrefix(p, dstRoot+string(filepath.Separator)))
		if !keep[rel] {
			stale = append(stale, p)
			if d.IsDir() {
				return fs.SkipDir
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	for _, p := range stale {
		if err := os.RemoveAll(p); err != nil {
			return fmt.Errorf("removing stale asset %s: %w", p, err)
		}
		a.Log.Info("removed stale asset", slog.String("path", p))
	}

	a.Log.Info("copied static assets", slog.String("dir", dstRoot), slog.Int("count", len(keep)))
	return nil
}

func (a *App) GenerateList(servers []Server) error {
	outPath := filepath.Join(a.OutputDir, "index.html")
	page := PageList{
		Page:    a.page(""), // index.html sits at the site root
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

// GenerateServerPages writes servers/<slug>/index.html (clean URLs on any
// static host) and removes pages for servers no longer in the data.
func (a *App) GenerateServerPages(servers []Server, slugs map[string]bool) error {
	serversDir := filepath.Join(a.OutputDir, "servers")
	if err := os.MkdirAll(serversDir, 0o755); err != nil {
		return fmt.Errorf("creating %s/: %w", serversDir, err)
	}

	for _, s := range servers {
		pageDir := filepath.Join(serversDir, s.Slug())
		if err := os.MkdirAll(pageDir, 0o755); err != nil {
			return fmt.Errorf("creating %s/: %w", pageDir, err)
		}
		page := PageServer{
			Page:   a.page("../../"), // two levels up to reach static/
			Server: s,
		}

		err := a.writeAtomic(filepath.Join(pageDir, "index.html"), func(out *os.File) error {
			return serverTmpl.ExecuteTemplate(out, "base.html", page)
		})
		if err != nil {
			return fmt.Errorf("server page %s: %w", s.Name, err)
		}
	}

	// Remove pages for servers that were deleted from the JSON files
	// (also catches old flat <slug>.html files from the previous layout).
	entries, err := os.ReadDir(serversDir)
	if err != nil {
		return fmt.Errorf("reading %s/: %w", serversDir, err)
	}
	for _, e := range entries {
		if slugs[e.Name()] {
			continue
		}
		stalePath := filepath.Join(serversDir, e.Name())
		if err := os.RemoveAll(stalePath); err != nil {
			return fmt.Errorf("removing stale page %s: %w", stalePath, err)
		}
		a.Log.Info("removed stale server page", slog.String("path", stalePath))
	}

	a.Log.Info("generated server pages", slog.String("dir", serversDir), slog.Int("count", len(servers)))
	return nil
}

// GenerateJSONIndex writes a machine-readable servers.json alongside the HTML.
func (a *App) GenerateJSONIndex(servers []Server) error {
	type entry struct {
		Name        string    `json:"name"`
		IP          string    `json:"ip"`
		Slug        string    `json:"slug"`
		Page        string    `json:"page"`
		DateCreated time.Time `json:"date_created"`
	}

	entries := make([]entry, 0, len(servers))
	for _, s := range servers {
		entries = append(entries, entry{
			Name:        s.Name,
			IP:          s.IP,
			Slug:        s.Slug(),
			Page:        "servers/" + s.Slug() + "/",
			DateCreated: s.DateCreated,
		})
	}

	outPath := filepath.Join(a.OutputDir, "servers.json")
	err := a.writeAtomic(outPath, func(out *os.File) error {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(entries)
	})
	if err != nil {
		return err
	}
	a.Log.Info("generated JSON index", slog.String("path", outPath))
	return nil
}

func (a *App) GenerateSitemap(servers []Server) error {
	type urlEntry struct {
		Loc     string `xml:"loc"`
		LastMod string `xml:"lastmod"`
	}
	type urlSet struct {
		XMLName xml.Name   `xml:"urlset"`
		Xmlns   string     `xml:"xmlns,attr"`
		URLs    []urlEntry `xml:"url"`
	}

	lastMod := a.Now.Format("2006-01-02")
	set := urlSet{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  []urlEntry{{Loc: a.BaseURL + "/", LastMod: lastMod}},
	}
	for _, s := range servers {
		set.URLs = append(set.URLs, urlEntry{
			Loc:     a.BaseURL + "/servers/" + s.Slug() + "/",
			LastMod: lastMod,
		})
	}

	outPath := filepath.Join(a.OutputDir, "sitemap.xml")
	err := a.writeAtomic(outPath, func(out *os.File) error {
		if _, err := out.WriteString(xml.Header); err != nil {
			return err
		}
		enc := xml.NewEncoder(out)
		enc.Indent("", "  ")
		return enc.Encode(set)
	})
	if err != nil {
		return err
	}
	a.Log.Info("generated sitemap", slog.String("path", outPath), slog.Int("urls", len(set.URLs)))
	return nil
}

// page builds the shared Page fields. root is the relative prefix back to
// the site root ("" for index.html, "../../" for server pages).
func (a *App) page(root string) Page {
	return Page{
		Root:           root,
		GenerationDate: a.Now.Format("January 2 2006 at 15:04:05 MST / -0700"),
	}
}

// writeAtomic renders fn into path via a temp file + rename, so readers
// never see a half-written page. CreateTemp gives each run a unique temp
// name, so concurrent builds can't trample each other.
func (a *App) writeAtomic(path string, render func(out *os.File) error) (err error) {
	out, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("creating temp for %s: %w", path, err)
	}
	tmpPath := out.Name()
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
	// CreateTemp uses 0600; published files should be world-readable.
	if err := os.Chmod(tmpPath, 0o644); err != nil {
		return fmt.Errorf("chmod %s: %w", tmpPath, err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("renaming to %s: %w", path, err)
	}
	return nil
}
