package main

import (
	"testing"

	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/net/url"
)

func TestUser2(t *testing.T) {
	var user2, user3 User2
	var user User

	user = DB.Users[0]

	user2.FirstName = user.FirstName
	user2.LastName = user.LastName
	user2.Email = user.Email
	user2.Password = user.Password
	user2.Courses = []int32{0}
	user2.CreatedOn = user.CreatedOn.Unix()

	if err := CreateUser(&DB2, &user2); err != nil {
		t.Fatal(err)
	}

	if err := GetUserByID(&DB2, 0, &user3); err != nil {
		t.Fatal(err)
	}

	if user2.FirstName != user3.FirstName {
		t.Errorf("First name: %s != %s", user2.FirstName, user3.FirstName)
	}
	if user2.LastName != user3.LastName {
		t.Errorf("Last name: %s != %s", user2.LastName, user3.LastName)
	}
	if user2.Email != user3.Email {
		t.Errorf("Email: %s != %s", user2.Email, user3.Email)
	}
	if user2.Password != user3.Password {
		t.Errorf("Password: %s != %s", user2.Password, user3.Password)
	}
	if user2.Courses[0] != user3.Courses[0] {
		t.Errorf("Courses[0]: %d != %d", user2.Courses[0], user3.Courses[0])
	}
	if user2.CreatedOn != user3.CreatedOn {
		t.Errorf("CreatedOn: %d != %d", user2.CreatedOn, user3.CreatedOn)
	}

	if err := GetUserByID(&DB2, 100, &user3); err == nil {
		t.Fatal("expected error, got nothing")
	}

	if err := DeleteUserByID(&DB2, 0); err != nil {
		t.Fatal(err)
	}
}

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
		"",
		"AntonAntonAntonAntonAntonAntonAntonAntonAntonA",
		"Anton1",
		" Anton",
	}

	for _, test := range expectedOK {
		if err := UserNameValid(test); err != nil {
			t.Errorf("Expected name %q to be valid, but got error %v", test, err)
		}
	}

	for _, test := range expectedError {
		if err := UserNameValid(test); err == nil {
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
		{{"FirstName", []string{""}}, {"LastName", []string{"Testovich"}}, {"Email", []string{"test@masters.com"}}, {"Password", []string{"testtest"}}, {"RepeatPassword", []string{"testtest"}}},
		{{"FirstName", []string{"TestTestTestTestTestTestTestTestTestTestTestTe"}}, {"LastName", []string{"Testovich"}}, {"Email", []string{"test@masters.com"}}, {"Password", []string{"testtest"}}, {"RepeatPassword", []string{"testtest"}}},
		{{"FirstName", []string{"Test"}}, {"LastName", []string{""}}, {"Email", []string{"test@masters.com"}}, {"Password", []string{"testtest"}}, {"RepeatPassword", []string{"testtest"}}},
		{{"FirstName", []string{"Test"}}, {"LastName", []string{"TestovichTestovichTestovichTestovichTestovichT"}}, {"Email", []string{"test@masters.com"}}, {"Password", []string{"testtest"}}, {"RepeatPassword", []string{"testtest"}}},
		{{"FirstName", []string{"Test"}}, {"LastName", []string{"Testovich"}}, {"Email", []string{"testmasters.com"}}, {"Password", []string{"testtest"}}, {"RepeatPassword", []string{"testtest"}}},
		{{"FirstName", []string{"Test"}}, {"LastName", []string{"Testovich"}}, {"Email", []string{"test@masters.com"}}, {"Password", []string{"test"}}, {"RepeatPassword", []string{"test"}}},
		{{"FirstName", []string{"Test"}}, {"LastName", []string{"Testovich"}}, {"Email", []string{"test@masters.com"}}, {"Password", []string{"testtesttesttesttesttesttesttesttesttesttestte"}}, {"RepeatPassword", []string{"testtesttesttesttesttesttesttesttesttesttestte"}}},
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

	CreateInitialDB()

	expectedOK := [...]url.Values{
		{{"ID", []string{"2"}}, {"FirstName", []string{"Test"}}, {"LastName", []string{"Testovich"}}, {"Email", []string{"test@masters.com"}}, {"Password", []string{"testtest"}}, {"RepeatPassword", []string{"testtest"}}},
	}

	expectedBadRequest := [...]url.Values{
		{{"ID", []string{"a"}}, {"FirstName", []string{"Test"}}, {"LastName", []string{"Testovich"}}, {"Email", []string{"test@masters.com"}}, {"Password", []string{"testtest"}}, {"RepeatPassword", []string{"testtest"}}},
		{{"ID", []string{"0"}}, {"FirstName", []string{""}}, {"LastName", []string{"Testovich"}}, {"Email", []string{"test@masters.com"}}, {"Password", []string{"testtest"}}, {"RepeatPassword", []string{"testtest"}}},
		{{"ID", []string{"0"}}, {"FirstName", []string{"TestTestTestTestTestTestTestTestTestTestTestTe"}}, {"LastName", []string{"Testovich"}}, {"Email", []string{"test@masters.com"}}, {"Password", []string{"testtest"}}, {"RepeatPassword", []string{"testtest"}}},
		{{"ID", []string{"0"}}, {"FirstName", []string{"Test"}}, {"LastName", []string{""}}, {"Email", []string{"test@masters.com"}}, {"Password", []string{"testtest"}}, {"RepeatPassword", []string{"testtest"}}},
		{{"ID", []string{"0"}}, {"FirstName", []string{"Test"}}, {"LastName", []string{"TestovichTestovichTestovichTestovichTestovichT"}}, {"Email", []string{"test@masters.com"}}, {"Password", []string{"testtest"}}, {"RepeatPassword", []string{"testtest"}}},
		{{"ID", []string{"0"}}, {"FirstName", []string{"Test"}}, {"LastName", []string{"Testovich"}}, {"Email", []string{"testmasters.com"}}, {"Password", []string{"testtest"}}, {"RepeatPassword", []string{"testtest"}}},
		{{"ID", []string{"0"}}, {"FirstName", []string{"Test"}}, {"LastName", []string{"Testovich"}}, {"Email", []string{"test@masters.com"}}, {"Password", []string{"test"}}, {"RepeatPassword", []string{"test"}}},
		{{"ID", []string{"0"}}, {"FirstName", []string{"Test"}}, {"LastName", []string{"Testovich"}}, {"Email", []string{"test@masters.com"}}, {"Password", []string{"testtesttesttesttesttesttesttesttesttesttestte"}}, {"RepeatPassword", []string{"testtesttesttesttesttesttesttesttesttesttestte"}}},
		{{"ID", []string{"0"}}, {"FirstName", []string{"Test"}}, {"LastName", []string{"Testovich"}}, {"Email", []string{"test@masters.com"}}, {"Password", []string{"testtest"}}, {"RepeatPassword", []string{"testtesttest"}}},
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

func TestUserSigninHandler(t *testing.T) {
	const endpoint = APIPrefix + "/user/signin"

	CreateInitialDB()

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
