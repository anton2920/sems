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

	testPostInvalidFormAuth(t, endpoint, testTokens[AdminID])

	testPost(t, endpoint, nil, http.StatusUnauthorized)
	testPostAuth(t, endpoint, testInvalidToken, nil, http.StatusUnauthorized)

	for i := AdminID + 1; i < len(DB.Users); i++ {
		testPostAuth(t, endpoint, testTokens[i], url.Values{{Key: "ID", Values: []string{"0"}}}, http.StatusForbidden)
	}
}

func TestUserSigninPageHandler(t *testing.T) {
	testGet(t, "/user/signin", http.StatusOK)
}

func TestUserCreateHandler(t *testing.T) {
	const endpoint = APIPrefix + "/user/create"

	expectedOK := [...]url.Values{
		{{Key: "FirstName", Values: []string{"Test"}}, {Key: "LastName", Values: []string{"Testovich"}}, {Key: "Email", Values: []string{"test@masters.com"}}, {Key: "Password", Values: []string{"testtest"}}, {Key: "RepeatPassword", Values: []string{"testtest"}}},
	}

	expectedBadRequest := [...]url.Values{
		{{Key: "FirstName", Values: []string{""}}, {Key: "LastName", Values: []string{"Testovich"}}, {Key: "Email", Values: []string{"test@masters.com"}}, {Key: "Password", Values: []string{"testtest"}}, {Key: "RepeatPassword", Values: []string{"testtest"}}},
		{{Key: "FirstName", Values: []string{"TestTestTestTestTestTestTestTestTestTestTestTe"}}, {Key: "LastName", Values: []string{"Testovich"}}, {Key: "Email", Values: []string{"test@masters.com"}}, {Key: "Password", Values: []string{"testtest"}}, {Key: "RepeatPassword", Values: []string{"testtest"}}},
		{{Key: "FirstName", Values: []string{"Test"}}, {Key: "LastName", Values: []string{""}}, {Key: "Email", Values: []string{"test@masters.com"}}, {Key: "Password", Values: []string{"testtest"}}, {Key: "RepeatPassword", Values: []string{"testtest"}}},
		{{Key: "FirstName", Values: []string{"Test"}}, {Key: "LastName", Values: []string{"TestovichTestovichTestovichTestovichTestovichT"}}, {Key: "Email", Values: []string{"test@masters.com"}}, {Key: "Password", Values: []string{"testtest"}}, {Key: "RepeatPassword", Values: []string{"testtest"}}},
		{{Key: "FirstName", Values: []string{"Test"}}, {Key: "LastName", Values: []string{"Testovich"}}, {Key: "Email", Values: []string{"testmasters.com"}}, {Key: "Password", Values: []string{"testtest"}}, {Key: "RepeatPassword", Values: []string{"testtest"}}},
		{{Key: "FirstName", Values: []string{"Test"}}, {Key: "LastName", Values: []string{"Testovich"}}, {Key: "Email", Values: []string{"test@masters.com"}}, {Key: "Password", Values: []string{"test"}}, {Key: "RepeatPassword", Values: []string{"test"}}},
		{{Key: "FirstName", Values: []string{"Test"}}, {Key: "LastName", Values: []string{"Testovich"}}, {Key: "Email", Values: []string{"test@masters.com"}}, {Key: "Password", Values: []string{"testtesttesttesttesttesttesttesttesttesttestte"}}, {Key: "RepeatPassword", Values: []string{"testtesttesttesttesttesttesttesttesttesttestte"}}},
		{{Key: "FirstName", Values: []string{"Test"}}, {Key: "LastName", Values: []string{"Testovich"}}, {Key: "Email", Values: []string{"test@masters.com"}}, {Key: "Password", Values: []string{"testtest"}}, {Key: "RepeatPassword", Values: []string{"testtesttest"}}},
	}

	expectedForbidden := expectedOK

	expectedConflict := expectedOK

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

	for _, test := range expectedForbidden {
		testPostAuth(t, endpoint, testTokens[1], test, http.StatusForbidden)
	}

	for _, test := range expectedConflict {
		testPostAuth(t, endpoint, testTokens[AdminID], test, http.StatusConflict)
	}
}

