package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Struct that stores all data that needs to be shown in a page
type Page struct {
	// Sessions
	IsLoggedIn bool

	// Theming
	//Theme string // "light" or "dark"

	// Errors
	Err string
}

func (a *App) SlashHandler(w http.ResponseWriter, r *http.Request) {
	page := strings.TrimPrefix(r.URL.Path, "/")
	if page == "" {
		page = "home"
	}

	path := "./templates/" + page + ".html"

	if _, err := os.Stat(path); os.IsNotExist(err) {
		a.Return404(w, r)
		return
	}

	data := Page{
		IsLoggedIn: a.HasSessionToken(r),
	}

	a.render(w, r, http.StatusOK, page+".html", data)
}

func (a *App) LogoutHandler(w http.ResponseWriter, r *http.Request) {
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

	// Expire every cookie the browser sent with this request
	for _, c := range r.Cookies() {
		http.SetCookie(w, &http.Cookie{
			Name:     c.Name,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			Expires:  time.Unix(0, 0),
			HttpOnly: c.HttpOnly,
			Secure:   c.Secure,
			SameSite: c.SameSite,
		})
	}

	// Also force-clear known app cookies in case they weren't sent
	known := []string{"session_token", "csrf_token"}
	for _, name := range known {
		http.SetCookie(w, &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
		})
	}

	u := User{
		ID: uid,
	}

	err = a.LogoutUser(u)
	if err != nil {
		a.Error(w, r, "Failed to log you out: "+err.Error())
		return
	}

	data := Page{IsLoggedIn: false}
	a.render(w, r, http.StatusOK, "loggedout.html", data)
}

func (a *App) RegisterGET(w http.ResponseWriter, r *http.Request) {
	data := Page{
		IsLoggedIn: a.HasSessionToken(r),
	}

	a.render(w, r, http.StatusOK, "register.html", data)
}

func (a *App) RegisterPOST(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	ok, err := a.ValidateFormFields(r.FormValue("username"), r.FormValue("password"), r.FormValue("email"), true)

	if !ok {
		a.Error(w, r, "Form field validation failed: "+err.Error())
		return
	}

	_, err = a.GetUserIDByUsername(r.FormValue("username"))
	if err != nil && err != sql.ErrNoRows {
		a.Error(w, r, "Failed to check user: "+err.Error())
		return
	}

	if err == nil {
		a.Error(w, r, "User already exists")
		return
	}

	u := User{
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
		Email:    r.FormValue("email"),
	}

	_, err = a.CreateUser(u)
	if err != nil {
		a.Error(w, r, "Failed to create user! ", err.Error())
		return
	}

	data := Page{
		IsLoggedIn: a.HasSessionToken(r),
	}

	a.render(w, r, http.StatusOK, "register_ok.html", data)
}

func (a *App) LoginGET(w http.ResponseWriter, r *http.Request) {
	data := Page{
		IsLoggedIn: a.HasSessionToken(r),
	}

	a.render(w, r, http.StatusOK, "login.html", data)
}

func (a *App) LoginPOST(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	id, err := a.GetUserIDByUsername(r.FormValue("username"))
	if err == sql.ErrNoRows {
		a.Error(w, r, "Invalid credentials")
		return
	}
	if err != nil {
		a.Error(w, r, "Failed to check user: "+err.Error())
		return
	}

	var ok bool
	ok, err = a.CheckPasswordDB(id, r.FormValue("password"))
	if err != nil {
		a.Error(w, r, "Failed to check if your password is correct: "+err.Error())
		return
	}

	if !ok {
		a.Error(w, r, "Invalid credentials")
		return
	}

	var token string
	token, err = a.GenerateSessionToken()
	if err != nil {
		a.Error(w, r, "Failed to create a session token to log you in: ", err.Error())
		return
	}

	expires := time.Now().Add(a.DefaultExpiry)
	err = a.SetSessionToken(id, token, expires)

	c := http.Cookie{
		Name:    "session_token",
		Value:   token,
		Expires: expires,
	}

	http.SetCookie(w, &c)

	http.Redirect(w, r, "/", http.StatusSeeOther)
	return
}

func (a *App) AccountGET(w http.ResponseWriter, r *http.Request) {
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

	user, err := a.GetUser(uid)
	if err != nil {
		a.Error(w, r, "Failed to load user: "+err.Error())
		return
	}

	servers, err := a.GetServersByUser(uid)
	if err != nil {
		a.Error(w, r, "Failed to load servers: "+err.Error())
		return
	}

	data := struct {
		Page
		User
		Servers []Server
	}{
		Page:    Page{IsLoggedIn: true},
		User:    *user,
		Servers: servers,
	}

	a.render(w, r, http.StatusOK, "account.html", data)
}

func (a *App) AccountPOST(w http.ResponseWriter, r *http.Request) {
	ok, err := a.CheckReqSessionTok(r)
	if err != nil || !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	_, err = a.GetUserIDByUsername(r.FormValue("username"))
	if err != nil && err != sql.ErrNoRows {
		a.Error(w, r, "Failed to check user: "+err.Error())
		return
	}

	if err == nil {
		a.Error(w, r, "Username is taken")
		return
	}

	uid, err := a.GetUIDFromToken(getSessionCookie(r))
	if err != nil {
		a.Error(w, r, "Failed to identify user: "+err.Error())
		return
	}

	r.ParseForm()

	user, err := a.GetUser(uid)
	if err != nil {
		a.Error(w, r, "Failed to load user: "+err.Error())
		return
	}

	user.Username = strings.TrimSpace(r.FormValue("username"))
	user.Email = strings.TrimSpace(r.FormValue("email"))
	user.ProfilePictureURL = strings.TrimSpace(r.FormValue("profile_picture_url"))
	user.Description = strings.TrimSpace(r.FormValue("description"))
	user.Pronouns = strings.TrimSpace(r.FormValue("pronouns"))

	if user.Username == "" || user.Email == "" {
		a.Error(w, r, "Username and email are required")
		return
	}

	ok, err = a.ValidateFormFields(user.Username, user.Email, "", false)
	if !ok {
		a.Error(w, r, "Form field validation failed: "+err.Error())
		return
	}

	if err := a.UpdateUserProfile(*user); err != nil {
		a.Error(w, r, "Failed to update account: "+err.Error())
		return
	}

	http.Redirect(w, r, "/account", http.StatusSeeOther)
}

