package main

import(
	"net/http"
)

/* ========== HANDLER FUNCTION NAMING SPEC ============ */
/* <METHOD><Path>                                       */
/* Example:                                             */
/* (app *App) GETSlash                                  */

func (app *App) GETSlash(w http.ResponseWriter, r *http.Request) {
	app.Render(w, r, http.StatusOK, "home.html", "Servers", nil)	
}