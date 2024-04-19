package main

import (
	"encoding/gob"
	"fmt"
	"strconv"
	"time"
	"unsafe"
)

type (
	PassedQuestion struct {
		SelectedAnswers []int
	}
	PassedTest struct {
		Test            *StepTest
		PassedQuestions []PassedQuestion
	}

	PassedProgramming struct {
		Task       *StepProgramming
		LanguageID int
		Solution   string
	}

	Submission struct {
		Name        string
		User        *User
		Steps       []interface{}
		StartedAt   time.Time
		PassedSteps []interface{}
		FinishedAt  time.Time
	}
)

const (
	MinSolutionLen = 1
	MaxSolutionLen = 1024
)

func init() {
	gob.Register(&PassedTest{})
	gob.Register(&PassedProgramming{})
}

func EvaluationPassTestVerifyRequest(vs URLValues, passedTest *PassedTest) error {
	test := passedTest.Test

	for i := 0; i < len(test.Questions); i++ {
		question := &test.Questions[i]
		passedQuestion := &passedTest.PassedQuestions[i]

		buffer := fmt.Appendf(make([]byte, 0, 30), "SelectedAnswer%d", i)
		selectedAnswers := vs.GetMany(unsafe.String(unsafe.SliceData(buffer), len(buffer)))
		if len(selectedAnswers) == 0 {
			return NewHTTPError(HTTPStatusBadRequest, fmt.Sprintf("question %d: select at least one answer", i+1))
		}
		if (len(question.CorrectAnswers) == 1) && (len(selectedAnswers) > 1) {
			return ReloadPageError
		}
		for j := 0; j < len(selectedAnswers); j++ {
			if j >= len(passedQuestion.SelectedAnswers) {
				passedQuestion.SelectedAnswers = append(passedQuestion.SelectedAnswers, 0)
			}

			var err error
			passedQuestion.SelectedAnswers[j], err = strconv.Atoi(selectedAnswers[j])
			if (err != nil) || (passedQuestion.SelectedAnswers[j] < 0) || (passedQuestion.SelectedAnswers[j] >= len(question.Answers)) {
				return ReloadPageError
			}
		}
	}

	return nil
}

func EvaluationPassProgrammingVerifyRequest(vs URLValues, passedTask *PassedProgramming) error {
	var err error

	passedTask.LanguageID, err = strconv.Atoi(vs.Get("LanguageID"))
	/* TODO(anton2920): replace with actual check for programming language. */
	if (err != nil) || (passedTask.LanguageID < 0) || (passedTask.LanguageID >= 2) {
		return ReloadPageError
	}

	passedTask.Solution = vs.Get("Solution")
	if !StringLengthInRange(passedTask.Solution, MinSolutionLen, MaxSolutionLen) {
		return NewHTTPError(HTTPStatusBadRequest, fmt.Sprintf("solution length must be between %d and %d characters long", MinSolutionLen, MaxSolutionLen))
	}

	return nil
}

func EvaluationPassMainPageHandler(w *HTTPResponse, r *HTTPRequest, submission *Submission) error {
	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>`)
	w.AppendString(`Evaluation for `)
	w.WriteHTMLString(submission.Name)
	w.AppendString(`</title></head>`)
	w.AppendString(`<body>`)

	w.AppendString(`<h1>`)
	w.AppendString(`Evaluation for `)
	w.WriteHTMLString(submission.Name)
	w.AppendString(`</h1>`)

	w.AppendString(`<form style="min-width: 300px; max-width: max-content;" method="POST" action="`)
	w.WriteString(r.URL.Path)
	w.AppendString(`">`)

	w.AppendString(`<input type="hidden" name="CurrentPage" value="Main">`)

	w.AppendString(`<input type="hidden" name="ID" value="`)
	w.WriteHTMLString(r.Form.Get("ID"))
	w.AppendString(`">`)

	w.AppendString(`<input type="hidden" name="LessonIndex" value="`)
	w.WriteHTMLString(r.Form.Get("LessonIndex"))
	w.AppendString(`">`)

	w.AppendString(`<input type="hidden" name="SubmissionIndex" value="`)
	w.WriteHTMLString(r.Form.Get("SubmissionIndex"))
	w.AppendString(`">`)

	for i := 0; i < len(submission.Steps); i++ {
		var name, stepType string

		step := submission.Steps[i]
		switch step := step.(type) {
		default:
			panic("invalid step type")
		case *StepTest:
			name = step.Name
			stepType = "Test"
		case *StepProgramming:
			name = step.Name
			stepType = "Programming task"
		}

		w.AppendString(`<fieldset>`)

		w.AppendString(`<legend>Step #`)
		w.WriteInt(i + 1)
		w.AppendString(`</legend>`)

		w.AppendString(`<p>Name: `)
		w.WriteHTMLString(name)
		w.AppendString(`</p>`)

		w.AppendString(`<p>Type: `)
		w.AppendString(stepType)
		w.AppendString(`</p>`)

		w.AppendString(`<input type="submit" name="Command`)
		w.WriteInt(i)
		w.AppendString(`" value="`)
		if submission.PassedSteps[i] == nil {
			w.AppendString(`Pass`)
		} else {
			w.AppendString(`Edit`)
		}
		w.AppendString(`">`)

		w.AppendString(`</fieldset>`)
		w.AppendString(`<br>`)
	}

	w.AppendString(`<input type="submit" name="NextPage" value="Finish">`)

	w.AppendString(`</form>`)

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil

}

