package main

import (
	"errors"
	"github.com/go-sql-driver/mysql"
	"net/http"
)

func (a *App) SparkUpdatePOST(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		a.ApiError(w, r, "[0] Failed to parse form: "+err.Error())
		return
	}

	token := r.FormValue("token")
	updateType := r.FormValue("update")

	ok, err := a.IsValidApiToken(token)
	if err != nil {
		a.ApiError(w, r, "[1] Failed to check if your token is valid: "+err.Error())
		return
	}

	if !ok {
		a.ApiError(w, r, "[2] Invalid token")
		return
	}

	ownID, err := a.GetOwnerFromToken(token)
	if err != nil {
		a.ApiError(w, r, "[1] Failed to get the owner of the token: "+err.Error())
		return
	}

	var isSuspended bool
	isSuspended, err = a.IsUserSuspended(ownID)
	if err != nil {
		a.ApiError(w, r, "[1] Failed to check your account status: "+err.Error())
		return
	}

	if isSuspended {
		a.ApiError(w, r, "Your account has been suspended by an administrator.")
		return
	}

	sid, err := a.GetServerFromToken(token)
	if err != nil {
		a.ApiError(w, r, "[3] Failed to get the server that corresponds to your token: "+err.Error())
		return
	}

	switch updateType {
	case "playerJoin":
		err := a.ApiIncServerPlayers(sid)
		if err != nil {
			a.ApiError(w, r, "[5] Failed to update your server! "+err.Error())
			return
		}

	case "playerLeave":
		err := a.ApiDecServerPlayers(sid)
		if err != nil {
			var mysqlErr *mysql.MySQLError
			if errors.As(err, &mysqlErr) && mysqlErr.Number == 1690 {
				a.ApiError(w, r, "[6] Nice try. You cannot set a negative playercount :)")
				return
			}
			a.ApiError(w, r, "[5] Failed to update your server! "+err.Error())
			return
		}

	case "playerDeath":
		// pass
		if err != nil {
			a.ApiError(w, r, "[4] Unsupported update type! "+err.Error())
			return
		}

	case "serverLag":
		// pass
		if err != nil {
			a.ApiError(w, r, "[4] Unsupported update type! "+err.Error())
			return
		}

	case "serverReady":
		err := a.ApiServerOn(sid)
		if err != nil {
			a.ApiError(w, r, "[5] Failed to update your server! "+err.Error())
			return
		}

	case "serverOff":
		err := a.ApiServerOff(sid)
		if err != nil {
			a.ApiError(w, r, "[5] Failed to update your server! "+err.Error())
			return
		}

	default:
		a.ApiError(w, r, "[4] Invalid Update type! "+updateType)
		return
	}
}
