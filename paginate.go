package paginate

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
)

const (
	DESC = 0
	ASC  = 1
)

type Item interface {
	PaginationValue(p *Pagination) string
}

func ConvertToItems(its interface{}) ([]Item, error) {
	itsValue := reflect.ValueOf(its)
	itsKind := itsValue.Kind()
	if itsKind != reflect.Array && itsKind != reflect.Slice {
		return nil, fmt.Errorf("Expected items to be an Array or a Slice, got %s", itsKind)
	}
	itsLength := itsValue.Len()
	items := make([]Item, itsLength)
	for i := 0; i < itsLength; i++ {
		itsItem := itsValue.Index(i)
		if item, ok := itsItem.Interface().(Item); ok {
			items[i] = item
		} else {
			return nil, fmt.Errorf("item #%d does not implement the Item interface: %s", i, itsItem)
		}
	}
	return items, nil
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
	errors := []error{}
	m, err := url.ParseQuery(query)
	if err != nil {
		errors = append(errors)
		goto exit
	}

	if v, ok := m["value"]; ok {
		c.Value = v[0]
	}
	if v, ok := m["offset"]; ok {
		offset, err := strconv.Atoi(v[0])
		if err != nil {
			errors = append(errors, fmt.Errorf("error when parsing offset %s: %s", v[0], err))
		} else {
			c.Offset = offset
		}
	}
	if v, ok := m["count"]; ok {
		count, err := strconv.Atoi(v[0])
		if err != nil {
			errors = append(errors, fmt.Errorf("error when parsing count %s: %s", v[0], err))
		} else {
			c.Count = count
		}
	}
	if v, ok := m["order"]; ok {
		c.Order = v[0]
	}
	if v, ok := m["direction"]; ok {
		direction, err := strconv.Atoi(v[0])
		if err != nil {
			errors = append(errors, fmt.Errorf("error when parsing direction %s: %s", v[0], err))
		} else {
			if direction == ASC || direction == DESC {
				c.Direction = direction
			} else {
				errors = append(errors, fmt.Errorf("'%s' in not a supported direction, use 0 (DESC) or 1 (ASC)", direction))
			}
		}
	}

exit:
	if len(errors) == 0 {
		return c, nil
	} else {
		return c, fmt.Errorf("errors while parsing query string %s: %s", query, errors)
	}
}

func (p *Pagination) lastItemIndex(items []Item) int {
	if len(items) <= p.Count {
		return len(items) - 1
	} else {
		// This is the index of the first item on the next page.
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
	if defaults.Count == 0 {
		defaults.Count = 10
	}
	if cursor.Count == 0 {
		cursor.Count = defaults.Count
	}

	if cursor.Value == "" {
		cursor.Value = defaults.Value
	}
	if cursor.Offset == 0 {
		cursor.Offset = defaults.Offset
	}
	if cursor.Order == "" {
		cursor.Order = defaults.Order
	}
	if cursor.Direction == 0 {
		cursor.Direction = defaults.Direction
	}

	return &Pagination{cursor, defaults}
}

func NewPaginationFromUrl(rawurl *url.URL, defaults *Cursor) (*Pagination, error) {
	cursor, err := NewCursorFromQuery(rawurl.RawQuery)
	return NewPagination(cursor, *defaults), err
}

func NewPaginationFromDefaults(defaults Cursor) *Pagination {
	return NewPagination(Cursor{}, defaults)
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

func (p *Pagination) Prev(its interface{}) (*Pagination, error) {
	items, err := ConvertToItems(its)
	if err != nil {
		return nil, err
	}
	lastItemIndex := 0
	var newDirection int
	if p.Direction == ASC {
		newDirection = DESC
	} else {
		newDirection = ASC
	}
	return p.after(items, lastItemIndex, newDirection), nil
}

func (p *Pagination) Next(its interface{}, next_page_prefetched bool) (*Pagination, error) {
	items, err := ConvertToItems(its)
	if err != nil {
		return nil, err
	}
	if next_page_prefetched && len(items) <= p.Count {
		return nil, nil
	}
	return p.after(items, p.lastItemIndex(items), p.Direction), nil
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
