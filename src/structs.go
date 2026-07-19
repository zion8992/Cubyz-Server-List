package main

import (
	"database/sql"
	"html/template"
	"log/slog"
	"time"
)

type App struct {
	Log                *slog.Logger
	DB                 *sql.DB
	BannedWords        map[string]struct{}
	DefaultTokenExpiry time.Duration
	TemplateCache      map[string]*template.Template
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
	PrivLevel           string
	AccountSuspended    bool
	TOTPSecret          string
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
	IconURL      string
	IP           string
	LastSpark    time.Time
	Status       string // Not in database, used only for list filtering
}

type APIToken struct {
	ID          uint64
	OwnerID     uint64
	ServerID    uint64
	DateCreated time.Time
	Expiry      time.Time
	Type        string
	TokenHash   string
}

type Token struct {
	ID          uint64
	Name        string
	OwnerID     uint64
	ServerID    uint64
	DateCreated time.Time
	Expiry      time.Time
	Type        string
	TokenHash   string
}

type ServerFilter struct {
	Search       string
	Version      string
	MinPlayers   uint64
	MaxPlayers   uint64
	Gamemodes    []string
	Languages    []string
	RequiresMods bool
	Status       string
	Sort         string
}

// UserFilter holds the query params for the admin users tab
type UserFilter struct {
	Search string // matches username
	Sort   string // players | servers | tokens | newest | oldest | username
	Page   int    // 1-based
}

// AdminUserRow is one row in the admin users table, with aggregated stats
type AdminUserRow struct {
	ID            uint64
	Username      string
	Email         string
	DateCreated   time.Time
	PrivLevel     string
	ServerCount   uint64
	TokenCount    uint64
	OnlinePlayers uint64
	Suspended     bool
}
