package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const maxRecentLogs = 8

type guiSettings struct {
	LastLogfile    string   `json:"last_logfile,omitempty"`
	RecentLogfiles []string `json:"recent_logfiles,omitempty"`
	OverlayVisible bool     `json:"overlay_visible,omitempty"`
	WaylandNotice  bool     `json:"wayland_overlay_notice_shown,omitempty"`
	MainFontScale  float32  `json:"main_font_scale,omitempty"`
	DPSFontScale   float32  `json:"dps_font_scale,omitempty"`
	DPSOpacity     float32  `json:"dps_opacity,omitempty"`
}

func (settings *guiSettings) normalize() {
	settings.MainFontScale = clampSetting(settings.MainFontScale, .75, 1.5, 1)
	settings.DPSFontScale = clampSetting(settings.DPSFontScale, .75, 1.5, 1)
	settings.DPSOpacity = clampSetting(settings.DPSOpacity, .35, 1, .8)
}

func clampSetting(value, minimum, maximum, fallback float32) float32 {
	if value == 0 {
		return fallback
	}
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}

func settingsPath() (string, error) {
	directory, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(directory, "eqdps", "gui.json"), nil
}

func loadSettings() (guiSettings, error) {
	path, err := settingsPath()
	if err != nil {
		return guiSettings{}, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return guiSettings{}, nil
	}
	if err != nil {
		return guiSettings{}, err
	}
	var settings guiSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return guiSettings{}, err
	}
	settings.normalize()
	return settings, nil
}

func saveSettings(settings guiSettings) error {
	path, err := settingsPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), "gui-*.json")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}

func (settings *guiSettings) rememberLog(path string) {
	path = filepath.Clean(path)
	settings.LastLogfile = path
	recent := []string{path}
	for _, candidate := range settings.RecentLogfiles {
		if candidate != path && len(recent) < maxRecentLogs {
			recent = append(recent, candidate)
		}
	}
	settings.RecentLogfiles = recent
}
