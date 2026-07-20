package main

import "testing"

func TestOverlayDisplaysLatestCompletedFightBetweenFights(t *testing.T) {
	overlay := combatOverlay{fights: []fakeFightSection{
		{name: "latest completed"},
		{name: "older completed"},
	}}

	if got := overlay.displayFight(); got == nil || got.name != "latest completed" {
		t.Fatalf("expected latest completed fight, got %#v", got)
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
