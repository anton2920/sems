package main

import (
	"strconv"
	"testing"

	"github.com/anton2920/gofa/database"
	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/net/url"
)

func TestCoursePageHandler(t *testing.T) {
	const endpoint = "/course/"

	expectedOK := [...]string{"0", "1"}

	expectedBadRequest := [...]string{"a", "b", "c"}

	expectedForbidden := [...]string{"0"}

	expectedNotFound := [...]string{"2", "3", "4"}

	for _, test := range expectedOK {
		testGetAuth(t, endpoint+test, testTokens[AdminID], http.StatusOK)
	}

	for _, test := range expectedBadRequest {
		testGetAuth(t, endpoint+test, testTokens[AdminID], http.StatusBadRequest)
	}

	testGet(t, endpoint, http.StatusUnauthorized)
	testGetAuth(t, endpoint, testInvalidToken, http.StatusUnauthorized)

	for _, test := range expectedForbidden {
		testGetAuth(t, endpoint+test, testTokens[1], http.StatusForbidden)
	}

	for _, test := range expectedNotFound {
		testGetAuth(t, endpoint+test, testTokens[AdminID], http.StatusNotFound)
	}
}

func testCourseCreateEditPageHandler(t *testing.T, endpoint string) {
	expectedOK := [...]url.Values{
		/* Create two lessons, move them around and then delete. */
		{{"CurrentPage", []string{"Course"}}, {"NextPage", []string{Ls(GL, "Add lesson")}}},
		{{"LessonIndex", []string{"0"}}, {"CurrentPage", []string{"Lesson"}}, {"Name", []string{"Test lesson #1"}}, {"Theory", []string{"This is test lesson #1's theory."}}, {"NextPage", []string{Ls(GL, "Next")}}},
		{{"CurrentPage", []string{"Course"}}, {"NextPage", []string{Ls(GL, "Add lesson")}}},
		{{"LessonIndex", []string{"1"}}, {"CurrentPage", []string{"Lesson"}}, {"Name", []string{"Test lesson #2"}}, {"Theory", []string{"This is test lesson #2's theory."}}, {"NextPage", []string{Ls(GL, "Next")}}},
		{{"CurrentPage", []string{"Course"}}, {"Command1", []string{Ls(GL, "^|")}}},
		{{"CurrentPage", []string{"Course"}}, {"Command0", []string{Ls(GL, "|v")}}},
		{{"CurrentPage", []string{"Course"}}, {"Command2", []string{Ls(GL, "^|")}}},
		{{"CurrentPage", []string{"Course"}}, {"Command2", []string{Ls(GL, "|v")}}},
		{{"CurrentPage", []string{"Course"}}, {"Command2", []string{Ls(GL, "Delete")}}},
		{{"CurrentPage", []string{"Course"}}, {"Command0", []string{Ls(GL, "Delete")}}},
		{{"CurrentPage", []string{"Course"}}, {"Command0", []string{Ls(GL, "Delete")}}},

		/* Create lesson, create test and programming task. */
		{{"CurrentPage", []string{"Course"}}, {"NextPage", []string{Ls(GL, "Add lesson")}}},
		{{"LessonIndex", []string{"0"}}, {"CurrentPage", []string{"Lesson"}}, {"NextPage", []string{Ls(GL, "Add test")}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{""}}, {"Command0", []string{Ls(GL, "Add another answer")}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{""}}, {"Command0", []string{Ls(GL, "Add another answer")}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{""}}, {"Command0", []string{Ls(GL, "Add another answer")}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{""}}, {"Command", []string{Ls(GL, "Add another question")}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{"", ""}}, {"Command1", []string{Ls(GL, "Add another answer")}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{"", ""}}, {"Command", []string{Ls(GL, "Add another question")}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{"", "", ""}}, {"Command2", []string{Ls(GL, "Delete")}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{"", "", ""}}, {"Command2", []string{Ls(GL, "Add another answer")}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{"", "", ""}}, {"Command2", []string{Ls(GL, "Add another answer")}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{"", "", ""}}, {"Command2", []string{Ls(GL, "Add another answer")}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Name", []string{"Back-end development basics"}}, {"Question", []string{"What is an API?", "To be or not to be?", "Third question"}}, {"Answer0", []string{"One", "Two", "Three", "Four"}}, {"CorrectAnswer0", []string{"2"}}, {"Answer1", []string{"To be", "Not to be"}}, {"CorrectAnswer1", []string{"0", "1"}}, {"Answer2", []string{"What?", "When?", "Where?", "Correct"}}, {"CorrectAnswer2", []string{"3"}}, {"NextPage", []string{Ls(GL, "Continue")}}},
		{{"LessonIndex", []string{"0"}}, {"CurrentPage", []string{"Lesson"}}, {"NextPage", []string{Ls(GL, "Add programming task")}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"Command", []string{Ls(GL, "Add example")}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"ExampleInput", []string{""}}, {"ExampleOutput", []string{""}}, {"Command", []string{Ls(GL, "Add example")}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"ExampleInput", []string{"", ""}}, {"ExampleOutput", []string{"", ""}}, {"Command1.0", []string{Ls(GL, "-")}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"ExampleInput", []string{"", ""}}, {"ExampleOutput", []string{"", ""}}, {"Command", []string{Ls(GL, "Add test")}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"Name", []string{"Introduction"}}, {"ExampleInput", []string{"aaa", "ccc"}}, {"ExampleOutput", []string{"bbb", "ddd"}}, {"TestInput", []string{"fff"}}, {"TestOutput", []string{"eee"}}, {"Description", []string{"Print 'hello, world' in your favourite language"}}, {"NextPage", []string{Ls(GL, "Continue")}}},
		{{"LessonIndex", []string{"0"}}, {"CurrentPage", []string{"Lesson"}}, {"Command1", []string{Ls(GL, "^|")}}},
		{{"LessonIndex", []string{"0"}}, {"CurrentPage", []string{"Lesson"}}, {"Command0", []string{Ls(GL, "|v")}}},
		{{"LessonIndex", []string{"0"}}, {"CurrentPage", []string{"Lesson"}}, {"Name", []string{"Introduction"}}, {"Theory", []string{"This is an introduction."}}, {"NextPage", []string{Ls(GL, "Next")}}},

		/* Edit lesson, add/remove another test and/move/remove another check to programming task. */
		{{"CurrentPage", []string{"Course"}}, {"Command0", []string{Ls(GL, "Edit")}}},
		{{"LessonIndex", []string{"0"}}, {"CurrentPage", []string{"Lesson"}}, {"NextPage", []string{Ls(GL, "Add test")}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"2"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{""}}, {"Command0", []string{Ls(GL, "Add another answer")}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"2"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{""}}, {"Answer0", []string{"", ""}}, {"Command0.1", []string{Ls(GL, "^|")}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"2"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{""}}, {"Answer0", []string{"", ""}}, {"Command0.0", []string{Ls(GL, "|v")}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"2"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{""}}, {"Answer0", []string{"", ""}}, {"CorrectAnswer0", []string{"0", "1"}}, {"Command0.0", []string{Ls(GL, "-")}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"2"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{"", ""}}, {"Answer0", []string{"", ""}}, {"CorrectAnswer0", []string{"0"}}, {"Command0.1", []string{Ls(GL, "^|")}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"2"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{"", ""}}, {"Answer0", []string{"", ""}}, {"CorrectAnswer0", []string{"1"}}, {"Command0.1", []string{Ls(GL, "^|")}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"2"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{"", ""}}, {"Answer0", []string{"", ""}}, {"CorrectAnswer0", []string{"0"}}, {"Command0.0", []string{Ls(GL, "|v")}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"2"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{"", ""}}, {"Answer0", []string{"", ""}}, {"CorrectAnswer0", []string{"1"}}, {"Command0.0", []string{Ls(GL, "|v")}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"2"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{"", ""}}, {"Answer0", []string{"", ""}}, {"CorrectAnswer0", []string{"0", "1"}}, {"Command1", []string{Ls(GL, "^|")}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"2"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{"", ""}}, {"Answer0", []string{"", ""}}, {"CorrectAnswer0", []string{"0", "1"}}, {"Command1", []string{Ls(GL, "|v")}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"2"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{""}}, {"Command0", []string{Ls(GL, "Add another answer")}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"2"}}, {"CurrentPage", []string{"Test"}}, {"Name", []string{"Simple test"}}, {"Question", []string{"Yes?"}}, {"Answer0", []string{"No", "Yes"}}, {"CorrectAnswer0", []string{"1"}}, {"NextPage", []string{Ls(GL, "Continue")}}},
		{{"LessonIndex", []string{"0"}}, {"CurrentPage", []string{"Lesson"}}, {"Command2", []string{Ls(GL, "Delete")}}},
		{{"LessonIndex", []string{"0"}}, {"CurrentPage", []string{"Lesson"}}, {"Command0", []string{Ls(GL, "Edit")}}},
		{{"LessonIndex", []string{"0"}}, {"CurrentPage", []string{"Lesson"}}, {"Command1", []string{Ls(GL, "Edit")}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"Command", []string{Ls(GL, "Add example")}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"ExampleInput", []string{"aaa", "ccc", ""}}, {"ExampleOutput", []string{"bbb", "ddd", ""}}, {"Command2.0", []string{Ls(GL, "^|")}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"ExampleInput", []string{"aaa", "", "ccc"}}, {"ExampleOutput", []string{"bbb", "", "ddd"}}, {"Command1.0", []string{Ls(GL, "|v")}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"ExampleInput", []string{"aaa", "ccc", ""}}, {"ExampleOutput", []string{"bbb", "ddd", ""}}, {"Command2.0", []string{Ls(GL, "-")}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"Name", []string{"Introduction"}}, {"ExampleInput", []string{"aaa", "ccc"}}, {"ExampleOutput", []string{"bbb", "ddd"}}, {"TestInput", []string{"fff"}}, {"TestOutput", []string{"eee"}}, {"Description", []string{"Print 'hello, world' in your favourite language"}}, {"NextPage", []string{Ls(GL, "Continue")}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Name", []string{"Back-end development basics"}}, {"Question", []string{"What is an API?", "To be or not to be?", "Third question"}}, {"Answer0", []string{"One", "Two", "Three", "Four"}}, {"CorrectAnswer0", []string{"2"}}, {"Answer1", []string{"To be", "Not to be"}}, {"CorrectAnswer1", []string{"0", "1"}}, {"Answer2", []string{"What?", "When?", "Where?", "Correct"}}, {"CorrectAnswer2", []string{"3"}}, {"NextPage", []string{Ls(GL, "Continue")}}},
		{{"LessonIndex", []string{"0"}}, {"CurrentPage", []string{"Lesson"}}, {"Name", []string{"Introduction"}}, {"Theory", []string{"This is an introduction."}}, {"NextPage", []string{Ls(GL, "Next")}}},
	}

	expectedBadRequest := [...]url.Values{
		/* Misc. */
		{{"ID", []string{"0"}}, {"Command", []string{Ls(GL, "Command")}}},
		{{"ID", []string{"0"}}, {"Commanda", []string{Ls(GL, "Command")}}},
		{{"ID", []string{"0"}}, {"Command0.a", []string{Ls(GL, "Command")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"a"}}, {"StepIndex", []string{"0"}}, {"NextPage", []string{Ls(GL, "Continue")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"a"}}, {"NextPage", []string{Ls(GL, "Continue")}}},

		/* Test page. */
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"a"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"a"}}, {"CurrentPage", []string{"Test"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Test"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Test"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"a"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Command", []string{Ls(GL, "Command")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Test"}}, {"Command", []string{Ls(GL, "Command")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"2"}}, {"CurrentPage", []string{"Test"}}, {"Command", []string{Ls(GL, "Command")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Command", []string{Ls(GL, "Command")}}, {"Name", []string{"Back-end development basics"}}, {"Question", []string{"What is an API?", "To be or not to be?", "Third question"}}, {"Answer0", []string{"One", "Two", "Three", "Four"}}, {"CorrectAnswer0", []string{"4"}}, {"Answer1", []string{"To be", "Not to be"}}, {"CorrectAnswer1", []string{"0", "1"}}, {"Answer2", []string{"What?", "When?", "Where?", "Correct"}}, {"CorrectAnswer2", []string{"3"}}, {"NextPage", []string{Ls(GL, "Continue")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Command1", []string{Ls(GL, "Add another answer")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Command1.0", []string{Ls(GL, "-")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{"Test question"}}, {"Command0.1", []string{Ls(GL, "-")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{"Test question"}}, {"Answer0", []string{"", ""}}, {"CorrectAnswer0", []string{"0", "1"}}, {"Command1.1", []string{Ls(GL, "^|")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{"Test question"}}, {"Answer0", []string{"", ""}}, {"CorrectAnswer0", []string{"0", "1"}}, {"Command1.0", []string{Ls(GL, "|v")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Name", []string{testString(MinStepNameLen - 1)}}, {"Question", []string{"What is an API?", "To be or not to be?", "Third question"}}, {"Answer0", []string{"One", "Two", "Three", "Four"}}, {"CorrectAnswer0", []string{"2"}}, {"Answer1", []string{"To be", "Not to be"}}, {"CorrectAnswer1", []string{"0", "1"}}, {"Answer2", []string{"What?", "When?", "Where?", "Correct"}}, {"CorrectAnswer2", []string{"3"}}, {"NextPage", []string{Ls(GL, "Continue")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Name", []string{testString(MaxStepNameLen + 1)}}, {"Question", []string{"What is an API?", "To be or not to be?", "Third question"}}, {"Answer0", []string{"One", "Two", "Three", "Four"}}, {"CorrectAnswer0", []string{"2"}}, {"Answer1", []string{"To be", "Not to be"}}, {"CorrectAnswer1", []string{"0", "1"}}, {"Answer2", []string{"What?", "When?", "Where?", "Correct"}}, {"CorrectAnswer2", []string{"3"}}, {"NextPage", []string{Ls(GL, "Continue")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Name", []string{"Back-end development basics"}}, {"Question", []string{testString(MinQuestionLen - 1), "To be or not to be?", "Third question"}}, {"Answer0", []string{"One", "Two", "Three", "Four"}}, {"CorrectAnswer0", []string{"2"}}, {"Answer1", []string{"To be", "Not to be"}}, {"CorrectAnswer1", []string{"0", "1"}}, {"Answer2", []string{"What?", "When?", "Where?", "Correct"}}, {"CorrectAnswer2", []string{"3"}}, {"NextPage", []string{Ls(GL, "Continue")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Name", []string{"Back-end development basics"}}, {"Question", []string{testString(MaxQuestionLen + 1), "To be or not to be?", "Third question"}}, {"Answer0", []string{"One", "Two", "Three", "Four"}}, {"CorrectAnswer0", []string{"2"}}, {"Answer1", []string{"To be", "Not to be"}}, {"CorrectAnswer1", []string{"0", "1"}}, {"Answer2", []string{"What?", "When?", "Where?", "Correct"}}, {"CorrectAnswer2", []string{"3"}}, {"NextPage", []string{Ls(GL, "Continue")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Name", []string{"Back-end development basics"}}, {"Question", []string{"What is an API?", "To be or not to be?", "Third question"}}, {"Answer0", []string{testString(MinAnswerLen - 1), "Two", "Three", "Four"}}, {"CorrectAnswer0", []string{"2"}}, {"Answer1", []string{"To be", "Not to be"}}, {"CorrectAnswer1", []string{"0", "1"}}, {"Answer2", []string{"What?", "When?", "Where?", "Correct"}}, {"CorrectAnswer2", []string{"3"}}, {"NextPage", []string{Ls(GL, "Continue")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Name", []string{"Back-end development basics"}}, {"Question", []string{"What is an API?", "To be or not to be?", "Third question"}}, {"Answer0", []string{testString(MaxAnswerLen + 1), "Two", "Three", "Four"}}, {"CorrectAnswer0", []string{"2"}}, {"Answer1", []string{"To be", "Not to be"}}, {"CorrectAnswer1", []string{"0", "1"}}, {"Answer2", []string{"What?", "When?", "Where?", "Correct"}}, {"CorrectAnswer2", []string{"3"}}, {"NextPage", []string{Ls(GL, "Continue")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Name", []string{"Back-end development basics"}}, {"Question", []string{"What is an API?", "To be or not to be?", "Third question"}}, {"Answer0", []string{"One", "Two", "Three", "Four"}}, {"CorrectAnswer0", []string{"4"}}, {"Answer1", []string{"To be", "Not to be"}}, {"CorrectAnswer1", []string{"0", "1"}}, {"Answer2", []string{"What?", "When?", "Where?", "Correct"}}, {"CorrectAnswer2", []string{"3"}}, {"NextPage", []string{Ls(GL, "Continue")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Name", []string{"Back-end development basics"}}, {"Question", []string{"What is an API?", "To be or not to be?", "Third question"}}, {"Answer0", []string{"One", "Two", "Three", "Four"}}, {"CorrectAnswer0", nil}, {"Answer1", []string{"To be", "Not to be"}}, {"CorrectAnswer1", []string{"0", "1"}}, {"Answer2", []string{"What?", "When?", "Where?", "Correct"}}, {"CorrectAnswer2", []string{"3"}}, {"NextPage", []string{Ls(GL, "Continue")}}},

		/* Programming page. */
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"a"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"a"}}, {"CurrentPage", []string{"Programming"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Programming"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"a"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Programming"}}, {"Command", []string{Ls(GL, "Command")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Programming"}}, {"Command", []string{Ls(GL, "Command")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"2"}}, {"CurrentPage", []string{"Programming"}}, {"Command", []string{Ls(GL, "Command")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"Command", []string{Ls(GL, "Command")}}, {"Name", []string{"Introduction"}}, {"ExampleInput", []string{"aaa", "ccc"}}, {"ExampleOutput", []string{"bbb"}}, {"TestInput", []string{"fff"}}, {"TestOutput", []string{"eee"}}, {"Description", []string{"Print 'hello, world' in your favourite language"}}, {"NextPage", []string{Ls(GL, "Continue")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"Command0.2", []string{Ls(GL, "-")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"Command0.2", []string{Ls(GL, "^|")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"Command0.2", []string{Ls(GL, "|v")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"Name", []string{testString(MinStepNameLen - 1)}}, {"ExampleInput", []string{"aaa", "ccc"}}, {"ExampleOutput", []string{"bbb", "ddd"}}, {"TestInput", []string{"fff"}}, {"TestOutput", []string{"eee"}}, {"Description", []string{"Print 'hello, world' in your favourite language"}}, {"NextPage", []string{Ls(GL, "Continue")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"Name", []string{testString(MaxStepNameLen + 1)}}, {"ExampleInput", []string{"aaa", "ccc"}}, {"ExampleOutput", []string{"bbb", "ddd"}}, {"TestInput", []string{"fff"}}, {"TestOutput", []string{"eee"}}, {"Description", []string{"Print 'hello, world' in your favourite language"}}, {"NextPage", []string{Ls(GL, "Continue")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"Name", []string{"Introduction"}}, {"ExampleInput", []string{"aaa", "ccc"}}, {"ExampleOutput", []string{"bbb", "ddd"}}, {"TestInput", []string{"fff"}}, {"TestOutput", []string{"eee"}}, {"Description", []string{testString(MinDescriptionLen - 1)}}, {"NextPage", []string{Ls(GL, "Continue")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"Name", []string{"Introduction"}}, {"ExampleInput", []string{"aaa", "ccc"}}, {"ExampleOutput", []string{"bbb", "ddd"}}, {"TestInput", []string{"fff"}}, {"TestOutput", []string{"eee"}}, {"Description", []string{testString(MaxDescriptionLen + 1)}}, {"NextPage", []string{Ls(GL, "Continue")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"Name", []string{"Introduction"}}, {"ExampleInput", []string{testString(MinCheckLen - 1), "ccc"}}, {"ExampleOutput", []string{"bbb", "ddd"}}, {"TestInput", []string{"fff"}}, {"TestOutput", []string{"eee"}}, {"Description", []string{"Print 'hello, world' in your favourite language"}}, {"NextPage", []string{Ls(GL, "Continue")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"Name", []string{"Introduction"}}, {"ExampleInput", []string{testString(MaxCheckLen + 1), "ccc"}}, {"ExampleOutput", []string{"bbb", "ddd"}}, {"TestInput", []string{"fff"}}, {"TestOutput", []string{"eee"}}, {"Description", []string{"Print 'hello, world' in your favourite language"}}, {"NextPage", []string{Ls(GL, "Continue")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"Name", []string{"Introduction"}}, {"ExampleInput", []string{"aaa", "ccc"}}, {"ExampleOutput", []string{testString(MinCheckLen - 1), "ddd"}}, {"TestInput", []string{"fff"}}, {"TestOutput", []string{"eee"}}, {"Description", []string{"Print 'hello, world' in your favourite language"}}, {"NextPage", []string{Ls(GL, "Continue")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"Name", []string{"Introduction"}}, {"ExampleInput", []string{"aaa", "ccc"}}, {"ExampleOutput", []string{testString(MaxCheckLen + 1), "ddd"}}, {"TestInput", []string{"fff"}}, {"TestOutput", []string{"eee"}}, {"Description", []string{"Print 'hello, world' in your favourite language"}}, {"NextPage", []string{Ls(GL, "Continue")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"Name", []string{"Introduction"}}, {"ExampleInput", []string{"aaa", "ccc"}}, {"ExampleOutput", []string{"bbb", "ddd"}}, {"TestInput", []string{testString(MinCheckLen - 1)}}, {"TestOutput", []string{"eee"}}, {"Description", []string{"Print 'hello, world' in your favourite language"}}, {"NextPage", []string{Ls(GL, "Continue")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"Name", []string{"Introduction"}}, {"ExampleInput", []string{"aaa", "ccc"}}, {"ExampleOutput", []string{"bbb", "ddd"}}, {"TestInput", []string{testString(MaxCheckLen + 1)}}, {"TestOutput", []string{"eee"}}, {"Description", []string{"Print 'hello, world' in your favourite language"}}, {"NextPage", []string{Ls(GL, "Continue")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"Name", []string{"Introduction"}}, {"ExampleInput", []string{"aaa", "ccc"}}, {"ExampleOutput", []string{"bbb", "ddd"}}, {"TestInput", []string{"fff"}}, {"TestOutput", []string{testString(MinCheckLen - 1)}}, {"Description", []string{"Print 'hello, world' in your favourite language"}}, {"NextPage", []string{Ls(GL, "Continue")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"Name", []string{"Introduction"}}, {"ExampleInput", []string{"aaa", "ccc"}}, {"ExampleOutput", []string{"bbb", "ddd"}}, {"TestInput", []string{"fff"}}, {"TestOutput", []string{testString(MaxCheckLen + 1)}}, {"Description", []string{"Print 'hello, world' in your favourite language"}}, {"NextPage", []string{Ls(GL, "Continue")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"Name", []string{"Introduction"}}, {"ExampleInput", []string{"aaa", "ccc"}}, {"ExampleOutput", []string{"bbb"}}, {"TestInput", []string{"fff"}}, {"TestOutput", []string{"eee"}}, {"Description", []string{"Print 'hello, world' in your favourite language"}}, {"NextPage", []string{Ls(GL, "Continue")}}},

		/* Lesson page. */
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"a"}}, {"CurrentPage", []string{"Lesson"}}, {"Command0", []string{Ls(GL, "Edit")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"CurrentPage", []string{"Lesson"}}, {"Command2", []string{Ls(GL, "Edit")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"a"}}, {"CurrentPage", []string{"Lesson"}}, {"NextPage", []string{Ls(GL, "Add test")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"a"}}, {"CurrentPage", []string{"Lesson"}}, {"NextPage", []string{Ls(GL, "Add programming task")}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"a"}}, {"CurrentPage", []string{"Lesson"}}, {"NextPage", []string{Ls(GL, "Next")}}, {"Name", []string{"Introduction"}}, {"Theory", []string{"This is an introduction."}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"a"}}, {"CurrentPage", []string{"Lesson"}}, {"NextPage", []string{Ls(GL, "Next")}}, {"Name", []string{"Introduction"}}, {"Theory", []string{"This is an introduction."}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"1"}}, {"CurrentPage", []string{"Lesson"}}, {"NextPage", []string{Ls(GL, "Next")}}, {"Name", []string{"Introduction"}}, {"Theory", []string{"This is an introduction."}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"CurrentPage", []string{"Lesson"}}, {"NextPage", []string{Ls(GL, "Next")}}, {"Name", []string{testString(MinNameLen - 1)}}, {"Theory", []string{"This is an introduction"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"CurrentPage", []string{"Lesson"}}, {"NextPage", []string{Ls(GL, "Next")}}, {"Name", []string{testString(MaxNameLen + 1)}}, {"Theory", []string{"This is an introduction"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"CurrentPage", []string{"Lesson"}}, {"NextPage", []string{Ls(GL, "Next")}}, {"Name", []string{"Introduction"}}, {"Theory", []string{testString(MinTheoryLen - 1)}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"CurrentPage", []string{"Lesson"}}, {"NextPage", []string{Ls(GL, "Next")}}, {"Name", []string{"Introduction"}}, {"Theory", []string{testString(MaxTheoryLen + 1)}}},

		/* Course page. */
		{{"ID", []string{"a"}}},
		{{"ID", []string{"0"}}, {"CurrentPage", []string{"Course"}}, {"NextPage", []string{Ls(GL, "Save")}}, {"Name", []string{testString(MinNameLen - 1)}}},
		{{"ID", []string{"0"}}, {"CurrentPage", []string{"Course"}}, {"NextPage", []string{Ls(GL, "Save")}}, {"Name", []string{testString(MaxNameLen + 1)}}},
		{{"ID", []string{"0"}}, {"CurrentPage", []string{"Course"}}, {"Command1", []string{Ls(GL, "Edit")}}},
	}

	expectedForbidden := [...]url.Values{
		{{"ID", []string{"0"}}},
	}

	expectedNotFound := [...]url.Values{
		{{"ID", []string{"4"}}},
	}

	if err := database.Drop(CoursesDB); err != nil {
		t.Fatalf("Failed to drop courses data: %v", err)
	}
	for i, token := range testTokens {
		testPostAuth(t, endpoint, token, url.Values{}, http.StatusOK)
		for _, test := range expectedOK {
			test.SetInt("ID", i)
			testPostAuth(t, endpoint, token, test, http.StatusOK)
		}
		testPostAuth(t, endpoint, token, url.Values{{"ID", []string{strconv.Itoa(i)}}, {"CurrentPage", []string{"Course"}}, {"Name", []string{"Programming basics"}}, {"NextPage", []string{Ls(GL, "Save")}}}, http.StatusSeeOther)
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

func TestCourseCreatePageHandler(t *testing.T) {
	testCourseCreateEditPageHandler(t, "/course/create")
}

func TestCourseEditPageHandler(t *testing.T) {
	testCourseCreateEditPageHandler(t, "/course/edit")
}
