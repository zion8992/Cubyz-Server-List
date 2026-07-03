package main

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"strings"
)

func (a *App) Error(w http.ResponseWriter, r *http.Request, errs ...string) {
	w.WriteHeader(http.StatusInternalServerError)
	msg := strings.Join(errs, "")

	a.Log.Error("request error",
		"path", r.URL.Path,
		"errors", msg,
	)

	w.Header().Set("Content-Type", "text/html")

	tmpl, err := template.ParseFiles(
		"./templates/base.html",
		"./templates/error.html",
	)
	if err != nil {
		fmt.Fprintf(w, "Failed to load ./templates/error.html! File not found")
	}

	data := Page{
		Err: msg,
	}

	if err := tmpl.Execute(w, data); err != nil {
		fmt.Fprintf(w, "Failed to render template: %s\n", err.Error())
	}
}

func (a *App) ApiError(w http.ResponseWriter, r *http.Request, errs ...string) {
	w.WriteHeader(http.StatusInternalServerError)
	msg := strings.Join(errs, "")

	a.Log.Error("request api error",
		"path", r.URL.Path,
		"errors", msg,
	)

	fmt.Fprint(w, msg)
}

func (a *App) ValidateFormFields(username, password, email string, validatePassword bool) (bool, error) {
	// Username:
	// - ASCII only
	// - No spaces, quotes, slashes/backslashes
	usernameRegex := regexp.MustCompile(`^[\x21\x23-\x26\x28-\x2E\x30-\x39\x3A-\x7E]+$`)
	// excludes: space (0x20), " (0x22), ' (0x27), / (0x2F), \ (0x5C)

	if !usernameRegex.MatchString(username) {
		return false, errors.New("invalid username format")
	}

	if validatePassword {
		// Password:
		// - min 11 chars
		// - at least 2 uppercase
		// - at least 4 digits
		// - at least 1 lowercase
		// - at least 1 symbol
		// - no / or \
		if len(password) < 11 {
			return false, errors.New("invalid password format")
		}

		var upper, lower, digit, symbol int

		for _, c := range password {
			switch {
			case c == '/' || c == '\\':
				return false, errors.New("invalid password format")
			case 'A' <= c && c <= 'Z':
				upper++
			case 'a' <= c && c <= 'z':
				lower++
			case '0' <= c && c <= '9':
				digit++
			default:
				symbol++
			}
		}

		if upper < 2 || lower < 1 || digit < 4 || symbol < 1 {
			return false, errors.New("invalid password format")
		}

		// Email:
		// - no spaces or backslashes before @
		// - simple <abc>@<whatever>.<domain>
		emailRegex := regexp.MustCompile(`^[^\\\s@]+@[^@\s]+\.[^@\s]+$`)

		if !emailRegex.MatchString(email) {
			return false, errors.New("invalid email format")
		}
	}

	return true, nil
}

func (a *App) Return404(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(
		"./templates/base.html",
		"./templates/404.html",
	)
	if err != nil {
		a.Error(w, r, fmt.Sprintf("failed to load template: %v (templates/404.html)", err))
	}

	data := Page{
		IsLoggedIn: a.HasSessionToken(r),
	}

	if err := tmpl.Execute(w, data); err != nil {
		a.Error(w, r, "template execution failed: ", err.Error())
		return
	}
}

func (a *App) HasSessionToken(r *http.Request) bool {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return false
	}

	if cookie.Value == "" {
		return false
	}

	_, err = a.CheckReqSessionTok(r)
	if err != nil {
		fmt.Printf("Failed to check users session token: %s\n", err.Error())
		return false
	}

	return true
}

func (a *App) CheckReqSessionTok(r *http.Request) (bool, error) {
	cookie, err := r.Cookie("session_token")

	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return false, nil
		}

		return false, fmt.Errorf(
			"failed to check the session token in your HTTP request: %w",
			err,
		)
	}

	if cookie.Value == "" {
		return false, errors.New("Failed to check the session token in your HTTP request: empty cookie value")
	}

	var uid uint64
	uid, err = a.GetUIDFromToken(cookie.Value)
	if err != nil {
		return false, errors.New("Failed to get the uid from your session token: " + err.Error())
	}

	if uid == 0 {
		return false, errors.New("Failed to get the username from your session token: uid is 0")
	}

	var ok bool
	ok, err = a.CheckSessionToken(uid, cookie.Value)
	if err != nil {
		return false, errors.New("Failed to check if the session token in your HTTP request is valid: " + err.Error())
	}

	if ok {
		return true, nil
	}

	return false, nil
}

func (a *App) NoEmbedrender(w http.ResponseWriter, r *http.Request, status int, page string, data any) {
	ts, ok := a.TemplateCache[page]
	if !ok {
		a.Error(w, r, fmt.Sprintf("the template %s does not exist", page))
		return
	}

	w.WriteHeader(status)

	if err := ts.Execute(w, data); err != nil {
		a.Error(w, r, "template execution failed: "+err.Error())
	}
}
