package httpx

import (
	"net/http"
	"strconv"
)

type Page struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

func (p Page) Offset() int { return (p.Page - 1) * p.Limit }

// ParsePage legge ?page= e ?limit= dalla query string, con defaults sicuri.
func ParsePage(r *http.Request) Page {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	switch {
	case limit < 1:
		limit = 20
	case limit > 200:
		limit = 200
	}
	return Page{Page: page, Limit: limit}
}

type PagedResponse[T any] struct {
	Items []T `json:"items"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}
