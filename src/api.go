package main

import (
	"errors"
	"github.com/go-sql-driver/mysql"
	"net/http"
	"strconv"
	"encoding/json"
)

// Max servers returned per page by the public API.
const ApiServersPerPage = 50

// apiServer is the public JSON shape for a server. It deliberately omits
// internal fields (ID, OwnerID, LastSpark, Status) that live on Server.
type apiServer struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	XMLFeedLink  string `json:"xmlFeedLink"`
	PlayerCount  uint64 `json:"playerCount"`
	Gamemodes    string `json:"gamemodes"`
	Version      string `json:"version"`
	Languages    string `json:"languages"`
	RequiresMods bool   `json:"requiresMods"`
	WebsiteURL   string `json:"websiteURL"`
	ChatURL      string `json:"chatURL"`
	IconURL      string `json:"iconURL"`
	IP           string `json:"ip"`
}

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

// ApiListServers handles GET /api/v1/listServers[?page=N].
// Returns up to ApiServersPerPage (50) servers per page as JSON.
func (a *App) ApiListServers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	// Parse the requested page (defaults to 1).
	page := 1
	if p, err := strconv.Atoi(q.Get("page")); err == nil && p > 0 {
		page = p
	}

	// Reuse the same visibility rules as the site (hides suspended owners).
	servers, err := a.ListServers(ServerFilter{})
	if err != nil {
		a.ApiError(w, r, "Failed to load servers: "+err.Error())
		return
	}

	// --- Pagination (max ApiServersPerPage per page) ---
	total := len(servers)
	totalPages := (total + ApiServersPerPage - 1) / ApiServersPerPage
	if totalPages < 1 {
		totalPages = 1
	}
	if page > totalPages {
		page = totalPages
	}

	start := (page - 1) * ApiServersPerPage
	if start > total {
		start = total
	}
	end := start + ApiServersPerPage
	if end > total {
		end = total
	}
	paged := servers[start:end]
	// ----------------------------------------------------

	out := make([]apiServer, 0, len(paged))
	for _, s := range paged {
		out = append(out, apiServer{
			Name:         s.Name,
			Description:  s.Description,
			XMLFeedLink:  s.XMLFeedLink,
			PlayerCount:  s.PlayerCount,
			Gamemodes:    s.Gamemodes,
			Version:      s.Version,
			Languages:    s.Languages,
			RequiresMods: s.RequiresMods,
			WebsiteURL:   s.WebsiteURL,
			ChatURL:      s.ChatURL,
			IconURL:      s.IconURL,
			IP:           s.IP,
		})
	}

	resp := struct {
		Page       int         `json:"page"`
		PerPage    int         `json:"perPage"`
		TotalPages int         `json:"totalPages"`
		Total      int         `json:"total"`
		HasPrev    bool        `json:"hasPrev"`
		HasNext    bool        `json:"hasNext"`
		Servers    []apiServer `json:"servers"`
	}{
		Page:       page,
		PerPage:    ApiServersPerPage,
		TotalPages: totalPages,
		Total:      total,
		HasPrev:    page > 1,
		HasNext:    page < totalPages,
		Servers:    out,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		a.ApiError(w, r, "Failed to encode servers: "+err.Error())
		return
	}
}