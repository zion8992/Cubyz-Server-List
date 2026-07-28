package main

import (
	"flag"
	"log/slog"
	"os"
	"github.com/zion8992/netgliss"
)

func main() {
	// === CLI Flags ===
	var (
		baseURL   = flag.String("base-url", "/", `prefix for every root-relative link (e.g. "/servers" or "https://example.com/servers")`)
		tmplDir   = flag.String("templates", "templates", "template directory (required for build)")
		staticDir = flag.String("static", "static", "static assets directory (required for build)")
		outDir    = flag.String("out", "public", "output directory for the static site")
		verbose   = flag.Bool("v", false, "debug logging")
	)
	flag.Parse()

	// === Logging Setup ===

	level := slog.LevelInfo
	if *verbose {
		level = slog.LevelDebug
	}
	
	// === Create App ===
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))

	app := NewApp(log, Config{
		BaseURL:     *baseURL,
		TemplateDir: *tmplDir,
		StaticDir:   *staticDir,
		OutDir:      *outDir,
	})

	// === Load Servers ===
	servers := netgliss.New()
	servers.LoadServers("servers-*.json")

	if err := app.Build(servers); err != nil {
		log.Error("build failed", "err", err)
		os.Exit(1)
	}
}
