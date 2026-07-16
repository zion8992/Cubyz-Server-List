package main

import (
	"net/http"
	"net/url"
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

	q := r.URL.Query()

	// Which tab should be shown on load (defaults to users).
	activeTab := q.Get("tab")
	switch activeTab {
	case "users", "servers", "tokens", "actions":
		// valid
	default:
		activeTab = "users"
	}

	// --- users tab: filters + paging ---
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

	// --- servers tab: filters + paging ---
	sf := ServerFilter{
		Search: strings.TrimSpace(q.Get("s_search")),
		Sort:   q.Get("s_sort"),
		Status: q.Get("s_status"), // "" (all) or "online"
	}
	if sf.Sort == "" {
		sf.Sort = "players" // default: most players
	}

	sPage, _ := strconv.Atoi(q.Get("s_page"))
	if sPage < 1 {
		sPage = 1
	}

	allServers, err := a.ListServersAdmin(sf)
	if err != nil {
		a.Error(w, r, "Failed to load servers: "+err.Error())
		return
	}

	// Resolve online status and apply the "online only" filter in Go.
	filtered := make([]Server, 0, len(allServers))
	for i := range allServers {
		online := a.IsServerOnline(allServers[i].ID)
		if online {
			allServers[i].Status = "Online"
		} else {
			allServers[i].Status = "Offline"
		}
		if sf.Status == "online" && !online {
			continue
		}
		filtered = append(filtered, allServers[i])
	}

	sTotal := len(filtered)
	sTotalPages := (sTotal + AdminServersPerPage - 1) / AdminServersPerPage
	if sTotalPages < 1 {
		sTotalPages = 1
	}
	if sPage > sTotalPages {
		sPage = sTotalPages
	}

	start := (sPage - 1) * AdminServersPerPage
	if start > sTotal {
		start = sTotal
	}
	end := start + AdminServersPerPage
	if end > sTotal {
		end = sTotal
	}
	servers := filtered[start:end]

	// Build server-tab page URLs that preserve the current filters and keep
	// the servers tab open.
	serverPageURL := func(p int) string {
		v := url.Values{}
		v.Set("tab", "servers")
		if sf.Search != "" {
			v.Set("s_search", sf.Search)
		}
		if sf.Sort != "" {
			v.Set("s_sort", sf.Sort)
		}
		if sf.Status != "" {
			v.Set("s_status", sf.Status)
		}
		v.Set("s_page", strconv.Itoa(p))
		return "/admin/panel?" + v.Encode()
	}

	data := struct {
		Page
		ActiveTab string

		// Users tab
		Users       []AdminUserRow
		Filter      UserFilter
		CurrentPage int
		TotalPages  int
		TotalUsers  int
		HasPrev     bool
		HasNext     bool
		PrevPage    int
		NextPage    int

		// Servers tab
		Servers      []Server
		SFilter      ServerFilter
		SCurrentPage int
		STotalPages  int
		STotalCount  int
		SHasPrev     bool
		SHasNext     bool
		SPrevURL     string
		SNextURL     string
	}{
		Page:      Page{IsLoggedIn: true, CSRFToken: a.GetCSRFTokenFromRequest(r)},
		ActiveTab: activeTab,

		Users:       users,
		Filter:      f,
		CurrentPage: f.Page,
		TotalPages:  totalPages,
		TotalUsers:  total,
		HasPrev:     f.Page > 1,
		HasNext:     f.Page < totalPages,
		PrevPage:    f.Page - 1,
		NextPage:    f.Page + 1,

		Servers:      servers,
		SFilter:      sf,
		SCurrentPage: sPage,
		STotalPages:  sTotalPages,
		STotalCount:  sTotal,
		SHasPrev:     sPage > 1,
		SHasNext:     sPage < sTotalPages,
		SPrevURL:     serverPageURL(sPage - 1),
		SNextURL:     serverPageURL(sPage + 1),
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

	suspendedUser := User {
		ID: targetID,
	}

	err = a.LogoutUser(suspendedUser)
	if err != nil {
		a.Error(w, r, "Failed to log the suspended user out: "+err.Error())
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
