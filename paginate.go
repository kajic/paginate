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

type Item interface {
	OrderValue(order string) string
}

type Interface interface {
	Equal(order string, i, j int) bool
	Value(order string, i int) string
	Len() int
}

type Page struct {
	Items []Item
}

func (p *Page) Equal(order string, i, j int) bool {
	return p.Value(order, i) == p.Value(order, j)
}

func (p *Page) Value(order string, i int) string {
	return p.Items[i].OrderValue(order)
}

func (p *Page) Len() int {
	return len(p.Items)
}

type Config struct {
	Count     int
	Order     string
	Direction int
}

type Cursor struct {
	Value     string
	Offset    int
	Count     int
	Order     string
	Direction int
}

func (c Cursor) DirectionString() string {
	if c.Direction == ASC {
		return "asc"
	} else {
		return "desc"
	}
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
		cursor.Count = config.Count
	}
	if cursor.Order == "" {
		cursor.Order = config.Order
	}
	if cursor.Direction == 0 {
		cursor.Direction = config.Direction
	}
	return &Pagination{cursor, config}
}

func FromUrl(rawurl *url.URL, config Config) (*Pagination, error) {
	cursor, err := NewCursorFromQuery(rawurl.RawQuery)
	if err != nil {
		return nil, err
	}
	return &Pagination{cursor, config}, nil
}

func (p *Pagination) after(items Interface, last, direction int) *Pagination {
	if items.Len() == 0 {
		return nil
	}
	value := items.Value(p.Order, last)
	offset := p.equalCount(items, p.Order, last)
	if offset == p.Count && value == p.Value {
		offset += p.Offset
	}
	cursor := Cursor{value, offset, p.Count, p.Order, direction}
	return &Pagination{cursor, p.config}
}

func (p *Pagination) Prev(items Interface) *Pagination {
	min := 0
	return p.after(items, min, p.Direction*-1)
}

func (p *Pagination) Next(items Interface, next_page_prefetched bool) *Pagination {
	if next_page_prefetched && items.Len() <= p.Count {
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
	query.Set("direction", p.DirectionString())
	newurl, err := url.Parse(baseurl.String())
	if err != nil {
		return nil, err
	}
	newurl.RawQuery = query.Encode()
	return newurl, nil
}
