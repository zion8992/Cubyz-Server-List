package main

import (
	"net/http"
	"strconv"
	"strings"
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

	// --- users tab: filters + paging ---
	q := r.URL.Query()

	f := UserFilter{
		Search: strings.TrimSpace(q.Get("search")),
		Sort:   q.Get("sort"),
	}
	if f.Sort == "" {
		f.Sort = "players" // default: most online players
	}

	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	f.Page = page

	users, total, err := a.ListUsersAdmin(f, AdminUsersPerPage)
	if err != nil {
		a.Error(w, r, "Failed to load users: "+err.Error())
		return
	}

	totalPages := (total + AdminUsersPerPage - 1) / AdminUsersPerPage
	if totalPages < 1 {
		totalPages = 1
	}
	// Clamp current page if it ran off the end (e.g. filters changed).
	if f.Page > totalPages {
		f.Page = totalPages
	}

	data := struct {
		Page
		Users       []AdminUserRow
		Filter      UserFilter
		CurrentPage int
		TotalPages  int
		TotalUsers  int
		HasPrev     bool
		HasNext     bool
		PrevPage    int
		NextPage    int
	}{
		Page:        Page{IsLoggedIn: true, CSRFToken: a.GetCSRFTokenFromRequest(r)},
		Users:       users,
		Filter:      f,
		CurrentPage: f.Page,
		TotalPages:  totalPages,
		TotalUsers:  total,
		HasPrev:     f.Page > 1,
		HasNext:     f.Page < totalPages,
		PrevPage:    f.Page - 1,
		NextPage:    f.Page + 1,
	}

	a.render(w, r, http.StatusOK, "admin_panel.html", data)
}

func (a *App) AdminSuspendUser(w http.ResponseWriter, r *http.Request) {
	ok, err := a.CheckReqSessionTok(r)
	if err != nil || !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

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
		a.Error(w, r, "This action requires admin access.")
		return
	}

	r.ParseForm()

	targetID, err := strconv.ParseUint(r.FormValue("user_id"), 10, 64)
	if err != nil {
		a.Error(w, r, "Invalid user ID: "+err.Error())
		return
	}

	// Don't let an admin lock themselves out.
	if targetID == uid {
		a.Error(w, r, "You cannot suspend your own account.")
		return
	}

	suspend := r.FormValue("suspend") == "true"

	if err := a.SetUserSuspended(targetID, suspend); err != nil {
		a.Error(w, r, "Failed to update suspension status: "+err.Error())
		return
	}

	// Return to the same filtered/paged view the admin came from.
	redirect := "/admin/panel"
	if ret := strings.TrimSpace(r.FormValue("return")); ret != "" &&
		strings.HasPrefix(ret, "/admin/panel") {
		redirect = ret
	}
	http.Redirect(w, r, redirect, http.StatusSeeOther)
}
