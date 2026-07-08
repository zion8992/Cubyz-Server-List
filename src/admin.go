package main

import (
	"net/http"
)

func (a *App) AdminPanel(w http.ResponseWriter, r *http.Request) {
	ok, err := a.CheckReqSessionTok(r)
	if err != nil || !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// get user UID from session tok
	uid, err := a.GetUIDFromToken(getSessionCookie(r))
	if err != nil {
		a.Error(w, r, "Failed to identify user: "+err.Error())
		return
	}

	ok, err = a.IsUserAdmin(uid)
	if err != nil {
		a.Error(w, r, "Failed to check your account status: "+err.Error())
		return
	}

	if !ok {
		a.Error(w, r, "This page requires admin access.")
		return
	}

	data := struct {
		Page
	}{
		Page: Page{IsLoggedIn: true, CSRFToken: a.GetCSRFTokenFromRequest(r)},
	}

	a.render(w, r, http.StatusOK, "admin_panel.html", data)
}
