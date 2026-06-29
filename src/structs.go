package main

import (
	"time"
)

type User struct {
	ID                  uint64
	Username            string
	Email               string
	Password            string
	DateCreated         time.Time
	ServersOwned        []Server
	SessionToken        string
	SessionTokenExpires time.Time
	ProfilePictureURL string
	Description       string
	Pronouns          string
}


type Server struct {
	ID          uint64
	Name        string
	Description string
	XMLFeedLink string
	PlayerCount uint64
	OwnerID     uint64
}
