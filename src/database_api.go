package main

import (
	"database/sql"
	"errors"
	"time"
)

var (
	ErrTokenNotFound = errors.New("token not found")
)

func (a *App) IsValidApiToken(hash string) (bool, error) {
	var exists bool
	err := a.DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM api_tokens WHERE token_hash = ? AND expiry > CURRENT_TIMESTAMP)",
		hash,
	).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (a *App) GetApiTokenType(hash string) (string, error) {
	var tokenType string
	err := a.DB.QueryRow(
		"SELECT type FROM api_tokens WHERE token_hash = ? AND expiry > CURRENT_TIMESTAMP",
		hash,
	).Scan(&tokenType)
	if err == sql.ErrNoRows {
		return "", ErrTokenNotFound
	}
	if err != nil {
		return "", err
	}
	return tokenType, nil
}

func (a *App) CheckApiTokenType(hash string, tokenType string) (bool, error) {
	var exists bool
	err := a.DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM api_tokens WHERE token_hash = ? AND type = ? AND expiry > CURRENT_TIMESTAMP)",
		hash, tokenType,
	).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (a *App) GetOwnerFromToken(hash string) (uint64, error) {
	var ownerID uint64
	err := a.DB.QueryRow(
		"SELECT owner_id FROM api_tokens WHERE token_hash = ? AND expiry > CURRENT_TIMESTAMP",
		hash,
	).Scan(&ownerID)
	if err == sql.ErrNoRows {
		return 0, ErrTokenNotFound
	}
	if err != nil {
		return 0, err
	}
	return ownerID, nil
}

func (a *App) GetServerFromToken(hash string) (uint64, error) {
	var serverID uint64
	err := a.DB.QueryRow(
		"SELECT server_id FROM api_tokens WHERE token_hash = ? AND expiry > CURRENT_TIMESTAMP",
		hash,
	).Scan(&serverID)
	if err == sql.ErrNoRows {
		return 0, ErrTokenNotFound
	}
	if err != nil {
		return 0, err
	}
	return serverID, nil
}

func (a *App) ApiIncServerPlayers(serverID uint64) error {
	_, err := a.DB.Exec(
		"UPDATE servers SET playercount = playercount + 1 WHERE id = ?",
		serverID,
	)
	return err
}

func (a *App) ApiDecServerPlayers(serverID uint64) error {
	_, err := a.DB.Exec(
		"UPDATE servers SET playercount = playercount - 1 WHERE id = ?",
		serverID,
	)
	return err
}

func (a *App) ApiServerOn(serverID uint64) error {
	_, err := a.DB.Exec(
		"UPDATE servers SET status = true, last_spark = CURRENT_TIMESTAMP WHERE id = ?",
		serverID,
	)
	return err
}

func (a *App) ApiServerOff(serverID uint64) error {
	_, err := a.DB.Exec(
		"UPDATE servers SET status = false, last_spark = CURRENT_TIMESTAMP WHERE id = ?",
		serverID,
	)
	return err
}

func (a *App) IsServerOnline(serverID uint64, lastSpark time.Time) bool {
	if time.Since(lastSpark) < time.Hour {
		return true
	}

	var status bool
	err := a.DB.QueryRow("SELECT status FROM servers WHERE id = ?", serverID).Scan(&status)
	if err != nil {
		return false
	}

	return status
}
