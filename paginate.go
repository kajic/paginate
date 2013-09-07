package main

import (
	"fmt"
	"net/url"
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
	pageSize  int
	order     string
	direction int
}

type Cursor struct {
	Value     string
	Offset    int
	Order     string
	Direction int
}

type Pagination struct {
	Cursor
	config Config
}

func (p *Pagination) max(items Interface) int {
	if items.Len() <= p.config.pageSize {
		return items.Len() - 1
	} else {
		return p.config.pageSize
	}
}

func (p *Pagination) equalCount(items Interface, order string, max int) int {
	c := 0
	for i := 0; i < p.config.pageSize; i++ {
		if items.Equal(order, i, max) {
			c += 1
		}
	}
	return c
}

func NewPagination(cursor Cursor, config Config) *Pagination {
	if cursor.Order == "" {
		cursor.Order = config.order
	}
	if cursor.Direction == 0 {
		cursor.Direction = config.direction
	}
	return &Pagination{cursor, config}
}

func NewPaginationFromUrl(u *url.URL, c Config) Pagination {
	return Pagination{config: c}
}

func (p *Pagination) after(items Interface, last, direction int) *Pagination {
	value := items.Value(p.Order, last)
	offset := p.equalCount(items, p.Order, last)

	if offset == p.config.pageSize && value == p.Value {
		offset += p.Offset
	}

	cursor := Cursor{value, offset, p.Order, direction}
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
	config := Config{pageSize: 2, order: "updated_at", direction: DESC}
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
}
