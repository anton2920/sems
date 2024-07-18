package main

import (
	"net/url"
	"strconv"
	"testing"

	"github.com/anton2920/gofa/net/http"
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
		testPostAuth(t, endpoint, token, url.Values{"ID": {strconv.Itoa(i)}}, http.StatusOK)
	}

	testPostAuth(t, endpoint, testTokens[AdminID], url.Values{"ID": {"a"}}, http.StatusBadRequest)
	testPostInvalidFormAuth(t, endpoint, testTokens[AdminID])

	testPost(t, endpoint, nil, http.StatusUnauthorized)
	testPostAuth(t, endpoint, testInvalidToken, nil, http.StatusUnauthorized)

	for _, token := range testTokens[AdminID+1:] {
		testPostAuth(t, endpoint, token, url.Values{"ID": {"0"}}, http.StatusForbidden)
	}
}

func TestUserSigninPageHandler(t *testing.T) {
	testGet(t, "/user/signin", http.StatusOK)
}

func TestUserCreateHandler(t *testing.T) {
	const endpoint = APIPrefix + "/user/create"

	expectedOK := [...]url.Values{
		{"FirstName": {"Test"}, "LastName": {"Testovich"}, "Email": {"test@masters.com"}, "Password": {"testtest"}, "RepeatPassword": {"testtest"}},
	}

	expectedBadRequest := [...]url.Values{
		{"FirstName": {testString(MinNameLen - 1)}, "LastName": {"Testovich"}, "Email": {"test@masters.com"}, "Password": {"testtest"}, "RepeatPassword": {"testtest"}},
		{"FirstName": {testString(MaxNameLen + 1)}, "LastName": {"Testovich"}, "Email": {"test@masters.com"}, "Password": {"testtest"}, "RepeatPassword": {"testtest"}},
		{"FirstName": {"Test"}, "LastName": {testString(MinNameLen - 1)}, "Email": {"test@masters.com"}, "Password": {"testtest"}, "RepeatPassword": {"testtest"}},
		{"FirstName": {"Test"}, "LastName": {testString(MaxNameLen + 1)}, "Email": {"test@masters.com"}, "Password": {"testtest"}, "RepeatPassword": {"testtest"}},
		{"FirstName": {"Test"}, "LastName": {"Testovich"}, "Email": {"testmasters.com"}, "Password": {"testtest"}, "RepeatPassword": {"testtest"}},
		{"FirstName": {"Test"}, "LastName": {"Testovich"}, "Email": {"test@masters.com"}, "Password": {testString(MinPasswordLen - 1)}, "RepeatPassword": {testString(MinPasswordLen - 1)}},
		{"FirstName": {"Test"}, "LastName": {"Testovich"}, "Email": {"test@masters.com"}, "Password": {testString(MaxPasswordLen + 1)}, "RepeatPassword": {testString(MaxPasswordLen + 1)}},
		{"FirstName": {"Test"}, "LastName": {"Testovich"}, "Email": {"test@masters.com"}, "Password": {"testtest"}, "RepeatPassword": {"testtesttest"}},
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
		{"ID": {"2"}, "FirstName": {"Test"}, "LastName": {"Testovich"}, "Email": {"test@masters.com"}, "Password": {"testtest"}, "RepeatPassword": {"testtest"}},
	}

	expectedBadRequest := [...]url.Values{
		{"ID": {"a"}, "FirstName": {"Test"}, "LastName": {"Testovich"}, "Email": {"test@masters.com"}, "Password": {"testtest"}, "RepeatPassword": {"testtest"}},
		{"ID": {"0"}, "FirstName": {testString(MinNameLen - 1)}, "LastName": {"Testovich"}, "Email": {"test@masters.com"}, "Password": {"testtest"}, "RepeatPassword": {"testtest"}},
		{"ID": {"0"}, "FirstName": {testString(MaxNameLen + 1)}, "LastName": {"Testovich"}, "Email": {"test@masters.com"}, "Password": {"testtest"}, "RepeatPassword": {"testtest"}},
		{"ID": {"0"}, "FirstName": {"Test"}, "LastName": {testString(MinNameLen - 1)}, "Email": {"test@masters.com"}, "Password": {"testtest"}, "RepeatPassword": {"testtest"}},
		{"ID": {"0"}, "FirstName": {"Test"}, "LastName": {testString(MaxNameLen + 1)}, "Email": {"test@masters.com"}, "Password": {"testtest"}, "RepeatPassword": {"testtest"}},
		{"ID": {"0"}, "FirstName": {"Test"}, "LastName": {"Testovich"}, "Email": {"testmasters.com"}, "Password": {"testtest"}, "RepeatPassword": {"testtest"}},
		{"ID": {"0"}, "FirstName": {"Test"}, "LastName": {"Testovich"}, "Email": {"test@masters.com"}, "Password": {testString(MinPasswordLen - 1)}, "RepeatPassword": {testString(MinPasswordLen - 1)}},
		{"ID": {"0"}, "FirstName": {"Test"}, "LastName": {"Testovich"}, "Email": {"test@masters.com"}, "Password": {testString(MaxPasswordLen + 1)}, "RepeatPassword": {testString(MaxPasswordLen - 1)}},
		{"ID": {"0"}, "FirstName": {"Test"}, "LastName": {"Testovich"}, "Email": {"test@masters.com"}, "Password": {"testtest"}, "RepeatPassword": {"testtesttest"}},
	}

	expectedForbidden := expectedOK

	expectedNotFound := [...]url.Values{
		{"ID": {"5"}, "FirstName": {"Test"}, "LastName": {"Testovich"}, "Email": {"test@masters.com"}, "Password": {"testtest"}, "RepeatPassword": {"testtest"}},
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
		test.Set("ID", strconv.Itoa(i))
		testPostAuth(t, endpoint, testTokens[AdminID], test, http.StatusConflict)
	}
}

func TestUserSigninHandler(t *testing.T) {
	const endpoint = APIPrefix + "/user/signin"

	testCreateInitialDBs()

	expectedOK := [...]url.Values{
		{"Email": {"admin@masters.com"}, "Password": {"admin"}},
		{"Email": {"teacher@masters.com"}, "Password": {"teacher"}},
		{"Email": {"student@masters.com"}, "Password": {"student"}},
		{"Email": {"student2@masters.com"}, "Password": {"student2"}},
	}

	expectedBadRequest := [...]url.Values{
		{"Email": {"adminmasters.com"}, "Password": {"admin"}},
	}

	expectedNotFound := [...]url.Values{
		{"Email": {"uncle-bob@masters.com"}, "Password": {"uncle-bob"}},
	}

	expectedConflict := [...]url.Values{
		{"Email": {"admin@masters.com"}, "Password": {"not-admin"}},
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
