# paginate

paginate is a Go library that lets you paginate any data source. Typically, it's intended to be used in a flow resembling something like this:

```
1. Create pagination object based on request url.
2. Query data source based on request parameters.
3. Create pagination urls based or returned items.
4. Respond with items and pagination urls.
```

## Example
Paginate comments stored in a database:

```go
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
```

You can see the full example in [examples/comments.go](examples/comments.go).
