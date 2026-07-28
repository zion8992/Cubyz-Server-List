package main

import (
	"flag"
	"log/slog"
	"os"
)

func main() {
	var (
		baseURL   = flag.String("base-url", "/", `prefix for every root-relative link (e.g. "/servers" or "https://example.com/servers")`)
		tmplDir   = flag.String("templates", "templates", "template directory")
		staticDir = flag.String("static", "static", "static asset directory")
		outDir    = flag.String("out", "public", "output directory")
		verbose   = flag.Bool("v", false, "debug logging")
	)
	flag.Parse()

	level := slog.LevelInfo
	if *verbose {
		level = slog.LevelDebug
	}
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))

	app := NewApp(log, Config{
		BaseURL:     *baseURL,
		TemplateDir: *tmplDir,
		StaticDir:   *staticDir,
		OutDir:      *outDir,
	})

	if err := app.Build(); err != nil {
		log.Error("build failed", "err", err)
		os.Exit(1)
	}
}
