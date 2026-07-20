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

func TestSettingsNormalizationAddsPreferenceDefaults(t *testing.T) {
	settings := guiSettings{}
	settings.normalize()
	if settings.MainFontScale != 1 || settings.DPSFontScale != 1 || settings.DPSOpacity != .8 || settings.IdleTimeoutSec != 15 {
		t.Fatalf("unexpected preference defaults: %#v", settings)
	}
	if settings.MainWidth != 1050 || settings.MainHeight != 700 || settings.OverlayWidth != 520 || settings.OverlayHeight != 310 {
		t.Fatalf("unexpected window size defaults: %#v", settings)
	}
}

func TestSettingsNormalizationClampsIdleTimeoutAndPreservesWindowSizes(t *testing.T) {
	settings := guiSettings{IdleTimeoutSec: 100, MainWidth: 1200, MainHeight: 800, OverlayWidth: 600, OverlayHeight: 400}
	settings.normalize()
	if settings.IdleTimeoutSec != 60 {
		t.Fatalf("expected timeout clamp, got %d", settings.IdleTimeoutSec)
	}
	if settings.MainWidth != 1200 || settings.MainHeight != 800 || settings.OverlayWidth != 600 || settings.OverlayHeight != 400 {
		t.Fatalf("expected saved window sizes, got %#v", settings)
	}
}

func TestSettingsNormalizationClampsPreferenceRanges(t *testing.T) {
	settings := guiSettings{MainFontScale: .1, DPSFontScale: 3, DPSOpacity: .1}
	settings.normalize()
	if settings.MainFontScale != .75 || settings.DPSFontScale != 1.5 || settings.DPSOpacity != .35 {
		t.Fatalf("unexpected clamped preferences: %#v", settings)
	}
}

func TestOverlayFontScaleAllowsFiftyPercent(t *testing.T) {
	settings := guiSettings{DPSFontScale: .5}
	settings.normalize()
	if settings.DPSFontScale != .5 {
		t.Fatalf("expected 50%% overlay scale, got %v", settings.DPSFontScale)
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
