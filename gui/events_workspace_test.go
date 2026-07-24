package main

import (
	"reflect"
	"strings"
	"testing"

	"github.com/uija/eqdps/internal/audio"
	"github.com/uija/eqdps/internal/catalog"
	"github.com/uija/eqdps/internal/event"
	"github.com/uija/eqdps/internal/eventstore"
)

func TestEventsRailOrder(t *testing.T) {
	items := eventRailItems()
	short := make([]string, len(items))
	for index := range items {
		short[index] = items[index].short
	}
	if want := []string{"DPS", "SKY", "EVENTS", "SET"}; !reflect.DeepEqual(short, want) {
		t.Fatalf("rail order = %#v, want %#v", short, want)
	}
}

func TestGUIIconPromptIsDeferredUntilEventsEntryWithLog(t *testing.T) {
	tests := []struct {
		state eventstore.IconSetup
		log   string
		want  bool
	}{
		{state: eventstore.IconSetupUnknown, log: "/eq/Logs/eqlog.txt", want: true},
		{state: eventstore.IconSetupUnknown, log: "", want: false},
		{state: eventstore.IconSetupEnabled, log: "/eq/Logs/eqlog.txt", want: false},
		{state: eventstore.IconSetupDeclined, log: "/eq/Logs/eqlog.txt", want: false},
	}
	for _, test := range tests {
		if got := shouldPromptGUIEventIcons(test.state, test.log); got != test.want {
			t.Errorf("shouldPromptGUIEventIcons(%q, %q) = %v, want %v", test.state, test.log, got, test.want)
		}
	}
}

func TestGUIIconPromptUsesExplicitChoices(t *testing.T) {
	ui := eventsGUI{iconAutomatic: true}
	actions := ui.iconActions()
	labels := make([]string, len(actions))
	for index, action := range actions {
		labels[index] = action.label
	}
	want := []string{"Extract icons", "Ask next time", "Don't ask again"}
	if !reflect.DeepEqual(labels, want) {
		t.Fatalf("automatic icon actions = %#v, want %#v", labels, want)
	}

	ui.iconAutomatic = false
	actions = ui.iconActions()
	labels = labels[:len(actions)]
	for index, action := range actions {
		labels[index] = action.label
	}
	want = []string{"Extract icons", "Close"}
	if !reflect.DeepEqual(labels, want) {
		t.Fatalf("manual icon actions = %#v, want %#v", labels, want)
	}
}

func TestSpellSelectorFiltersByClass(t *testing.T) {
	ui := eventsGUI{
		spells: []catalog.Spell{
			{Name: "Bard Song", Classes: []string{"Bard"}},
			{Name: "Wizard Spell", Classes: []string{"Wizard"}},
		},
		classes:       []string{"Bard", "Wizard"},
		classSelected: 1,
	}
	ui.updateVisibleSpells()
	if len(ui.visibleSpells) != 1 || ui.visibleSpells[0].Name != "Bard Song" {
		t.Fatalf("visible spells = %#v", ui.visibleSpells)
	}
}

func TestInlinePickerAppearsBesideSelectedField(t *testing.T) {
	ui := eventsGUI{editingKind: event.TriggerSpell, picker: "class"}
	items := ui.editorItems()
	if got := strings.Join(items, ","); !strings.Contains(got, "class,picker,spell") {
		t.Fatalf("class picker is not inline with class selector: %q", got)
	}
	ui.picker = "sound"
	items = ui.editorItems()
	if got := strings.Join(items, ","); !strings.Contains(got, "sound,picker") {
		t.Fatalf("sound picker is not inline with sound selector: %q", got)
	}
}

func TestRegexpEditorRejectsInvalidPatternBeforePersistence(t *testing.T) {
	ui := eventsGUI{
		editingKind:   event.TriggerRegexp,
		editingActive: true,
		sounds:        []audio.Sound{{Label: "[No Sound]"}},
	}
	ui.titleEditor.SetText("Invalid")
	ui.patternEditor.SetText("[")
	ui.saveEditor()
	if !strings.Contains(ui.error, "Invalid regular expression") {
		t.Fatalf("editor error = %q", ui.error)
	}
}
