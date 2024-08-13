package main

import (
	"strconv"

	"github.com/anton2920/gofa/database"
	"github.com/anton2920/gofa/errors"
	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/net/url"
	"github.com/anton2920/gofa/strings"
	"github.com/anton2920/gofa/trace"
)

func GetIDFromURL(l Language, u url.URL, prefix string) (database.ID, error) {
	defer trace.End(trace.Begin(""))

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
	defer trace.End(trace.Begin(""))

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
	defer trace.End(trace.Begin(""))

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
	defer trace.End(trace.Begin(""))

	i, err := strconv.Atoi(si)
	if err != nil {
		return -1, err
	}
	if (i < 0) || (i >= len) {
		return -1, errors.New("index out of range")
	}
	return i, nil
}
