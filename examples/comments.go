package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/kajic/paginate"
)

type Comment struct {
	text       string
	created_at int
	updated_at int
}

func (c *Comment) PaginationValue(p *paginate.Pagination) string {
	switch {
	case p.Order == "created_at":
		return strconv.Itoa(c.created_at)
	case p.Order == "updated_at":
		return strconv.Itoa(c.updated_at)
	default:
		return ""
	}
}

func OpenDatabase(addr string) (*sql.DB, error) {
	db, err := sql.Open("mysql", addr)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func GetComments(p *paginate.Pagination) ([]paginate.Pager, error) {
	var where string
	if p.Direction == paginate.ASC {
		where = fmt.Sprintf("%s >= %s", p.Order, p.Value)
	} else {
		where = fmt.Sprintf("%s <= %s", p.Order, p.Value)
	}
	var direction string
	if p.Direction == paginate.ASC {
		direction = "ASC"
	} else {
		direction = "DESC"
	}
	order := fmt.Sprintf("%s %s", p.Order, direction)

	q := `
	SELECT text, created_at, updated_at
	FROM   comments
	WHERE  ` + where + `
	ORDER BY ` + order + `
	LIMIT ?, ?
	`
	db, err := OpenDatabase("database url")
	if err != nil {
		return nil, err
	}
	// Note we fech an additional item to allow the pagination library to immedialely determine
	// if there is a next page after the current one.
	rows, err := db.Query(q, p.Offset, p.Count+1)
	if err != nil {
		return nil, err
	}

	var items []paginate.Pager
	for rows.Next() {
		var c *Comment
		if err := rows.Scan(&c.text, &c.created_at, &c.updated_at); err != nil {
			panic(err)
		}
		items = append(items, c)
	}
	return items, nil
}

func commentsHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Create pagination object based on request url.
	c := paginate.Cursor{Count: 5, Order: "updated_at", Direction: paginate.DESC}
	p, _ := paginate.FromUrl(r.URL, c)

	// 2. Query data source based on request parameters.
	items, _ := GetComments(p)

	// 3. Create pagination urls based on returned items.
	next := p.Next(items, true)
	prev := p.Prev(items)

	// 4. Respond with items and pagination urls.
	comments := make([]*Comment, len(items))
	for i, item := range items {
		comments[i] = item.(*Comment)
	}
	fmt.Fprint(w, comments)
	if next != nil {
		nexturl, _ := next.ToUrl(r.URL)
		fmt.Fprint(w, nexturl)
	}
	if prev != nil {
		prevurl, _ := prev.ToUrl(r.URL)
		fmt.Fprint(w, prevurl)
	}
}

func main() {
	http.HandleFunc("/comments", commentsHandler)
	http.ListenAndServe(":8080", nil)
}
