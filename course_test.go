package main

import (
	"testing"

	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/net/url"
)

func testCourseCreateEditPageHandler(t *testing.T, endpoint string) {
	expectedOK := [...]url.Values{
		{},

		/* Create two lessons, move them around and then delete. */
		{{"ID", []string{"0"}}, {"CurrentPage", []string{"Course"}}, {"NextPage", []string{"Add lesson"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"CurrentPage", []string{"Lesson"}}, {"Name", []string{"Test lesson #1"}}, {"Theory", []string{"This is test lesson #1's theory."}}, {"NextPage", []string{"Next"}}},
		{{"ID", []string{"0"}}, {"CurrentPage", []string{"Course"}}, {"NextPage", []string{"Add lesson"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"1"}}, {"CurrentPage", []string{"Lesson"}}, {"Name", []string{"Test lesson #2"}}, {"Theory", []string{"This is test lesson #2's theory."}}, {"NextPage", []string{"Next"}}},
		{{"ID", []string{"0"}}, {"CurrentPage", []string{"Course"}}, {"Command1", []string{"^|"}}},
		{{"ID", []string{"0"}}, {"CurrentPage", []string{"Course"}}, {"Command0", []string{"|v"}}},
		{{"ID", []string{"0"}}, {"CurrentPage", []string{"Course"}}, {"Command0", []string{"Delete"}}},
		{{"ID", []string{"0"}}, {"CurrentPage", []string{"Course"}}, {"Command0", []string{"Delete"}}},

		/* Create lesson, create test and programming task. */
		{{"ID", []string{"0"}}, {"CurrentPage", []string{"Course"}}, {"NextPage", []string{"Add lesson"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"CurrentPage", []string{"Lesson"}}, {"NextPage", []string{"Add test"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{""}}, {"Command0", []string{"Add another answer"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{""}}, {"Command0", []string{"Add another answer"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{""}}, {"Command0", []string{"Add another answer"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{""}}, {"Command", []string{"Add another question"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{"", ""}}, {"Command1", []string{"Add another answer"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{"", ""}}, {"Command", []string{"Add another question"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{"", "", ""}}, {"Command2", []string{"Add another answer"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{"", "", ""}}, {"Command2", []string{"Add another answer"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{"", "", ""}}, {"Command2", []string{"Add another answer"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"0"}}, {"CurrentPage", []string{"Test"}}, {"Name", []string{"Back-end development basics"}}, {"Question", []string{"What is an API?", "To be or not to be?", "Third question"}}, {"Answer0", []string{"One", "Two", "Three", "Four"}}, {"CorrectAnswer0", []string{"2"}}, {"Answer1", []string{"To be", "Not to be"}}, {"CorrectAnswer1", []string{"0", "1"}}, {"Answer2", []string{"What?", "When?", "Where?", "Correct"}}, {"CorrectAnswer2", []string{"3"}}, {"NextPage", []string{"Continue"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"CurrentPage", []string{"Lesson"}}, {"NextPage", []string{"Add programming task"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"Command", []string{"Add example"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"Command", []string{"Add example"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"Command", []string{"Add test"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"Name", []string{"Introduction"}}, {"ExampleInput", []string{"aaa", "ccc"}}, {"ExampleOutput", []string{"bbb", "ddd"}}, {"TestInput", []string{"fff"}}, {"TestOutput", []string{"eee"}}, {"Description", []string{"Print 'hello, world' in your favourite language"}}, {"NextPage", []string{"Continue"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"CurrentPage", []string{"Lesson"}}, {"Name", []string{"Introduction"}}, {"Theory", []string{"This is an introduction."}}, {"NextPage", []string{"Next"}}},

		/* Edit lesson, add/remove another test and/move/remove another check to programming task. */
		{{"ID", []string{"0"}}, {"CurrentPage", []string{"Course"}}, {"Command0", []string{"Edit"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"CurrentPage", []string{"Lesson"}}, {"NextPage", []string{"Add test"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"2"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{""}}, {"Command0", []string{"Add another answer"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"2"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{""}}, {"Answer0", []string{"", ""}}, {"Command0.1", []string{"^|"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"2"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{""}}, {"Answer0", []string{"", ""}}, {"Command0.0", []string{"|v"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"2"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{""}}, {"Answer0", []string{"", ""}}, {"Command0.0", []string{"-"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"2"}}, {"CurrentPage", []string{"Test"}}, {"Question", []string{""}}, {"Command0", []string{"Add another answer"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"2"}}, {"CurrentPage", []string{"Test"}}, {"Name", []string{"Simple test"}}, {"Question", []string{"Yes?"}}, {"Answer0", []string{"No", "Yes"}}, {"CorrectAnswer0", []string{"1"}}, {"NextPage", []string{"Continue"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"CurrentPage", []string{"Lesson"}}, {"Command2", []string{"Delete"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"CurrentPage", []string{"Lesson"}}, {"Command1", []string{"Edit"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"Command", []string{"Add example"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"ExampleInput", []string{"aaa", "ccc", ""}}, {"ExampleOutput", []string{"bbb", "ddd", ""}}, {"Command2.0", []string{"^|"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"ExampleInput", []string{"aaa", "", "ccc"}}, {"ExampleOutput", []string{"bbb", "", "ddd"}}, {"Command1.0", []string{"|v"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"ExampleInput", []string{"aaa", "ccc", ""}}, {"ExampleOutput", []string{"bbb", "ddd", ""}}, {"Command2.0", []string{"-"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"StepIndex", []string{"1"}}, {"CurrentPage", []string{"Programming"}}, {"Name", []string{"Introduction"}}, {"ExampleInput", []string{"aaa", "ccc"}}, {"ExampleOutput", []string{"bbb", "ddd"}}, {"TestInput", []string{"fff"}}, {"TestOutput", []string{"eee"}}, {"Description", []string{"Print 'hello, world' in your favourite language"}}, {"NextPage", []string{"Continue"}}},
		{{"ID", []string{"0"}}, {"LessonIndex", []string{"0"}}, {"CurrentPage", []string{"Lesson"}}, {"Name", []string{"Introduction"}}, {"Theory", []string{"This is an introduction."}}, {"NextPage", []string{"Next"}}},
	}

	for i, token := range testTokens {
		DB.Users[i].Courses = nil
		for _, test := range expectedOK {
			testPostAuth(t, endpoint, token, test, http.StatusOK)
		}
		testPostAuth(t, endpoint, token, url.Values{{"ID", []string{"0"}}, {"CurrentPage", []string{"Course"}}, {"Name", []string{"Programming basics"}}, {"NextPage", []string{"Save"}}}, http.StatusSeeOther)
	}

	testPostInvalidFormAuth(t, endpoint, testTokens[AdminID])

	testPost(t, endpoint, nil, http.StatusUnauthorized)
	testPostAuth(t, endpoint, testInvalidToken, nil, http.StatusUnauthorized)
}

func TestCourseCreatePageHandler(t *testing.T) {
	testCourseCreateEditPageHandler(t, "/course/create")
}

func TestCourseEditPageHandler(t *testing.T) {
	testCourseCreateEditPageHandler(t, "/course/edit")
}
