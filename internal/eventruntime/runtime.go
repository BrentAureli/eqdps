package eventruntime

import (
	"context"
	"fmt"
	"math"
	"os"
	"sync"
	"sync/atomic"

	"github.com/uija/eqdps/internal/audio"
	"github.com/uija/eqdps/internal/catalog"
	"github.com/uija/eqdps/internal/event"
	"github.com/uija/eqdps/internal/eventstore"
	"github.com/uija/eqdps/internal/notify"
	"github.com/uija/eqdps/internal/spellicon"
)

const deliveryQueueSize = 32

type notificationSender interface {
	Send(context.Context, string, string, string, bool) error
}

type soundPlayer interface {
	Play(context.Context, string, float64, func(error)) error
}

type playerFactory func(string) (soundPlayer, error)

type Runtime struct {
	dispatcher *event.Dispatcher
	audioDir   string
	iconDir    string
	iconIDs    map[string]int
	notifier   notificationSender
	newPlayer  playerFactory
	onError    func(error)
	startOnce  sync.Once
	volumeBits atomic.Uint64
}

func Open(onError func(error)) (*Runtime, *eventstore.Store, error) {
	store, err := eventstore.Open()
	if err != nil {
		return nil, nil, err
	}
	events, err := store.Load()
	if err != nil {
		return nil, nil, err
	}
	volume, err := store.AudioVolume()
	if err != nil {
		return nil, nil, err
	}
	spells, err := catalog.Load()
	if err != nil {
		return nil, nil, err
	}
	runtime, err := newRuntime(events, store.AudioDir(), store.IconDir(), spells, notify.Desktop{}, func(audioDir string) (soundPlayer, error) {
		return audio.NewPlayback(audioDir)
	}, onError)
	if err != nil {
		return nil, nil, err
	}
	runtime.SetAudioVolume(volume)
	return runtime, store, nil
}

func newRuntime(
	events []event.Event,
	audioDir, iconDir string,
	spells []catalog.Spell,
	notifier notificationSender,
	newPlayer playerFactory,
	onError func(error),
) (*Runtime, error) {
	dispatcher, err := event.NewDispatcher(events, deliveryQueueSize, onError)
	if err != nil {
		return nil, err
	}
	iconIDs := make(map[string]int, len(spells))
	for _, spell := range spells {
		iconIDs[spell.Name] = spell.IconID
	}
	runtime := &Runtime{
		dispatcher: dispatcher,
		audioDir:   audioDir,
		iconDir:    iconDir,
		iconIDs:    iconIDs,
		notifier:   notifier,
		newPlayer:  newPlayer,
		onError:    onError,
	}
	runtime.SetAudioVolume(1)
	return runtime, nil
}

func (r *Runtime) Start(ctx context.Context) {
	r.startOnce.Do(func() {
		go r.runNotifications(ctx)
		go r.runSounds(ctx)
	})
}

func (r *Runtime) Replace(events []event.Event) error {
	return r.dispatcher.Replace(events)
}

func (r *Runtime) ObserveLiveLine(line string) {
	r.dispatcher.ObserveLiveLine(line)
}

func (r *Runtime) SetAudioVolume(volume float64) {
	if math.IsNaN(volume) || volume < 0 {
		volume = 0
	} else if volume > 1 {
		volume = 1
	}
	r.volumeBits.Store(math.Float64bits(volume))
}

func (r *Runtime) AudioVolume() float64 {
	return math.Float64frombits(r.volumeBits.Load())
}

func (r *Runtime) runNotifications(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case delivery := <-r.dispatcher.Notifications():
			icon := ""
			if delivery.Event.TriggerType == event.TriggerSpell {
				if iconID, ok := r.iconIDs[delivery.Event.SpellName]; ok {
					candidate := spellicon.IconPath(r.iconDir, iconID)
					if info, err := os.Stat(candidate); err == nil && info.Mode().IsRegular() {
						icon = candidate
					}
				}
			}
			if err := r.notifier.Send(
				ctx,
				delivery.Event.Title,
				delivery.NotificationText,
				icon,
				delivery.Event.RequestPersistence,
			); err != nil {
				r.report(fmt.Errorf("desktop notification: %w", err))
			}
		}
	}
}

func (r *Runtime) runSounds(ctx context.Context) {
	player, err := r.newPlayer(r.audioDir)
	if err != nil {
		r.report(fmt.Errorf("initialize event audio: %w", err))
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case soundID := <-r.dispatcher.Sounds():
			if err := player.Play(ctx, soundID, r.AudioVolume(), func(err error) {
				r.report(fmt.Errorf("event audio: %w", err))
			}); err != nil {
				r.report(fmt.Errorf("event audio: %w", err))
			}
		}
	}
}

func (r *Runtime) report(err error) {
	if r.onError != nil {
		r.onError(err)
	}
}
