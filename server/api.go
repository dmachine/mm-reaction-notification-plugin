package main

import (
	"encoding/json"
	"net/http"
)

// initAPI registers all HTTP routes onto the plugin's router.
// Called from OnActivate in plugin.go.
func (p *Plugin) initAPI() {
	apiRouter := p.router.PathPrefix("/api/v1").Subrouter()
	apiRouter.HandleFunc("/prefs", p.handleGetPrefs).Methods(http.MethodGet)
	apiRouter.HandleFunc("/prefs", p.handlePostPrefs).Methods(http.MethodPost)
}

// handleGetPrefs returns the authenticated user's notification preferences.
//
// GET /plugins/com.github.reaction-notification/api/v1/prefs
//
// Response: UserPreferences JSON
// The Mattermost-User-Id header is automatically injected by Mattermost for
// authenticated sessions; no manual token validation is required.
func (p *Plugin) handleGetPrefs(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-Id")
	if userID == "" {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	prefs, err := p.GetUserPreferences(userID)
	if err != nil {
		p.API.LogError("handleGetPrefs: GetUserPreferences failed",
			"user_id", userID, "err", err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(prefs); err != nil {
		p.API.LogError("handleGetPrefs: encode failed", "err", err.Error())
	}
}

// handlePostPrefs updates the authenticated user's notification preferences.
//
// POST /plugins/com.github.reaction-notification/api/v1/prefs
//
// Body: UserPreferences JSON
// Response: updated UserPreferences JSON
func (p *Plugin) handlePostPrefs(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-Id")
	if userID == "" {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	var prefs UserPreferences
	if err := json.NewDecoder(r.Body).Decode(&prefs); err != nil {
		http.Error(w, "Bad request: invalid JSON", http.StatusBadRequest)
		return
	}

	if err := p.SetUserPreferences(userID, &prefs); err != nil {
		p.API.LogError("handlePostPrefs: SetUserPreferences failed",
			"user_id", userID, "err", err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(prefs); err != nil {
		p.API.LogError("handlePostPrefs: encode failed", "err", err.Error())
	}
}
