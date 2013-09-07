package main

import (
	"fmt"
	"net/url"
)

const (
	ASC  = 1
	DESC = 2
)

type Item interface {
	PaginationValue(key string) interface{}
}

type Interface interface {
	Equal(order string, i, j int) bool
	Value(order string, i int) interface{}
	Len() int
}

type Page struct {
	items []Item
}

func (p *Page) Equal(order string, i, j int) bool {
	return p.Value(order, i) == p.Value(order, j)
}

func (p *Page) Value(order string, i int) interface{} {
	return p.items[i].PaginationValue(order)
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
	Value  interface{}
	Offset int
	Order  string
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

func (p *Pagination) next(items Interface) *Pagination {
	max := p.max(items)
	value := items.Value(p.Order, max)
	offset := p.equalCount(items, p.Order, max)

	if offset == p.config.pageSize && value == p.Value {
		offset += p.Offset
	}

	return &Pagination{Cursor{value, offset, p.Order}, p.config}
}

type Comment struct {
	text       string
	created_at int
	updated_at int
}

func (c *Comment) PaginationValue(key string) interface{} {
	switch {
	case key == "created_at":
		return c.created_at
	case key == "updated_at":
		return c.updated_at
	default:
		return nil
	}
}

func main() {
	pagination := Pagination{config: Config{pageSize: 2, order: "created_at"}}

	items := &Page{[]Item{
		&Comment{"a", 0, 4},
		&Comment{"b", 1, 4},
		&Comment{"c", 2, 5},
		&Comment{"d", 3, 5},
	}}
	next := pagination.next(items)

	fmt.Printf("next value: %i, offset: %i\n", next.Value, next.Offset)
}
