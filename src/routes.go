package main

import (
	"database/sql"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Struct that stores all data that needs to be shown in a page
type Page struct {
	// Sessions
	IsLoggedIn bool
	CSRFToken  string

	// Theming
	//Theme string // "light" or "dark"

	// Errors
	Err string
}

func (a *App) SlashHandler(w http.ResponseWriter, r *http.Request) {
	ok, err := a.CheckReqSessionTok(r)
	if err != nil || ok {
		http.Redirect(w, r, "/list", http.StatusSeeOther)
		return
	}

	page := strings.TrimPrefix(r.URL.Path, "/")
	if page == "" {
		page = "home"
	}

	if _, ok := a.TemplateCache[page+".html"]; !ok {
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
		CSRFToken:  a.GetCSRFTokenFromRequest(r),
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

	if banned, word := a.CheckBanned(r.FormValue("username")); banned {
		a.Error(w, r, "Username contains a banned word: "+word)
		return
	}

	if a.CheckTinyString(r.FormValue("username")) {
		a.Error(w, r, "Username too long. Max characters is 20.")
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
		CSRFToken:  a.GetCSRFTokenFromRequest(r),
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

	if banned, word := a.CheckBanned(r.FormValue("username")); banned {
		a.Error(w, r, "Username contains a banned word: "+word+" You will no longer be able to login into this acccount.")
		return
	}

	if a.CheckTinyString(r.FormValue("username")) {
		a.Error(w, r, "Username too long. Max characters is 20. You will no longer be able to login into this account.")
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

	// Block suspended accounts from logging in.
	suspended, err := a.IsUserSuspended(id)
	if err != nil {
		a.Error(w, r, "Failed to check account status: "+err.Error())
		return
	}
	if suspended {
		a.Error(w, r, "This account has been suspended by an administrator.")
		return
	}

	var token string
	token, err = a.GenerateSessionToken()
	if err != nil {
		a.Error(w, r, "Failed to create a session token to log you in: ", err.Error())
		return
	}

	expires := time.Now().Add(a.DefaultTokenExpiry)
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

	tokens, err := a.GetTokensFromUser(uid)
	if err != nil {
		a.Error(w, r, "Failed to load users tokens: "+err.Error())
		return
	}

	servers, err := a.GetServersByUser(uid)
	if err != nil {
		a.Error(w, r, "Failed to load users servers: "+err.Error())
		return
	}

	data := struct {
		Page
		User
		Servers []Server
		Tokens  []Token
	}{
		Page:    Page{IsLoggedIn: true, CSRFToken: a.GetCSRFTokenFromRequest(r)},
		User:    *user,
		Servers: servers,
		Tokens:  tokens,
	}

	a.render(w, r, http.StatusOK, "account.html", data)
}

func (a *App) AccountPOST(w http.ResponseWriter, r *http.Request) {
	ok, err := a.CheckReqSessionTok(r)
	if err != nil || !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if banned, word := a.CheckBanned(r.FormValue("username")); banned {
		a.Error(w, r, "Username contains a banned word: "+word+" You will no longer be able to login into this acccount.")
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

	_, err = a.GetUserIDByUsername(r.FormValue("username"))
	if err != nil && err != sql.ErrNoRows {
		a.Error(w, r, "Failed to check user: "+err.Error())
		return
	}

	if err == nil && strings.TrimSpace(r.FormValue("username")) != user.Username {
		a.Error(w, r, "Username is taken")
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

	/* check string sizes */
	if a.CheckTinyString(user.Username) {
		a.Error(w, r, "Username too long. Max characters is 20.")
		return
	}

	if a.CheckTinyString(user.Email) {
		a.Error(w, r, "Email too long. Max characters is 20.")
		return
	}

	if a.CheckSmallString(user.Description) {
		a.Error(w, r, "Description too long. Max characters is 123.")
		return
	}

	if a.CheckTinyString(user.Pronouns) {
		a.Error(w, r, "Pronouns too long. Max characters is 20.")
		return
	}

	if a.CheckSmallString(user.ProfilePictureURL) {
		a.Error(w, r, "Profile Picture URL too long. Max characters is 123.")
		return
	}

	ok, err = a.ValidateFormFields(user.Username, user.Password, user.Email, false)
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

	if a.CheckPubkeyString(pubkey) {
		a.Error(w, r, "Public Key is too long. Max characters is 60.")
		return
	}

	if banned, word := a.CheckBanned(pubkey); banned {
		a.Error(w, r, "Public key contains a banned word: "+word)
		return
	}

	if err := a.UpdateUserPubkey(uid, pubkey); err != nil {
		a.Error(w, r, "Failed to update account: "+err.Error())
		return
	}

	http.Redirect(w, r, "/account?tab=verified", http.StatusSeeOther)
}

func (a *App) AccountChangePassPOST(w http.ResponseWriter, r *http.Request) {
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

	u, err := a.GetUser(uid)
	if err != nil {
		a.Error(w, r, "Failed to get user data: "+err.Error())
		return
	}

	r.ParseForm()

	oldPass := strings.TrimSpace(r.FormValue("oldPass"))
	newPass := strings.TrimSpace(r.FormValue("newPass"))

	if oldPass == "" {
		a.Error(w, r, "Missing old password!")
		return
	}

	if newPass == "" {
		a.Error(w, r, "Missing new password!")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(oldPass)); err != nil {
		a.Error(w, r, "Old password is incorrect!")
		return
	}

	hash, err := a.HashPassword(newPass)
	if err != nil {
		a.Error(w, r, "Failed to hash password: "+err.Error())
		return
	}

	if err := a.UpdateUserPass(uid, hash); err != nil {
		a.Error(w, r, "Failed to change your password: "+err.Error())
		return
	}

	http.Redirect(w, r, "/account?tab=security", http.StatusSeeOther)
}

func (a *App) DeleteAccountGET(w http.ResponseWriter, r *http.Request) {
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
		Page:    Page{IsLoggedIn: true, CSRFToken: a.GetCSRFTokenFromRequest(r)},
		User:    *user,
		Servers: servers,
	}

	a.render(w, r, http.StatusOK, "delete-account.html", data)
}

func (a *App) DeleteAccountPOST(w http.ResponseWriter, r *http.Request) {
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

	var servers []Server
	servers, err = a.GetServersByUser(uid)
	for _, s := range servers {
		if err := a.DeleteServer(s.ID, uid); err != nil {
			a.Log.Error("failed to delete server: ", err.Error())
			return
		}
	}

	var tkns []Token
	tkns, err = a.GetTokensFromUser(uid)
	for _, t := range tkns {
		if err := a.ApiDeleteToken(t.ID); err != nil {
			a.Log.Error("failed to delete token: ", err.Error())
			return
		}
	}

	err = a.DeleteUser(uid)
	if err != nil {
		a.Error(w, r, "Failed to delete your account! "+err.Error())
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
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
		CSRFToken:  a.GetCSRFTokenFromRequest(r),
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

	if a.CheckMediumString(name) {
		a.Error(w, r, "Server name too long. Max characters is 52.")
		return
	}

	ip := strings.TrimSpace(r.FormValue("ip"))
	if ip == "" {
		a.Error(w, r, "Server IP is required")
		return
	}

	if a.CheckMediumString(ip) {
		a.Error(w, r, "Server IP too long. Max characters is 52.")
		return
	}

	if banned, word := a.CheckBanned(name); banned {
		a.Error(w, r, "Server name contains a banned word: "+word)
		return
	}

	description := strings.TrimSpace(r.FormValue("description"))
	if a.CheckSmallString(description) {
		a.Error(w, r, "Description is too long. Max characters is 123.")
		return
	}

	if banned, word := a.CheckBanned(description); banned {
		a.Error(w, r, "Server description contains a banned word: "+word)
		return
	}

	if banned, word := a.CheckBanned(ip); banned {
		a.Error(w, r, "Server IP contains a banned word: "+word)
		return
	}

	s := Server{
		Name:        name,
		Description: description,
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

	ownerUser, err := a.GetUser(server.OwnerID)
	if err != nil {
		a.Error(w, r, "Failed to load owner user profile"+err.Error())
		return
	}

	// Show error if server owner account was suspended
	suspended, err := a.IsUserSuspended(server.OwnerID)
	if err != nil {
		a.Error(w, r, "Failed to check account status: "+err.Error())
		return
	}
	if suspended {
		a.Error(w, r, "The owner of the server you are trying to view was suspended by an administrator.")
		return
	}

	data := struct {
		Page
		Server
		*User        // user who OWNS the server
		IsOwner      bool
		ServerStatus bool
	}{
		Page:         Page{IsLoggedIn: a.HasSessionToken(r)},
		Server:       *server,
		User:         ownerUser,
		IsOwner:      isOwner,
		ServerStatus: a.IsServerOnline(server.ID),
	}

	a.render(w, r, http.StatusOK, "server_info.html", data)
}

func (a *App) UserInfo(w http.ResponseWriter, r *http.Request) {
	// user trying to view

	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		a.Return404(w, r)
		return
	}

	user, err := a.GetUser(id)
	if err != nil {
		a.Error(w, r, "Failed to load owner user profile"+err.Error())
		return
	}

	// Block suspended accounts from showing up.
	suspended, err := a.IsUserSuspended(id)
	if err != nil {
		a.Error(w, r, "Failed to check account status: "+err.Error())
		return
	}
	if suspended {
		a.Error(w, r, "The account profile you are trying to view was supended by an administrator.")
		return
	}

	servers, err := a.GetServersByUser(id)
	if err != nil {
		a.Error(w, r, "Failed to load users servers: "+err.Error())
		return
	}

	data := struct {
		Page
		*User   // user on the page
		Servers []Server
	}{
		Page:    Page{IsLoggedIn: a.HasSessionToken(r)},
		User:    user,
		Servers: servers,
	}

	a.render(w, r, http.StatusOK, "user_info.html", data)
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
		Page:   Page{IsLoggedIn: true, CSRFToken: a.GetCSRFTokenFromRequest(r)},
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
	if a.CheckMediumString(name) {
		a.Error(w, r, "Server Name is too long. Max characters is 52.")
		return
	}

	ip := strings.TrimSpace(r.FormValue("ip"))
	if ip == "" {
		a.Error(w, r, "Server IP is required")
		return
	}
	if a.CheckMediumString(ip) {
		a.Error(w, r, "Server IP is too long. Max characters is 52.")
		return
	}

	if banned, word := a.CheckBanned(name); banned {
		a.Error(w, r, "Server name contains a banned word: "+word)
		return
	}

	fields := []struct {
		label string
		value string
	}{
		{"Name", name},
		{"Description", strings.TrimSpace(r.FormValue("description"))},
		{"XML Feed Link", strings.TrimSpace(r.FormValue("xml_feed_link"))},
		{"Gamemodes", strings.TrimSpace(r.FormValue("gamemodes"))},
		{"Version", strings.TrimSpace(r.FormValue("version"))},
		{"Languages", strings.TrimSpace(r.FormValue("languages"))},
		{"Website URL", strings.TrimSpace(r.FormValue("website_url"))},
		{"Chat URL", strings.TrimSpace(r.FormValue("chat_url"))},
		{"Icon URL", strings.TrimSpace(r.FormValue("icon_url"))},
	}

	for _, f := range fields {
		if banned, word := a.CheckBanned(f.value); banned {
			a.Error(w, r, f.label+" contains a banned word: "+word)
			return
		}

		if f.label == "Description" {
			if a.CheckSmallString(f.value) {
				a.Error(w, r, "Description is too long. Max characters is 123.")
				return
			}
			continue // skip the 52-char check
		}

		if a.CheckMediumString(f.value) {
			a.Error(w, r, fmt.Sprintf("%s is too long. Max characters is 52.", f.label))
			return
		}
	}


	s := Server{
		ID:           id,
		Name:         name,
		Description:  strings.TrimSpace(r.FormValue("description")),
		XMLFeedLink:  strings.TrimSpace(r.FormValue("xml_feed_link")),
		IP:           ip,
		Gamemodes:    strings.TrimSpace(r.FormValue("gamemodes")),
		Version:      strings.TrimSpace(r.FormValue("version")),
		Languages:    strings.TrimSpace(r.FormValue("languages")),
		RequiresMods: r.FormValue("requires_mods") == "on",
		WebsiteURL:   strings.TrimSpace(r.FormValue("website_url")),
		ChatURL:      strings.TrimSpace(r.FormValue("chat_url")),
		IconURL:      strings.TrimSpace(r.FormValue("icon_url")),
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

func (a *App) CreateTokenUI(w http.ResponseWriter, r *http.Request) {
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

	servers, err := a.GetServersByUser(uid)
	if err != nil {
		a.Error(w, r, "Failed to load servers: "+err.Error())
		return
	}

	data := struct {
		Page
		Servers []Server
	}{
		Page:    Page{IsLoggedIn: a.HasSessionToken(r), CSRFToken: a.GetCSRFTokenFromRequest(r)},
		Servers: servers,
	}

	a.render(w, r, http.StatusOK, "create-token-ui.html", data)
}

func (a *App) ApiCreateTokenHandler(w http.ResponseWriter, r *http.Request) {
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

	serverID, err := strconv.ParseUint(r.FormValue("server"), 10, 64)
	if err != nil {
		a.Error(w, r, "Invalid server ID: "+err.Error())
		return
	}

	typ := r.FormValue("typ")
	name := strings.TrimSpace(r.FormValue("name"))

	if a.CheckMediumString(name) {
		a.Error(w, r, "Token name is too long. Max characters is 52.")
		return
	}

	if a.CheckMediumString(typ) {
		a.Error(w, r, "Nice try :)")
		return
	}

	if banned, word := a.CheckBanned(name); banned {
		a.Error(w, r, "Token name contains a banned word: "+word)
		return
	}

	tokenHash, err := generateToken(15)
	if err != nil {
		a.Error(w, r, "Failed to generate token hash: "+err.Error())
		return
	}

	switch typ {
	case "spark":
		// valid type, proceed
	case "votifier":
		a.Error(w, r, "Votifier tokens are yet not supported.")
		return
	default:
		a.Error(w, r, "Invalid token type.")
		return
	}

	err = a.ApiCreateToken(serverID, uid, typ, name, tokenHash)
	if err != nil {
		a.Error(w, r, "Failed to create token: "+err.Error())
		return
	}

	data := struct {
		Page
		TokenHash string
	}{
		Page:      Page{IsLoggedIn: a.HasSessionToken(r)},
		TokenHash: tokenHash,
	}

	a.render(w, r, http.StatusOK, "your-api-token.html", data)
}

func (a *App) ApiDeleteTokenHandler(w http.ResponseWriter, r *http.Request) {
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

	tokenID, err := strconv.ParseUint(r.FormValue("tokenID"), 10, 64)
	if err != nil {
		a.Error(w, r, "Invalid token ID: "+err.Error())
		return
	}

	owned, err := a.IsTokenOwnedByUser(tokenID, uid)
	if err != nil {
		a.Error(w, r, "Failed to verify token ownership: "+err.Error())
		return
	}
	if !owned {
		a.Error(w, r, "Token not found or not owned by you.")
		return
	}

	err = a.ApiDeleteToken(tokenID)
	if err != nil {
		a.Error(w, r, "Failed to delete token: "+err.Error())
		return
	}

	http.Redirect(w, r, "/account?tab=api", http.StatusSeeOther)
}

func (a *App) ServerListGET(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	f := ServerFilter{
		Search:       q.Get("search"),
		Version:      q.Get("version"),
		Gamemodes:    q["gamemode"],
		Languages:    q["language"],
		RequiresMods: q.Get("requires_mods") == "on",
		Status:       q.Get("status"),
		Sort:         q.Get("sort"),
	}

	if mp, err := strconv.ParseUint(q.Get("min_players"), 10, 64); err == nil {
		f.MinPlayers = mp
	}
	if mp, err := strconv.ParseUint(q.Get("max_players"), 10, 64); err == nil {
		f.MaxPlayers = mp
	}

	servers, err := a.ListServers(f)
	if err != nil {
		a.Error(w, r, "Failed to load servers: "+err.Error())
		return
	}

	// --- Pagination (max ServersPerPage shown at a time) ---
	page := 1
	if p, err := strconv.Atoi(q.Get("page")); err == nil && p > 0 {
		page = p
	}

	total := len(servers)
	totalPages := (total + ServersPerPage - 1) / ServersPerPage
	if totalPages < 1 {
		totalPages = 1
	}
	if page > totalPages {
		page = totalPages
	}

	start := (page - 1) * ServersPerPage
	if start > total {
		start = total
	}
	end := start + ServersPerPage
	if end > total {
		end = total
	}
	servers = servers[start:end]

	// Build page URLs that preserve the current filters.
	pq := r.URL.Query()
	pq.Del("page")
	base := pq.Encode()
	pageURL := func(p int) string {
		if base == "" {
			return fmt.Sprintf("/list?page=%d", p)
		}
		return fmt.Sprintf("/list?%s&page=%d", base, p)
	}
	// -------------------------------------------------------

	for i := range servers {
		if a.IsServerOnline(servers[i].ID) {
			servers[i].Status = "Online"
		} else {
			servers[i].Status = "Offline"
		}
	}

	gmChecks := map[string]bool{}
	for _, g := range f.Gamemodes {
		gmChecks[g] = true
	}
	langChecks := map[string]bool{}
	for _, l := range f.Languages {
		langChecks[l] = true
	}

	data := struct {
		Page
		Servers        []Server
		Filter         ServerFilter
		GamemodeChecks map[string]bool
		LanguageChecks map[string]bool
		CurrentPage    int
		TotalPages     int
		HasPrev        bool
		HasNext        bool
		PrevURL        string
		NextURL        string
	}{
		Page:           Page{IsLoggedIn: a.HasSessionToken(r)},
		Servers:        servers,
		Filter:         f,
		GamemodeChecks: gmChecks,
		LanguageChecks: langChecks,
		CurrentPage:    page,
		TotalPages:     totalPages,
		HasPrev:        page > 1,
		HasNext:        page < totalPages,
		PrevURL:        pageURL(page - 1),
		NextURL:        pageURL(page + 1),
	}

	a.render(w, r, http.StatusOK, "list.html", data)
}

// ToggleTheme toggles the "light-mode" cookie on every request.
// - empty/unset -> "true"
// - "true"      -> "false"
// - "false"     -> "true"
func (a *App) ToggleTheme(w http.ResponseWriter, r *http.Request) {
	// Default value when the cookie is missing or empty.
	next := "true"

	if cookie, err := r.Cookie("light-mode"); err == nil {
		switch cookie.Value {
		case "true":
			next = "false"
		case "false":
			next = "true"
		default:
			// Any other/empty value falls back to "true".
			next = "true"
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "light-mode",
		Value:    next,
		Path:     "/",
		HttpOnly: false, // must be readable by client-side JS (getCookie)
		SameSite: http.SameSiteLaxMode,
	})

	// Send the user back where they came from, or to home.
	referer := r.Header.Get("Referer")
	if referer == "" {
		referer = "/"
	}
	http.Redirect(w, r, referer, http.StatusSeeOther)
}