func EvaluationPassTestPageHandler(w *HTTPResponse, r *HTTPRequest, passedTest *PassedTest) error {
	test := passedTest.Test

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>`)
	w.AppendString(`Test: `)
	w.WriteHTMLString(test.Name)
	w.AppendString(`</title></head>`)
	w.AppendString(`<body>`)

	w.AppendString(`<h1>`)
	w.AppendString(`Test: `)
	w.WriteHTMLString(test.Name)
	w.AppendString(`</h1>`)

	ErrorDiv(w, r.Form.Get("Error"))

	w.AppendString(`<form style="min-width: 300px; max-width: max-content;" method="POST" action="`)
	w.WriteString(r.URL.Path)
	w.AppendString(`">`)

	w.AppendString(`<input type="hidden" name="CurrentPage" value="Test">`)

	w.AppendString(`<input type="hidden" name="ID" value="`)
	w.WriteHTMLString(r.Form.Get("ID"))
	w.AppendString(`">`)

	w.AppendString(`<input type="hidden" name="LessonIndex" value="`)
	w.WriteHTMLString(r.Form.Get("LessonIndex"))
	w.AppendString(`">`)

	w.AppendString(`<input type="hidden" name="SubmissionIndex" value="`)
	w.WriteHTMLString(r.Form.Get("SubmissionIndex"))
	w.AppendString(`">`)

	w.AppendString(`<input type="hidden" name="StepIndex" value="`)
	w.WriteHTMLString(r.Form.Get("StepIndex"))
	w.AppendString(`">`)

	if passedTest.PassedQuestions == nil {
		passedTest.PassedQuestions = make([]PassedQuestion, len(test.Questions))
	}
	for i := 0; i < len(test.Questions); i++ {
		question := &test.Questions[i]

		w.AppendString(`<fieldset>`)
		w.AppendString(`<legend>Question #`)
		w.WriteInt(i + 1)
		w.AppendString(`</legend>`)
		w.AppendString(`<p>`)
		w.WriteHTMLString(question.Name)
		w.AppendString(`</p>`)

		w.AppendString(`<ol>`)
		for j := 0; j < len(question.Answers); j++ {
			answer := question.Answers[j]

			if j > 0 {
				w.AppendString(`<br>`)
			}

			w.AppendString(`<li>`)

			w.AppendString(`<input type="`)
			if len(question.CorrectAnswers) > 1 {
				w.AppendString(`checkbox`)
			} else {
				w.AppendString(`radio`)
			}
			w.AppendString(`" name="SelectedAnswer`)
			w.WriteInt(i)
			w.AppendString(`" value="`)
			w.WriteInt(j)
			w.AppendString(`"`)

			for k := 0; k < len(passedTest.PassedQuestions[i].SelectedAnswers); k++ {
				selectedAnswer := passedTest.PassedQuestions[i].SelectedAnswers[k]
				if j == selectedAnswer {
					w.AppendString(` checked`)
					break
				}
			}

			w.AppendString(`>`)
			w.AppendString("\r\n")

			w.AppendString(`<span>`)
			w.WriteHTMLString(answer)
			w.AppendString(`</span>`)

			w.AppendString(`</li>`)
		}
		w.AppendString(`</ol>`)

		w.AppendString(`</fieldset>`)
		w.AppendString(`<br>`)
	}

	w.AppendString(`<input type="submit" name="NextPage" value="Save">`)
	w.AppendString("\r\n")
	w.AppendString(`<input type="submit" name="NextPage" value="Discard">`)

	w.AppendString(`</form>`)

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func EvaluationPassProgrammingPageHandler(w *HTTPResponse, r *HTTPRequest, passedTask *PassedProgramming) error {
	task := passedTask.Task

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>`)
	w.AppendString(`Programming task: `)
	w.WriteHTMLString(task.Name)
	w.AppendString(`</title></head>`)
	w.AppendString(`<body>`)

	w.AppendString(`<h1>`)
	w.AppendString(`Programming task: `)
	w.WriteHTMLString(task.Name)
	w.AppendString(`</h1>`)

	ErrorDiv(w, r.Form.Get("Error"))

	w.AppendString(`<h2>Description</h2>`)
	w.AppendString(`<p>`)
	w.WriteHTMLString(task.Description)
	w.AppendString(`</p>`)

	w.AppendString(`<h2>Examples</h2>`)

	w.AppendString(`<ol>`)
	for i := 0; i < len(task.Checks[CheckTypeExample]); i++ {
		check := &task.Checks[CheckTypeExample][i]

		w.AppendString(`<li>`)

		w.AppendString(`<label>Input: `)

		w.AppendString(`<textarea rows="1" readonly>`)
		w.WriteHTMLString(check.Input)
		w.AppendString(`</textarea>`)

		w.AppendString(`</label>`)
		w.AppendString("\r\n")

		w.AppendString(`<label>output: `)

		w.AppendString(`<textarea rows="1" readonly>`)
		w.WriteHTMLString(check.Output)
		w.AppendString(`</textarea>`)

		w.AppendString(`</label>`)

		w.AppendString(`</li>`)
	}
	w.AppendString(`</ol>`)

	w.AppendString(`<form method="POST" action="`)
	w.WriteString(r.URL.Path)
	w.AppendString(`">`)

	w.AppendString(`<input type="hidden" name="CurrentPage" value="Programming">`)

	w.AppendString(`<input type="hidden" name="ID" value="`)
	w.WriteHTMLString(r.Form.Get("ID"))
	w.AppendString(`">`)

	w.AppendString(`<input type="hidden" name="LessonIndex" value="`)
	w.WriteHTMLString(r.Form.Get("LessonIndex"))
	w.AppendString(`">`)

	w.AppendString(`<input type="hidden" name="SubmissionIndex" value="`)
	w.WriteHTMLString(r.Form.Get("SubmissionIndex"))
	w.AppendString(`">`)

	w.AppendString(`<input type="hidden" name="StepIndex" value="`)
	w.WriteHTMLString(r.Form.Get("StepIndex"))
	w.AppendString(`">`)

	w.AppendString(`<h2>Solution</h2>`)

	w.AppendString(`<label>Programming language: `)
	/* TODO(anton2920): add list of programming languages. */
	w.AppendString(`<select name="LanguageID"><option value="0">C</option><option value="1">Go</option></select>`)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<textarea cols="80" rows="24" name="Solution">`)
	w.WriteHTMLString(passedTask.Solution)
	w.AppendString(`</textarea>`)

	w.AppendString(`<br><br>`)
	w.AppendString(`<input type="submit" name="NextPage" value="Save">`)
	w.WriteHTMLString("\r\n")
	w.AppendString(`<input type="submit" name="NextPage" value="Discard">`)

	w.AppendString(`</form>`)
	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func EvaluationPassHandleCommand(w *HTTPResponse, r *HTTPRequest, submission *Submission, currentPage, k, command string) error {
	pindex, spindex, _, _, err := GetIndicies(k[len("Command"):])
	if err != nil {
		return ReloadPageError
	}

	switch currentPage {
	default:
		return ReloadPageError
	case "Main":
		switch command {
		default:
			return ReloadPageError
		case "Pass", "Edit":
			if (pindex < 0) || (pindex >= len(submission.Steps)) {
				return ReloadPageError
			}

			switch step := submission.Steps[pindex].(type) {
			default:
				panic("invalid step type")
			case *StepTest:
				passedStep, ok := submission.PassedSteps[pindex].(*PassedTest)
				if !ok {
					passedStep = new(PassedTest)
					passedStep.Test = step
					submission.PassedSteps[pindex] = passedStep
				}

				r.Form.Set("StepIndex", spindex)
				return EvaluationPassTestPageHandler(w, r, passedStep)
			case *StepProgramming:
				passedStep, ok := submission.PassedSteps[pindex].(*PassedProgramming)
				if !ok {
					passedStep = new(PassedProgramming)
					passedStep.Task = step
					submission.PassedSteps[pindex] = passedStep
				}

				r.Form.Set("StepIndex", spindex)
				return EvaluationPassProgrammingPageHandler(w, r, passedStep)
			}
		}
	}
}

func EvaluationPassPageHandler(w *HTTPResponse, r *HTTPRequest) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}

	if err := r.ParseForm(); err != nil {
		return ReloadPageError
	}

	subjectID, err := strconv.Atoi(r.Form.Get("ID"))
	if (err != nil) || (subjectID < 0) || (subjectID >= len(DB.Subjects)) {
		return ReloadPageError
	}
	subject := &DB.Subjects[subjectID]

	li, err := strconv.Atoi(r.Form.Get("LessonIndex"))
	if (err != nil) || (li < 0) || (li >= len(subject.Lessons)) {
		return ReloadPageError
	}
	lesson := subject.Lessons[li]

	if WhoIsUserInSubject(session.ID, subject) != SubjectUserStudent {
		return ForbiddenError
	}

	var submission *Submission
	submissionIndex := r.Form.Get("SubmissionIndex")
	if submissionIndex == "" {
		submission = new(Submission)
		submission.Name = lesson.Name
		submission.StartedAt = time.Now()
		submission.User = &DB.Users[session.ID]
		StepsDeepCopy(&submission.Steps, lesson.Steps)
		submission.PassedSteps = make([]interface{}, len(lesson.Steps))

		lesson.Submissions = append(lesson.Submissions, submission)
		r.Form.Set("SubmissionIndex", strconv.Itoa(len(lesson.Submissions)-1))
	} else {
		si, err := strconv.Atoi(submissionIndex)
		if (err != nil) || (si < 0) || (si >= len(lesson.Submissions)) {
			return ReloadPageError
		}
		submission = lesson.Submissions[si]
	}

	currentPage := r.Form.Get("CurrentPage")
	nextPage := r.Form.Get("NextPage")

	for i := 0; i < len(r.Form); i++ {
		k := r.Form[i].Key
		v := r.Form[i].Values[0]

		/* 'command' is button, which modifies content of a current page. */
		if StringStartsWith(k, "Command") {
			/* NOTE(anton2920): after command is executed, function must return. */
			return EvaluationPassHandleCommand(w, r, submission, currentPage, k, v)
		}
	}

	stepIndex := r.Form.Get("StepIndex")
	if stepIndex != "" {
		si, err := strconv.Atoi(r.Form.Get("StepIndex"))
		if (err != nil) || (si < 0) || (si >= len(lesson.Steps)) {
			return ReloadPageError
		}
		if nextPage != "Discard" {
			switch currentPage {
			case "Test":
				passedTest, ok := submission.PassedSteps[si].(*PassedTest)
				if !ok {
					return ReloadPageError
				}

				if err := EvaluationPassTestVerifyRequest(r.Form, passedTest); err != nil {
					return WritePageEx(w, r, EvaluationPassTestPageHandler, passedTest, err)
				}
			case "Programming":
				passedProgramming, ok := submission.PassedSteps[si].(*PassedProgramming)
				if !ok {
					return ReloadPageError
				}

				if err := EvaluationPassProgrammingVerifyRequest(r.Form, passedProgramming); err != nil {
					return WritePageEx(w, r, EvaluationPassProgrammingPageHandler, passedProgramming, err)
				}
			}
		} else {
			submission.PassedSteps[si] = nil
		}
	}

	switch nextPage {
	default:
		return EvaluationPassMainPageHandler(w, r, submission)
	case "Finish":
		submission.FinishedAt = time.Now()
		return EvaluationPassHandler(w, r)
	}

	return nil
}

func EvaluationPassHandler(w *HTTPResponse, r *HTTPRequest) error {
	id := r.Form.Get("ID")
	w.Redirect(fmt.Appendf(make([]byte, 0, 30), "/subject/%s", id), HTTPStatusSeeOther)
	return nil
}
