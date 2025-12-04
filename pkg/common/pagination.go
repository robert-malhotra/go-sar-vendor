package common

import (
	"fmt"
	"iter"
	"net/url"
	"strconv"
)

// OffsetParams holds offset/limit pagination parameters.
type OffsetParams struct {
	Limit  int
	Offset int
}

// ToQuery converts pagination params to URL query values.
func (p OffsetParams) ToQuery(q url.Values) {
	if p.Limit > 0 {
		q.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.Offset > 0 {
		q.Set("offset", strconv.Itoa(p.Offset))
	}
}

// PageParams holds page-based pagination parameters.
type PageParams struct {
	Page  int
	Limit int
	Sort  string
}

// ToQuery converts pagination params to URL query values.
func (p PageParams) ToQuery(q url.Values) {
	if p.Page > 0 {
		q.Set("page", strconv.Itoa(p.Page))
	}
	if p.Limit > 0 {
		q.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.Sort != "" {
		q.Set("sort", p.Sort)
	}
}

// CursorParams holds cursor-based pagination parameters.
type CursorParams struct {
	Cursor string
	Limit  int
}

// ToQuery converts pagination params to URL query values.
func (p CursorParams) ToQuery(q url.Values) {
	if p.Cursor != "" {
		q.Set("cursor", p.Cursor)
	}
	if p.Limit > 0 {
		q.Set("limit", strconv.Itoa(p.Limit))
	}
}

// Paginate creates an iterator for cursor-based pagination.
// The fetch function should return a slice of items, the next cursor (or nil if done), and any error.
func Paginate[T any](fetch func(cursor *string) ([]T, *string, error)) iter.Seq2[[]T, error] {
	return func(yield func([]T, error) bool) {
		var cursor *string
		for {
			data, next, err := fetch(cursor)
			if !yield(data, err) {
				return
			}
			if err != nil || next == nil || *next == "" {
				return
			}
			cursor = next
		}
	}
}

// PaginateOffset creates an iterator for offset-based pagination.
// The fetch function should return a slice of items and any error.
// pageSize determines how many items to request per page.
func PaginateOffset[T any](pageSize int, fetch func(offset, limit int) ([]T, error)) iter.Seq2[[]T, error] {
	return func(yield func([]T, error) bool) {
		offset := 0
		for {
			data, err := fetch(offset, pageSize)
			if !yield(data, err) {
				return
			}
			if err != nil || len(data) < pageSize {
				return
			}
			offset += len(data)
		}
	}
}

// PaginatePage creates an iterator for page-based pagination.
// The fetch function should return a slice of items and any error.
// pageSize determines how many items to request per page.
func PaginatePage[T any](pageSize int, fetch func(page, limit int) ([]T, error)) iter.Seq2[[]T, error] {
	return func(yield func([]T, error) bool) {
		page := 1
		for {
			data, err := fetch(page, pageSize)
			if !yield(data, err) {
				return
			}
			if err != nil || len(data) < pageSize {
				return
			}
			page++
		}
	}
}

// AddPaginationParams adds common pagination parameters to a URL query.
func AddPaginationParams(q url.Values, limit, offset int) {
	if limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", limit))
	}
	if offset > 0 {
		q.Set("offset", fmt.Sprintf("%d", offset))
	}
}
