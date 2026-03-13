package fq

import (
	"strings"
	"reflect"
	"testing"
	"github.com/shaco-go/tomato-terminal/config"
)

func TestNextThenPrevDoesNotRepeatCurrentPage(t *testing.T) {
	old := *config.Conf
	defer func() {
		*config.Conf = old
	}()

	config.Conf.ItemID = "chapter-1"
	config.Conf.Line = 2
	config.Conf.Cursor = 4

	r := &Reader{
		chapter: &chapterItem{
			ItemID:     "chapter-1",
			PrevItemID: "chapter-0",
			NextItemID: "chapter-2",
			Content:    []string{"0", "1", "2", "3", "4", "5", "6", "7"},
		},
	}

	next := r.Next()
	if !reflect.DeepEqual(next, []string{"4", "5"}) {
		t.Fatalf("unexpected next page: %v", next)
	}

	prev := r.Prev()
	if !reflect.DeepEqual(prev, []string{"2", "3"}) {
		t.Fatalf("prev should go to previous page, got: %v", prev)
	}
}

func TestPrevThenNextDoesNotRepeatCurrentPage(t *testing.T) {
	old := *config.Conf
	defer func() {
		*config.Conf = old
	}()

	config.Conf.ItemID = "chapter-1"
	config.Conf.Line = 2
	config.Conf.Cursor = 4

	r := &Reader{
		chapter: &chapterItem{
			ItemID:     "chapter-1",
			PrevItemID: "chapter-0",
			NextItemID: "chapter-2",
			Content:    []string{"0", "1", "2", "3", "4", "5", "6", "7"},
		},
	}

	prev := r.Prev()
	if !reflect.DeepEqual(prev, []string{"2", "3"}) {
		t.Fatalf("unexpected prev page: %v", prev)
	}

	next := r.Next()
	if !reflect.DeepEqual(next, []string{"4", "5"}) {
		t.Fatalf("next should go to next page, got: %v", next)
	}
}

func TestNextAtLastLineShouldNotSwitchChapterImmediately(t *testing.T) {
	old := *config.Conf
	defer func() {
		*config.Conf = old
	}()

	config.Conf.ItemID = "chapter-1"
	config.Conf.Line = 3
	config.Conf.Cursor = 7

	r := &Reader{
		chapter: &chapterItem{
			ItemID:     "chapter-1",
			PrevItemID: "chapter-0",
			NextItemID: "chapter-2",
			Content:    []string{"0", "1", "2", "3", "4", "5", "6", "7"},
		},
	}

	next := r.Next()
	if !reflect.DeepEqual(next, []string{"7"}) {
		t.Fatalf("next should return current chapter tail first, got: %v", next)
	}
	if config.Conf.ItemID != "chapter-1" {
		t.Fatalf("should not switch chapter yet, got itemID: %s", config.Conf.ItemID)
	}
	if config.Conf.Cursor != 8 {
		t.Fatalf("unexpected cursor after reading tail, got: %d", config.Conf.Cursor)
	}
}

func TestNextWithNonPositiveLineFallsBackToOne(t *testing.T) {
	old := *config.Conf
	defer func() {
		*config.Conf = old
	}()

	config.Conf.ItemID = "chapter-1"
	config.Conf.Line = 0
	config.Conf.Cursor = 0

	r := &Reader{
		chapter: &chapterItem{
			ItemID:     "chapter-1",
			PrevItemID: "chapter-0",
			NextItemID: "chapter-2",
			Content:    []string{"0", "1", "2"},
		},
	}

	next := r.Next()
	if !reflect.DeepEqual(next, []string{"0"}) {
		t.Fatalf("line<=0 should fallback to one line, got: %v", next)
	}
	if config.Conf.Cursor != 1 {
		t.Fatalf("cursor should move by one, got: %d", config.Conf.Cursor)
	}
}

func TestParseContentReturnsClearErrorWhenScriptMissing(t *testing.T) {
	r := NewReader()
	_, err := r.parseContent("")
	if err == nil {
		t.Fatal("expected error when script is empty")
	}
	if !strings.Contains(err.Error(), "__INITIAL_STATE__") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPrevAtFirstChapterThenNextShouldStartFromFirstPage(t *testing.T) {
	old := *config.Conf
	defer func() {
		*config.Conf = old
	}()

	config.Conf.ItemID = "chapter-1"
	config.Conf.Line = 2
	config.Conf.Cursor = 0

	r := &Reader{
		chapter: &chapterItem{
			ItemID:     "chapter-1",
			PrevItemID: "",
			NextItemID: "chapter-2",
			Content:    []string{"0", "1", "2", "3", "4"},
		},
		lastMove: movePrev,
	}

	prev := r.Prev()
	if !reflect.DeepEqual(prev, []string{"没有上一章节"}) {
		t.Fatalf("unexpected prev response: %v", prev)
	}

	next := r.Next()
	if !reflect.DeepEqual(next, []string{"0", "1"}) {
		t.Fatalf("next should start from first page, got: %v", next)
	}
}

func TestNextAtLatestChapterThenPrevShouldShowLastPage(t *testing.T) {
	old := *config.Conf
	defer func() {
		*config.Conf = old
	}()

	config.Conf.ItemID = "chapter-2"
	config.Conf.Line = 2
	config.Conf.Cursor = 5

	r := &Reader{
		chapter: &chapterItem{
			ItemID:     "chapter-2",
			PrevItemID: "chapter-1",
			NextItemID: "",
			Content:    []string{"0", "1", "2", "3", "4"},
		},
		lastMove: moveNext,
	}

	next := r.Next()
	if !reflect.DeepEqual(next, []string{"已到最新章节"}) {
		t.Fatalf("unexpected next response: %v", next)
	}

	prev := r.Prev()
	if !reflect.DeepEqual(prev, []string{"3", "4"}) {
		t.Fatalf("prev should show last page, got: %v", prev)
	}
}
