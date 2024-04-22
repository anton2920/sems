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

func GetIndicies(indicies string) (pindex int, spindex string, sindex int, ssindex string, err error) {
	if len(indicies) == 0 {
		return
	}

	spindex = indicies
	if i := FindChar(indicies, '.'); i != -1 {
		ssindex = indicies[i+1:]
		sindex, err = strconv.Atoi(ssindex)
		if err != nil {
			return
		}
		spindex = indicies[:i]
	}
	pindex, err = strconv.Atoi(spindex)
	return
}

func GetValidIndex[T any](vs URLValues, key string, ts []T) (int, error) {
	i, err := vs.GetInt(key)
	if err != nil {
		return 0, err
	}

	if (i < 0) || (i >= len(ts)) {
		return 0, NewError("slice index out of range")
	}

	return i, nil
}

func MoveDown[T any](vs []T, i int) {
	if (i >= 0) && (i < len(vs)-1) {
		vs[i], vs[i+1] = vs[i+1], vs[i]
	}
}

func MoveUp[T any](vs []T, i int) {
	if (i > 0) && (i <= len(vs)-1) {
		vs[i-1], vs[i] = vs[i], vs[i-1]
	}
}

func RemoveAtIndex[T any](ts []T, i int) []T {
	if (len(ts) == 0) || (i < 0) || (i >= len(ts)) {
		return ts
	}

	if i < len(ts)-1 {
		copy(ts[i:], ts[i+1:])
	}
	return ts[:len(ts)-1]
}

func StringLengthInRange(s string, min, max int) bool {
	return (utf8.RuneCountInString(s) >= min) && (utf8.RuneCountInString(s) <= max)
}

func StringStartsWith(s, prefix string) bool {
	return (len(s) >= len(prefix)) && (s[:len(prefix)] == prefix)
}