func (a *App) AccountVerifyPOST(w http.ResponseWriter, r *http.Request) {
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

	r.ParseForm()

	pubkey := strings.TrimSpace(r.FormValue("pubkey"))

	if pubkey == "" {
		a.Error(w, r, "Public Key is required.")
		return
	}

	if err := a.UpdateUserPubkey(uid, pubkey); err != nil {
		a.Error(w, r, "Failed to update account: "+err.Error())
		return
	}

	http.Redirect(w, r, "/account?tab=verified", http.StatusSeeOther)
}

func getSessionCookie(r *http.Request) string {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return ""
	}
	return cookie.Value
}

/** SERVERS **/

func (a *App) ServerCreateGET(w http.ResponseWriter, r *http.Request) {
	ok, err := a.CheckReqSessionTok(r)
	if err != nil || !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := Page{
		IsLoggedIn: true,
	}

	a.render(w, r, http.StatusOK, "server_create.html", data)
}

func (a *App) ServerCreatePOST(w http.ResponseWriter, r *http.Request) {
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

	r.ParseForm()

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		a.Error(w, r, "Server name is required")
		return
	}

	ip := strings.TrimSpace(r.FormValue("ip"))
	if ip == "" {
		a.Error(w, r, "Server IP is required")
		return
	}

	s := Server{
		Name:        name,
		Description: strings.TrimSpace(r.FormValue("description")),
		IP:          ip,
		OwnerID:     uid,
	}

	id, err := a.CreateServer(s)
	if err != nil {
		a.Error(w, r, "Failed to create server: "+err.Error())
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/servers/%d", id), http.StatusSeeOther)
}

func (a *App) ServerInfo(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		a.Return404(w, r)
		return
	}

	server, err := a.GetServer(id)
	if err != nil {
		if err == sql.ErrNoRows {
			a.Return404(w, r)
			return
		}
		a.Error(w, r, "Failed to load server: "+err.Error())
		return
	}

	isOwner := false
	if ok, _ := a.CheckReqSessionTok(r); ok {
		uid, err := a.GetUIDFromToken(getSessionCookie(r))
		if err == nil && uid == server.OwnerID {
			isOwner = true
		}
	}

	data := struct {
		Page
		Server
		IsOwner bool
	}{
		Page:    Page{IsLoggedIn: a.HasSessionToken(r)},
		Server:  *server,
		IsOwner: isOwner,
	}

	a.render(w, r, http.StatusOK, "server_info.html", data)
}

func (a *App) ServerEditGET(w http.ResponseWriter, r *http.Request) {
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

	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		a.Return404(w, r)
		return
	}

	server, err := a.GetServer(id)
	if err != nil {
		if err == sql.ErrNoRows {
			a.Return404(w, r)
			return
		}
		a.Error(w, r, "Failed to load server: "+err.Error())
		return
	}

	if server.OwnerID != uid {
		a.Return404(w, r)
		return
	}

	data := struct {
		Page
		Server
	}{
		Page:   Page{IsLoggedIn: true},
		Server: *server,
	}

	a.render(w, r, http.StatusOK, "server_edit.html", data)
}

func (a *App) ServerEditPOST(w http.ResponseWriter, r *http.Request) {
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

	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		a.Return404(w, r)
		return
	}

	server, err := a.GetServer(id)
	if err != nil {
		if err == sql.ErrNoRows {
			a.Return404(w, r)
			return
		}
		a.Error(w, r, "Failed to load server: "+err.Error())
		return
	}

	if server.OwnerID != uid {
		a.Return404(w, r)
		return
	}

	r.ParseForm()

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		a.Error(w, r, "Server name is required")
		return
	}

	s := Server{
		ID:           id,
		Name:         name,
		Description:  strings.TrimSpace(r.FormValue("description")),
		XMLFeedLink:  strings.TrimSpace(r.FormValue("xml_feed_link")),
		IP:           strings.TrimSpace(r.FormValue("ip")),
		Gamemodes:    strings.TrimSpace(r.FormValue("gamemodes")),
		Version:      strings.TrimSpace(r.FormValue("version")),
		Languages:    strings.TrimSpace(r.FormValue("languages")),
		RequiresMods: r.FormValue("requires_mods") == "on",
		WebsiteURL:   strings.TrimSpace(r.FormValue("website_url")),
		ChatURL:      strings.TrimSpace(r.FormValue("chat_url")),
		OwnerID:      uid,
	}

	if err := a.UpdateServer(s); err != nil {
		a.Error(w, r, "Failed to update server: "+err.Error())
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/servers/%d", id), http.StatusSeeOther)
}

func (a *App) ServerDeletePOST(w http.ResponseWriter, r *http.Request) {
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

	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		a.Return404(w, r)
		return
	}

	if err := a.DeleteServer(id, uid); err != nil {
		a.Error(w, r, "Failed to delete server: "+err.Error())
		return
	}

	http.Redirect(w, r, "/account", http.StatusSeeOther)
}
