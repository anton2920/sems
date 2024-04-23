package main

import (
	"encoding/gob"
	"fmt"
	"strconv"
	"time"
	"unsafe"
)

type (
	SubmittedQuestion struct {
		SelectedAnswers []int
	}
	SubmittedTest struct {
		Test               *StepTest
		SubmittedQuestions []SubmittedQuestion
	}

	ProgrammingLanguage struct {
		Name         string
		Compiler     string
		CompilerArgs []string
		Runner       string
		RunnerArgs   []string
		SourceFile   string
		Executable   string
		Available    bool
	}
	SubmittedProgramming struct {
		Task       *StepProgramming
		LanguageID int
		Solution   string
	}

	Submission struct {
		Name           string
		User           *User
		Steps          []interface{}
		StartedAt      time.Time
		SubmittedSteps []interface{}
		FinishedAt     time.Time

		/* TODO(anton2920): I don't like this. Replace with 'pointer|1'. */
		Draft bool
	}
)

const (
	MinSolutionLen = 1
	MaxSolutionLen = 1024
)

var ProgrammingLanguages = []ProgrammingLanguage{
	{"c", "cc", nil, "", nil, "main.c", "./a.out", true},
	{"c++", "c++", nil, "", nil, "main.cpp", "./a.out", true},
	{"go", "sh", []string{"-c", "/usr/local/bin/go-build"}, "", nil, "main.go", "./main", false},
	{"php", "php", []string{"-l"}, "php", nil, "main.php", "", true},
	{"python3", "python3", []string{"-c", `import ast; ast.parse(open("main.py").read())`}, "python3", nil, "main.py", "", true},
}

func init() {
	gob.Register(&SubmittedTest{})
	gob.Register(&SubmittedProgramming{})
}

func SubmissionMainPageHandler(w *HTTPResponse, r *HTTPRequest, submission *Submission) error {
	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>`)
	w.AppendString(`Submission for `)
	w.WriteHTMLString(submission.Name)
	w.AppendString(` by `)
	w.WriteHTMLString(submission.User.LastName)
	w.AppendString(` `)
	w.WriteHTMLString(submission.User.FirstName)
	w.AppendString(`</title></head>`)
	w.AppendString(`<body>`)

	w.AppendString(`<h1>`)
	w.AppendString(`Submission for `)
	w.WriteHTMLString(submission.Name)
	w.AppendString(` by `)
	w.WriteHTMLString(submission.User.LastName)
	w.AppendString(` `)
	w.WriteHTMLString(submission.User.FirstName)
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

		if submission.SubmittedSteps[i] == nil {
			w.AppendString(`<p><i>Student skipped this step.</i></p>`)
		} else {
			w.AppendString(`<input type="submit" name="Command`)
			w.WriteInt(i)
			w.AppendString(`" value="Open">`)
		}

		w.AppendString(`</fieldset>`)
		w.AppendString(`<br>`)
	}

	w.AppendString(`</form>`)

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil

}

