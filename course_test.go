package main

import (
	"strconv"
	"testing"

	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/net/url"
)

func testCourseCreateEditPageHandler(t *testing.T, endpoint string) {
	expectedOK := [...]url.Values{
		/* Create two lessons, move them around and then delete. */
		{{"CurrentPage", []string{"Course"}}, {"NextPage", []string{"Add lesson"}}},
		{{"LessonIndex", []string{"0"}}, {"CurrentPage", []string{"Lesson"}}, {"Name", []string{"Test lesson #1"}}, {"Theory", []string{"This is test lesson #1's theory."}}, {"NextPage", []string{"Next"}}},
		{{"CurrentPage", []string{"Course"}}, {"NextPage", []string{"Add lesson"}}},
		{{"LessonIndex", []string{"1"}}, {"CurrentPage", []string{"Lesson"}}, {"Name", []string{"Test lesson #2"}}, {"Theory", []string{"This is test lesson #2's theory."}}, {"NextPage", []string{"Next"}}},
		{{"CurrentPage", []string{"Course"}}, {"Command1", []string{"^|"}}},
		{{"CurrentPage", []string{"Course"}}, {"Command0", []string{"|v"}}},
		{{"CurrentPage", []string{"Course"}}, {"Command2", []string{"^|"}}},
		{{"CurrentPage", []string{"Course"}}, {"Command2", []string{"|v"}}},
		{{"CurrentPage", []string{"Course"}}, {"Command2", []string{"Delete"}}},
		{{"CurrentPage", []string{"Course"}}, {"Command0", []string{"Delete"}}},
		{{"CurrentPage", []string{"Course"}}, {"Command0", []string{"Delete"}}},

		/* Create lesson, create test and programming task. */
		{{"CurrentPage", []string{"Course"}}, {"NextPage", []string{"Add lesson"}}},
		{{"LessonIndex", []string{"0"}}, {"CurrentPage", []string{"Lesson"}}, {"NextPage", []string{"Add test"}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{""}}, {"Command0", []string{"Add another answer"}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{""}}, {"Command0", []string{"Add another answer"}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{""}}, {"Command0", []string{"Add another answer"}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{""}}, {"Command", []string{"Add another question"}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{"", ""}}, {"Command1", []string{"Add another answer"}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{"", ""}}, {"Command", []string{"Add another question"}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{"", "", ""}}, {"Command2", []string{"Add another answer"}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{"", "", ""}}, {"Command2", []string{"Add another answer"}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{"", "", ""}}, {"Command2", []string{"Add another answer"}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Name", []string{"Back-end development basics"}}, {"Question", []string{"What is an API?", "To be or not to be?", "Third question"}}, {"Answer0", []string{"One", "Two", "Three", "Four"}}, {"CorrectAnswer0", []string{"2"}}, {"Answer1", []string{"To be", "Not to be"}}, {"CorrectAnswer1", []string{"0", "1"}}, {"Answer2", []string{"What?", "When?", "Where?", "Correct"}}, {"CorrectAnswer2", []string{"3"}}, {"NextPage", []string{"Continue"}}},
		{{"LessonIndex", []string{"0"}}, {"CurrentPage", []string{"Lesson"}}, {"NextPage", []string{"Add programming task"}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"Command", []string{"Add example"}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"Command", []string{"Add example"}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"Command", []string{"Add test"}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"Name", []string{"Introduction"}}, {"ExampleInput", []string{"aaa", "ccc"}}, {"ExampleOutput", []string{"bbb", "ddd"}}, {"TestInput", []string{"fff"}}, {"TestOutput", []string{"eee"}}, {"Description", []string{"Print 'hello, world' in your favourite language"}}, {"NextPage", []string{"Continue"}}},
		{{"LessonIndex", []string{"0"}}, {"CurrentPage", []string{"Lesson"}}, {"Name", []string{"Introduction"}}, {"Theory", []string{"This is an introduction."}}, {"NextPage", []string{"Next"}}},

		/* Edit lesson, add/remove another test and/move/remove another check to programming task. */
		{{"CurrentPage", []string{"Course"}}, {"Command0", []string{"Edit"}}},
		{{"LessonIndex", []string{"0"}}, {"CurrentPage", []string{"Lesson"}}, {"NextPage", []string{"Add test"}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"2"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{""}}, {"Command0", []string{"Add another answer"}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"2"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{""}}, {"Answer0", []string{"", ""}}, {"Command0.1", []string{"^|"}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"2"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{""}}, {"Answer0", []string{"", ""}}, {"Command0.0", []string{"|v"}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"2"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{""}}, {"Answer0", []string{"", ""}}, {"Command0.0", []string{"-"}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"2"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{""}}, {"Command0", []string{"Add another answer"}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"2"}}, {"CurrentPage", []string{"Test"}}, {"Name", []string{"Simple test"}}, {"Question", []string{"Yes?"}}, {"Answer0", []string{"No", "Yes"}}, {"CorrectAnswer0", []string{"1"}}, {"NextPage", []string{"Continue"}}},
		{{"LessonIndex", []string{"0"}}, {"CurrentPage", []string{"Lesson"}}, {"Command2", []string{"Delete"}}},
		{{"LessonIndex", []string{"0"}}, {"CurrentPage", []string{"Lesson"}}, {"Command1", []string{"Edit"}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"Command", []string{"Add example"}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"ExampleInput", []string{"aaa", "ccc", ""}}, {"ExampleOutput", []string{"bbb", "ddd", ""}}, {"Command2.0", []string{"^|"}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"ExampleInput", []string{"aaa", "", "ccc"}}, {"ExampleOutput", []string{"bbb", "", "ddd"}}, {"Command1.0", []string{"|v"}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"ExampleInput", []string{"aaa", "ccc", ""}}, {"ExampleOutput", []string{"bbb", "ddd", ""}}, {"Command2.0", []string{"-"}}},
		{{"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"Name", []string{"Introduction"}}, {"ExampleInput", []string{"aaa", "ccc"}}, {"ExampleOutput", []string{"bbb", "ddd"}}, {"TestInput", []string{"fff"}}, {"TestOutput", []string{"eee"}}, {"Description", []string{"Print 'hello, world' in your favourite language"}}, {"NextPage", []string{"Continue"}}},
		{{"LessonIndex", []string{"0"}}, {"CurrentPage", []string{"Lesson"}}, {"Name", []string{"Introduction"}}, {"Theory", []string{"This is an introduction."}}, {"NextPage", []string{"Next"}}},
	}

	expectedBadRequest := [...]url.Values{
		/* Course page. */
		{{"ID", []string{"a"}}},
		{{"ID", []string{"100"}}},
		{{"ID", []string{"0"}}, {"CurrentPage", []string{"Course"}}, {"NextPage", []string{"Save"}}, {"Name", []string{""}}},
		{{"ID", []string{"0"}}, {"CurrentPage", []string{"Course"}}, {"NextPage", []string{"Save"}}, {"Name", []string{"TestTestTestTestTestTestTestTestTestTestTestTe"}}},
		{{"ID", []string{"0"}}, {"CurrentPage", []string{"Course"}}, {"Command1", []string{"Edit"}}},

		/* Lesson page. */
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"a"}}, {"CurrentPage", []string{"Lesson"}}, {"Name", []string{"Introduction"}}, {"Theory", []string{"This is an introduction."}}, {"NextPage", []string{"Next"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"1"}}, {"CurrentPage", []string{"Lesson"}}, {"Name", []string{"Introduction"}}, {"Theory", []string{"This is an introduction."}}, {"NextPage", []string{"Next"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"CurrentPage", []string{"Lesson"}}, {"NextPage", []string{"Save"}}, {"Name", []string{""}}, {"Description", []string{"This is an introduction"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"CurrentPage", []string{"Lesson"}}, {"NextPage", []string{"Save"}}, {"Name", []string{"TestTestTestTestTestTestTestTestTestTestTestTe"}}, {"Description", []string{"This is an introduction"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"CurrentPage", []string{"Lesson"}}, {"NextPage", []string{"Save"}}, {"Name", []string{"Introduction"}}, {"Description", []string{""}}},

		/* Test page. */

		/* Programming page. */
	}

	DB.Courses = nil
	for i, token := range testTokens {
		DB.Users[i].Courses = nil

		testPostAuth(t, endpoint, token, url.Values{}, http.StatusOK)
		for _, test := range expectedOK {
			test.SetInt("ID", i)
			testPostAuth(t, endpoint, token, test, http.StatusOK)
		}
		testPostAuth(t, endpoint, token, url.Values{{"ID", []string{strconv.Itoa(i)}}, {"CurrentPage", []string{"Course"}}, {"Name", []string{"Programming basics"}}, {"NextPage", []string{"Save"}}}, http.StatusSeeOther)
	}

	for _, test := range expectedBadRequest {
		testPostAuth(t, endpoint, testTokens[AdminID], test, http.StatusBadRequest)
	}
	testPostInvalidFormAuth(t, endpoint, testTokens[AdminID])

	testPost(t, endpoint, nil, http.StatusUnauthorized)
	testPostAuth(t, endpoint, testInvalidToken, nil, http.StatusUnauthorized)
}

func TestCourseCreatePageHandler(t *testing.T) {
	testCourseCreateEditPageHandler(t, "/course/create")
}

func TestCourseEditPageHandler(t *testing.T) {
	//testCourseCreateEditPageHandler(t, "/course/edit")
}
