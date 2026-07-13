package combat

import (
	"fmt"
	"testing"
	"time"
)

func TestMeterPlayersStayInFirstSeenOrder(t *testing.T) {
	meter := NewMeter()
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	meter.Add(Event{Time: now, Source: "You", Target: "mob", Amount: 10})
	meter.Add(Event{Time: now.Add(time.Second), Source: "A rock golem", Target: "YOU", Amount: 1000})

	players := meter.Players()
	if len(players) != 2 || players[0].Name != "You" || players[1].Name != "A rock golem" {
		t.Fatalf("unexpected players: %#v", players)
	}
}

func TestMeterTracksDamageBreakdownAndCombinesDots(t *testing.T) {
	meter := NewMeter()
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	meter.Add(Event{Time: now, Source: "You", Target: "mob", Amount: 10})
	meter.Add(Event{Time: now.Add(time.Second), Source: "You", Target: "mob", Amount: 40, Ability: "Spell1"})
	meter.Add(Event{Time: now.Add(2 * time.Second), Source: "You", Target: "mob", Amount: 90})
	meter.Add(Event{Time: now.Add(3 * time.Second), Source: "You", Target: "mob", Amount: 50, Ability: "Tuyen's Chant of Disease", DamageOverTime: true})
	meter.Add(Event{Time: now.Add(4 * time.Second), Source: "You", Target: "mob", Amount: 60, Ability: "Tuyen's Chant of Flame", DamageOverTime: true})

	players := meter.Players()
	if len(players) != 1 {
		t.Fatalf("expected one combatant, got %#v", players)
	}
	expected := map[string]int{"Melee": 100, "Spell1": 40, "DoTs": 110}
	for _, entry := range players[0].DamageBreakdown() {
		if expected[entry.Name] != entry.Damage {
			t.Fatalf("unexpected breakdown entry: %#v", entry)
		}
		delete(expected, entry.Name)
	}
	if len(expected) != 0 {
		t.Fatalf("missing breakdown entries: %#v", expected)
	}
}

func TestMeterMergesPossessivePetWhenOwnerIsPresent(t *testing.T) {
	meter := NewMeter()
	now := time.Date(2026, 7, 8, 12, 0, 0, 0, time.UTC)
	meter.Add(Event{Time: now, Source: "Sobatin`s warder", Target: "an orc raider", Amount: 4})
	meter.Add(Event{Time: now.Add(time.Second), Source: "Sobatin", Target: "an orc raider", Amount: 11})

	players := meter.Players()
	if len(players) != 1 || players[0].Name != "Sobatin" || players[0].Damage != 15 || players[0].DamageTypes["Pet: warder"] != 4 {
		t.Fatalf("unexpected merged stats: %#v", players)
	}
}

func TestMeterKeepsPossessiveMobWhenOwnerIsAbsent(t *testing.T) {
	meter := NewMeter()
	now := time.Date(2026, 7, 8, 12, 0, 0, 0, time.UTC)
	meter.Add(Event{Time: now, Source: "Innoruuk`s Chosen", Target: "YOU", Amount: 37})

	players := meter.Players()
	if len(players) != 1 || players[0].Name != "Innoruuk`s Chosen" {
		t.Fatalf("unexpected players: %#v", players)
	}
}

func TestFightTrackerKeepsConcurrentMobsSeparate(t *testing.T) {
	tracker := NewFightTracker()
	now := time.Date(2026, 7, 13, 14, 0, 0, 0, time.UTC)

	tracker.AddDamage(Event{Time: now, Source: "You", Target: "first mob", Amount: 100})
	tracker.AddDamage(Event{Time: now.Add(time.Second), Source: "Alice", Target: "first mob", Amount: 50})
	tracker.AddDamage(Event{Time: now.Add(2 * time.Second), Source: "Bob", Target: "second mob", Amount: 80})
	tracker.AddDamage(Event{Time: now.Add(3 * time.Second), Source: "second mob", Target: "Bob", Amount: 20})

	sections := tracker.DisplaySections()
	if len(sections) != 2 || sections[0].Fight.Mob != "first mob" || sections[1].Fight.Mob != "second mob" {
		t.Fatalf("unexpected mob sections: %#v", sections)
	}
	if sections[0].Fight.Meter.Events() != 2 || sections[1].Fight.Meter.Events() != 2 {
		t.Fatalf("events were not assigned per mob: %#v", sections)
	}
}

func TestFightTrackerDoesNotLimitPlayersPerMob(t *testing.T) {
	tracker := NewFightTracker()
	now := time.Date(2026, 7, 13, 14, 0, 0, 0, time.UTC)
	for index := range 100 {
		tracker.AddDamage(Event{
			Time:   now.Add(time.Duration(index) * time.Millisecond),
			Source: fmt.Sprintf("Player%03d", index),
			Target: "raid mob",
			Amount: index + 1,
		})
	}

	fight, current := tracker.DisplayFight()
	if fight == nil || !current || len(fight.Meter.Players()) != 100 {
		t.Fatalf("expected all 100 players in one mob record, got fight=%#v current=%v", fight, current)
	}
}

func TestFightTrackerUsesSharedMobDurationForPlayerDPS(t *testing.T) {
	tracker := NewFightTracker()
	now := time.Date(2026, 7, 13, 14, 0, 0, 0, time.UTC)
	tracker.AddDamage(Event{Time: now, Source: "You", Target: "mob", Amount: 100})
	tracker.AddDamage(Event{Time: now.Add(9 * time.Second), Source: "Alice", Target: "mob", Amount: 50})

	fight, _ := tracker.DisplayFight()
	if fight.ActiveDuration() != 10*time.Second {
		t.Fatalf("unexpected mob duration: %s", fight.ActiveDuration())
	}
	players := fight.Meter.Players()
	if got := players[0].DPSForDuration(fight.ActiveDuration()); got != 10 {
		t.Fatalf("expected You DPS over mob duration, got %.2f", got)
	}
	if got := players[1].DPSForDuration(fight.ActiveDuration()); got != 5 {
		t.Fatalf("expected Alice DPS over mob duration, got %.2f", got)
	}
}

