package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"strconv"
	"time"
)

/**
This file contains:

--- Users ---
- CreateUser(u User) (int64, error)
- GetUser(id uint64) (*User, error)
- UpdateUserProfile(u User) error
- UpdateUser(u User) error
- DeleteUser(id uint64) error
- GetServersByUser(userID uint64) ([]Server, error)
- UserExists(id uint64) (bool, error)
- GetUserIDByUsername(username string) (uint64, error)

--- Passwords ---
- HashPassword(password string) (string, error)
- CheckPassword(hash, password string) (bool, error)
- CheckPasswordDB(userID uint64, password string) (bool, error)

--- Session Tokens ---
- GenerateSessionToken() (string, error)
- SetSessionToken(userID uint64, token string, expires time.Time) error
- CheckSessionToken(userID uint64, token string) (bool, error)
- GetUIDFromToken(sessionToken string) (uint64, error)
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
	var profilePictureURL, description, pronouns, pubkey sql.NullString

	err := a.DB.QueryRow(
		`SELECT id,username,email,password,date_created,session_token,session_token_expires,
		        profile_picture_url, description, pronouns, pubkey
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
		&pubkey,
	)

	u.ProfilePictureURL = profilePictureURL.String
	u.Description = description.String
	u.Pronouns = pronouns.String
	u.Pubkey = pubkey.String

	return &u, err
}

func (a *App) UpdateUserProfile(u User) error {
	_, err := a.DB.Exec(
		`UPDATE users SET username=?, email=?, profile_picture_url=?, description=?, pronouns=? WHERE id=?`,
		u.Username, u.Email, u.ProfilePictureURL, u.Description, u.Pronouns, u.ID,
	)
	return err
}

func (a *App) UpdateUserPubkey(id uint64, pubkey string) error {
	_, err := a.DB.Exec(
		`UPDATE users SET pubkey=? WHERE id=?`,
		pubkey, id,
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

func (a *App) LogoutUser(u User) error {
	_, err := a.DB.Exec(
		`UPDATE users SET session_token = NULL WHERE id = ?`,
		u.ID,
	)
	return err
}

func (a *App) DeleteUser(id uint64) error {
	_, err := a.DB.Exec("DELETE FROM users WHERE id=?", id)
	return err
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

	ok, err := a.CheckPassword(hash, password)
	if err != nil {
		return false, err
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
	var converted uint64
	converted, err = strconv.ParseUint(uid, 10, 64)
	if err != nil {
		return 0, errors.New("failed to convert uid to an uint64: " + err.Error())
	}

	return converted, nil
}

/** SERVERS **/

func (a *App) CreateServer(s Server) (int64, error) {
	res, err := a.DB.Exec(
		`INSERT INTO servers (
			name, description, xml_feed_link, ip, playercount, owner_id,
			gamemodes, version, languages, requires_mods, website_url, chat_url
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.Name,
		s.Description,
		s.XMLFeedLink,
		s.IP,
		s.PlayerCount,
		s.OwnerID,
		s.Gamemodes,
		s.Version,
		s.Languages,
		s.RequiresMods,
		s.WebsiteURL,
		s.ChatURL,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (a *App) GetServer(id uint64) (*Server, error) {
	var s Server
	var description, xmlFeedLink, ip, gamemodes, version, languages, websiteURL, chatURL sql.NullString
	var lastSpark sql.NullTime

	err := a.DB.QueryRow(
		`SELECT id, name, description, xml_feed_link, ip, playercount, owner_id,
			gamemodes, version, languages, requires_mods, website_url, chat_url, last_spark, status
		 FROM servers WHERE id=?`,
		id,
	).Scan(
		&s.ID,
		&s.Name,
		&description,
		&xmlFeedLink,
		&ip,
		&s.PlayerCount,
		&s.OwnerID,
		&gamemodes,
		&version,
		&languages,
		&s.RequiresMods,
		&websiteURL,
		&chatURL,
		&lastSpark,
		&s.Status,
	)
	if err != nil {
		return nil, err
	}

	s.Description = description.String
	s.XMLFeedLink = xmlFeedLink.String
	s.IP = ip.String
	s.Gamemodes = gamemodes.String
	s.Version = version.String
	s.Languages = languages.String
	s.WebsiteURL = websiteURL.String
	s.ChatURL = chatURL.String
	s.LastSpark = lastSpark.Time

	return &s, nil
}

func (a *App) GetServersByUser(userID uint64) ([]Server, error) {
	rows, err := a.DB.Query(
		"SELECT id FROM servers WHERE owner_id=?",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []Server
	for rows.Next() {
		var id uint64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		s, err := a.GetServer(id)
		if err != nil {
			return nil, err
		}

		servers = append(servers, *s)
	}

	return servers, nil
}

func (a *App) UpdateServer(s Server) error {
	_, err := a.DB.Exec(
		`UPDATE servers SET
			name = ?,
			description = ?,
			xml_feed_link = ?,
			ip = ?,
			gamemodes = ?,
			version = ?,
			languages = ?,
			requires_mods = ?,
			website_url = ?,
			chat_url = ?
		 WHERE id = ? AND owner_id = ?`,
		s.Name,
		s.Description,
		s.XMLFeedLink,
		s.IP,
		s.Gamemodes,
		s.Version,
		s.Languages,
		s.RequiresMods,
		s.WebsiteURL,
		s.ChatURL,
		s.ID,
		s.OwnerID,
	)
	return err
}

func (a *App) DeleteServer(id, ownerID uint64) error {
	_, err := a.DB.Exec(`DELETE FROM servers WHERE id = ? AND owner_id = ?`, id, ownerID)
	return err
}