func TestUserEditHandler(t *testing.T) {
	const endpoint = APIPrefix + "/user/edit"

	CreateInitialDB()

	expectedOK := [...]url.Values{
		{{Key: "ID", Values: []string{"2"}}, {Key: "FirstName", Values: []string{"Test"}}, {Key: "LastName", Values: []string{"Testovich"}}, {Key: "Email", Values: []string{"test@masters.com"}}, {Key: "Password", Values: []string{"testtest"}}, {Key: "RepeatPassword", Values: []string{"testtest"}}},
	}

	expectedBadRequest := [...]url.Values{
		{{Key: "ID", Values: []string{"0"}}, {Key: "FirstName", Values: []string{""}}, {Key: "LastName", Values: []string{"Testovich"}}, {Key: "Email", Values: []string{"test@masters.com"}}, {Key: "Password", Values: []string{"testtest"}}, {Key: "RepeatPassword", Values: []string{"testtest"}}},
		{{Key: "ID", Values: []string{"0"}}, {Key: "FirstName", Values: []string{"TestTestTestTestTestTestTestTestTestTestTestTe"}}, {Key: "LastName", Values: []string{"Testovich"}}, {Key: "Email", Values: []string{"test@masters.com"}}, {Key: "Password", Values: []string{"testtest"}}, {Key: "RepeatPassword", Values: []string{"testtest"}}},
		{{Key: "ID", Values: []string{"0"}}, {Key: "FirstName", Values: []string{"Test"}}, {Key: "LastName", Values: []string{""}}, {Key: "Email", Values: []string{"test@masters.com"}}, {Key: "Password", Values: []string{"testtest"}}, {Key: "RepeatPassword", Values: []string{"testtest"}}},
		{{Key: "ID", Values: []string{"0"}}, {Key: "FirstName", Values: []string{"Test"}}, {Key: "LastName", Values: []string{"TestovichTestovichTestovichTestovichTestovichT"}}, {Key: "Email", Values: []string{"test@masters.com"}}, {Key: "Password", Values: []string{"testtest"}}, {Key: "RepeatPassword", Values: []string{"testtest"}}},
		{{Key: "ID", Values: []string{"0"}}, {Key: "FirstName", Values: []string{"Test"}}, {Key: "LastName", Values: []string{"Testovich"}}, {Key: "Email", Values: []string{"testmasters.com"}}, {Key: "Password", Values: []string{"testtest"}}, {Key: "RepeatPassword", Values: []string{"testtest"}}},
		{{Key: "ID", Values: []string{"0"}}, {Key: "FirstName", Values: []string{"Test"}}, {Key: "LastName", Values: []string{"Testovich"}}, {Key: "Email", Values: []string{"test@masters.com"}}, {Key: "Password", Values: []string{"test"}}, {Key: "RepeatPassword", Values: []string{"test"}}},
		{{Key: "ID", Values: []string{"0"}}, {Key: "FirstName", Values: []string{"Test"}}, {Key: "LastName", Values: []string{"Testovich"}}, {Key: "Email", Values: []string{"test@masters.com"}}, {Key: "Password", Values: []string{"testtesttesttesttesttesttesttesttesttesttestte"}}, {Key: "RepeatPassword", Values: []string{"testtesttesttesttesttesttesttesttesttesttestte"}}},
		{{Key: "ID", Values: []string{"0"}}, {Key: "FirstName", Values: []string{"Test"}}, {Key: "LastName", Values: []string{"Testovich"}}, {Key: "Email", Values: []string{"test@masters.com"}}, {Key: "Password", Values: []string{"testtest"}}, {Key: "RepeatPassword", Values: []string{"testtesttest"}}},
	}

	expectedForbidden := expectedOK

	expectedConflict := expectedOK

	for _, test := range expectedOK {
		testPostAuth(t, endpoint, testTokens[AdminID], test, http.StatusSeeOther)
		testPostAuth(t, endpoint, testTokens[2], test, http.StatusSeeOther)
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

	for _, test := range expectedForbidden {
		testPostAuth(t, endpoint, testTokens[1], test, http.StatusForbidden)
	}

	for i, test := range expectedConflict {
		test.SetInt("ID", i)
		testPostAuth(t, endpoint, testTokens[AdminID], test, http.StatusConflict)
	}
}
