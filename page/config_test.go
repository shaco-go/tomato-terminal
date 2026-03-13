package page

import (
	"testing"
	"github.com/shaco-go/tomato-terminal/config"
)

func TestApplyReadConfigResetsCursorAndUsesParsedLine(t *testing.T) {
	old := *config.Conf
	defer func() {
		*config.Conf = old
	}()

	config.Conf.ItemID = "old"
	config.Conf.Cursor = 50
	config.Conf.Line = 5

	applyReadConfig("new", "12")

	if config.Conf.ItemID != "new" {
		t.Fatalf("unexpected itemID: %s", config.Conf.ItemID)
	}
	if config.Conf.Line != 12 {
		t.Fatalf("unexpected line: %d", config.Conf.Line)
	}
	if config.Conf.Cursor != 0 {
		t.Fatalf("cursor should reset to 0, got: %d", config.Conf.Cursor)
	}
}

func TestApplyReadConfigKeepsCursorWhenItemIDNotChanged(t *testing.T) {
	old := *config.Conf
	defer func() {
		*config.Conf = old
	}()

	config.Conf.ItemID = "same"
	config.Conf.Line = 7
	config.Conf.Cursor = 10

	applyReadConfig("same", "abc")

	if config.Conf.Line != 7 {
		t.Fatalf("invalid line should keep old value, got: %d", config.Conf.Line)
	}
	if config.Conf.Cursor != 10 {
		t.Fatalf("cursor should keep old value when itemID unchanged, got: %d", config.Conf.Cursor)
	}
}

func TestApplyReadConfigTrimsItemIDAndResetsCursor(t *testing.T) {
	old := *config.Conf
	defer func() {
		*config.Conf = old
	}()

	config.Conf.ItemID = "old"
	config.Conf.Cursor = 9
	config.Conf.Line = 5

	applyReadConfig("  new-id  ", "5")

	if config.Conf.ItemID != "new-id" {
		t.Fatalf("itemID should be trimmed, got: %q", config.Conf.ItemID)
	}
	if config.Conf.Cursor != 0 {
		t.Fatalf("cursor should reset to 0 when itemID changes, got: %d", config.Conf.Cursor)
	}
}

func TestApplyReadConfigFallsBackToDefaultWhenNoValidLine(t *testing.T) {
	old := *config.Conf
	defer func() {
		*config.Conf = old
	}()

	config.Conf.Line = 0

	applyReadConfig("same", "0")

	if config.Conf.Line != 10 {
		t.Fatalf("expected default line 10, got: %d", config.Conf.Line)
	}
}
