package main

import (
	"testing"

	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/net/url"
)

func TestUserNameValid(t *testing.T) {
	expectedOK := [...]string{
		/* Make sure basic English works. */
		"Admin",
		"Ivan",
		"Anton",
		"Sidorov",

		/* TODO(anton2920): make sure basic Russian works. */

		/* Make sure less common cases also work. */
		"St. Peter",
		"Mr. Smith",
		"Von Neumann",
		"Adelson-Velskii",
		"De'Wayne",
	}

	expectedError := [...]string{
		testString(MinNameLen - 1),
		testString(MaxNameLen + 1),
		"Anton1",
		" Anton",
	}

	for _, test := range expectedOK {
		if err := UserNameValid(GL, test); err != nil {
			t.Errorf("Expected name %q to be valid, but got error %v", test, err)
		}
	}

	for _, test := range expectedError {
		if err := UserNameValid(GL, test); err == nil {
			t.Errorf("Expected name %q to be invalid, but got no error", test)
		}
	}
}

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
			t.Run(id, func(t *testing.T) {
				t.Parallel()
				testGetAuth(t, endpoint+id, testTokens[i], http.StatusOK)
			})
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

	for _, token := range testTokens[AdminID+1:] {
		testGetAuth(t, endpoint, token, http.StatusForbidden)
	}
}

func TestUserEditPageHandler(t *testing.T) {
	const endpoint = "/user/edit"

	for i, token := range testTokens {
		var vs url.Values
		vs.SetInt("ID", i)
		testPostAuth(t, endpoint, token, vs, http.StatusOK)
	}

	testPostAuth(t, endpoint, testTokens[AdminID], url.Values{{"ID", []string{"a"}}}, http.StatusBadRequest)
	testPostInvalidFormAuth(t, endpoint, testTokens[AdminID])

	testPost(t, endpoint, nil, http.StatusUnauthorized)
	testPostAuth(t, endpoint, testInvalidToken, nil, http.StatusUnauthorized)

	for _, token := range testTokens[AdminID+1:] {
		testPostAuth(t, endpoint, token, url.Values{{"ID", []string{"0"}}}, http.StatusForbidden)
	}
}

func TestUserSigninPageHandler(t *testing.T) {
	testGet(t, "/user/signin", http.StatusOK)
}

