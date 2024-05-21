package main

import (
	"testing"

	"github.com/anton2920/gofa/net/http"
)

func TestGroupPageHandler(t *testing.T) {
	const endpoint = "/group/"

	expectedBadRequest := [...]string{"", "a", "b", "c"}

	expectedNotFound := [...]string{"1", "2", "3"}

	testGetAuth(t, endpoint+"0", testTokens[AdminID], http.StatusOK)
	for _, token := range testTokens[2:] {
		testGetAuth(t, endpoint+"0", token, http.StatusOK)
	}

	for _, id := range expectedBadRequest {
		testGetAuth(t, endpoint+id, testTokens[0], http.StatusBadRequest)
	}

	testGet(t, endpoint, http.StatusUnauthorized)
	testGetAuth(t, endpoint, testInvalidToken, http.StatusUnauthorized)

	for _, token := range testTokens[1:2] {
		testGetAuth(t, endpoint+"0", token, http.StatusForbidden)
	}

	for _, id := range expectedNotFound {
		testGetAuth(t, endpoint+id, testTokens[0], http.StatusNotFound)
	}
}

func testGroupCreateEditPageHandler(t *testing.T, endpoint string) {
	testGetAuth(t, endpoint, testTokens[AdminID], http.StatusOK)

	testGet(t, endpoint, http.StatusUnauthorized)
	testGetAuth(t, endpoint, testInvalidToken, http.StatusUnauthorized)

	for _, token := range testTokens[AdminID+1:] {
		testGetAuth(t, endpoint, token, http.StatusForbidden)
	}
}

func TestGroupCreatePageHandler(t *testing.T) {
	testGroupCreateEditPageHandler(t, "/group/create")
}

func TestGroupEditPageHandler(t *testing.T) {
	testGroupCreateEditPageHandler(t, "/group/edit")
}
