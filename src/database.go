package main

import (
	"crypto/rand"
	"encoding/hex"
	"golang.org/x/crypto/bcrypt"
	"time"
	"strconv"
	"errors"
	"database/sql"
)

/** This file contains:
--- Users ---
- create user (u User) (id int64, error error)
- get user (uid uint64) returns a *User, error
- update user (u User) error
- delete user (uid uint64) error
- get servers by user (uid uint64) returns []Server, error
- check if user exists (id uint64) returns bool, error
- get user id by username (username string) returns uint64, error

--- Passwords ---
- hash password (password string) returns string, error
- check password (hash, password string) bool
- check password with db (uint64, password) returns bool, error

--- Session Tokens ---
- generate session token () returns string, error
- set session token (uid uint64, token string, expires time.Time) error
- check session token (uid uint64, token string) returns bool, error
- get username from session token (token string) returns string, error
**/

/** USERS **/

func (a *App) CreateUser(u User) (int64, error) {
	hash, err := a.HashPassword(u.Password)
	if err != nil {
		return 0, err
	}

	res, err := a.DB.Exec(
		"INSERT INTO users(username,email,password,date_created) VALUES (?,?,?,?)",
		u.Username, u.Email, hash, time.Now(),
	)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (a *App) GetUser(id uint64) (*User, error) {
	var u User
	var profilePictureURL, description, pronouns sql.NullString

	err := a.DB.QueryRow(
		`SELECT id,username,email,password,date_created,session_token,session_token_expires,
		        profile_picture_url, description, pronouns
		 FROM users WHERE id=?`,
		id,
	).Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.Password,
		&u.DateCreated,
		&u.SessionToken,
		&u.SessionTokenExpires,
		&profilePictureURL,
		&description,
		&pronouns,
	)

	u.ProfilePictureURL = profilePictureURL.String
	u.Description = description.String
	u.Pronouns = pronouns.String

	return &u, err
}


func (a *App) UpdateUserProfile(u User) error {
	_, err := a.DB.Exec(
		`UPDATE users SET username=?, email=?, profile_picture_url=?, description=?, pronouns=? WHERE id=?`,
		u.Username, u.Email, u.ProfilePictureURL, u.Description, u.Pronouns, u.ID,
	)
	return err
}



func (a *App) UpdateUser(u User) error {
	_, err := a.DB.Exec(
		`UPDATE users SET username=?, email=? WHERE id=?`,
		u.Username, u.Email, u.ID,
	)
	return err
}

func (a *App) DeleteUser(id uint64) error {
	_, err := a.DB.Exec("DELETE FROM users WHERE id=?", id)
	return err
}

func (a *App) GetServersByUser(userID uint64) ([]Server, error) {
	rows, err := a.DB.Query(
		"SELECT id,name,description,xml_feed_link,playercount,owner_id FROM servers WHERE owner_id=?",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []Server
	for rows.Next() {
		var s Server
		if err := rows.Scan(
			&s.ID,
			&s.Name,
			&s.Description,
			&s.XMLFeedLink,
			&s.PlayerCount,
			&s.OwnerID,
		); err != nil {
			return nil, err
		}
		servers = append(servers, s)
	}

	return servers, nil
}

func (a *App) UserExists(id uint64) (bool, error) {
	var exists bool

	err := a.DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM users WHERE id=?)",
		id,
	).Scan(&exists)

	if err != nil {
		return false, err
	}

	return exists, nil
}

func (a *App) GetUserIDByUsername(username string) (uint64, error) {
	var id uint64

	err := a.DB.QueryRow(
		"SELECT id FROM users WHERE username=?",
		username,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

/** PASSWORDS **/

func (a *App) HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), err
}

func (a *App) CheckPassword(hash, password string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return false, err
	}

	return true, nil
}

func (a *App) CheckPasswordDB(userID uint64, password string) (bool, error) {
	var hash string

	err := a.DB.QueryRow(
		"SELECT password FROM users WHERE id=?",
		userID,
	).Scan(&hash)

	if err != nil {
		return false, err
	}

	ok, e := a.CheckPassword(hash, password)
	if err != nil {
		return false, e
	}

	if !ok {
		return false, nil
	}

	return true, nil
}

/** SESSION TOKENS **/

func (a *App) GenerateSessionToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (a *App) SetSessionToken(userID uint64, token string, expires time.Time) error {
	_, err := a.DB.Exec(
		"UPDATE users SET session_token=?, session_token_expires=? WHERE id=?",
		token, expires, userID,
	)
	return err
}

func (a *App) CheckSessionToken(userID uint64, token string) (bool, error) {
	var stored string
	var expires time.Time

	err := a.DB.QueryRow(
		"SELECT session_token, session_token_expires FROM users WHERE id=?",
		userID,
	).Scan(&stored, &expires)

	if err != nil {
		return false, err
	}

	if stored != token {
		return false, nil
	}

	return time.Now().Before(expires), nil
}

func (a *App) GetUIDFromToken(sessionToken string) (uint64, error) {
	var uid string // converted to uint64 on at the bottom

	err := a.DB.QueryRow(
		`select id from users where session_token=?;`,
		sessionToken,
	).Scan(&uid)

	if err != nil {
		return 0, err
	}

	// conversions
	var converted int
	converted, err = strconv.Atoi(uid)
	if err != nil {
		return 0, errors.New("failed to convert uid to an int64: "+err.Error())
	}

	// conversions #2
	return uint64(converted), nil
}