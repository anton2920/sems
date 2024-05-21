package main

import (
	"testing"

	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/net/url"
)

func TestUserPageHandler(t *testing.T) {
	const endpoint = "/user/"

	expectedOK := [...]string{"0", "1", "2", "3"}

	expectedBadRequest := [...]string{"", "a", "b", "c"}

	expectedForbidden := expectedOK[:3]

	expectedNotFound := [...]string{"4", "5", "6"}

	t.Run("expectedOK", func(t *testing.T) {
		for i, id := range expectedOK {
			i := i
			id := id
			t.Run(id, func(t *testing.T) { t.Parallel(); testGetAuth(t, endpoint+id, testTokens[i], http.StatusOK) })
		}
	})

	for _, id := range expectedBadRequest {
		testGetAuth(t, endpoint+id, testTokens[0], http.StatusBadRequest)
	}

	testGet(t, endpoint, http.StatusUnauthorized)
	testGetAuth(t, endpoint, testInvalidToken, http.StatusUnauthorized)

	for _, id := range expectedForbidden {
		testGetAuth(t, endpoint+id, testTokens[3], http.StatusForbidden)
	}

	for _, id := range expectedNotFound {
		testGetAuth(t, endpoint+id, testTokens[0], http.StatusNotFound)
	}
}

func TestUserCreatePageHandler(t *testing.T) {
	const endpoint = "/user/create"

	testGetAuth(t, endpoint, testTokens[AdminID], http.StatusOK)

	testGet(t, endpoint, http.StatusUnauthorized)
	testGetAuth(t, endpoint, testInvalidToken, http.StatusUnauthorized)

	for i := AdminID + 1; i < len(DB.Users); i++ {
		testGetAuth(t, endpoint, testTokens[i], http.StatusForbidden)
	}
}

func TestUserEditPageHandler(t *testing.T) {
	const endpoint = "/user/edit"

	for i := 0; i < len(DB.Users); i++ {
		var vs url.Values
		vs.SetInt("ID", i)
		testPostAuth(t, endpoint, testTokens[i], vs, http.StatusOK)
	}

	testPost(t, endpoint, nil, http.StatusUnauthorized)
	testPostAuth(t, endpoint, testInvalidToken, nil, http.StatusUnauthorized)

	for i := AdminID + 1; i < len(DB.Users); i++ {
		testPostAuth(t, endpoint, testTokens[i], url.Values{{Key: "ID", Values: []string{"0"}}}, http.StatusForbidden)
	}
}
