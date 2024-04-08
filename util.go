package main

import (
	"fmt"
	"strconv"
	"unicode/utf8"
)

func GetIDFromURL(u URL, prefix string) (int, error) {
	path := u.Path

	if !StringStartsWith(path, prefix) {
		return 0, NotFoundError
	}

	id, err := strconv.Atoi(path[len(prefix):])
	if err != nil {
		return 0, NewHTTPError(HTTPStatusBadRequest, fmt.Sprintf("invalid ID for '%s'", prefix))
	}

	return id, nil
}

func StringLengthInRange(s string, min, max int) bool {
	return (utf8.RuneCountInString(s) >= min) && (utf8.RuneCountInString(s) <= max)
}

func StringStartsWith(s, prefix string) bool {
	return (len(s) >= len(prefix)) && (s[:len(prefix)] == prefix)
}

func WritePage(w *HTTPResponse, r *HTTPRequest, handler func(*HTTPResponse, *HTTPRequest) error, err HTTPError) error {
	r.Form.Set("Error", err.Message)
	w.StatusCode = err.StatusCode
	return handler(w, r)
}
