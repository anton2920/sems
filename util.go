package main

import (
	"strconv"
	"unicode/utf8"

	"github.com/anton2920/gofa/database"
	"github.com/anton2920/gofa/errors"
	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/net/url"
	"github.com/anton2920/gofa/strings"
)

func DisplayShortenedString(w *http.Response, s string, maxVisibleLen int) {
	if utf8.RuneCountInString(s) < maxVisibleLen {
		w.WriteHTMLString(s)
	} else {
		space := strings.FindCharReverse(s[:maxVisibleLen], ' ')
		if space == -1 {
			w.WriteHTMLString(s[:maxVisibleLen])
		} else {
			w.WriteHTMLString(s[:space])
		}
		w.AppendString(`...`)
	}
}

func GetIDFromURL(l Language, u url.URL, prefix string) (database.ID, error) {
	if !strings.StartsWith(u.Path, prefix) {
		return 0, http.NotFound(Ls(l, "requested page does not exist"))
	}

	id, err := strconv.Atoi(u.Path[len(prefix):])
	if (err != nil) || (id < 0) || (id >= (1 << 31)) {
		return 0, http.BadRequest(Ls(l, "invalid ID for %q"), prefix)
	}

	return database.ID(id), nil
}

func GetIndicies(indicies string) (pindex int, spindex string, sindex int, ssindex string, err error) {
	if len(indicies) == 0 {
		return
	}

	spindex = indicies
	if i := strings.FindChar(indicies, '.'); i != -1 {
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

func GetValidID(si string, nextID database.ID) (database.ID, error) {
	id, err := strconv.Atoi(si)
	if err != nil {
		return -1, err
	}
	if (id < database.MinValidID) || (id >= int(nextID)) {
		return -1, errors.New("ID out of range")
	}
	return database.ID(id), nil
}

func GetValidIndex(si string, len int) (int, error) {
	i, err := strconv.Atoi(si)
	if err != nil {
		return -1, err
	}
	if (i < 0) || (i >= len) {
		return -1, errors.New("index out of range")
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
