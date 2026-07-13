package main

import (
	"testing"
	"time"

	"github.com/rivo/tview"
	"github.com/uija/eqdps/internal/combat"
)

func TestFitTextTruncatesWithEllipsis(t *testing.T) {
	if got := fitText("an exceptionally long target name", 12); got != "an except..." {
		t.Fatalf("unexpected truncated text: %q", got)
	}
}

func TestTableLayoutForNarrowWidthKeepsTextColumnsUsable(t *testing.T) {
	layout := tableLayoutForWidth(70)
	if layout.combatantWidth < 10 {
		t.Fatalf("combatant width too small: %d", layout.combatantWidth)
	}
	if layout.targetWidth < 8 {
		t.Fatalf("target width too small: %d", layout.targetWidth)
	}
}

func TestHistoryDuration(t *testing.T) {
	tests := map[string]time.Duration{
		"Now":          0,
		"Last Hour":    time.Hour,
		"Last 4 Hours": 4 * time.Hour,
		"Last 8 Hours": 8 * time.Hour,
		"Last Day":     24 * time.Hour,
	}
	for label, expected := range tests {
		got, ok := historyDuration(label)
		if !ok {
			t.Fatalf("expected %q to be recognized", label)
		}
		if got != expected {
			t.Fatalf("expected %q to map to %s, got %s", label, expected, got)
		}
	}
}

func TestDamageBreakdownShowsDPSAndPercentInExpectedColumns(t *testing.T) {
	started := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
	player := combat.PlayerStats{
		Name:        "You",
		Damage:      100,
		FirstSeen:   started,
		LastSeen:    started.Add(9 * time.Second),
		DamageTypes: map[string]int{"Tuyen's Chant of Flame": 40},
	}
	table := tview.NewTable()

	nextRow := addDamageBreakdownRows(table, 0, player, tableLayoutForWidth(100))
	if nextRow != 1 {
		t.Fatalf("expected one detail row, got next row %d", nextRow)
	}
	if got := table.GetCell(0, 2).Text; got != "4.00" {
		t.Fatalf("expected ability DPS in DPS column, got %q", got)
	}
	if got := table.GetCell(0, 6).Text; got != "40.0%" {
		t.Fatalf("expected percentage in Last Target column, got %q", got)
	}
}
