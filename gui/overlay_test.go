package main

import (
	"testing"
	"time"
)

func TestOverlayDisplaysLatestCompletedFightBetweenFights(t *testing.T) {
	overlay := combatOverlay{fights: []fakeFightSection{
		{name: "latest completed"},
		{name: "older completed"},
	}}

	if got := overlay.displayFight(); got == nil || got.name != "latest completed" {
		t.Fatalf("expected latest completed fight, got %#v", got)
	}
}

func TestOverlayExpiresCompletedFightAfterIdleTimeout(t *testing.T) {
	now := time.Date(2026, 7, 20, 12, 0, 0, 0, time.UTC)
	overlay := combatOverlay{
		fights:      []fakeFightSection{{name: "completed"}},
		idleTimeout: 15 * time.Second,
		completedAt: now,
	}
	if got := overlay.displayFightAt(now.Add(14 * time.Second)); got == nil {
		t.Fatal("completed fight expired too early")
	}
	if got := overlay.displayFightAt(now.Add(15 * time.Second)); got != nil {
		t.Fatalf("completed fight remained after timeout: %#v", got)
	}
}

func TestOverlayPrefersNewestOfConcurrentCurrentFights(t *testing.T) {
	started := time.Date(2026, 7, 20, 12, 0, 0, 0, time.UTC)
	overlay := combatOverlay{fights: []fakeFightSection{
		{name: "older current", current: true, started: started},
		{name: "newer current", current: true, started: started.Add(time.Second)},
	}}

	if got := overlay.displayFight(); got == nil || got.name != "newer current" {
		t.Fatalf("expected newest current fight, got %#v", got)
	}
}

func TestOverlayPrefersCurrentFightMostRecentlyDamagedIntentionallyByYou(t *testing.T) {
	now := time.Date(2026, 7, 20, 12, 0, 0, 0, time.UTC)
	overlay := combatOverlay{fights: []fakeFightSection{
		{name: "newer incidental fight", current: true, started: now.Add(10 * time.Second)},
		{name: "your target", current: true, started: now, lastYouIntentionalOrder: 2},
		{name: "your previous target", current: true, started: now.Add(2 * time.Second), lastYouIntentionalOrder: 1},
	}}

	if got := overlay.displayFight(); got == nil || got.name != "your target" {
		t.Fatalf("expected latest intentional target, got %#v", got)
	}
}

func TestOverlayRetainsMostRecentlyAttackedFightAfterConcurrentKills(t *testing.T) {
	now := time.Date(2026, 7, 20, 12, 0, 0, 0, time.UTC)
	overlay := combatOverlay{fights: []fakeFightSection{
		{name: "first tracker section", lastYouIntentionalOrder: 1},
		{name: "last target killed", lastYouIntentionalOrder: 3},
		{name: "middle target", lastYouIntentionalOrder: 2},
	}}

	if got := overlay.displayFightAt(now.Add(3 * time.Second)); got == nil || got.name != "last target killed" {
		t.Fatalf("expected most recently attacked completed fight, got %#v", got)
	}
}

func TestOverlayKeepsLastIntentionalTargetWhileOnlyIncidentalFightRemains(t *testing.T) {
	now := time.Date(2026, 7, 20, 12, 0, 0, 0, time.UTC)
	overlay := combatOverlay{fights: []fakeFightSection{
		{name: "incidental active fight", current: true, started: now.Add(time.Second)},
		{name: "last intentional target", lastYouIntentionalOrder: 4},
	}}

	if got := overlay.displayFightAt(now.Add(2 * time.Second)); got == nil || got.name != "last intentional target" {
		t.Fatalf("expected last intentional target to remain selected, got %#v", got)
	}
}

func TestWaylandSessionDetection(t *testing.T) {
	t.Setenv("XDG_SESSION_TYPE", "wayland")
	t.Setenv("WAYLAND_DISPLAY", "")
	if !isWaylandSession() {
		t.Fatal("expected XDG Wayland session to be detected")
	}

	t.Setenv("XDG_SESSION_TYPE", "x11")
	t.Setenv("WAYLAND_DISPLAY", "wayland-1")
	if !isWaylandSession() {
		t.Fatal("expected WAYLAND_DISPLAY to be detected")
	}

	t.Setenv("WAYLAND_DISPLAY", "")
	if isWaylandSession() {
		t.Fatal("did not expect X11 session to be detected as Wayland")
	}
}

func TestOverlayPrefersCurrentFightOverHistory(t *testing.T) {
	overlay := combatOverlay{fights: []fakeFightSection{
		{name: "latest completed"},
		{name: "current", current: true},
	}}

	if got := overlay.displayFight(); got == nil || got.name != "current" {
		t.Fatalf("expected current fight, got %#v", got)
	}
}
