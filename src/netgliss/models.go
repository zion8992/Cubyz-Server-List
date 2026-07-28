package netgliss

import(
	"time"
)

type Server struct {
	// === Core Fields ===
	Name        string    `json:"name"`
	IP          string    `json:"ip"`
	DateCreated time.Time `json:"date_created"`

	// === Server-Provided Fields ===
	Description string `json:"description"`
	Version string `json:"version"`
	Gamemodes string `json:"gamemodes"`
	Languages string `json:"languages"`
	RequiresMods bool `json:"requires_mods"`
}