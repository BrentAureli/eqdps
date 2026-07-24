package eventruntime

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/uija/eqdps/internal/catalog"
	"github.com/uija/eqdps/internal/event"
)

type fakeNotifier struct {
	delivered chan string
}

func (f fakeNotifier) Send(_ context.Context, title, message, _ string, _ bool) error {
	f.delivered <- title + ": " + message
	return nil
}

type fakePlayer struct {
	played  chan string
	volumes chan float64
}

func (f fakePlayer) Play(_ context.Context, id string, volume float64, _ func(error)) error {
	f.played <- id
	if f.volumes != nil {
		f.volumes <- volume
	}
	return nil
}

func TestRuntimeDeliversNotificationsAndSoundsIndependently(t *testing.T) {
	notifications := make(chan string, 1)
	sounds := make(chan string, 1)
	runtime, err := newRuntime(
		[]event.Event{{
			ID: "spell", Title: "Buff", Active: true, TriggerType: event.TriggerSpell,
			Pattern: "Your speed returns to normal.", SpellName: "Alacrity",
			Notification: "%s faded.", Sound: "embedded:Chord.mp3",
		}},
		t.TempDir(),
		t.TempDir(),
		[]catalog.Spell{{Name: "Alacrity", IconID: 1}},
		fakeNotifier{delivered: notifications},
		func(string) (soundPlayer, error) { return fakePlayer{played: sounds}, nil },
		func(err error) { t.Errorf("runtime error: %v", err) },
	)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	runtime.Start(ctx)
	runtime.ObserveLiveLine("[Fri Jul 24 20:15:01 2026] Your speed returns to normal.")

	select {
	case got := <-notifications:
		if got != "Buff: Alacrity faded." {
			t.Fatalf("notification = %q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("notification was not delivered")
	}
	select {
	case got := <-sounds:
		if got != "embedded:Chord.mp3" {
			t.Fatalf("sound = %q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("sound was not delivered")
	}
}

func TestRuntimeAppliesCurrentVolumeToPlayback(t *testing.T) {
	sounds := make(chan string, 1)
	volumes := make(chan float64, 1)
	runtime, err := newRuntime(
		[]event.Event{{
			ID: "sound", Title: "Sound", Active: true, TriggerType: event.TriggerText,
			Pattern: "ready", Sound: "embedded:Chord.mp3",
		}},
		t.TempDir(), t.TempDir(), nil,
		fakeNotifier{delivered: make(chan string, 1)},
		func(string) (soundPlayer, error) {
			return fakePlayer{played: sounds, volumes: volumes}, nil
		},
		func(err error) { t.Errorf("runtime error: %v", err) },
	)
	if err != nil {
		t.Fatal(err)
	}
	runtime.SetAudioVolume(.25)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	runtime.Start(ctx)
	runtime.ObserveLiveLine("[Fri Jul 24 20:15:01 2026] ready")

	select {
	case got := <-volumes:
		if got != .25 {
			t.Fatalf("playback volume = %v, want .25", got)
		}
	case <-time.After(time.Second):
		t.Fatal("sound was not played")
	}
}

func TestRuntimeAudioFailureDoesNotStopNotifications(t *testing.T) {
	notifications := make(chan string, 1)
	var errorsMu sync.Mutex
	var errorsSeen []error
	runtime, err := newRuntime(
		[]event.Event{{
			ID: "text", Title: "Ready", Active: true, TriggerType: event.TriggerText,
			Pattern: "ready", Notification: "Ready", Sound: "missing",
		}},
		t.TempDir(), t.TempDir(), nil,
		fakeNotifier{delivered: notifications},
		func(string) (soundPlayer, error) { return nil, context.Canceled },
		func(err error) {
			errorsMu.Lock()
			errorsSeen = append(errorsSeen, err)
			errorsMu.Unlock()
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	runtime.Start(ctx)
	runtime.ObserveLiveLine("[Fri Jul 24 20:15:01 2026] ready")
	select {
	case <-notifications:
	case <-time.After(time.Second):
		t.Fatal("notification stopped after audio initialization failure")
	}
}
