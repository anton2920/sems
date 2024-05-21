package main

import (
	"testing"

	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/net/url"
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
	testPostAuth(t, endpoint, testTokens[AdminID], url.Values{{Key: "Name", Values: []string{"Test group"}}, {Key: "StudentID", Values: []string{"a"}}}, http.StatusOK)

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
		{{Key: "Name", Values: []string{"Test group"}}, {Key: "StudentID", Values: []string{"1", "2", "3"}}},
	}

	expectedBadRequest := [...]url.Values{
		{{Key: "Name", Values: []string{"Test"}}, {Key: "StudentID", Values: []string{"1", "2", "3"}}},
		{{Key: "Name", Values: []string{"TestTestTestTestTestTestTestTestTestTestTestT"}}, {Key: "StudentID", Values: []string{"1", "2", "3"}}},
		{{Key: "Name", Values: []string{"Test group"}}, {Key: "StudentID", Values: []string{"0"}}},
		{{Key: "Name", Values: []string{"Test group"}}, {Key: "StudentID", Values: []string{"a"}}},
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
		{{Key: "ID", Values: []string{"1"}}, {Key: "Name", Values: []string{"Test group"}}, {Key: "StudentID", Values: []string{"1", "2", "3"}}},
	}

	expectedBadRequest := [...]url.Values{
		{{Key: "ID", Values: []string{"a"}}, {Key: "Name", Values: []string{"Test group"}}, {Key: "StudentID", Values: []string{"1", "2", "3"}}},
		{{Key: "ID", Values: []string{"1"}}, {Key: "Name", Values: []string{"Test"}}, {Key: "StudentID", Values: []string{"1", "2", "3"}}},
		{{Key: "ID", Values: []string{"1"}}, {Key: "Name", Values: []string{"TestTestTestTestTestTestTestTestTestTestTestT"}}, {Key: "StudentID", Values: []string{"1", "2", "3"}}},
		{{Key: "ID", Values: []string{"1"}}, {Key: "Name", Values: []string{"Test group"}}, {Key: "StudentID", Values: []string{"0"}}},
		{{Key: "ID", Values: []string{"1"}}, {Key: "Name", Values: []string{"Test group"}}, {Key: "StudentID", Values: []string{"a"}}},
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
