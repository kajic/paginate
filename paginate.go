package paginate

import (
	"net/url"
	"strconv"
)

const (
	DESC = 0
	ASC  = 1
)

type Cursor struct {
	Value     string
	Offset    int
	Count     int
	Order     string
	Direction int
	Url       *url.URL
}

func NewCursorFromDefaults(c *Cursor) *Cursor {
	if c == nil {
		c = &Cursor{}
	}
	if c.Count == 0 {
		c.Count = 10
	}
	return c
}

type Item interface {
	PaginationValue(c *Cursor) string
}

type Response struct {
	Next *string `json:"next"`
}

func NewCursorFromUrl(u *url.URL, defaults *Cursor) *Cursor {
	c := NewCursorFromDefaults(defaults)
	c.Url = u

	values := u.Query()
	if v, ok := values["value"]; ok {
		c.Value = v[0]
	}
	if v, ok := values["offset"]; ok {
		offset, err := strconv.Atoi(v[0])
		if err != nil {
			c.Offset = offset
		}
	}
	if v, ok := values["count"]; ok {
		count, err := strconv.Atoi(v[0])
		if err == nil {
			c.Count = count
		}
	}
	if v, ok := values["order"]; ok {
		c.Order = v[0]
	}
	if v, ok := values["direction"]; ok {
		direction, err := strconv.Atoi(v[0])
		if err == nil && (direction == ASC || direction == DESC) {
			c.Direction = direction
		}
	}
	return c
}

func (c *Cursor) lastItemIndex(items []Item) int {
	if len(items) <= c.Count {
		return len(items) - 1
	} else {
		// This is the index of the first item on the next page.
		return c.Count
	}
}

func (c *Cursor) equalCount(items []Item, lastItemIndex int) int {
	count := 0
	for i := 0; i < c.Count; i++ {
		if items[i].PaginationValue(c) == items[lastItemIndex].PaginationValue(c) {
			count += 1
		}
	}
	return count
}

func (c *Cursor) after(items []Item, last, direction int) *Cursor {
	if len(items) == 0 {
		return nil
	}
	value := items[last].PaginationValue(c)
	offset := c.equalCount(items, last)
	if offset == c.Count && value == c.Value {
		offset += c.Offset
	}
	return &Cursor{value, offset, c.Count, c.Order, direction, c.Url}
}

func (c *Cursor) Prev(items []Item) *Cursor {
	lastItemIndex := 0
	var newDirection int
	if c.Direction == ASC {
		newDirection = DESC
	} else {
		newDirection = ASC
	}
	return c.after(items, lastItemIndex, newDirection)
}

func (c *Cursor) Next(items []Item, nextPagePrefetched bool) *Cursor {
	if nextPagePrefetched && len(items) <= c.Count {
		return nil
	}
	return c.after(items, c.lastItemIndex(items), c.Direction)
}

func (c *Cursor) ToUrl() *url.URL {
	newUrl := *c.Url
	values := newUrl.Query()
	values.Set("value", c.Value)
	values.Set("offset", strconv.Itoa(c.Offset))
	values.Set("count", strconv.Itoa(c.Count))
	values.Set("order", c.Order)
	values.Set("direction", strconv.Itoa(c.Direction))
	newUrl.RawQuery = values.Encode()
	return &newUrl
}

func (c *Cursor) ToResponse(items []Item, nextPagePrefetched bool) *Response {
	response := &Response{}
	next := c.Next(items, nextPagePrefetched)
	if next != nil {
		nextUrlString := next.ToUrl().String()
		response.Next = &nextUrlString
	}
	return response
}
