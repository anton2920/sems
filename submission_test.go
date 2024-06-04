package main

import (
	"testing"

	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/net/url"
)

func TestSubmissionNewPageHandler(t *testing.T) {
	const endpoint = "/submission/new"

	expectedOK := [...]url.Values{
		/* Test. */
		{{"CurrentPage", []string{"Main"}}, {"Command0", []string{"Pass"}}},
		{{"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"NextPage", []string{"Discard"}}},
		{{"CurrentPage", []string{"Main"}}, {"Command0", []string{"Pass"}}},
		{{"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"NextPage", []string{"Next"}}, {"SelectedAnswer0", []string{"2"}}, {"SelectedAnswer1", []string{"0", "1"}}, {"SelectedAnswer2", []string{"3"}}},
		{{"CurrentPage", []string{"Main"}}, {"Command0", []string{"Edit"}}},
		{{"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"NextPage", []string{"Next"}}, {"SelectedAnswer0", []string{"2"}}, {"SelectedAnswer1", []string{"0", "1"}}, {"SelectedAnswer2", []string{"3"}}},

		/* Programming task. */
		{{"CurrentPage", []string{"Main"}}, {"Command1", []string{"Pass"}}},
		{{"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"NextPage", []string{"Next"}}, {"LanguageID", []string{"4"}}, {"Solution", []string{"s = input()\nif s == \"aaa\":\n	print(\"bbb\")\nelif s == \"ccc\":\n	print(\"ddd\")\nelse:\n	print(\"eee\")\n"}}},
		{{"CurrentPage", []string{"Main"}}, {"Command1", []string{"Edit"}}},
		{{"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"NextPage", []string{"Next"}}, {"LanguageID", []string{"4"}}, {"Solution", []string{"s = input()\nif s == \"aaa\":\n	print(\"bbb\")\nelif s == \"ccc\":\n	print(\"ddd\")\nelse:\n	print(\"eee\")\n"}}},
	}

	testPostAuth(t, endpoint, testTokens[2], url.Values{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}}, http.StatusOK)
	for _, test := range expectedOK {
		test.SetInt("ID", 0)
		test.SetInt("LessonIndex", 0)
		test.SetInt("SubmissionIndex", 0)
		testPostAuth(t, endpoint, testTokens[2], test, http.StatusOK)
	}
	testPostAuth(t, endpoint, testTokens[2], url.Values{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"SubmissionIndex", []string{"0"}}, {"NextPage", []string{"Finish"}}}, http.StatusSeeOther)

	testPostInvalidFormAuth(t, endpoint, testTokens[AdminID])

	testPost(t, endpoint, nil, http.StatusUnauthorized)
	testPostAuth(t, endpoint, testInvalidToken, nil, http.StatusUnauthorized)
}
