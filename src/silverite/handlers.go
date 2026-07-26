package main

import(
	"net/http"
)

func (app *App) hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!\n"))
}