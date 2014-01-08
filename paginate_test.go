package paginate

import (
	"net/url"
	"testing"
)

type Word struct {
	Value string
}

func (w *Word) PaginationValue(c *Cursor) string {
	return w.Value
}

func wordsToItems(words []*Word) []Item {
	items := make([]Item, len(words))
	for i := 0; i < len(items); i++ {
		items[i] = words[i]
	}
	return items
}

func TestNextNoItems(t *testing.T) {
	cursor := NewCursorFromUrl(&url.URL{}, nil, &Options{Prefetch: true})

	words := []*Word{}
	cursor.Items = wordsToItems(words)
	next := cursor.Next()

	if next != nil {
		t.Fatalf("expected next to be nil, got %s", next)
	}
}

func TestNextWithOnePage(t *testing.T) {
	cursor := NewCursorFromUrl(&url.URL{}, nil, &Options{Prefetch: true})

	words := []*Word{&Word{"Hello"}, &Word{"world!"}}
	cursor.Items = wordsToItems(words)
	next := cursor.Next()

	if next != nil {
		t.Fatalf("expected next to be nil, got %s", next)
	}
}

func TestNextWithMultiplePages(t *testing.T) {
	cursor := NewCursorFromUrl(&url.URL{}, &Defaults{Count: 2}, &Options{Prefetch: true})

	words := []*Word{&Word{"Hello"}, &Word{"world!"}, &Word{"world!"}, &Word{"how"}, &Word{"are"}}
	cursor.Items = wordsToItems(words[0:3])

	cursor = cursor.Next()

	if cursor == nil {
		t.Fatal("expected cursor be non-nil")
	}
	if cursor.Offset != 1 {
		t.Fatalf("expected offset to be 1, got %d", cursor.Offset)
	}
	if cursor.Value != "world!" {
		t.Fatalf("expected value to be 'world!', got '%s'", cursor.Value)
	}

	cursor.Items = wordsToItems(words[2:5])
	cursor = cursor.Next()

	if cursor == nil {
		t.Fatal("expected cursor be non-nil")
	}
	if cursor.Offset != 0 {
		t.Fatalf("expected offset to be 0, got %d", cursor.Offset)
	}
	if cursor.Value != "are" {
		t.Fatalf("expected value to be 'are', got '%s'", cursor.Value)
	}
}
