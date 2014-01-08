package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	_ "github.com/mattn/go-sqlite3"

	"github.com/kajic/paginate"
)

type Comment struct {
	text       string
	created_at int
	updated_at int
}

func (c *Comment) PaginationValue(p *paginate.Cursor) string {
	switch {
	case p.Order == "created_at":
		return strconv.Itoa(c.created_at)
	case p.Order == "updated_at":
		return strconv.Itoa(c.updated_at)
	default:
		return ""
	}
}

func OpenDatabase(driver, addr string) (*sql.DB, error) {
	db, err := sql.Open(driver, addr)
	if err != nil {
		return nil, err
	}
	return db, db.Ping()
}

func GetComments(cursor *paginate.Cursor) ([]*Comment, error) {
	var where string
	if cursor.Direction == paginate.ASC {
		where = fmt.Sprintf("%s >= %s", cursor.Order, cursor.Value)
	} else {
		where = fmt.Sprintf("%s <= %s", cursor.Order, cursor.Value)
	}
	var direction string
	if cursor.Direction == paginate.ASC {
		direction = "ASC"
	} else {
		direction = "DESC"
	}
	order := fmt.Sprintf("%s %s", cursor.Order, direction)

	q := `
	SELECT text, created_at, updated_at
	FROM   comments
	WHERE  ` + where + `
	ORDER BY ` + order + `
	LIMIT ?, ?
	`
	db, err := OpenDatabase("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	// Use PrefetchCount() to fetch one extra comment when the cursor has
	// prefetching enabled.
	// This allows the cursor to immediately determine if there is a next page.
	rows, err := db.Query(q, cursor.Offset, cursor.PrefetchCount())
	if err != nil {
		return nil, err
	}

	var comments []*Comment
	for rows.Next() {
		var c *Comment
		if err := rows.Scan(&c.text, &c.created_at, &c.updated_at); err != nil {
			panic(err)
		}
		comments = append(comments, c)
	}

	// Inform cursor about the fetched comments so that it can later generate the
	// Next and Previous cursors.
	items := make([]paginate.Item, len(comments))
	for i, comment := range comments {
		items[i] = comment
	}
	cursor.Items = items

	// Drop the one extra comment that was prefetched for pagination.
	if len(comments) > cursor.Count {
		comments = comments[0:cursor.Count]
	}
	return comments, nil
}

func commentsHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Create cursor based on request url.
	defaults := &paginate.Defaults{Count: 5, Order: "updated_at", Direction: paginate.DESC}
	options := &paginate.Options{Prefetch: true}
	cursor := paginate.NewCursorFromUrl(r.URL, defaults, options)

	// 2. Fetch data based on cursor.
	comments, _ := GetComments(cursor)

	// 3. Respond with items and pagination urls.
	fmt.Fprint(w, comments)
	if cursor.Next() != nil {
		next := cursor.Next().ToUrl()
		fmt.Fprint(w, next)
	}
	if cursor.Prev() != nil {
		prev := cursor.Prev().ToUrl()
		fmt.Fprint(w, prev)
	}
}

func main() {
	http.HandleFunc("/comments", commentsHandler)
	http.ListenAndServe(":8080", nil)
}
