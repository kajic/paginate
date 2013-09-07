package main

import (
	"fmt"
	"net/url"
	"strconv"
)

const (
	ASC  = 1
	DESC = -1
)

type Item interface {
	OrderValue(key string) string
}

type Interface interface {
	Equal(order string, i, j int) bool
	Value(order string, i int) string
	Len() int
}

type Page struct {
	items []Item
}

func (p *Page) Equal(order string, i, j int) bool {
	return p.Value(order, i) == p.Value(order, j)
}

func (p *Page) Value(order string, i int) string {
	return p.items[i].OrderValue(order)
}

func (p *Page) Len() int {
	return len(p.items)
}

type Config struct {
	count     int
	order     string
	direction int
}

type Cursor struct {
	Value     string
	Offset    int
	Count     int
	Order     string
	Direction int
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
		switch {
		case v[0] == "desc":
			c.Direction = DESC
		case v[0] == "asc":
			c.Direction = ASC
		default:
			return c, fmt.Errorf("'%s' in not a supported direction, use asc or desc", v)
		}
	}
	return c, nil
}

type Pagination struct {
	Cursor
	config Config
}

func (p *Pagination) max(items Interface) int {
	if items.Len() <= p.Count {
		return items.Len() - 1
	} else {
		return p.Count
	}
}

func (p *Pagination) equalCount(items Interface, order string, max int) int {
	c := 0
	for i := 0; i < p.Count; i++ {
		if items.Equal(order, i, max) {
			c += 1
		}
	}
	return c
}

func NewPagination(cursor Cursor, config Config) *Pagination {
	if cursor.Count == 0 {
		cursor.Count = config.count
	}
	if cursor.Order == "" {
		cursor.Order = config.order
	}
	if cursor.Direction == 0 {
		cursor.Direction = config.direction
	}
	return &Pagination{cursor, config}
}

func (p *Pagination) after(items Interface, last, direction int) *Pagination {
	value := items.Value(p.Order, last)
	offset := p.equalCount(items, p.Order, last)

	if offset == p.Count && value == p.Value {
		offset += p.Offset
	}

	cursor := Cursor{value, offset, p.Count, p.Order, direction}
	return &Pagination{cursor, p.config}
}

func (p *Pagination) prev(items Interface) *Pagination {
	min := 0
	return p.after(items, min, p.Direction*-1)
}

func (p *Pagination) next(items Interface) *Pagination {
	max := p.max(items)
	return p.after(items, max, p.Direction)
}

type Comment struct {
	text       string
	created_at int
	updated_at int
}

func (c *Comment) OrderValue(order string) string {
	switch {
	case order == "created_at":
		return strconv.Itoa(c.created_at)
	case order == "updated_at":
		return strconv.Itoa(c.updated_at)
	default:
		return ""
	}
}

func main() {
	cursor := Cursor{Order: "created_at"}
	config := Config{count: 2, order: "updated_at", direction: DESC}
	pagination := NewPagination(cursor, config)

	items := &Page{[]Item{
		&Comment{"e", 3, 5},
		&Comment{"d", 3, 5},
		&Comment{"c", 2, 5},
		&Comment{"b", 1, 4},
		&Comment{"a", 0, 4},
	}}
	next := pagination.next(items)
	prev := pagination.prev(items)

	fmt.Printf("next value: %i, offset: %i, direction: %i\n", next.Value, next.Offset, next.Direction)
	fmt.Printf("prev value: %i, offset: %i, direction: %i\n", prev.Value, prev.Offset, prev.Direction)

	query := "order=updated_at&direction=desc&value=5&offset=0&count=5"
	cursor, err := NewCursorFromQuery(query)
	if err != nil {
		panic(err)
	}
	pagination = NewPagination(cursor, config)

	items = &Page{[]Item{
		&Comment{"e", 3, 5},
		&Comment{"d", 3, 5},
		&Comment{"c", 2, 5},
		&Comment{"b", 1, 4},
		&Comment{"a", 0, 4},
	}}
	next = pagination.next(items)
	prev = pagination.prev(items)

	fmt.Printf("next value: %i, offset: %i, direction: %i\n", next.Value, next.Offset, next.Direction)
	fmt.Printf("prev value: %i, offset: %i, direction: %i\n", prev.Value, prev.Offset, prev.Direction)

}
