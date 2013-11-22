package paginate

import (
	"fmt"
	"net/url"
	"strconv"
)

const (
	ASC  = 1
	DESC = -1
)

type Pager interface {
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

func NewCursorFromQuery(query string) (Cursor, error) {
	c := Cursor{}
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
			return c, fmt.Errorf("'%s' in not a supported direction, use -1 (ASC) or 1 (DESC)", direction)
		}
	}
	return c, nil
}

func (p *Pagination) max(items []Pager) int {
	if len(items) <= p.Count {
		return len(items) - 1
	} else {
		return p.Count
	}
}

func (p *Pagination) equalCount(items []Pager, max int) int {
	c := 0
	for i := 0; i < p.Count; i++ {
		if items[i].PaginationValue(p) == items[max].PaginationValue(p) {
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

func (p *Pagination) after(items []Pager, last, direction int) *Pagination {
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

func (p *Pagination) Prev(items []Pager) *Pagination {
	min := 0
	return p.after(items, min, p.Direction*-1)
}

func (p *Pagination) Next(items []Pager, next_page_prefetched bool) *Pagination {
	if next_page_prefetched && len(items) <= p.Count {
		return nil
	}
	max := p.max(items)
	return p.after(items, max, p.Direction)
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
