package main

import (
	"testing"

	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/net/url"
)

func TestSubjectPageHandler(t *testing.T) {
	const endpoint = "/subject/"

	expectedOK := [...]string{"0", "1"}

	expectedBadRequest := [...]string{"a", "b", "c"}

	expectedNotFound := [...]string{"2", "3", "4"}

	for i, id := range expectedOK {
		testGetAuth(t, endpoint+id, testTokens[i], http.StatusOK)
		testGetAuth(t, endpoint+id, testTokens[2], http.StatusOK)
		testGetAuth(t, endpoint+id, testTokens[3], http.StatusOK)
	}

	for _, id := range expectedBadRequest {
		testGetAuth(t, endpoint+id, testTokens[AdminID], http.StatusBadRequest)
	}

	testGet(t, endpoint, http.StatusUnauthorized)
	testGetAuth(t, endpoint, testInvalidToken, http.StatusUnauthorized)

	testGetAuth(t, endpoint+"0", testTokens[1], http.StatusForbidden)

	for _, id := range expectedNotFound {
		testGetAuth(t, endpoint+id, testTokens[AdminID], http.StatusNotFound)
	}
}

func TestSubjectCreatePageHandler(t *testing.T) {
	const endpoint = "/subject/create"

	expectedOK := [...]url.Values{
		{},
		{{"GroupID", []string{"0", "a"}}, {"TeacherID", []string{"0", "a"}}},
	}

	for _, test := range expectedOK {
		testPostAuth(t, endpoint, testTokens[AdminID], test, http.StatusOK)
	}

	testPostInvalidFormAuth(t, endpoint, testTokens[AdminID])

	testPost(t, endpoint, nil, http.StatusUnauthorized)
	testPostAuth(t, endpoint, testInvalidToken, nil, http.StatusUnauthorized)

	testPostAuth(t, endpoint, testTokens[1], nil, http.StatusForbidden)
}

func TestSubjectEditPageHandler(t *testing.T) {
	const endpoint = "/subject/edit"

	expectedOK := [...]url.Values{
		{{"ID", []string{"0"}}},
		{{"ID", []string{"0"}}, {"GroupID", []string{"0", "a"}}, {"TeacherID", []string{"0", "a"}}},
	}

	for _, test := range expectedOK {
		testPostAuth(t, endpoint, testTokens[AdminID], test, http.StatusOK)
	}

	testPostInvalidFormAuth(t, endpoint, testTokens[AdminID])

	testPost(t, endpoint, nil, http.StatusUnauthorized)
	testPostAuth(t, endpoint, testInvalidToken, nil, http.StatusUnauthorized)

	testPostAuth(t, endpoint, testTokens[1], nil, http.StatusForbidden)
}
