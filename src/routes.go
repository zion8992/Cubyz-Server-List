package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"
	"database/sql"
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
		a.Return404(w, r,)
		return
	}

	tmpl, err := template.ParseFiles(
		"./templates/base.html",
		"./templates/"+page+".html",
	)
	if err != nil {
		a.Error(w, r, fmt.Sprintf("failed to load template: %v (page=%s)", err, page))
	}

	data := Page{
		IsLoggedIn: a.HasSessionToken(r),
	}

	if err := tmpl.Execute(w, data); err != nil {
		a.Error(w, r, "template execution failed: ", err.Error())
	}
}

func (a *App) RegisterGET(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(
		"./templates/base.html",
		"./templates/auth/register.html",
	)
	if err != nil {
		a.Error(w, r, fmt.Sprintf("failed to load template: %v (templates/auth/login.html)", err))
	}

	data := Page{
		IsLoggedIn: a.HasSessionToken(r),
	}

	if err := tmpl.Execute(w, data); err != nil {
		a.Error(w, r, "template execution failed: ", err.Error())
	}
}

func (a *App) RegisterPOST(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	ok, err := a.ValidateFormFields(r.FormValue("username"), r.FormValue("password"), r.FormValue("email"))

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

	u := User {
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
		Email: r.FormValue("email"),
	}

	_, err = a.CreateUser(u)
	if err != nil {
		a.Error(w, r, "Failed to create user! ", err.Error())
		return
	}

	tmpl, err := template.ParseFiles(
		"./templates/base.html",
		"./templates/auth/register_ok.html",
	)
	if err != nil {
		a.Error(w, r, fmt.Sprintf("failed to load template: %v (templates/auth/login.html)", err))
		return
	}

	data := Page{
		IsLoggedIn: a.HasSessionToken(r),
	}

	if err := tmpl.Execute(w, data); err != nil {
		a.Error(w, r, "template execution failed: ", err.Error())
	}
}

func (a *App) LoginGET(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(
		"./templates/base.html",
		"./templates/auth/login.html",
	)
	if err != nil {
		a.Error(w, r, fmt.Sprintf("failed to load template: %v (templates/auth/login.html)", err))
	}

	data := Page{
		IsLoggedIn: a.HasSessionToken(r),
	}

	if err := tmpl.Execute(w, data); err != nil {
		a.Error(w, r, "template execution failed: ", err.Error())
	}
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
		a.Error(w, r, "Failed to check if your password is correct: " + err.Error())
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

	c := http.Cookie {
		Name: "session_token",
		Value: token,
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

	tmpl, err := template.ParseFiles(
		"./templates/base.html",
		"./templates/account.html",
	)
	if err != nil {
		a.Error(w, r, fmt.Sprintf("failed to load template: %v (templates/account.html)", err))
		return
	}

	data := struct {
		Page
		User
	}{
		Page: Page{IsLoggedIn: true},
		User: *user,
	}

	if err := tmpl.Execute(w, data); err != nil {
		a.Error(w, r, "template execution failed: ", err.Error())
	}
}

func (a *App) AccountPOST(w http.ResponseWriter, r *http.Request) {
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

	ok, err = a.ValidateFormFields(user.Username, user.Password, user.Email)
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

func getSessionCookie(r *http.Request) string {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return ""
	}
	return cookie.Value
}
