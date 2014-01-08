# paginate

paginate is a Go library that lets you paginate any data source. Typically, it's intended to be used in a flow resembling something like this:

```
1. Create cursor based on request url.
2. Fetch data based on cursor.
3. Pass data to cursor.
4. Respond with items and pagination urls.
```

## Example
Paginate comments stored in a database:

```go
func commentsHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Create cursor based on request url.
	defaults := &paginate.Defaults{Count: 5, Order: "updated_at", Direction: paginate.DESC}
	options := &paginate.Options{Prefetch: true}
	cursor := paginate.NewCursorFromUrl(r.URL, defaults, options)

	// 2. Fetch data based on cursor.
  // Note: Use cursor.PrefetchCount() instead of cursor.Count if the cursor was
  // created with the Prefetch option set to true.
  comments := ...

  // 3. Pass data to cursor so that it can generate the next and previous cursors.
	items := make([]paginate.Item, len(comments))
	for i, comment := range comments {
		items[i] = comment
	}
	cursor.Items = items

	// 3.5. (Optional) Drop extra comment that was prefetched for pagination.
  // This is only necessary when the cursor is created with the Prefetch
  // option.
	if len(comments) > cursor.Count {
		comments = comments[0:cursor.Count]
	}
	return comments, nil

	// 4. Respond with comments and pagination urls.
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
```

You can see the full example in [examples/comments.go](examples/comments.go).
