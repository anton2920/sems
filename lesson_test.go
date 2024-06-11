package main

import (
	"testing"

	"github.com/anton2920/gofa/net/http"
)

func TestLessonPageHandler(t *testing.T) {
	const endpoint = "/lesson/"

	CreateInitialDBs()

	expectedOK := [...]string{"0", "1", "2"}

	expectedBadRequest := [...]string{"a", "b", "c"}

	expectedForbidden := expectedOK[:2]

	expectedNotFound := [...]string{"3", "4", "5"}

	for _, id := range expectedOK {
		testGetAuth(t, endpoint+id, testTokens[AdminID], http.StatusOK)
	}
	testGetAuth(t, endpoint+"2", testTokens[2], http.StatusOK)

	for _, id := range expectedBadRequest {
		testGetAuth(t, endpoint+id, testTokens[AdminID], http.StatusBadRequest)
	}

	testGet(t, endpoint, http.StatusUnauthorized)
	testGetAuth(t, endpoint, testInvalidToken, http.StatusUnauthorized)

	for _, id := range expectedForbidden {
		testGetAuth(t, endpoint+id, testTokens[3], http.StatusForbidden)
	}

	for _, id := range expectedNotFound {
		testGetAuth(t, endpoint+id, testTokens[AdminID], http.StatusNotFound)
	}
}
