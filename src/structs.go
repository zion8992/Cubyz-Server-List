package main

import (
	"database/sql"
	"html/template"
	"log/slog"
	"time"
)

type App struct {
	Log           *slog.Logger
	DB            *sql.DB
	BlockedWords  string
	DefaultExpiry time.Duration
	TemplateCache map[string]*template.Template
}

type User struct {
	ID                  uint64
	Username            string
	Email               string
	Password            string
	DateCreated         time.Time
	ServersOwned        []Server
	SessionToken        string
	SessionTokenExpires time.Time
	ProfilePictureURL   string
	Description         string
	Pronouns            string
	Pubkey              string
}

type Server struct {
	ID           uint64
	Name         string
	Description  string
	XMLFeedLink  string
	PlayerCount  uint64
	OwnerID      uint64
	Gamemodes    string
	Version      string
	Languages    string
	RequiresMods bool
	WebsiteURL   string
	ChatURL      string
	IP           string
	LastSpark    time.Time
	Status   bool
}

type APIToken struct {
	ID          uint64    `json:"id" db:"id"`
	OwnerID     uint64    `json:"ownerID" db:"owner_id"`
	ServerID    uint64    `json:"serverID" db:"server_id"`
	DateCreated time.Time `json:"dateCreated" db:"date_created"`
	Expiry      time.Time `json:"expiry" db:"expiry"`
	Type        string    `json:"type" db:"type"`
	TokenHash   string    `json:"-" db:"token_hash"`
}
