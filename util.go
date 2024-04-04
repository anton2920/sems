package main

import (
	"fmt"
	"strconv"
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

func StringStartsWith(s, prefix string) bool {
	return (len(s) >= len(prefix)) && (s[:len(prefix)] == prefix)
}
