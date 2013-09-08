# paginate

paginate is a Go library that lets you paginate any data source. For example comments stored in a database:

```go
func commentsHandler(w http.ResponseWriter, r *http.Request) {
	c := paginate.Config{Count: 5, Order: "updated_at", Direction: paginate.DESC}
	p, _ := paginate.FromUrl(r.URL, c)

	comments := &paginate.Page{GetComments(p)}

	next := p.Next(comments, true)
	if next != nil {
		nexturl, _ := next.ToUrl(r.URL)
		fmt.Fprintf(w, "next: %s<br>", nexturl)
	}

	prev := p.Prev(comments)
	if prev != nil {
		prevurl, _ := prev.ToUrl(r.URL)
		fmt.Fprintf(w, "prev: %s<br>", prevurl)
	}
}
```

You can see the full example in [examples/comments.go](examples/comments.go).
