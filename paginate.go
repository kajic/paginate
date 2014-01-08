package paginate

import (
	"net/url"
	"strconv"
)

const (
	DESC = 0
	ASC  = 1

	DEFAULT_COUNT = 10
)

// Item is the iterface that must be implemented by objects that can be paginated.
type Item interface {
	// PaginationValue should return a pagination value for this item.
	// For example, when pagination SQL rows this could simply be the id of the row.
	// If paginating a redis zset, this might be the score of the zset member.
	// When paginating an array this would be the index of the array element.
	PaginationValue(c *Cursor) string
}

type Defaults struct {
	Value     string
	Offset    int
	Count     int
	Order     string
	Direction int
}

type Options struct {
	// Prefetch can be set to true to indicate to the cursor that Items will contain one
	// additional item, i.e. the first item from the next page.
	Prefetch bool
}

type Cursor struct {
	// Value should be used an an inclusive value when making queries based on this
	// cursor.
	// For example, in SQL: WHERE created_at <= cursor.Value (assuming descending
	// sort order).
	Value string

	// Offset should be as offset when making queries based on this cursor.
	// For example, in SQL: LIMIT cursor.Offset, 10.
	Offset int

	// Count is the number of items per page. Instead of using this directly you should
	// use PrefetchCount() to account for prefetching.
	// For example, in SQL: LIMIT 0, cursor.PrefetchCount()
	Count int

	// Order is an arbitrary string that should be used by Item.PaginationValue to
	// decide what pagination value to return.
	Order string

	// Direction should control the comparison order when making queries based on this
	// cursor.
	// For example, in SQL:
	// when Direction is DESC: WHERE created_at <= cursor.Value
	// when Direction is ASC:  WHERE created_at >= cursor.Value
	Direction int

	// Url is the url that was used to create this cursor. It's used to construct the
	// next and previous urls.
	Url *url.URL

	Options *Options

	// Items is a slice of Item objects on which we should paginate. It must be set before
	// calling Next, Prev or ToPagination.
	Items []Item
}

func NewCursorFromDefaultsAndOptions(defaults *Defaults, options *Options) *Cursor {
	c := &Cursor{}

	if defaults != nil {
		c.Value = defaults.Value
		c.Offset = defaults.Offset
		c.Count = defaults.Count
		c.Order = defaults.Order
		c.Direction = defaults.Direction
	}

	if c.Count == 0 {
		c.Count = DEFAULT_COUNT
	}

	if options != nil {
		c.Options = options
	} else {
		c.Options = &Options{}
	}

	return c
}

type Pagination struct {
	Next *string `json:"next"`
}

func NewCursorFromUrl(u *url.URL, defaults *Defaults, options *Options) *Cursor {
	c := NewCursorFromDefaultsAndOptions(defaults, options)
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
	for i := 0; i < len(items) && i < c.Count; i++ {
		if items[i].PaginationValue(c) == items[lastItemIndex].PaginationValue(c) {
			count += 1
		}
	}
	return count
}

func (c *Cursor) after(items []Item, lastItemIndex, direction int) *Cursor {
	if items == nil {
		return nil
	}
	if len(items) == 0 {
		return nil
	}
	value := items[lastItemIndex].PaginationValue(c)
	offset := c.equalCount(items, lastItemIndex)
	if offset == c.Count && value == c.Value {
		offset += c.Offset
	}
	return &Cursor{value, offset, c.Count, c.Order, direction, c.Url, c.Options, nil}
}

func (c *Cursor) PrefetchCount() int {
	if c.Options.Prefetch {
		return c.Count + 1
	} else {
		return c.Count
	}
}

func (c *Cursor) Prev() *Cursor {
	lastItemIndex := 0
	var newDirection int
	if c.Direction == ASC {
		newDirection = DESC
	} else {
		newDirection = ASC
	}
	return c.after(c.Items, lastItemIndex, newDirection)
}

func (c *Cursor) Next() *Cursor {
	if c.Options.Prefetch && len(c.Items) <= c.Count {
		return nil
	}
	return c.after(c.Items, c.lastItemIndex(c.Items), c.Direction)
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

func (c *Cursor) ToPagination() *Pagination {
	pagination := &Pagination{}
	if c.Items != nil {
		next := c.Next()
		if next != nil {
			nextUrlString := next.ToUrl().String()
			pagination.Next = &nextUrlString
		}
	}
	return pagination
}
