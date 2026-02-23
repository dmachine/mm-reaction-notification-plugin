package main

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

// kvKeyUserPrefs returns the KV store key for a given user's preferences.
// This is the single source of truth for the key naming convention.
func kvKeyUserPrefs(userID string) string {
	return fmt.Sprintf("user_prefs_%s", userID)
}

// UserPreferences holds per-user notification configuration stored in the KV
// store. The default (Enabled=true, no exclusions) means the plugin is opt-out.
type UserPreferences struct {
	Enabled       bool     `json:"enabled"`
	ExcludedEmoji []string `json:"excluded_emoji"`
}

// GetUserPreferences retrieves preferences for userID from the KV store.
// If no record exists yet, it returns the default preferences with Enabled=true
// so that the plugin is opt-out rather than opt-in.
func (p *Plugin) GetUserPreferences(userID string) (*UserPreferences, error) {
	raw, appErr := p.API.KVGet(kvKeyUserPrefs(userID))
	if appErr != nil {
		return nil, errors.Wrap(appErr, "KVGet failed")
	}
	if raw == nil {
		return &UserPreferences{Enabled: true, ExcludedEmoji: []string{}}, nil
	}
	var prefs UserPreferences
	if err := json.Unmarshal(raw, &prefs); err != nil {
		return nil, errors.Wrap(err, "unmarshal UserPreferences failed")
	}
	// Normalise: ensure ExcludedEmoji is never nil so JSON responses always
	// return [] rather than null.
	if prefs.ExcludedEmoji == nil {
		prefs.ExcludedEmoji = []string{}
	}
	return &prefs, nil
}

// SetUserPreferences writes preferences for userID to the KV store.
func (p *Plugin) SetUserPreferences(userID string, prefs *UserPreferences) error {
	if prefs.ExcludedEmoji == nil {
		prefs.ExcludedEmoji = []string{}
	}
	raw, err := json.Marshal(prefs)
	if err != nil {
		return errors.Wrap(err, "marshal UserPreferences failed")
	}
	if appErr := p.API.KVSet(kvKeyUserPrefs(userID), raw); appErr != nil {
		return errors.Wrap(appErr, "KVSet failed")
	}
	return nil
}
