package main

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"math/big"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyz0123456789" // charset for generating completely randomized tokens

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
		"UPDATE servers SET last_spark = CURRENT_TIMESTAMP WHERE id = ?",
		serverID,
	)
	return err
}

func (a *App) ApiServerOff(serverID uint64) error {
	_, err := a.DB.Exec(
		"UPDATE servers SET last_spark = NULL, playercount = 0 WHERE id = ?",
		serverID,
	)
	return err
}

func (a *App) IsServerOnline(serverID uint64) bool {
	var lastSpark sql.NullTime

	err := a.DB.QueryRow(
		`select last_spark from servers where id=?;`,
		serverID,
	).Scan(&lastSpark)
	if err != nil {
		a.Log.Error("Failed to get a server status: " + err.Error())
		return false
	}

	if !lastSpark.Valid {
		return false
	}

	return time.Since(lastSpark.Time) < time.Hour
}

func (a *App) ApiCreateToken(serverID uint64, ownerID uint64, typ string, name string, tokenHash string) error {
	dateCreated := time.Now()
	expiry := dateCreated.Add(2160 * time.Hour) // expires in 90 days

	_, err := a.DB.Exec(
		`INSERT INTO api_tokens (
			owner_id, server_id, type, token_hash, expiry, date_created, name
		) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		ownerID,
		serverID,
		typ,
		tokenHash,
		expiry,
		dateCreated,
		name,
	)
	if err != nil {
		return err
	}
	return nil
}

func generateToken(length int) (string, error) {
	token := make([]byte, length)
	for i := range token {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		token[i] = charset[n.Int64()]
	}
	return string(token), nil
}

/** API AND TOKENS **/

func (a *App) GetTokensFromUser(uid uint64) ([]Token, error) {
	rows, err := a.DB.Query(
		`SELECT id, owner_id, server_id, date_created, expiry, type, token_hash, name
		FROM api_tokens
		WHERE owner_id = ?`,
		uid,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []Token
	for rows.Next() {
		var t Token
		if err := rows.Scan(
			&t.ID,
			&t.OwnerID,
			&t.ServerID,
			&t.DateCreated,
			&t.Expiry,
			&t.Type,
			&t.TokenHash,
			&t.Name,
		); err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tokens, nil
}

func (a *App) ApiDeleteToken(tokenId uint64) error {
	_, err := a.DB.Exec(
		`DELETE FROM api_tokens WHERE id = ?`,
		tokenId,
	)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) IsTokenOwnedByUser(tokenID uint64, uid uint64) (bool, error) {
	var ownerID uint64
	err := a.DB.QueryRow(
		`SELECT owner_id FROM api_tokens WHERE id = ?`,
		tokenID,
	).Scan(&ownerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return ownerID == uid, nil
}
