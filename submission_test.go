package main

import (
	"testing"

	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/net/url"
)

func TestSubmissionNewPageHandler(t *testing.T) {
	const endpoint = "/submission/new"

	CreateInitialDBs()
	testPostAuth(t, "/subject/lessons", testTokens[AdminID], url.Values{{"ID", []string{"0"}}, {"CourseID", []string{"0"}}, {"Action", []string{"give as is"}}}, http.StatusSeeOther)

	expectedOK := [...]url.Values{
		/* Test page. */
		{{"CurrentPage", []string{"Main"}}, {"Command0", []string{"Pass"}}},
		{{"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"NextPage", []string{"Discard"}}},
		{{"CurrentPage", []string{"Main"}}, {"Command0", []string{"Pass"}}},
		{{"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"NextPage", []string{"Next"}}, {"SelectedAnswer0", []string{"2"}}, {"SelectedAnswer1", []string{"0", "1"}}, {"SelectedAnswer2", []string{"3"}}},
		{{"CurrentPage", []string{"Main"}}, {"Command0", []string{"Edit"}}},
		{{"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"NextPage", []string{"Next"}}, {"SelectedAnswer0", []string{"2"}}, {"SelectedAnswer1", []string{"0", "1"}}, {"SelectedAnswer2", []string{"3"}}},

		/* Programming page. */
		{{"CurrentPage", []string{"Main"}}, {"Command1", []string{"Pass"}}},
		{{"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"NextPage", []string{"Next"}}, {"LanguageID", []string{"4"}}, {"Solution", []string{"s = input()\nif s == \"aaa\":\n	print(\"bbb\")\nelif s == \"ccc\":\n	print(\"ddd\")\nelse:\n	print(\"eee\")\n"}}},
		{{"CurrentPage", []string{"Main"}}, {"Command1", []string{"Edit"}}},
		{{"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"NextPage", []string{"Next"}}, {"LanguageID", []string{"4"}}, {"Solution", []string{"s = input()\nif s == \"aaa\":\n	print(\"bbb\")\nelif s == \"ccc\":\n	print(\"ddd\")\nelse:\n	print(\"eee\")\n"}}},
	}

	expectedBadRequest := [...]url.Values{
		/* Test page. */
		{{"ID", []string{"3"}}, {"SubmissionIndex", []string{"1"}}, {"StepIndex", []string{"a"}}, {"CurrentPage", []string{"Test"}}, {"NextPage", []string{"Next"}}, {"SelectedAnswer0", []string{"2"}}, {"SelectedAnswer1", []string{"0", "1"}}, {"SelectedAnswer2", []string{"3"}}},
		{{"ID", []string{"3"}}, {"SubmissionIndex", []string{"1"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Test"}}, {"NextPage", []string{"Next"}}, {"SelectedAnswer0", []string{"2"}}, {"SelectedAnswer1", []string{"0", "1"}}, {"SelectedAnswer2", []string{"3"}}},
		{{"ID", []string{"3"}}, {"SubmissionIndex", []string{"1"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"NextPage", []string{"Next"}}, {"SelectedAnswer0", []string{"4"}}, {"SelectedAnswer1", []string{"0", "1"}}, {"SelectedAnswer2", []string{"3"}}},
		{{"ID", []string{"3"}}, {"SubmissionIndex", []string{"1"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"NextPage", []string{"Next"}}, {"SelectedAnswer0", []string{"2", "3"}}, {"SelectedAnswer1", []string{"0", "1"}}, {"SelectedAnswer2", []string{"3"}}},
		{{"ID", []string{"3"}}, {"SubmissionIndex", []string{"1"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"NextPage", []string{"Next"}}, {"SelectedAnswer0", []string{}}, {"SelectedAnswer1", []string{"0", "1"}}, {"SelectedAnswer2", []string{"3"}}},

		/* Programming page. */
		{{"ID", []string{"3"}}, {"SubmissionIndex", []string{"1"}}, {"StepIndex", []string{"a"}}, {"CurrentPage", []string{"Programming"}}, {"NextPage", []string{"Next"}}, {"LanguageID", []string{"4"}}, {"Solution", []string{"s = input()\nif s == \"aaa\":\n	print(\"bbb\")\nelif s == \"ccc\":\n	print(\"ddd\")\nelse:\n	print(\"eee\")\n"}}},
		{{"ID", []string{"3"}}, {"SubmissionIndex", []string{"1"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Programming"}}, {"NextPage", []string{"Next"}}, {"LanguageID", []string{"4"}}, {"Solution", []string{"s = input()\nif s == \"aaa\":\n	print(\"bbb\")\nelif s == \"ccc\":\n	print(\"ddd\")\nelse:\n	print(\"eee\")\n"}}},
		{{"ID", []string{"3"}}, {"SubmissionIndex", []string{"1"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"NextPage", []string{"Next"}}, {"LanguageID", []string{"a"}}, {"Solution", []string{"s = input()\nif s == \"aaa\":\n	print(\"bbb\")\nelif s == \"ccc\":\n	print(\"ddd\")\nelse:\n	print(\"eee\")\n"}}},
		{{"ID", []string{"3"}}, {"SubmissionIndex", []string{"1"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"NextPage", []string{"Next"}}, {"LanguageID", []string{"4"}}, {"Solution", []string{testString(MinSolutionLen - 1)}}},
		{{"ID", []string{"3"}}, {"SubmissionIndex", []string{"1"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"NextPage", []string{"Next"}}, {"LanguageID", []string{"4"}}, {"Solution", []string{testString(MaxSolutionLen + 1)}}},
		{{"ID", []string{"3"}}, {"SubmissionIndex", []string{"1"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"NextPage", []string{"Next"}}, {"LanguageID", []string{"4"}}, {"Solution", []string{"= input()\nif s == \"aaa\":\n	print(\"bbb\")\nelif s == \"ccc\":\n	print(\"ddd\")\nelse:\n	print(\"eee\")\n"}}},
		{{"ID", []string{"3"}}, {"SubmissionIndex", []string{"1"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"NextPage", []string{"Next"}}, {"LanguageID", []string{"4"}}, {"Solution", []string{"s = input()\nif s == \"aaa\":\n	print(\"aaa\")\nelif s == \"ccc\":\n	print(\"ddd\")\nelse:\n	print(\"eee\")\n"}}},

		/* Main page. */
		{{"ID", []string{"a"}}},
		{{"ID", []string{"3"}}, {"SubmissionIndex", []string{"a"}}},
		{{"ID", []string{"3"}}, {"SubmissionIndex", []string{"1"}}, {"CurrentPage", []string{"Main"}}, {"Commanda", []string{"Pass"}}},
		{{"ID", []string{"3"}}, {"SubmissionIndex", []string{"1"}}, {"CurrentPage", []string{"Main"}}, {"Command2", []string{"Pass"}}},
		{{"ID", []string{"3"}}, {"SubmissionIndex", []string{"1"}}, {"CurrentPage", []string{"Main"}}, {"Command", []string{"Command"}}},
		{{"ID", []string{"3"}}, {"SubmissionIndex", []string{"1"}}, {"CurrentPage", []string{"Main"}}, {"Command", nil}, {"Command2", []string{"Pass"}}},
		{{"ID", []string{"3"}}, {"SubmissionIndex", []string{"1"}}, {"NextPage", []string{"Finish"}}},
	}

	testPostAuth(t, endpoint, testTokens[2], url.Values{{"ID", []string{"3"}}}, http.StatusOK)
	for _, test := range expectedOK {
		test.SetInt("ID", 3)
		test.SetInt("SubmissionIndex", 0)
		testPostAuth(t, endpoint, testTokens[2], test, http.StatusOK)
	}
	testPostAuth(t, endpoint, testTokens[2], url.Values{{"ID", []string{"3"}}, {"SubmissionIndex", []string{"0"}}, {"NextPage", []string{"Finish"}}}, http.StatusSeeOther)

	testPostAuth(t, endpoint, testTokens[3], url.Values{{"ID", []string{"3"}}}, http.StatusOK)
	testPostAuth(t, endpoint, testTokens[3], url.Values{{"ID", []string{"3"}}, {"SubmissionIndex", []string{"1"}}, {"NextPage", []string{"Finish"}}}, http.StatusBadRequest)
	testPostAuth(t, endpoint, testTokens[3], url.Values{{"ID", []string{"3"}}, {"SubmissionIndex", []string{"1"}}, {"Command0", []string{"Pass"}}}, http.StatusBadRequest)
	testPostAuth(t, endpoint, testTokens[3], url.Values{{"ID", []string{"3"}}, {"SubmissionIndex", []string{"1"}}, {"CurrentPage", []string{"Main"}}, {"Command0", []string{"Pass"}}}, http.StatusOK)
	testPostAuth(t, endpoint, testTokens[3], url.Values{{"ID", []string{"3"}}, {"SubmissionIndex", []string{"1"}}, {"CurrentPage", []string{"Main"}}, {"Command1", []string{"Pass"}}}, http.StatusOK)
	for _, test := range expectedBadRequest {
		testPostAuth(t, endpoint, testTokens[3], test, http.StatusBadRequest)
	}
	ProgrammingLanguages[4].Available = false
	testPostAuth(t, endpoint, testTokens[3], url.Values{{"ID", []string{"3"}}, {"SubmissionIndex", []string{"1"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"NextPage", []string{"Next"}}, {"LanguageID", []string{"4"}}, {"Solution", []string{"s = input()\nif s == \"aaa\":\n	print(\"bbb\")\nelif s == \"ccc\":\n	print(\"ddd\")\nelse:\n	print(\"eee\")\n"}}}, http.StatusBadRequest)
	ProgrammingLanguages[4].Available = true
	testPostInvalidFormAuth(t, endpoint, testTokens[AdminID])

	testPost(t, endpoint, nil, http.StatusUnauthorized)
	testPostAuth(t, endpoint, testInvalidToken, nil, http.StatusUnauthorized)

	testPostAuth(t, endpoint, testTokens[1], url.Values{{"ID", []string{"3"}}}, http.StatusForbidden)
}
