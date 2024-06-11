package main

import (
	"testing"

	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/net/url"
)

func TestSubjectPageHandler(t *testing.T) {
	const endpoint = "/subject/"

	CreateInitialDBs()

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

func TestSubjectCreateHandler(t *testing.T) {
	const endpoint = APIPrefix + "/subject/create"

	expectedOK := [...]url.Values{
		{{"Name", []string{"Chemistry"}}, {"TeacherID", []string{"1"}}, {"GroupID", []string{"0"}}},
	}

	expectedBadRequest := [...]url.Values{
		{{"Name", []string{testString(MinNameLen - 1)}}, {"TeacherID", []string{"1"}}, {"GroupID", []string{"0"}}},
		{{"Name", []string{testString(MaxNameLen + 1)}}, {"TeacherID", []string{"1"}}, {"GroupID", []string{"0"}}},
		{{"Name", []string{"Chemistry"}}, {"TeacherID", []string{"a"}}, {"GroupID", []string{"0"}}},
		{{"Name", []string{"Chemistry"}}, {"TeacherID", []string{"10"}}, {"GroupID", []string{"0"}}},
		{{"Name", []string{"Chemistry"}}, {"TeacherID", []string{"1"}}, {"GroupID", []string{"a"}}},
		{{"Name", []string{"Chemistry"}}, {"TeacherID", []string{"1"}}, {"GroupID", []string{"10"}}},
	}

	expectedForbidden := expectedOK

	for _, test := range expectedOK {
		testPostAuth(t, endpoint, testTokens[AdminID], test, http.StatusSeeOther)
	}

	for _, test := range expectedBadRequest {
		testPostAuth(t, endpoint, testTokens[AdminID], test, http.StatusBadRequest)
	}

	testPostInvalidFormAuth(t, endpoint, testTokens[AdminID])

	testPost(t, endpoint, nil, http.StatusUnauthorized)
	testPostAuth(t, endpoint, testInvalidToken, nil, http.StatusUnauthorized)

	for _, test := range expectedForbidden {
		testPostAuth(t, endpoint, testTokens[1], test, http.StatusForbidden)
	}
}

func TestSubjectEditHandler(t *testing.T) {
	const endpoint = APIPrefix + "/subject/edit"

	expectedOK := [...]url.Values{
		{{"ID", []string{"0"}}, {"Name", []string{"Chemistry"}}, {"TeacherID", []string{"1"}}, {"GroupID", []string{"0"}}},
	}

	expectedBadRequest := [...]url.Values{
		{{"ID", []string{"a"}}, {"Name", []string{testString(MinNameLen - 1)}}, {"TeacherID", []string{"1"}}, {"GroupID", []string{"0"}}},
		{{"ID", []string{"0"}}, {"Name", []string{testString(MinNameLen - 1)}}, {"TeacherID", []string{"1"}}, {"GroupID", []string{"0"}}},
		{{"ID", []string{"0"}}, {"Name", []string{testString(MaxNameLen + 1)}}, {"TeacherID", []string{"1"}}, {"GroupID", []string{"0"}}},
		{{"ID", []string{"0"}}, {"Name", []string{"Chemistry"}}, {"TeacherID", []string{"a"}}, {"GroupID", []string{"0"}}},
		{{"ID", []string{"0"}}, {"Name", []string{"Chemistry"}}, {"TeacherID", []string{"10"}}, {"GroupID", []string{"0"}}},
		{{"ID", []string{"0"}}, {"Name", []string{"Chemistry"}}, {"TeacherID", []string{"1"}}, {"GroupID", []string{"a"}}},
		{{"ID", []string{"0"}}, {"Name", []string{"Chemistry"}}, {"TeacherID", []string{"1"}}, {"GroupID", []string{"10"}}},
	}

	expectedForbidden := expectedOK

	expectedNotFound := [...]url.Values{
		{{"ID", []string{"3"}}, {"Name", []string{"Chemistry"}}, {"TeacherID", []string{"1"}}, {"GroupID", []string{"0"}}},
	}

	for _, test := range expectedOK {
		testPostAuth(t, endpoint, testTokens[AdminID], test, http.StatusSeeOther)
	}

	for _, test := range expectedBadRequest {
		testPostAuth(t, endpoint, testTokens[AdminID], test, http.StatusBadRequest)
	}

	testPostInvalidFormAuth(t, endpoint, testTokens[AdminID])

	testPost(t, endpoint, nil, http.StatusUnauthorized)
	testPostAuth(t, endpoint, testInvalidToken, nil, http.StatusUnauthorized)

	for _, test := range expectedForbidden {
		testPostAuth(t, endpoint, testTokens[1], test, http.StatusForbidden)
	}

	for _, test := range expectedNotFound {
		testPostAuth(t, endpoint, testTokens[AdminID], test, http.StatusNotFound)
	}
}
