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

func OpenDatabase(addr string) (*sql.DB, error) {
	db, err := sql.Open("mysql", addr)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func GetComments(p *paginate.Pagination) []paginate.Item {
	var where string
	if p.Direction == paginate.ASC {
		where = fmt.Sprintf("%s >= %s", p.Order, p.Value)
	} else {
		where = fmt.Sprintf("%s <= %s", p.Order, p.Value)
	}
	order := fmt.Sprintf("%s %s", p.Order, p.DirectionString())

	q := `
	SELECT text, created_at, updated_at
	FROM   comments
	WHERE  ` + where + `
	ORDER BY ` + order + `
	LIMIT ?, ?
	`
	db, _ := OpenDatabase("database url")
	rows, _ := db.Query(q, p.Offset, p.Count+1)

	var items []paginate.Item
	for rows.Next() {
		var c *Comment
		if err := rows.Scan(&c.text, &c.created_at, &c.updated_at); err != nil {
			panic(err)
		}
		items = append(items, c)
	}
	return items
}

func commentsHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Create pagination object based on request url.
	c := paginate.Config{Count: 5, Order: "updated_at", Direction: paginate.DESC}
	p, _ := paginate.FromUrl(r.URL, c)

	// 2. Query data source based on request parameters.
	comments := &paginate.Page{GetComments(p)}

	// 3. Create pagination urls based on data.
	next := p.Next(comments, true)
	prev := p.Prev(comments)

	// 4. Respond with data and pagination urls.
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
