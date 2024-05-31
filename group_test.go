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
	testPostAuth(t, endpoint, testTokens[AdminID], url.Values{{"Name", []string{"Test group"}}, {"StudentID", []string{"a"}}}, http.StatusOK)

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
		{{"Name", []string{"Test group"}}, {"StudentID", []string{"1", "2", "3"}}},
	}

	expectedBadRequest := [...]url.Values{
		{{"Name", []string{testString(MinGroupNameLen - 1)}}, {"StudentID", []string{"1", "2", "3"}}},
		{{"Name", []string{testString(MaxGroupNameLen + 1)}}, {"StudentID", []string{"1", "2", "3"}}},
		{{"Name", []string{"Test group"}}},
		{{"Name", []string{"Test group"}}, {"StudentID", []string{"0"}}},
		{{"Name", []string{"Test group"}}, {"StudentID", []string{"a"}}},
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
		{{"ID", []string{"1"}}, {"Name", []string{"Test group"}}, {"StudentID", []string{"1", "2", "3"}}},
	}

	expectedBadRequest := [...]url.Values{
		{{"ID", []string{"a"}}, {"Name", []string{"Test group"}}, {"StudentID", []string{"1", "2", "3"}}},
		{{"ID", []string{"1"}}, {"Name", []string{testString(MinGroupNameLen - 1)}}, {"StudentID", []string{"1", "2", "3"}}},
		{{"ID", []string{"1"}}, {"Name", []string{testString(MaxGroupNameLen + 1)}}, {"StudentID", []string{"1", "2", "3"}}},
		{{"ID", []string{"1"}}, {"Name", []string{"Test group"}}},
		{{"ID", []string{"1"}}, {"Name", []string{"Test group"}}, {"StudentID", []string{"0"}}},
		{{"ID", []string{"1"}}, {"Name", []string{"Test group"}}, {"StudentID", []string{"a"}}},
	}

	expectedForbidden := expectedOK[0]

	expectedNotFound := [...]url.Values{
		{{"ID", []string{"2"}}, {"Name", []string{"Test group"}}, {"StudentID", []string{"1", "2", "3"}}},
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