func TestFightTrackerAOStrikesDoNotChangeAnotherMobsLifecycle(t *testing.T) {
	tracker := NewFightTracker()
	now := time.Date(2026, 7, 13, 14, 7, 30, 0, time.UTC)
	tracker.AddDamage(Event{Time: now, Source: "You", Target: "Hoptor Thaggelum", Amount: 42})
	tracker.AddDamage(Event{Time: now, Source: "You", Target: "a zol ghoul knight", Amount: 56})
	tracker.AddDamage(Event{Time: now, Source: "You", Target: "a bok ghoul knight", Amount: 4})
	tracker.AddDamage(Event{Time: now, Source: "You", Target: "Hoptor Thaggelum pet", Amount: 106})
	tracker.AddDeath(Death{Time: now, Victim: "Hoptor Thaggelum pet", Killer: "You"})

	sections := tracker.DisplaySections()
	if len(sections) != 3 {
		t.Fatalf("expected owner/pet plus two adds, got %#v", sections)
	}
	for _, section := range sections {
		if !section.Current {
			t.Fatalf("pet death must not close a mob section: %#v", section)
		}
	}
	hoptor := sections[0]
	if hoptor.Fight.Mob != "Hoptor Thaggelum" || hoptor.Fight.Meter.Events() != 2 {
		t.Fatalf("expected pet damage in owner record: %#v", hoptor)
	}

	tracker.AddDeath(Death{Time: now.Add(time.Second), Victim: "Hoptor Thaggelum", Killer: "You"})
	sections = tracker.DisplaySections()
	if sections[0].Current || sections[0].Fight.Death.Victim != "Hoptor Thaggelum" {
		t.Fatalf("expected only Hoptor section pending death: %#v", sections)
	}
	if !sections[1].Current || !sections[2].Current {
		t.Fatalf("add sections should remain active: %#v", sections)
	}
}

func TestFightTrackerKeepsLateDamageWithSlainMob(t *testing.T) {
	tracker := NewFightTracker()
	now := time.Date(2026, 7, 5, 17, 22, 20, 0, time.UTC)
	tracker.AddDamage(Event{Time: now, Source: "You", Target: "a rock golem", Amount: 10})
	tracker.AddDeath(Death{Time: now.Add(time.Second), Victim: "a rock golem", Killer: "You"})
	tracker.AddDamage(Event{Time: now.Add(2 * time.Second), Source: "a rock golem", Target: "YOU", Amount: 20})

	fight, current := tracker.DisplayFight()
	if fight == nil || current || fight.Meter.Events() != 2 {
		t.Fatalf("expected late damage in pending mob record, got fight=%#v current=%v", fight, current)
	}
}

func TestFightTrackerPlayerDeathClosesEveryActiveMob(t *testing.T) {
	tracker := NewFightTracker()
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	tracker.AddDamage(Event{Time: now, Source: "You", Target: "first mob", Amount: 10})
	tracker.AddDamage(Event{Time: now, Source: "You", Target: "second mob", Amount: 10})
	tracker.AddDeath(Death{Time: now.Add(time.Second), Victim: "You", Killer: "second mob"})

	sections := tracker.DisplaySections()
	if len(sections) != 2 || sections[0].Current || sections[1].Current {
		t.Fatalf("expected two completed mob records: %#v", sections)
	}
}

func TestFightTrackerEndsEachMobIndependentlyAtLogIdle(t *testing.T) {
	tracker := NewFightTracker()
	now := time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC)
	tracker.AddDamageWithIdle(Event{Time: now, Source: "You", Target: "first mob", Amount: 10}, DefaultIdleTimeout)
	tracker.AddDamageWithIdle(Event{Time: now.Add(10 * time.Second), Source: "You", Target: "second mob", Amount: 10}, DefaultIdleTimeout)
	tracker.AddDamageWithIdle(Event{Time: now.Add(16 * time.Second), Source: "You", Target: "second mob", Amount: 10}, DefaultIdleTimeout)

	sections := tracker.DisplaySections()
	if len(sections) != 2 || sections[0].Fight.Mob != "second mob" || !sections[0].Current {
		t.Fatalf("expected only second mob active first: %#v", sections)
	}
	if sections[1].Fight.Mob != "first mob" || sections[1].Fight.EndReason != "idle timeout" {
		t.Fatalf("expected first mob idle-completed: %#v", sections)
	}
}

func TestFightTrackerHistoryLimitCountsCompletedMobs(t *testing.T) {
	tracker := NewFightTrackerWithHistory(2)
	now := time.Date(2026, 7, 5, 17, 0, 0, 0, time.UTC)
	for i, mob := range []string{"first", "second", "third"} {
		start := now.Add(time.Duration(i) * time.Minute)
		tracker.AddDamage(Event{Time: start, Source: "You", Target: mob, Amount: 10})
		tracker.AddDeath(Death{Time: start.Add(time.Second), Victim: mob, Killer: "You"})
		tracker.EndIdleAtLogTime(start.Add(deathGracePeriod+time.Second), DefaultIdleTimeout)
	}
	sections := tracker.DisplaySections()
	if len(sections) != 2 || sections[0].Fight.Mob != "third" || sections[1].Fight.Mob != "second" {
		t.Fatalf("unexpected limited history: %#v", sections)
	}
}
