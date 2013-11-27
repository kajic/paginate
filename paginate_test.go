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

func WordsToItems(words []*Word) []Item {
	items := make([]Item, len(words))
	for i := 0; i < len(items); i++ {
		items[i] = words[i]
	}
	return items
}

func TestNextNoItems(t *testing.T) {
	cursor := NewCursorFromUrl(&url.URL{}, nil)

	words := []*Word{}
	next := cursor.Next(WordsToItems(words), true)

	if next != nil {
		t.Fatal("expected next to be nil, got %s", next)
	}
}

func TestNextWithOnePage(t *testing.T) {
	cursor := NewCursorFromUrl(&url.URL{}, nil)

	words := []*Word{&Word{"Hello"}, &Word{"world!"}}
	items := WordsToItems(words)
	next := cursor.Next(items, true)

	if next != nil {
		t.Fatal("expected next to be nil, got %s", next)
	}
}