func TestUserCreateHandler(t *testing.T) {
	const endpoint = APIPrefix + "/user/create"

	expectedOK := [...]url.Values{
		{{"FirstName", []string{"Test"}}, {"LastName", []string{"Testovich"}}, {"Email", []string{"test@masters.com"}}, {"Password", []string{"testtest"}}, {"RepeatPassword", []string{"testtest"}}},
	}

	expectedBadRequest := [...]url.Values{
		{{"FirstName", []string{testString(MinNameLen - 1)}}, {"LastName", []string{"Testovich"}}, {"Email", []string{"test@masters.com"}}, {"Password", []string{"testtest"}}, {"RepeatPassword", []string{"testtest"}}},
		{{"FirstName", []string{testString(MaxNameLen + 1)}}, {"LastName", []string{"Testovich"}}, {"Email", []string{"test@masters.com"}}, {"Password", []string{"testtest"}}, {"RepeatPassword", []string{"testtest"}}},
		{{"FirstName", []string{"Test"}}, {"LastName", []string{testString(MinNameLen - 1)}}, {"Email", []string{"test@masters.com"}}, {"Password", []string{"testtest"}}, {"RepeatPassword", []string{"testtest"}}},
		{{"FirstName", []string{"Test"}}, {"LastName", []string{testString(MaxNameLen + 1)}}, {"Email", []string{"test@masters.com"}}, {"Password", []string{"testtest"}}, {"RepeatPassword", []string{"testtest"}}},
		{{"FirstName", []string{"Test"}}, {"LastName", []string{"Testovich"}}, {"Email", []string{"testmasters.com"}}, {"Password", []string{"testtest"}}, {"RepeatPassword", []string{"testtest"}}},
		{{"FirstName", []string{"Test"}}, {"LastName", []string{"Testovich"}}, {"Email", []string{"test@masters.com"}}, {"Password", []string{testString(MinPasswordLen - 1)}}, {"RepeatPassword", []string{testString(MinPasswordLen - 1)}}},
		{{"FirstName", []string{"Test"}}, {"LastName", []string{"Testovich"}}, {"Email", []string{"test@masters.com"}}, {"Password", []string{testString(MaxPasswordLen + 1)}}, {"RepeatPassword", []string{testString(MaxPasswordLen + 1)}}},
		{{"FirstName", []string{"Test"}}, {"LastName", []string{"Testovich"}}, {"Email", []string{"test@masters.com"}}, {"Password", []string{"testtest"}}, {"RepeatPassword", []string{"testtesttest"}}},
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

	testCreateInitialDBs()

	expectedOK := [...]url.Values{
		{{"ID", []string{"2"}}, {"FirstName", []string{"Test"}}, {"LastName", []string{"Testovich"}}, {"Email", []string{"test@masters.com"}}, {"Password", []string{"testtest"}}, {"RepeatPassword", []string{"testtest"}}},
	}

	expectedBadRequest := [...]url.Values{
		{{"ID", []string{"a"}}, {"FirstName", []string{"Test"}}, {"LastName", []string{"Testovich"}}, {"Email", []string{"test@masters.com"}}, {"Password", []string{"testtest"}}, {"RepeatPassword", []string{"testtest"}}},
		{{"ID", []string{"0"}}, {"FirstName", []string{testString(MinNameLen - 1)}}, {"LastName", []string{"Testovich"}}, {"Email", []string{"test@masters.com"}}, {"Password", []string{"testtest"}}, {"RepeatPassword", []string{"testtest"}}},
		{{"ID", []string{"0"}}, {"FirstName", []string{testString(MaxNameLen + 1)}}, {"LastName", []string{"Testovich"}}, {"Email", []string{"test@masters.com"}}, {"Password", []string{"testtest"}}, {"RepeatPassword", []string{"testtest"}}},
		{{"ID", []string{"0"}}, {"FirstName", []string{"Test"}}, {"LastName", []string{testString(MinNameLen - 1)}}, {"Email", []string{"test@masters.com"}}, {"Password", []string{"testtest"}}, {"RepeatPassword", []string{"testtest"}}},
		{{"ID", []string{"0"}}, {"FirstName", []string{"Test"}}, {"LastName", []string{testString(MaxNameLen + 1)}}, {"Email", []string{"test@masters.com"}}, {"Password", []string{"testtest"}}, {"RepeatPassword", []string{"testtest"}}},
		{{"ID", []string{"0"}}, {"FirstName", []string{"Test"}}, {"LastName", []string{"Testovich"}}, {"Email", []string{"testmasters.com"}}, {"Password", []string{"testtest"}}, {"RepeatPassword", []string{"testtest"}}},
		{{"ID", []string{"0"}}, {"FirstName", []string{"Test"}}, {"LastName", []string{"Testovich"}}, {"Email", []string{"test@masters.com"}}, {"Password", []string{testString(MinPasswordLen - 1)}}, {"RepeatPassword", []string{testString(MinPasswordLen - 1)}}},
		{{"ID", []string{"0"}}, {"FirstName", []string{"Test"}}, {"LastName", []string{"Testovich"}}, {"Email", []string{"test@masters.com"}}, {"Password", []string{testString(MaxPasswordLen + 1)}}, {"RepeatPassword", []string{testString(MaxPasswordLen - 1)}}},
		{{"ID", []string{"0"}}, {"FirstName", []string{"Test"}}, {"LastName", []string{"Testovich"}}, {"Email", []string{"test@masters.com"}}, {"Password", []string{"testtest"}}, {"RepeatPassword", []string{"testtesttest"}}},
	}

	expectedForbidden := expectedOK

	expectedNotFound := [...]url.Values{
		{{"ID", []string{"5"}}, {"FirstName", []string{"Test"}}, {"LastName", []string{"Testovich"}}, {"Email", []string{"test@masters.com"}}, {"Password", []string{"testtest"}}, {"RepeatPassword", []string{"testtest"}}},
	}

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

	for _, test := range expectedNotFound {
		testPostAuth(t, endpoint, testTokens[AdminID], test, http.StatusNotFound)
	}

	for i, test := range expectedConflict {
		test.SetInt("ID", i)
		testPostAuth(t, endpoint, testTokens[AdminID], test, http.StatusConflict)
	}
}

func TestUserSigninHandler(t *testing.T) {
	const endpoint = APIPrefix + "/user/signin"

	testCreateInitialDBs()

	expectedOK := [...]url.Values{
		{{"Email", []string{"admin@masters.com"}}, {"Password", []string{"admin"}}},
		{{"Email", []string{"teacher@masters.com"}}, {"Password", []string{"teacher"}}},
		{{"Email", []string{"student@masters.com"}}, {"Password", []string{"student"}}},
		{{"Email", []string{"student2@masters.com"}}, {"Password", []string{"student2"}}},
	}

	expectedBadRequest := [...]url.Values{
		{{"Email", []string{"adminmasters.com"}}, {"Password", []string{"admin"}}},
	}

	expectedNotFound := [...]url.Values{
		{{"Email", []string{"uncle-bob@masters.com"}}, {"Password", []string{"uncle-bob"}}},
	}

	expectedConflict := [...]url.Values{
		{{"Email", []string{"admin@masters.com"}}, {"Password", []string{"not-admin"}}},
	}

	t.Run("expectedOK", func(t *testing.T) {
		for _, test := range expectedOK {
			test := test
			t.Run("", func(t *testing.T) {
				t.Parallel()
				testPost(t, endpoint, test, http.StatusSeeOther)
			})
		}
	})

	for _, test := range expectedBadRequest {
		testPost(t, endpoint, test, http.StatusBadRequest)
	}
	testPostInvalidFormAuth(t, endpoint, testTokens[AdminID])

	for _, test := range expectedNotFound {
		testPost(t, endpoint, test, http.StatusNotFound)
	}

	for _, test := range expectedConflict {
		testPost(t, endpoint, test, http.StatusConflict)
	}
}

func TestUserSignoutHandler(t *testing.T) {
	const endpoint = APIPrefix + "/user/signout"

	backup := make(map[string]*Session)
	for k, v := range Sessions {
		backup[k] = v
	}

	t.Run("expectedOK", func(t *testing.T) {
		for i, token := range testTokens {
			i := i
			token := token

			t.Run("", func(t *testing.T) {
				t.Parallel()

				testGetAuth(t, endpoint, token, http.StatusSeeOther)
				if _, err := GetSessionFromToken(token); err == nil {
					t.Errorf("User with ID=%d is still authorized", i)
				}
			})
		}
	})

	testGet(t, endpoint, http.StatusUnauthorized)
	testGetAuth(t, endpoint, testInvalidToken, http.StatusUnauthorized)

	for k, v := range backup {
		Sessions[k] = v
	}
}
