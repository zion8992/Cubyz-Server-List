package main

import(
	"time"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string // alphanumeric
	Password string // hash

	Created time.Time
}