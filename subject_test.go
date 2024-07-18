package main

import (
	"net/url"
	"strconv"
	"testing"

	"github.com/anton2920/gofa/net/http"
)

func TestSubjectPageHandler(t *testing.T) {
	const endpoint = "/subject/"

	testCreateInitialDBs()

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
		{"GroupID": {"0", "a"}, "TeacherID": {"0", "a"}},
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
		{"ID": {"0"}},
		{"ID": {"0"}, "GroupID": {"0", "a"}, "TeacherID": {"0", "a"}},
	}

	for _, test := range expectedOK {
		testPostAuth(t, endpoint, testTokens[AdminID], test, http.StatusOK)
	}

	testPostInvalidFormAuth(t, endpoint, testTokens[AdminID])

	testPost(t, endpoint, nil, http.StatusUnauthorized)
	testPostAuth(t, endpoint, testInvalidToken, nil, http.StatusUnauthorized)

	testPostAuth(t, endpoint, testTokens[1], nil, http.StatusForbidden)
}

func TestSubjectLessonsPageHandler(t *testing.T) {
	const endpoint = "/subject/lessons"

	expectedOK := [...]url.Values{
		{},

		/* Create two lessons, move them around and then delete. */
		{"CurrentPage": {"Main"}, "NextPage": {Ls(GL, "Add lesson")}},
		{"LessonIndex": {"0"}, "CurrentPage": {"Lesson"}, "Name": {"Test lesson #1"}, "Theory": {"This is test lesson #1's theory."}, "NextPage": {Ls(GL, "Next")}},
		{"CurrentPage": {"Main"}, "NextPage": {Ls(GL, "Add lesson")}},
		{"LessonIndex": {"1"}, "CurrentPage": {"Lesson"}, "Name": {"Test lesson #2"}, "Theory": {"This is test lesson #2's theory."}, "NextPage": {Ls(GL, "Next")}},
		{"CurrentPage": {"Main"}, "Command1": {Ls(GL, "^|")}},
		{"CurrentPage": {"Main"}, "Command0": {Ls(GL, "|v")}},
		{"CurrentPage": {"Main"}, "Command2": {Ls(GL, "^|")}},
		{"CurrentPage": {"Main"}, "Command2": {Ls(GL, "|v")}},
		{"CurrentPage": {"Main"}, "Command2": {Ls(GL, "Delete")}},
		{"CurrentPage": {"Main"}, "Command0": {Ls(GL, "Delete")}},
		{"CurrentPage": {"Main"}, "Command0": {Ls(GL, "Delete")}},

		/* Create lesson, create test and programming task. */
		{"CurrentPage": {"Main"}, "NextPage": {Ls(GL, "Add lesson")}},
		{"LessonIndex": {"0"}, "CurrentPage": {"Lesson"}, "NextPage": {Ls(GL, "Add test")}},
		{"LessonIndex": {"0"}, "StepIndex": {"0"}, "CurrentPage": {"Test"}, "Question": {""}, "Command0": {Ls(GL, "Add another answer")}},
		{"LessonIndex": {"0"}, "StepIndex": {"0"}, "CurrentPage": {"Test"}, "Question": {""}, "Command0": {Ls(GL, "Add another answer")}},
		{"LessonIndex": {"0"}, "StepIndex": {"0"}, "CurrentPage": {"Test"}, "Question": {""}, "Command0": {Ls(GL, "Add another answer")}},
		{"LessonIndex": {"0"}, "StepIndex": {"0"}, "CurrentPage": {"Test"}, "Question": {""}, "Command": {Ls(GL, "Add another question")}},
		{"LessonIndex": {"0"}, "StepIndex": {"0"}, "CurrentPage": {"Test"}, "Question": {"", ""}, "Command1": {Ls(GL, "Add another answer")}},
		{"LessonIndex": {"0"}, "StepIndex": {"0"}, "CurrentPage": {"Test"}, "Question": {"", ""}, "Command": {Ls(GL, "Add another question")}},
		{"LessonIndex": {"0"}, "StepIndex": {"0"}, "CurrentPage": {"Test"}, "Question": {"", "", ""}, "Command2": {Ls(GL, "Delete")}},
		{"LessonIndex": {"0"}, "StepIndex": {"0"}, "CurrentPage": {"Test"}, "Question": {"", "", ""}, "Command2": {Ls(GL, "Add another answer")}},
		{"LessonIndex": {"0"}, "StepIndex": {"0"}, "CurrentPage": {"Test"}, "Question": {"", "", ""}, "Command2": {Ls(GL, "Add another answer")}},
		{"LessonIndex": {"0"}, "StepIndex": {"0"}, "CurrentPage": {"Test"}, "Question": {"", "", ""}, "Command2": {Ls(GL, "Add another answer")}},
		{"LessonIndex": {"0"}, "StepIndex": {"0"}, "CurrentPage": {"Test"}, "Name": {"Back-end development basics"}, "Question": {"What is an API?", "To be or not to be?", "Third question"}, "Answer0": {"One", "Two", "Three", "Four"}, "CorrectAnswer0": {"2"}, "Answer1": {"To be", "Not to be"}, "CorrectAnswer1": {"0", "1"}, "Answer2": {"What?", "When?", "Where?", "Correct"}, "CorrectAnswer2": {"3"}, "NextPage": {Ls(GL, "Continue")}},
		{"LessonIndex": {"0"}, "CurrentPage": {"Lesson"}, "NextPage": {Ls(GL, "Add programming task")}},
		{"LessonIndex": {"0"}, "StepIndex": {"1"}, "CurrentPage": {"Programming"}, "Command": {Ls(GL, "Add example")}},
		{"LessonIndex": {"0"}, "StepIndex": {"1"}, "CurrentPage": {"Programming"}, "ExampleInput": {""}, "ExampleOutput": {""}, "Command": {Ls(GL, "Add example")}},
		{"LessonIndex": {"0"}, "StepIndex": {"1"}, "CurrentPage": {"Programming"}, "ExampleInput": {"", ""}, "ExampleOutput": {"", ""}, "Command1.0": {Ls(GL, "-")}},
		{"LessonIndex": {"0"}, "StepIndex": {"1"}, "CurrentPage": {"Programming"}, "ExampleInput": {"", ""}, "ExampleOutput": {"", ""}, "Command": {Ls(GL, "Add test")}},
		{"LessonIndex": {"0"}, "StepIndex": {"1"}, "CurrentPage": {"Programming"}, "Name": {"Introduction"}, "ExampleInput": {"aaa", "ccc"}, "ExampleOutput": {"bbb", "ddd"}, "TestInput": {"fff"}, "TestOutput": {"eee"}, "Description": {"Print 'hello, world' in your favourite language"}, "NextPage": {Ls(GL, "Continue")}},
		{"LessonIndex": {"0"}, "CurrentPage": {"Lesson"}, "Command1": {Ls(GL, "^|")}},
		{"LessonIndex": {"0"}, "CurrentPage": {"Lesson"}, "Command0": {Ls(GL, "|v")}},
		{"LessonIndex": {"0"}, "CurrentPage": {"Lesson"}, "Name": {"Introduction"}, "Theory": {"This is an introduction."}, "NextPage": {Ls(GL, "Next")}},

		/* Edit lesson, add/remove another test and/move/remove another check to programming task. */
		{"CurrentPage": {"Main"}, "Command0": {Ls(GL, "Edit")}},
		{"LessonIndex": {"0"}, "CurrentPage": {"Lesson"}, "NextPage": {Ls(GL, "Add test")}},
		{"LessonIndex": {"0"}, "StepIndex": {"2"}, "CurrentPage": {"Test"}, "Question": {""}, "Command0": {Ls(GL, "Add another answer")}},
		{"LessonIndex": {"0"}, "StepIndex": {"2"}, "CurrentPage": {"Test"}, "Question": {""}, "Answer0": {"", ""}, "Command0.1": {Ls(GL, "^|")}},
		{"LessonIndex": {"0"}, "StepIndex": {"2"}, "CurrentPage": {"Test"}, "Question": {""}, "Answer0": {"", ""}, "Command0.0": {Ls(GL, "|v")}},
		{"LessonIndex": {"0"}, "StepIndex": {"2"}, "CurrentPage": {"Test"}, "Question": {""}, "Answer0": {"", ""}, "CorrectAnswer0": {"0", "1"}, "Command0.0": {Ls(GL, "-")}},
		{"LessonIndex": {"0"}, "StepIndex": {"2"}, "CurrentPage": {"Test"}, "Question": {"", ""}, "Answer0": {"", ""}, "CorrectAnswer0": {"0"}, "Command0.1": {Ls(GL, "^|")}},
		{"LessonIndex": {"0"}, "StepIndex": {"2"}, "CurrentPage": {"Test"}, "Question": {"", ""}, "Answer0": {"", ""}, "CorrectAnswer0": {"1"}, "Command0.1": {Ls(GL, "^|")}},
		{"LessonIndex": {"0"}, "StepIndex": {"2"}, "CurrentPage": {"Test"}, "Question": {"", ""}, "Answer0": {"", ""}, "CorrectAnswer0": {"0"}, "Command0.0": {Ls(GL, "|v")}},
		{"LessonIndex": {"0"}, "StepIndex": {"2"}, "CurrentPage": {"Test"}, "Question": {"", ""}, "Answer0": {"", ""}, "CorrectAnswer0": {"1"}, "Command0.0": {Ls(GL, "|v")}},
		{"LessonIndex": {"0"}, "StepIndex": {"2"}, "CurrentPage": {"Test"}, "Question": {"", ""}, "Answer0": {"", ""}, "CorrectAnswer0": {"0", "1"}, "Command1": {Ls(GL, "^|")}},
		{"LessonIndex": {"0"}, "StepIndex": {"2"}, "CurrentPage": {"Test"}, "Question": {"", ""}, "Answer0": {"", ""}, "CorrectAnswer0": {"0", "1"}, "Command1": {Ls(GL, "|v")}},
		{"LessonIndex": {"0"}, "StepIndex": {"2"}, "CurrentPage": {"Test"}, "Question": {""}, "Command0": {Ls(GL, "Add another answer")}},
		{"LessonIndex": {"0"}, "StepIndex": {"2"}, "CurrentPage": {"Test"}, "Name": {"Simple test"}, "Question": {"Yes?"}, "Answer0": {"No", "Yes"}, "CorrectAnswer0": {"1"}, "NextPage": {Ls(GL, "Continue")}},
		{"LessonIndex": {"0"}, "CurrentPage": {"Lesson"}, "Command2": {Ls(GL, "Delete")}},
		{"LessonIndex": {"0"}, "CurrentPage": {"Lesson"}, "Command0": {Ls(GL, "Edit")}},
		{"LessonIndex": {"0"}, "CurrentPage": {"Lesson"}, "Command1": {Ls(GL, "Edit")}},
		{"LessonIndex": {"0"}, "StepIndex": {"1"}, "CurrentPage": {"Programming"}, "Command": {Ls(GL, "Add example")}},
		{"LessonIndex": {"0"}, "StepIndex": {"1"}, "CurrentPage": {"Programming"}, "ExampleInput": {"aaa", "ccc", ""}, "ExampleOutput": {"bbb", "ddd", ""}, "Command2.0": {Ls(GL, "^|")}},
		{"LessonIndex": {"0"}, "StepIndex": {"1"}, "CurrentPage": {"Programming"}, "ExampleInput": {"aaa", "", "ccc"}, "ExampleOutput": {"bbb", "", "ddd"}, "Command1.0": {Ls(GL, "|v")}},
		{"LessonIndex": {"0"}, "StepIndex": {"1"}, "CurrentPage": {"Programming"}, "ExampleInput": {"aaa", "ccc", ""}, "ExampleOutput": {"bbb", "ddd", ""}, "Command2.0": {Ls(GL, "-")}},
		{"LessonIndex": {"0"}, "StepIndex": {"1"}, "CurrentPage": {"Programming"}, "Name": {"Introduction"}, "ExampleInput": {"aaa", "ccc"}, "ExampleOutput": {"bbb", "ddd"}, "TestInput": {"fff"}, "TestOutput": {"eee"}, "Description": {"Print 'hello, world' in your favourite language"}, "NextPage": {Ls(GL, "Continue")}},
		{"LessonIndex": {"0"}, "StepIndex": {"0"}, "CurrentPage": {"Test"}, "Name": {"Back-end development basics"}, "Question": {"What is an API?", "To be or not to be?", "Third question"}, "Answer0": {"One", "Two", "Three", "Four"}, "CorrectAnswer0": {"2"}, "Answer1": {"To be", "Not to be"}, "CorrectAnswer1": {"0", "1"}, "Answer2": {"What?", "When?", "Where?", "Correct"}, "CorrectAnswer2": {"3"}, "NextPage": {Ls(GL, "Continue")}},
		{"LessonIndex": {"0"}, "CurrentPage": {"Lesson"}, "Name": {"Introduction"}, "Theory": {"This is an introduction."}, "NextPage": {Ls(GL, "Next")}},
	}

	expectedBadRequest := [...]url.Values{
		/* Misc. */
		{"ID": {"0"}, "Command": {Ls(GL, "Command")}},
		{"ID": {"0"}, "Commanda": {Ls(GL, "Command")}},
		{"ID": {"0"}, "Command0.a": {Ls(GL, "Command")}},
		{"ID": {"0"}, "LessonIndex": {"a"}, "StepIndex": {"0"}, "NextPage": {Ls(GL, "Continue")}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"a"}, "NextPage": {Ls(GL, "Continue")}},
		{"ID": {"0"}, "CourseID": {"a"}, "Action": {Ls(GL, "create from")}},
		{"ID": {"0"}, "CourseID": {"a"}, "Action": {Ls(GL, "give as is")}},

		/* Test page. */
		{"ID": {"0"}, "LessonIndex": {"a"}, "StepIndex": {"0"}, "CurrentPage": {"Test"}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"a"}, "CurrentPage": {"Test"}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"1"}, "CurrentPage": {"Test"}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"1"}, "CurrentPage": {"Test"}},
		{"ID": {"0"}, "LessonIndex": {"a"}, "StepIndex": {"0"}, "CurrentPage": {"Test"}, "Command": {Ls(GL, "Command")}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"1"}, "CurrentPage": {"Test"}, "Command": {Ls(GL, "Command")}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"2"}, "CurrentPage": {"Test"}, "Command": {Ls(GL, "Command")}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"0"}, "CurrentPage": {"Test"}, "Command": {Ls(GL, "Command")}, "Name": {"Back-end development basics"}, "Question": {"What is an API?", "To be or not to be?", "Third question"}, "Answer0": {"One", "Two", "Three", "Four"}, "CorrectAnswer0": {"4"}, "Answer1": {"To be", "Not to be"}, "CorrectAnswer1": {"0", "1"}, "Answer2": {"What?", "When?", "Where?", "Correct"}, "CorrectAnswer2": {"3"}, "NextPage": {Ls(GL, "Continue")}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"0"}, "CurrentPage": {"Test"}, "Command1": {Ls(GL, "Add another answer")}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"0"}, "CurrentPage": {"Test"}, "Command1.0": {Ls(GL, "-")}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"0"}, "CurrentPage": {"Test"}, "Question": {"Test question"}, "Command0.1": {Ls(GL, "-")}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"0"}, "CurrentPage": {"Test"}, "Question": {"Test question"}, "Answer0": {"", ""}, "CorrectAnswer0": {"0", "1"}, "Command1.1": {Ls(GL, "^|")}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"0"}, "CurrentPage": {"Test"}, "Question": {"Test question"}, "Answer0": {"", ""}, "CorrectAnswer0": {"0", "1"}, "Command1.0": {Ls(GL, "|v")}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"0"}, "CurrentPage": {"Test"}, "Name": {testString(MinStepNameLen - 1)}, "Question": {"What is an API?", "To be or not to be?", "Third question"}, "Answer0": {"One", "Two", "Three", "Four"}, "CorrectAnswer0": {"2"}, "Answer1": {"To be", "Not to be"}, "CorrectAnswer1": {"0", "1"}, "Answer2": {"What?", "When?", "Where?", "Correct"}, "CorrectAnswer2": {"3"}, "NextPage": {Ls(GL, "Continue")}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"0"}, "CurrentPage": {"Test"}, "Name": {testString(MaxStepNameLen + 1)}, "Question": {"What is an API?", "To be or not to be?", "Third question"}, "Answer0": {"One", "Two", "Three", "Four"}, "CorrectAnswer0": {"2"}, "Answer1": {"To be", "Not to be"}, "CorrectAnswer1": {"0", "1"}, "Answer2": {"What?", "When?", "Where?", "Correct"}, "CorrectAnswer2": {"3"}, "NextPage": {Ls(GL, "Continue")}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"0"}, "CurrentPage": {"Test"}, "Name": {"Back-end development basics"}, "Question": {testString(MinQuestionLen - 1), "To be or not to be?", "Third question"}, "Answer0": {"One", "Two", "Three", "Four"}, "CorrectAnswer0": {"2"}, "Answer1": {"To be", "Not to be"}, "CorrectAnswer1": {"0", "1"}, "Answer2": {"What?", "When?", "Where?", "Correct"}, "CorrectAnswer2": {"3"}, "NextPage": {Ls(GL, "Continue")}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"0"}, "CurrentPage": {"Test"}, "Name": {"Back-end development basics"}, "Question": {testString(MaxQuestionLen + 1), "To be or not to be?", "Third question"}, "Answer0": {"One", "Two", "Three", "Four"}, "CorrectAnswer0": {"2"}, "Answer1": {"To be", "Not to be"}, "CorrectAnswer1": {"0", "1"}, "Answer2": {"What?", "When?", "Where?", "Correct"}, "CorrectAnswer2": {"3"}, "NextPage": {Ls(GL, "Continue")}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"0"}, "CurrentPage": {"Test"}, "Name": {"Back-end development basics"}, "Question": {"What is an API?", "To be or not to be?", "Third question"}, "Answer0": {testString(MinAnswerLen - 1), "Two", "Three", "Four"}, "CorrectAnswer0": {"2"}, "Answer1": {"To be", "Not to be"}, "CorrectAnswer1": {"0", "1"}, "Answer2": {"What?", "When?", "Where?", "Correct"}, "CorrectAnswer2": {"3"}, "NextPage": {Ls(GL, "Continue")}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"0"}, "CurrentPage": {"Test"}, "Name": {"Back-end development basics"}, "Question": {"What is an API?", "To be or not to be?", "Third question"}, "Answer0": {testString(MaxAnswerLen + 1), "Two", "Three", "Four"}, "CorrectAnswer0": {"2"}, "Answer1": {"To be", "Not to be"}, "CorrectAnswer1": {"0", "1"}, "Answer2": {"What?", "When?", "Where?", "Correct"}, "CorrectAnswer2": {"3"}, "NextPage": {Ls(GL, "Continue")}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"0"}, "CurrentPage": {"Test"}, "Name": {"Back-end development basics"}, "Question": {"What is an API?", "To be or not to be?", "Third question"}, "Answer0": {"One", "Two", "Three", "Four"}, "CorrectAnswer0": {"4"}, "Answer1": {"To be", "Not to be"}, "CorrectAnswer1": {"0", "1"}, "Answer2": {"What?", "When?", "Where?", "Correct"}, "CorrectAnswer2": {"3"}, "NextPage": {Ls(GL, "Continue")}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"0"}, "CurrentPage": {"Test"}, "Name": {"Back-end development basics"}, "Question": {"What is an API?", "To be or not to be?", "Third question"}, "Answer0": {"One", "Two", "Three", "Four"}, "CorrectAnswer0": nil, "Answer1": {"To be", "Not to be"}, "CorrectAnswer1": {"0", "1"}, "Answer2": {"What?", "When?", "Where?", "Correct"}, "CorrectAnswer2": {"3"}, "NextPage": {Ls(GL, "Continue")}},

		/* Programming page. */
		{"ID": {"0"}, "LessonIndex": {"a"}, "StepIndex": {"1"}, "CurrentPage": {"Programming"}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"a"}, "CurrentPage": {"Programming"}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"0"}, "CurrentPage": {"Programming"}},
		{"ID": {"0"}, "LessonIndex": {"a"}, "StepIndex": {"0"}, "CurrentPage": {"Programming"}, "Command": {Ls(GL, "Command")}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"0"}, "CurrentPage": {"Programming"}, "Command": {Ls(GL, "Command")}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"2"}, "CurrentPage": {"Programming"}, "Command": {Ls(GL, "Command")}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"1"}, "CurrentPage": {"Programming"}, "Command": {Ls(GL, "Command")}, "Name": {"Introduction"}, "ExampleInput": {"aaa", "ccc"}, "ExampleOutput": {"bbb"}, "TestInput": {"fff"}, "TestOutput": {"eee"}, "Description": {"Print 'hello, world' in your favourite language"}, "NextPage": {Ls(GL, "Continue")}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"1"}, "CurrentPage": {"Programming"}, "Command0.2": {Ls(GL, "-")}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"1"}, "CurrentPage": {"Programming"}, "Command0.2": {Ls(GL, "^|")}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"1"}, "CurrentPage": {"Programming"}, "Command0.2": {Ls(GL, "|v")}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"1"}, "CurrentPage": {"Programming"}, "Name": {testString(MinStepNameLen - 1)}, "ExampleInput": {"aaa", "ccc"}, "ExampleOutput": {"bbb", "ddd"}, "TestInput": {"fff"}, "TestOutput": {"eee"}, "Description": {"Print 'hello, world' in your favourite language"}, "NextPage": {Ls(GL, "Continue")}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"1"}, "CurrentPage": {"Programming"}, "Name": {testString(MaxStepNameLen + 1)}, "ExampleInput": {"aaa", "ccc"}, "ExampleOutput": {"bbb", "ddd"}, "TestInput": {"fff"}, "TestOutput": {"eee"}, "Description": {"Print 'hello, world' in your favourite language"}, "NextPage": {Ls(GL, "Continue")}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"1"}, "CurrentPage": {"Programming"}, "Name": {"Introduction"}, "ExampleInput": {"aaa", "ccc"}, "ExampleOutput": {"bbb", "ddd"}, "TestInput": {"fff"}, "TestOutput": {"eee"}, "Description": {testString(MinDescriptionLen - 1)}, "NextPage": {Ls(GL, "Continue")}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"1"}, "CurrentPage": {"Programming"}, "Name": {"Introduction"}, "ExampleInput": {"aaa", "ccc"}, "ExampleOutput": {"bbb", "ddd"}, "TestInput": {"fff"}, "TestOutput": {"eee"}, "Description": {testString(MaxDescriptionLen + 1)}, "NextPage": {Ls(GL, "Continue")}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"1"}, "CurrentPage": {"Programming"}, "Name": {"Introduction"}, "ExampleInput": {testString(MinCheckLen - 1), "ccc"}, "ExampleOutput": {"bbb", "ddd"}, "TestInput": {"fff"}, "TestOutput": {"eee"}, "Description": {"Print 'hello, world' in your favourite language"}, "NextPage": {Ls(GL, "Continue")}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"1"}, "CurrentPage": {"Programming"}, "Name": {"Introduction"}, "ExampleInput": {testString(MaxCheckLen + 1), "ccc"}, "ExampleOutput": {"bbb", "ddd"}, "TestInput": {"fff"}, "TestOutput": {"eee"}, "Description": {"Print 'hello, world' in your favourite language"}, "NextPage": {Ls(GL, "Continue")}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"1"}, "CurrentPage": {"Programming"}, "Name": {"Introduction"}, "ExampleInput": {"aaa", "ccc"}, "ExampleOutput": {testString(MinCheckLen - 1), "ddd"}, "TestInput": {"fff"}, "TestOutput": {"eee"}, "Description": {"Print 'hello, world' in your favourite language"}, "NextPage": {Ls(GL, "Continue")}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"1"}, "CurrentPage": {"Programming"}, "Name": {"Introduction"}, "ExampleInput": {"aaa", "ccc"}, "ExampleOutput": {testString(MaxCheckLen + 1), "ddd"}, "TestInput": {"fff"}, "TestOutput": {"eee"}, "Description": {"Print 'hello, world' in your favourite language"}, "NextPage": {Ls(GL, "Continue")}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"1"}, "CurrentPage": {"Programming"}, "Name": {"Introduction"}, "ExampleInput": {"aaa", "ccc"}, "ExampleOutput": {"bbb", "ddd"}, "TestInput": {testString(MinCheckLen - 1)}, "TestOutput": {"eee"}, "Description": {"Print 'hello, world' in your favourite language"}, "NextPage": {Ls(GL, "Continue")}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"1"}, "CurrentPage": {"Programming"}, "Name": {"Introduction"}, "ExampleInput": {"aaa", "ccc"}, "ExampleOutput": {"bbb", "ddd"}, "TestInput": {testString(MaxCheckLen + 1)}, "TestOutput": {"eee"}, "Description": {"Print 'hello, world' in your favourite language"}, "NextPage": {Ls(GL, "Continue")}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"1"}, "CurrentPage": {"Programming"}, "Name": {"Introduction"}, "ExampleInput": {"aaa", "ccc"}, "ExampleOutput": {"bbb", "ddd"}, "TestInput": {"fff"}, "TestOutput": {testString(MinCheckLen - 1)}, "Description": {"Print 'hello, world' in your favourite language"}, "NextPage": {Ls(GL, "Continue")}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"1"}, "CurrentPage": {"Programming"}, "Name": {"Introduction"}, "ExampleInput": {"aaa", "ccc"}, "ExampleOutput": {"bbb", "ddd"}, "TestInput": {"fff"}, "TestOutput": {testString(MaxCheckLen + 1)}, "Description": {"Print 'hello, world' in your favourite language"}, "NextPage": {Ls(GL, "Continue")}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "StepIndex": {"1"}, "CurrentPage": {"Programming"}, "Name": {"Introduction"}, "ExampleInput": {"aaa", "ccc"}, "ExampleOutput": {"bbb"}, "TestInput": {"fff"}, "TestOutput": {"eee"}, "Description": {"Print 'hello, world' in your favourite language"}, "NextPage": {Ls(GL, "Continue")}},

		/* Lesson page. */
		{"ID": {"0"}, "LessonIndex": {"a"}, "CurrentPage": {"Lesson"}, "Command0": {Ls(GL, "Edit")}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "CurrentPage": {"Lesson"}, "Command2": {Ls(GL, "Edit")}},
		{"ID": {"0"}, "LessonIndex": {"a"}, "CurrentPage": {"Lesson"}, "NextPage": {Ls(GL, "Add test")}},
		{"ID": {"0"}, "LessonIndex": {"a"}, "CurrentPage": {"Lesson"}, "NextPage": {Ls(GL, "Add programming task")}},
		{"ID": {"0"}, "LessonIndex": {"a"}, "CurrentPage": {"Lesson"}, "NextPage": {Ls(GL, "Next")}, "Name": {"Introduction"}, "Theory": {"This is an introduction."}},
		{"ID": {"0"}, "LessonIndex": {"a"}, "CurrentPage": {"Lesson"}, "NextPage": {Ls(GL, "Next")}, "Name": {"Introduction"}, "Theory": {"This is an introduction."}},
		{"ID": {"0"}, "LessonIndex": {"1"}, "CurrentPage": {"Lesson"}, "NextPage": {Ls(GL, "Next")}, "Name": {"Introduction"}, "Theory": {"This is an introduction."}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "CurrentPage": {"Lesson"}, "NextPage": {Ls(GL, "Next")}, "Name": {testString(MinNameLen - 1)}, "Theory": {"This is an introduction"}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "CurrentPage": {"Lesson"}, "NextPage": {Ls(GL, "Next")}, "Name": {testString(MaxNameLen + 1)}, "Theory": {"This is an introduction"}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "CurrentPage": {"Lesson"}, "NextPage": {Ls(GL, "Next")}, "Name": {"Introduction"}, "Theory": {testString(MinTheoryLen - 1)}},
		{"ID": {"0"}, "LessonIndex": {"0"}, "CurrentPage": {"Lesson"}, "NextPage": {Ls(GL, "Next")}, "Name": {"Introduction"}, "Theory": {testString(MaxTheoryLen + 1)}},

		/* Main page. */
		{"ID": {"a"}},
		{"ID": {"0"}, "CurrentPage": {"Main"}, "Command1": {Ls(GL, "Edit")}},
	}

	expectedForbidden := [...]url.Values{
		{"ID": {"1"}, "CourseID": {"0"}, "Action": {Ls(GL, "create from")}},
		{"ID": {"1"}, "CourseID": {"0"}, "Action": {Ls(GL, "give as is")}},
	}

	expectedNotFound := [...]url.Values{
		{"ID": {"4"}},
		{"ID": {"0"}, "CourseID": {"10"}, "Action": {Ls(GL, "create from")}},
		{"ID": {"0"}, "CourseID": {"10"}, "Action": {Ls(GL, "give as is")}},
	}

	testPostAuth(t, endpoint, testTokens[AdminID], url.Values{"ID": {"0"}, "CourseID": {"0"}, "Action": {Ls(GL, "create from")}}, http.StatusOK)
	for i, token := range testTokens[:2] {
		for _, test := range expectedOK {
			test.Set("ID", strconv.Itoa(i))
			testPostAuth(t, endpoint, token, test, http.StatusOK)
		}
		testPostAuth(t, endpoint, token, url.Values{"ID": {strconv.Itoa(i)}, "NextPage": {Ls(GL, "Save")}}, http.StatusSeeOther)
	}

	var subject Subject
	if err := GetSubjectByID(0, &subject); err != nil {
		t.Fatalf("Failed to get subject by ID 0: %v", err)
	}
	subject.Lessons = nil
	if err := SaveSubject(&subject); err != nil {
		t.Fatalf("Failed to save subject: %v", err)
	}
	testPostAuth(t, endpoint, testTokens[AdminID], url.Values{"ID": {"0"}, "CourseID": {"0"}, "Action": {Ls(GL, "give as is")}}, http.StatusSeeOther)

	for _, test := range expectedBadRequest {
		testPostAuth(t, endpoint, testTokens[AdminID], test, http.StatusBadRequest)
	}

	if err := GetSubjectByID(1, &subject); err != nil {
		t.Fatalf("Failed to get subject by ID 1: %v", err)
	}
	var lesson Lesson
	if err := GetLessonByID(subject.Lessons[0], &lesson); err != nil {
		t.Fatalf("Failed to get lesson by ID: %v", err)
	}
	lesson.Flags = LessonDraft
	if err := SaveLesson(&lesson); err != nil {
		t.Fatalf("Failed to save lesson: %v", err)
	}
	testPostAuth(t, endpoint, testTokens[AdminID], url.Values{"ID": {"1"}, "NextPage": {Ls(GL, "Save")}}, http.StatusBadRequest)
	subject.Lessons = nil
	if err := SaveSubject(&subject); err != nil {
		t.Fatalf("Failed to save subject: %v", err)
	}
	testPostAuth(t, endpoint, testTokens[AdminID], url.Values{"ID": {"1"}, "NextPage": {Ls(GL, "Save")}}, http.StatusBadRequest)
	testPostInvalidFormAuth(t, endpoint, testTokens[AdminID])

	testPost(t, endpoint, nil, http.StatusUnauthorized)
	testPostAuth(t, endpoint, testInvalidToken, nil, http.StatusUnauthorized)

	for _, test := range expectedForbidden {
		testPostAuth(t, endpoint, testTokens[1], test, http.StatusForbidden)
	}
	testPostAuth(t, endpoint, testTokens[2], url.Values{"ID": {"0"}}, http.StatusForbidden)

	for _, test := range expectedNotFound {
		testPostAuth(t, endpoint, testTokens[AdminID], test, http.StatusNotFound)
	}
}

