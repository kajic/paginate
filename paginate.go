package paginate

import (
	"fmt"
	"net/url"
	"strconv"
)

const (
	DESC = 0
	ASC  = 1
)

type Item interface {
	PaginationValue(p *Pagination) string
}

type Cursor struct {
	Value     string
	Offset    int
	Count     int
	Order     string
	Direction int
}

type Pagination struct {
	Cursor
	defaults Cursor
}

func NewCursor(defaults *Cursor) Cursor {
	var cursor Cursor
	if defaults == nil {
		cursor = Cursor{}
	} else {
		cursor = *defaults
	}

	if cursor.Count == 0 {
		cursor.Count = 10
	}
	return cursor
}

func NewCursorFromQuery(query string) (Cursor, error) {
	c := NewCursor(nil)
	m, err := url.ParseQuery(query)
	if err != nil {
		return c, err
	}

	if v, ok := m["value"]; ok {
		c.Value = v[0]
	}
	if v, ok := m["offset"]; ok {
		offset, err := strconv.Atoi(v[0])
		if err != nil {
			return c, err
		}
		c.Offset = offset
	}
	if v, ok := m["count"]; ok {
		count, err := strconv.Atoi(v[0])
		if err != nil {
			return c, err
		}
		c.Count = count
	}
	if v, ok := m["order"]; ok {
		c.Order = v[0]
	}
	if v, ok := m["direction"]; ok {
		direction, err := strconv.Atoi(v[0])
		if err != nil {
			return c, err
		}
		if direction == ASC || direction == DESC {
			c.Direction = direction
		} else {
			return c, fmt.Errorf("'%s' in not a supported direction, use 0 (DESC) or 1 (ASC)", direction)
		}
	}
	return c, nil
}

func (p *Pagination) lastItemIndex(items []Item) int {
	if len(items) <= p.Count {
		return len(items) - 1
	} else {
		return p.Count
	}
}

func (p *Pagination) equalCount(items []Item, lastItemIndex int) int {
	c := 0
	for i := 0; i < p.Count; i++ {
		if items[i].PaginationValue(p) == items[lastItemIndex].PaginationValue(p) {
			c += 1
		}
	}
	return c
}

func NewPagination(cursor, defaults Cursor) *Pagination {
	if cursor.Value == "" {
		cursor.Value = defaults.Value
	}
	if cursor.Offset == 0 {
		cursor.Offset = defaults.Offset
	}
	if cursor.Count == 0 {
		cursor.Count = defaults.Count
	}
	if cursor.Order == "" {
		cursor.Order = defaults.Order
	}
	if cursor.Direction == 0 {
		cursor.Direction = defaults.Direction
	}
	return &Pagination{cursor, defaults}
}

func FromUrl(rawurl *url.URL, defaults Cursor) (*Pagination, error) {
	cursor, err := NewCursorFromQuery(rawurl.RawQuery)
	if err != nil {
		return nil, err
	}
	return NewPagination(cursor, defaults), nil
}

func (p *Pagination) after(items []Item, last, direction int) *Pagination {
	if len(items) == 0 {
		return nil
	}
	value := items[last].PaginationValue(p)
	offset := p.equalCount(items, last)
	if offset == p.Count && value == p.Value {
		offset += p.Offset
	}
	cursor := Cursor{value, offset, p.Count, p.Order, direction}
	return NewPagination(cursor, p.defaults)
}

func (p *Pagination) Prev(items []Item) *Pagination {
	lastItemIndex := 0
	var newDirection int
	if p.Direction == ASC {
		newDirection = DESC
	} else {
		newDirection = ASC
	}
	return p.after(items, lastItemIndex, newDirection)
}

func (p *Pagination) Next(items []Item, next_page_prefetched bool) *Pagination {
	if next_page_prefetched && len(items) <= p.Count {
		return nil
	}
	return p.after(items, p.lastItemIndex(items), p.Direction)
}

func (p *Pagination) ToUrl(baseurl *url.URL) (*url.URL, error) {
	query, err := url.ParseQuery(baseurl.RawQuery)
	if err != nil {
		return nil, err
	}
	query.Set("value", p.Value)
	query.Set("offset", strconv.Itoa(p.Offset))
	query.Set("count", strconv.Itoa(p.Count))
	query.Set("order", p.Order)
	query.Set("direction", strconv.Itoa(p.Direction))
	newurl, err := url.Parse(baseurl.String())
	if err != nil {
		return nil, err
	}
	newurl.RawQuery = query.Encode()
	return newurl, nil
}
