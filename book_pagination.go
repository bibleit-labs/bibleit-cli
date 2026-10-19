package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/bibleit-labs/bibleit-go/rest"
)

func bookPageOptions(args []string, options map[string]string) (url.Values, error) {
	limit, hasLimit := options["--limit"]
	cursor, hasCursor := options["--cursor"]
	if !hasLimit && (!hasCursor || len(args) > 0 && args[0] == "organization") {
		return nil, nil
	}
	if len(args) != 3 || args[0] != "read" {
		return nil, fmt.Errorf("--limit and read --cursor require a whole-book reference: bibleit read <translation> <book>")
	}
	q := url.Values{}
	if hasLimit {
		n, err := strconv.Atoi(limit)
		if err != nil || n < 1 || n > 500 {
			return nil, fmt.Errorf("--limit must be a whole number between 1 and 500")
		}
		q.Set("limit", strconv.Itoa(n))
	}
	if hasCursor {
		if cursor == "" {
			return nil, fmt.Errorf("--cursor requires a nonempty continuation cursor from a previous book page")
		}
		q.Set("after", cursor)
	}
	return q, nil
}

func (a httpClient) completeBook(plan httpPlan) (httpResult, error) {
	pages := a.client.ReadPersonalBookPages(a.ctx, plan.params["account_id"], plan.params["slug"], plan.params["book_id"], nil)
	if a.service {
		pages = a.client.ReadBookPages(a.ctx, plan.params["organization_id"], plan.params["slug"], plan.params["book_id"], nil)
	}
	var complete rest.ScriptureBookRead
	for page, err := range pages {
		if err != nil {
			return httpResult{}, err
		}
		if complete.Verses == nil {
			complete = page
		} else {
			complete.Verses = append(complete.Verses, page.Verses...)
		}
	}
	encoded, err := json.Marshal(complete)
	if err != nil {
		return httpResult{}, err
	}
	var data httpData
	if err := json.Unmarshal(encoded, &data); err != nil {
		return httpResult{}, err
	}
	// Preserve the existing complete-book output. Cursors belong to explicit pages.
	delete(data, "next_cursor")
	return httpResult{data: map[string]any(data)}, nil
}