func TestSubjectCreateHandler(t *testing.T) {
	const endpoint = APIPrefix + "/subject/create"

	expectedOK := [...]url.Values{
		{"Name": {"Chemistry"}, "TeacherID": {"1"}, "GroupID": {"0"}},
	}

	expectedBadRequest := [...]url.Values{
		{"Name": {testString(MinNameLen - 1)}, "TeacherID": {"1"}, "GroupID": {"0"}},
		{"Name": {testString(MaxNameLen + 1)}, "TeacherID": {"1"}, "GroupID": {"0"}},
		{"Name": {"Chemistry"}, "TeacherID": {"a"}, "GroupID": {"0"}},
		{"Name": {"Chemistry"}, "TeacherID": {"10"}, "GroupID": {"0"}},
		{"Name": {"Chemistry"}, "TeacherID": {"1"}, "GroupID": {"a"}},
		{"Name": {"Chemistry"}, "TeacherID": {"1"}, "GroupID": {"10"}},
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
		{"ID": {"0"}, "Name": {"Chemistry"}, "TeacherID": {"1"}, "GroupID": {"0"}},
	}

	expectedBadRequest := [...]url.Values{
		{"ID": {"a"}, "Name": {testString(MinNameLen - 1)}, "TeacherID": {"1"}, "GroupID": {"0"}},
		{"ID": {"0"}, "Name": {testString(MinNameLen - 1)}, "TeacherID": {"1"}, "GroupID": {"0"}},
		{"ID": {"0"}, "Name": {testString(MaxNameLen + 1)}, "TeacherID": {"1"}, "GroupID": {"0"}},
		{"ID": {"0"}, "Name": {"Chemistry"}, "TeacherID": {"a"}, "GroupID": {"0"}},
		{"ID": {"0"}, "Name": {"Chemistry"}, "TeacherID": {"10"}, "GroupID": {"0"}},
		{"ID": {"0"}, "Name": {"Chemistry"}, "TeacherID": {"1"}, "GroupID": {"a"}},
		{"ID": {"0"}, "Name": {"Chemistry"}, "TeacherID": {"1"}, "GroupID": {"10"}},
	}

	expectedForbidden := expectedOK

	expectedNotFound := [...]url.Values{
		{"ID": {"3"}, "Name": {"Chemistry"}, "TeacherID": {"1"}, "GroupID": {"0"}},
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