func SubmissionTestPageHandler(w *HTTPResponse, r *HTTPRequest, submittedTest *SubmittedTest) error {
	test := submittedTest.Test
	teacher := r.Form.Get("Teacher") != ""

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>`)
	w.AppendString(`Submitted test: `)
	w.WriteHTMLString(test.Name)
	w.AppendString(`</title></head>`)
	w.AppendString(`<body>`)

	w.AppendString(`<h1>`)
	w.AppendString(`Submitted test: `)
	w.WriteHTMLString(test.Name)
	w.AppendString(`</h1>`)

	if teacher {
		w.AppendString(`<p><i>Note: answers marked with [x] are correct.</i></p>`)
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

			for k := 0; k < len(submittedTest.SubmittedQuestions[i].SelectedAnswers); k++ {
				selectedAnswer := submittedTest.SubmittedQuestions[i].SelectedAnswers[k]
				if j == selectedAnswer {
					w.AppendString(` checked`)
					break
				}
			}

			w.AppendString(` disabled>`)
			w.AppendString("\r\n")

			w.AppendString(`<span>`)
			w.WriteHTMLString(answer)
			w.AppendString(`</span>`)

			if teacher {
				for k := 0; k < len(question.CorrectAnswers); k++ {
					correctAnswer := question.CorrectAnswers[k]
					if j == correctAnswer {
						w.AppendString("\r\n")
						w.AppendString(`<span>[x]</span>`)
						break
					}
				}
			}

			w.AppendString(`</li>`)
		}
		w.AppendString(`</ol>`)

		w.AppendString(`</fieldset>`)
		w.AppendString(`<br>`)
	}

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func SubmissionProgrammingPageHandler(w *HTTPResponse, r *HTTPRequest, submittedTask *SubmittedProgramming) error {
	task := submittedTask.Task
	teacher := r.Form.Get("Teacher") != ""

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>`)
	w.AppendString(`Submitted programming task: `)
	w.WriteHTMLString(task.Name)
	w.AppendString(`</title></head>`)
	w.AppendString(`<body>`)

	w.AppendString(`<h1>`)
	w.AppendString(`Submitted programming task: `)
	w.WriteHTMLString(task.Name)
	w.AppendString(`</h1>`)

	DisplayErrorMessage(w, r.Form.Get("Error"))

	w.AppendString(`<h2>Description</h2>`)
	w.AppendString(`<p>`)
	w.WriteHTMLString(task.Description)
	w.AppendString(`</p>`)

	w.AppendString(`<h2>Examples</h2>`)
	SubmissionNewProgrammingDisplayChecks(w, task, CheckTypeExample)

	w.AppendString(`<h2>Solution</h2>`)

	w.AppendString(`<label>Programming language: `)
	/* TODO(anton2920): add list of programming languages. */
	w.AppendString(`<select name="LanguageID" disabled><option value="0">C</option><option value="1">Go</option></select>`)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<textarea cols="80" rows="24" name="Solution" readonly>`)
	w.WriteHTMLString(submittedTask.Solution)
	w.AppendString(`</textarea>`)

	if teacher {
		w.AppendString(`<h2>Tests</h2>`)
		SubmissionNewProgrammingDisplayChecks(w, task, CheckTypeExample)
	}

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func SubmissionHandleCommand(w *HTTPResponse, r *HTTPRequest, submission *Submission, currentPage, k, command string) error {
	pindex, _, _, _, err := GetIndicies(k[len("Command"):])
	if err != nil {
		return ClientError(err)
	}

	switch currentPage {
	default:
		return ClientError(nil)
	case "Main":
		switch command {
		default:
			return ClientError(nil)
		case "Open":
			if (pindex < 0) || (pindex >= len(submission.Steps)) {
				return ClientError(nil)
			}

			switch submittedStep := submission.SubmittedSteps[pindex].(type) {
			default:
				panic("invalid step type")
			case *SubmittedTest:
				return SubmissionTestPageHandler(w, r, submittedStep)
			case *SubmittedProgramming:
				return SubmissionProgrammingPageHandler(w, r, submittedStep)
			}
		}
	}
}

func SubmissionPageHandler(w *HTTPResponse, r *HTTPRequest) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}

	if err := r.ParseForm(); err != nil {
		return ClientError(err)
	}

	subjectID, err := GetValidIndex(r.Form.Get("ID"), DB.Subjects)
	if err != nil {
		return ClientError(err)
	}
	subject := &DB.Subjects[subjectID]
	switch WhoIsUserInSubject(session.ID, subject) {
	default:
		r.Form.Set("Teacher", "")
	case SubjectUserAdmin, SubjectUserTeacher:
		r.Form.Set("Teacher", "yay")
	}

	li, err := GetValidIndex(r.Form.Get("LessonIndex"), subject.Lessons)
	if err != nil {
		return ClientError(err)
	}
	lesson := subject.Lessons[li]

	si, err := GetValidIndex(r.Form.Get("SubmissionIndex"), lesson.Submissions)
	if err != nil {
		return ClientError(err)
	}
	submission := lesson.Submissions[si]

	currentPage := r.Form.Get("CurrentPage")
	nextPage := r.Form.Get("NextPage")

	for i := 0; i < len(r.Form); i++ {
		k := r.Form[i].Key
		v := r.Form[i].Values[0]

		/* 'command' is button, which modifies content of a current page. */
		if StringStartsWith(k, "Command") {
			/* NOTE(anton2920): after command is executed, function must return. */
			return SubmissionHandleCommand(w, r, submission, currentPage, k, v)
		}
	}

	switch nextPage {
	default:
		return SubmissionMainPageHandler(w, r, submission)
	case "Discard":
		return SubmissionDiscardHandler(w, r)
	}
}

func SubmissionNewTestVerifyRequest(vs URLValues, submittedTest *SubmittedTest) error {
	test := submittedTest.Test

	selectedAnswerKey := make([]byte, 30)
	copy(selectedAnswerKey, "SelectedAnswer")

	for i := 0; i < len(test.Questions); i++ {
		question := &test.Questions[i]
		passedQuestion := &submittedTest.SubmittedQuestions[i]

		n := SlicePutInt(selectedAnswerKey[len("SelectedAnswer"):], i)
		selectedAnswers := vs.GetMany(unsafe.String(unsafe.SliceData(selectedAnswerKey), len("SelectedAnswer")+n))
		if len(selectedAnswers) == 0 {
			return BadRequest(fmt.Sprintf("question %d: select at least one answer", i+1))
		}
		if (len(question.CorrectAnswers) == 1) && (len(selectedAnswers) > 1) {
			return ClientError(nil)
		}
		for j := 0; j < len(selectedAnswers); j++ {
			if j >= len(passedQuestion.SelectedAnswers) {
				passedQuestion.SelectedAnswers = append(passedQuestion.SelectedAnswers, 0)
			}

			var err error
			passedQuestion.SelectedAnswers[j], err = GetValidIndex(selectedAnswers[j], question.Answers)
			if err != nil {
				return ClientError(err)
			}
		}
	}

	return nil
}

func SubmissionNewProgrammingVerifyRequest(vs URLValues, submittedTask *SubmittedProgramming) error {
	var err error

	/* TODO(anton2920): replace with actual check for programming language. */
	submittedTask.LanguageID, err = strconv.Atoi(vs.Get("LanguageID"))
	if (err != nil) || (submittedTask.LanguageID < 0) || (submittedTask.LanguageID >= 2) {
		return ClientError(err)
	}

	submittedTask.Solution = vs.Get("Solution")
	if !StringLengthInRange(submittedTask.Solution, MinSolutionLen, MaxSolutionLen) {
		return BadRequest(fmt.Sprintf("solution length must be between %d and %d characters long", MinSolutionLen, MaxSolutionLen))
	}

	return nil
}

func SubmissionNewMainPageHandler(w *HTTPResponse, r *HTTPRequest, submission *Submission) error {
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

	DisplayErrorMessage(w, r.Form.Get("Error"))

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
		if submission.SubmittedSteps[i] == nil {
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

func SubmissionNewTestPageHandler(w *HTTPResponse, r *HTTPRequest, submittedTest *SubmittedTest) error {
	test := submittedTest.Test

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

	DisplayErrorMessage(w, r.Form.Get("Error"))

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

	if submittedTest.SubmittedQuestions == nil {
		submittedTest.SubmittedQuestions = make([]SubmittedQuestion, len(test.Questions))
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

			for k := 0; k < len(submittedTest.SubmittedQuestions[i].SelectedAnswers); k++ {
				selectedAnswer := submittedTest.SubmittedQuestions[i].SelectedAnswers[k]
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

func SubmissionNewProgrammingDisplayChecks(w *HTTPResponse, task *StepProgramming, checkType CheckType) {
	w.AppendString(`<ol>`)
	for i := 0; i < len(task.Checks[checkType]); i++ {
		check := &task.Checks[checkType][i]

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
}

func SubmissionNewProgrammingPageHandler(w *HTTPResponse, r *HTTPRequest, submittedTask *SubmittedProgramming) error {
	task := submittedTask.Task

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

	DisplayErrorMessage(w, r.Form.Get("Error"))

	w.AppendString(`<h2>Description</h2>`)
	w.AppendString(`<p>`)
	w.WriteHTMLString(task.Description)
	w.AppendString(`</p>`)

	w.AppendString(`<h2>Examples</h2>`)
	SubmissionNewProgrammingDisplayChecks(w, task, CheckTypeExample)

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
	w.WriteHTMLString(submittedTask.Solution)
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

func SubmissionNewHandleCommand(w *HTTPResponse, r *HTTPRequest, submission *Submission, currentPage, k, command string) error {
	pindex, spindex, _, _, err := GetIndicies(k[len("Command"):])
	if err != nil {
		return ClientError(err)
	}

	switch currentPage {
	default:
		return ClientError(nil)
	case "Main":
		switch command {
		default:
			return ClientError(nil)
		case "Pass", "Edit":
			if (pindex < 0) || (pindex >= len(submission.Steps)) {
				return ClientError(nil)
			}

			switch step := submission.Steps[pindex].(type) {
			default:
				panic("invalid step type")
			case *StepTest:
				submittedStep, ok := submission.SubmittedSteps[pindex].(*SubmittedTest)
				if !ok {
					submittedStep = new(SubmittedTest)
					submittedStep.Test = step
					submission.SubmittedSteps[pindex] = submittedStep
				}

				r.Form.Set("StepIndex", spindex)
				return SubmissionNewTestPageHandler(w, r, submittedStep)
			case *StepProgramming:
				submittedStep, ok := submission.SubmittedSteps[pindex].(*SubmittedProgramming)
				if !ok {
					submittedStep = new(SubmittedProgramming)
					submittedStep.Task = step
					submission.SubmittedSteps[pindex] = submittedStep
				}

				r.Form.Set("StepIndex", spindex)
				return SubmissionNewProgrammingPageHandler(w, r, submittedStep)
			}
		}
	}
}

func SubmissionNewPageHandler(w *HTTPResponse, r *HTTPRequest) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}

	if err := r.ParseForm(); err != nil {
		return ClientError(err)
	}

	subjectID, err := GetValidIndex(r.Form.Get("ID"), DB.Subjects)
	if err != nil {
		return ClientError(err)
	}
	subject := &DB.Subjects[subjectID]

	li, err := GetValidIndex(r.Form.Get("LessonIndex"), subject.Lessons)
	if err != nil {
		return ClientError(err)
	}
	lesson := subject.Lessons[li]

	if WhoIsUserInSubject(session.ID, subject) != SubjectUserStudent {
		return ForbiddenError
	}

	submissionIndex := r.Form.Get("SubmissionIndex")
	var submission *Submission
	if submissionIndex == "" {
		submission = new(Submission)
		submission.Draft = true
		submission.Name = lesson.Name
		submission.StartedAt = time.Now()
		submission.User = &DB.Users[session.ID]
		StepsDeepCopy(&submission.Steps, lesson.Steps)
		submission.SubmittedSteps = make([]interface{}, len(lesson.Steps))

		lesson.Submissions = append(lesson.Submissions, submission)
		r.Form.SetInt("SubmissionIndex", len(lesson.Submissions)-1)
	} else {
		si, err := GetValidIndex(submissionIndex, lesson.Submissions)
		if err != nil {
			return ClientError(err)
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
			return SubmissionNewHandleCommand(w, r, submission, currentPage, k, v)
		}
	}

	stepIndex := r.Form.Get("StepIndex")
	if stepIndex != "" {
		si, err := GetValidIndex(r.Form.Get("StepIndex"), lesson.Steps)
		if err != nil {
			return ClientError(err)
		}
		if nextPage != "Discard" {
			switch currentPage {
			case "Test":
				submittedTest, ok := submission.SubmittedSteps[si].(*SubmittedTest)
				if !ok {
					return ClientError(nil)
				}

				if err := SubmissionNewTestVerifyRequest(r.Form, submittedTest); err != nil {
					return WritePageEx(w, r, SubmissionNewTestPageHandler, submittedTest, err)
				}
			case "Programming":
				passedProgramming, ok := submission.SubmittedSteps[si].(*SubmittedProgramming)
				if !ok {
					return ClientError(nil)
				}

				if err := SubmissionNewProgrammingVerifyRequest(r.Form, passedProgramming); err != nil {
					return WritePageEx(w, r, SubmissionNewProgrammingPageHandler, passedProgramming, err)
				}
			}
		} else {
			submission.SubmittedSteps[si] = nil
		}
	}

	switch nextPage {
	default:
		return SubmissionNewMainPageHandler(w, r, submission)
	case "Finish":
		return SubmissionNewHandler(w, r)
	}
}

func SubmissionDiscardHandler(w *HTTPResponse, r *HTTPRequest) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}

	if err := r.ParseForm(); err != nil {
		return ClientError(err)
	}

	subjectID, err := GetValidIndex(r.Form.Get("ID"), DB.Subjects)
	if err != nil {
		return ClientError(err)
	}
	subject := &DB.Subjects[subjectID]
	if (session.ID != AdminID) && (session.ID != subject.Teacher.ID) {
		return ForbiddenError
	}

	li, err := GetValidIndex(r.Form.Get("LessonIndex"), subject.Lessons)
	if err != nil {
		return ClientError(err)
	}
	lesson := subject.Lessons[li]

	si, err := GetValidIndex(r.Form.Get("SubmissionIndex"), lesson.Submissions)
	if err != nil {
		return ClientError(err)
	}
	lesson.Submissions = RemoveAtIndex(lesson.Submissions, si)

	w.RedirectID("/subject/", subjectID, HTTPStatusSeeOther)
	return nil
}

func SubmissionNewHandler(w *HTTPResponse, r *HTTPRequest) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}

	if err := r.ParseForm(); err != nil {
		return ClientError(err)
	}

	subjectID, err := GetValidIndex(r.Form.Get("ID"), DB.Subjects)
	if err != nil {
		return ClientError(err)
	}
	subject := &DB.Subjects[subjectID]
	if WhoIsUserInSubject(session.ID, subject) != SubjectUserStudent {
		return ForbiddenError
	}

	li, err := GetValidIndex(r.Form.Get("LessonIndex"), subject.Lessons)
	if err != nil {
		return ClientError(err)
	}
	lesson := subject.Lessons[li]

	si, err := GetValidIndex(r.Form.Get("SubmissionIndex"), lesson.Submissions)
	if err != nil {
		return ClientError(err)
	}
	submission := lesson.Submissions[si]

	empty := true
	for i := 0; i < len(submission.SubmittedSteps); i++ {
		if submission.SubmittedSteps[i] != nil {
			empty = false
			break
		}
	}
	if empty {
		return WritePageEx(w, r, SubmissionNewMainPageHandler, submission, BadRequest("you have to pass at least one step"))
	}

	submission.Draft = false
	submission.FinishedAt = time.Now()

	/* TODO(anton2920): send for verification. */

	w.RedirectID("/subject/", subjectID, HTTPStatusSeeOther)
	return nil
}
