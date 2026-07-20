package main

import (
	"encoding/json"
	"testing"
)

func TestOverlayVisibilityRoundTripsThroughSettingsJSON(t *testing.T) {
	settings := guiSettings{OverlayVisible: true}
	data, err := json.Marshal(settings)
	if err != nil {
		t.Fatal(err)
	}
	var decoded guiSettings
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.OverlayVisible {
		t.Fatal("expected overlay visibility to survive settings round trip")
	}
}

func TestRememberLogMovesPathToFrontWithoutDuplicates(t *testing.T) {
	settings := guiSettings{RecentLogfiles: []string{"/one", "/two", "/three"}}
	settings.rememberLog("/two")
	if settings.LastLogfile != "/two" || len(settings.RecentLogfiles) != 3 || settings.RecentLogfiles[0] != "/two" || settings.RecentLogfiles[1] != "/one" {
		t.Fatalf("unexpected settings: %#v", settings)
	}
}

func TestRememberLogLimitsRecentPaths(t *testing.T) {
	settings := guiSettings{RecentLogfiles: []string{"1", "2", "3", "4", "5", "6", "7", "8"}}
	settings.rememberLog("new")
	if len(settings.RecentLogfiles) != maxRecentLogs || settings.RecentLogfiles[0] != "new" {
		t.Fatalf("unexpected recent paths: %#v", settings.RecentLogfiles)
	}
}
