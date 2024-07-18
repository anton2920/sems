package main

import (
	"net/url"
	"testing"

	"github.com/anton2920/gofa/net/http"
)

func TestSubmissionNewPageHandler(t *testing.T) {
	const endpoint = "/submission/new"

	testCreateInitialDBs()
	testPostAuth(t, "/subject/lessons", testTokens[AdminID], url.Values{"ID": {"0"}, "CourseID": {"0"}, "Action": {Ls(GL, "give as is")}}, http.StatusSeeOther)

	expectedOK := [...]url.Values{
		/* Test page. */
		{"CurrentPage": {"Main"}, "Command0": {Ls(GL, "Pass")}},
		{"StepIndex": {"0"}, "CurrentPage": {"Test"}, "NextPage": {Ls(GL, "Discard")}},
		{"CurrentPage": {"Main"}, "Command0": {Ls(GL, "Pass")}},
		{"StepIndex": {"0"}, "CurrentPage": {"Test"}, "NextPage": {Ls(GL, "Save")}, "SelectedAnswer0": {"2"}, "SelectedAnswer1": {"0", "1"}, "SelectedAnswer2": {"3"}},
		{"CurrentPage": {"Main"}, "Command0": {Ls(GL, "Edit")}},
		{"StepIndex": {"0"}, "CurrentPage": {"Test"}, "NextPage": {Ls(GL, "Save")}, "SelectedAnswer0": {"2"}, "SelectedAnswer1": {"0", "1"}, "SelectedAnswer2": {"3"}},

		/* Programming page. */
		{"CurrentPage": {"Main"}, "Command1": {Ls(GL, "Pass")}},
		{"StepIndex": {"1"}, "CurrentPage": {"Programming"}, "NextPage": {Ls(GL, "Save")}, "LanguageID": {"4"}, "Solution": {"s = input()\nif s == \"aaa\":\n	print(\"bbb\")\nelif s == \"ccc\":\n	print(\"ddd\")\nelse:\n	print(\"eee\")\n"}},
		{"CurrentPage": {"Main"}, "Command1": {Ls(GL, "Edit")}},
		{"StepIndex": {"1"}, "CurrentPage": {"Programming"}, "NextPage": {Ls(GL, "Save")}, "LanguageID": {"4"}, "Solution": {"s = input()\nif s == \"aaa\":\n	print(\"bbb\")\nelif s == \"ccc\":\n	print(\"ddd\")\nelse:\n	print(\"eee\")\n"}},
	}

	expectedBadRequest := [...]url.Values{
		/* Test page. */
		{"ID": {"3"}, "SubmissionIndex": {"1"}, "StepIndex": {"a"}, "CurrentPage": {"Test"}, "NextPage": {Ls(GL, "Save")}, "SelectedAnswer0": {"2"}, "SelectedAnswer1": {"0", "1"}, "SelectedAnswer2": {"3"}},
		{"ID": {"3"}, "SubmissionIndex": {"1"}, "StepIndex": {"1"}, "CurrentPage": {"Test"}, "NextPage": {Ls(GL, "Save")}, "SelectedAnswer0": {"2"}, "SelectedAnswer1": {"0", "1"}, "SelectedAnswer2": {"3"}},
		{"ID": {"3"}, "SubmissionIndex": {"1"}, "StepIndex": {"0"}, "CurrentPage": {"Test"}, "NextPage": {Ls(GL, "Save")}, "SelectedAnswer0": {"4"}, "SelectedAnswer1": {"0", "1"}, "SelectedAnswer2": {"3"}},
		{"ID": {"3"}, "SubmissionIndex": {"1"}, "StepIndex": {"0"}, "CurrentPage": {"Test"}, "NextPage": {Ls(GL, "Save")}, "SelectedAnswer0": {"2", "3"}, "SelectedAnswer1": {"0", "1"}, "SelectedAnswer2": {"3"}},
		{"ID": {"3"}, "SubmissionIndex": {"1"}, "StepIndex": {"0"}, "CurrentPage": {"Test"}, "NextPage": {Ls(GL, "Save")}, "SelectedAnswer0": nil, "SelectedAnswer1": {"0", "1"}, "SelectedAnswer2": {"3"}},

		/* Programming page. */
		{"ID": {"3"}, "SubmissionIndex": {"1"}, "StepIndex": {"a"}, "CurrentPage": {"Programming"}, "NextPage": {Ls(GL, "Save")}, "LanguageID": {"4"}, "Solution": {"s = input()\nif s == \"aaa\":\n	print(\"bbb\")\nelif s == \"ccc\":\n	print(\"ddd\")\nelse:\n	print(\"eee\")\n"}},
		{"ID": {"3"}, "SubmissionIndex": {"1"}, "StepIndex": {"0"}, "CurrentPage": {"Programming"}, "NextPage": {Ls(GL, "Save")}, "LanguageID": {"4"}, "Solution": {"s = input()\nif s == \"aaa\":\n	print(\"bbb\")\nelif s == \"ccc\":\n	print(\"ddd\")\nelse:\n	print(\"eee\")\n"}},
		{"ID": {"3"}, "SubmissionIndex": {"1"}, "StepIndex": {"1"}, "CurrentPage": {"Programming"}, "NextPage": {Ls(GL, "Save")}, "LanguageID": {"a"}, "Solution": {"s = input()\nif s == \"aaa\":\n	print(\"bbb\")\nelif s == \"ccc\":\n	print(\"ddd\")\nelse:\n	print(\"eee\")\n"}},
		{"ID": {"3"}, "SubmissionIndex": {"1"}, "StepIndex": {"1"}, "CurrentPage": {"Programming"}, "NextPage": {Ls(GL, "Save")}, "LanguageID": {"4"}, "Solution": {testString(MinSolutionLen - 1)}},
		{"ID": {"3"}, "SubmissionIndex": {"1"}, "StepIndex": {"1"}, "CurrentPage": {"Programming"}, "NextPage": {Ls(GL, "Save")}, "LanguageID": {"4"}, "Solution": {testString(MaxSolutionLen + 1)}},
		{"ID": {"3"}, "SubmissionIndex": {"1"}, "StepIndex": {"1"}, "CurrentPage": {"Programming"}, "NextPage": {Ls(GL, "Save")}, "LanguageID": {"4"}, "Solution": {"= input()\nif s == \"aaa\":\n	print(\"bbb\")\nelif s == \"ccc\":\n	print(\"ddd\")\nelse:\n	print(\"eee\")\n"}},
		{"ID": {"3"}, "SubmissionIndex": {"1"}, "StepIndex": {"1"}, "CurrentPage": {"Programming"}, "NextPage": {Ls(GL, "Save")}, "LanguageID": {"4"}, "Solution": {"s = input()\nif s == \"aaa\":\n	print(\"aaa\")\nelif s == \"ccc\":\n	print(\"ddd\")\nelse:\n	print(\"eee\")\n"}},

		/* Main page. */
		{"ID": {"a"}},
		{"ID": {"3"}, "SubmissionIndex": {"a"}},
		{"ID": {"3"}, "SubmissionIndex": {"1"}, "CurrentPage": {"Main"}, "Commanda": {Ls(GL, "Pass")}},
		{"ID": {"3"}, "SubmissionIndex": {"1"}, "CurrentPage": {"Main"}, "Command2": {Ls(GL, "Pass")}},
		{"ID": {"3"}, "SubmissionIndex": {"1"}, "CurrentPage": {"Main"}, "Command": {Ls(GL, "Command")}},
		{"ID": {"3"}, "SubmissionIndex": {"1"}, "CurrentPage": {"Main"}, "Command": nil, "Command2": {Ls(GL, "Pass")}},
		{"ID": {"3"}, "SubmissionIndex": {"1"}, "NextPage": {Ls(GL, "Finish")}},
	}

	testPostAuth(t, endpoint, testTokens[2], url.Values{"ID": {"3"}}, http.StatusOK)
	for _, test := range expectedOK {
		test.Set("ID", "3")
		test.Set("SubmissionIndex", "0")
		testPostAuth(t, endpoint, testTokens[2], test, http.StatusOK)
	}
	testPostAuth(t, endpoint, testTokens[2], url.Values{"ID": {"3"}, "SubmissionIndex": {"0"}, "NextPage": {Ls(GL, "Finish")}}, http.StatusSeeOther)

	testPostAuth(t, endpoint, testTokens[3], url.Values{"ID": {"3"}}, http.StatusOK)
	testPostAuth(t, endpoint, testTokens[3], url.Values{"ID": {"3"}, "SubmissionIndex": {"1"}, "NextPage": {Ls(GL, "Finish")}}, http.StatusBadRequest)
	testPostAuth(t, endpoint, testTokens[3], url.Values{"ID": {"3"}, "SubmissionIndex": {"1"}, "CurrentPage": {"Main"}, "Command0": {Ls(GL, "Pass")}}, http.StatusOK)
	testPostAuth(t, endpoint, testTokens[3], url.Values{"ID": {"3"}, "SubmissionIndex": {"1"}, "CurrentPage": {"Main"}, "Command1": {Ls(GL, "Pass")}}, http.StatusOK)
	for _, test := range expectedBadRequest {
		testPostAuth(t, endpoint, testTokens[3], test, http.StatusBadRequest)
	}
	ProgrammingLanguages[4].Available = false
	testPostAuth(t, endpoint, testTokens[3], url.Values{"ID": {"3"}, "SubmissionIndex": {"1"}, "StepIndex": {"1"}, "CurrentPage": {"Programming"}, "NextPage": {Ls(GL, "Save")}, "LanguageID": {"4"}, "Solution": {"s = input()\nif s == \"aaa\":\n	print(\"bbb\")\nelif s == \"ccc\":\n	print(\"ddd\")\nelse:\n	print(\"eee\")\n"}}, http.StatusBadRequest)
	ProgrammingLanguages[4].Available = true
	testPostInvalidFormAuth(t, endpoint, testTokens[AdminID])

	testPost(t, endpoint, nil, http.StatusUnauthorized)
	testPostAuth(t, endpoint, testInvalidToken, nil, http.StatusUnauthorized)

	testPostAuth(t, endpoint, testTokens[1], url.Values{"ID": {"3"}}, http.StatusForbidden)
}
