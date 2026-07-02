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
	Status       bool
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
	Name string
	OwnerID     uint64
	ServerID    uint64
	DateCreated time.Time
	Expiry      time.Time
	Type        string
	TokenHash   string
}
