package main

import (
	"net/url"
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
	testPostAuth(t, endpoint, testTokens[AdminID], url.Values{"ID": {"0"}, "Name": {"Test group"}, "StudentID": {"a"}}, http.StatusOK)

	testPostInvalidFormAuth(t, endpoint, testTokens[AdminID])

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

func TestGroupCreateHandler(t *testing.T) {
	const endpoint = APIPrefix + "/group/create"

	expectedOK := [...]url.Values{
		{"Name": {"Test group"}, "StudentID": {"1", "2", "3"}},
	}

	expectedBadRequest := [...]url.Values{
		{"Name": {testString(MinGroupNameLen - 1)}, "StudentID": {"1", "2", "3"}},
		{"Name": {testString(MaxGroupNameLen + 1)}, "StudentID": {"1", "2", "3"}},
		{"Name": {"Test group"}},
		{"Name": {"Test group"}, "StudentID": {"0"}},
		{"Name": {"Test group"}, "StudentID": {"a"}},
	}

	expectedForbidden := expectedOK[0]

	for _, test := range expectedOK {
		testPostAuth(t, endpoint, testTokens[AdminID], test, http.StatusSeeOther)
	}

	t.Run("expectedBadRequest", func(t *testing.T) {
		for _, test := range expectedBadRequest {
			test := test
			t.Run("", func(t *testing.T) {
				t.Parallel()
				testPostAuth(t, endpoint, testTokens[AdminID], test, http.StatusBadRequest)
			})
		}
	})
	testPostInvalidFormAuth(t, endpoint, testTokens[AdminID])

	testPost(t, endpoint, nil, http.StatusUnauthorized)
	testPostAuth(t, endpoint, testInvalidToken, nil, http.StatusUnauthorized)

	testPostAuth(t, endpoint, testTokens[1], expectedForbidden, http.StatusForbidden)
}

func TestGroupEditHandler(t *testing.T) {
	const endpoint = APIPrefix + "/group/edit"

	expectedOK := [...]url.Values{
		{"ID": {"1"}, "Name": {"Test group"}, "StudentID": {"1", "2", "3"}},
	}

	expectedBadRequest := [...]url.Values{
		{"ID": {"a"}, "Name": {"Test group"}, "StudentID": {"1", "2", "3"}},
		{"ID": {"1"}, "Name": {testString(MinGroupNameLen - 1)}, "StudentID": {"1", "2", "3"}},
		{"ID": {"1"}, "Name": {testString(MaxGroupNameLen + 1)}, "StudentID": {"1", "2", "3"}},
		{"ID": {"1"}, "Name": {"Test group"}},
		{"ID": {"1"}, "Name": {"Test group"}, "StudentID": {"0"}},
		{"ID": {"1"}, "Name": {"Test group"}, "StudentID": {"a"}},
	}

	expectedForbidden := expectedOK[0]

	expectedNotFound := [...]url.Values{
		{"ID": {"2"}, "Name": {"Test group"}, "StudentID": {"1", "2", "3"}},
	}

	for _, test := range expectedOK {
		testPostAuth(t, endpoint, testTokens[AdminID], test, http.StatusSeeOther)
	}

	t.Run("expectedBadRequest", func(t *testing.T) {
		for _, test := range expectedBadRequest {
			test := test
			t.Run("", func(t *testing.T) {
				t.Parallel()
				testPostAuth(t, endpoint, testTokens[AdminID], test, http.StatusBadRequest)
			})
		}
	})
	testPostInvalidFormAuth(t, endpoint, testTokens[AdminID])

	testPost(t, endpoint, nil, http.StatusUnauthorized)
	testPostAuth(t, endpoint, testInvalidToken, nil, http.StatusUnauthorized)

	testPostAuth(t, endpoint, testTokens[1], expectedForbidden, http.StatusForbidden)

	for _, test := range expectedNotFound {
		testPostAuth(t, endpoint, testTokens[AdminID], test, http.StatusNotFound)
	}
}
