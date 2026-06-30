package main

import(
	"net/http"
	"fmt"
)

func (a *App) SparkUpdatePOST(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	token := r.FormValue("token")
	
	// check if token is valid
}