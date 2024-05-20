package main

import (
	"encoding/gob"
	"time"
	"unsafe"

	"github.com/anton2920/gofa/net/url"
	"github.com/anton2920/gofa/slices"

	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/strings"
)

type (
	SubmittedQuestion struct {
		SelectedAnswers []int
	}
	SubmittedTest struct {
		Test               *StepTest
		SubmittedQuestions []SubmittedQuestion

		Scores []int

		Status SubmissionCheckStatus

		/* TODO(anton2920): I don't like this. Replace with 'pointer|1'. */
		Draft bool
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
		Task     *StepProgramming
		Language *ProgrammingLanguage
		Solution string

		Scores   [2][]int
		Messages [2][]string
		Error    error

		Status SubmissionCheckStatus

		/* TODO(anton2920): I don't like this. Replace with 'pointer|1'. */
		Draft bool
	}

	Submission struct {
		LessonName string
		User       *User

		Steps          []interface{}
		StartedAt      time.Time
		SubmittedSteps []interface{}
		FinishedAt     time.Time

		Status SubmissionCheckStatus

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
	{"go", "sh", []string{"-c", "/usr/local/bin/go-build"}, "", nil, "main.go", "./main", true},
	{"php", "php", []string{"-l"}, "php", nil, "main.php", "", true},
	{"python3", "python3", []string{"-c", `import ast; ast.parse(open("main.py").read())`}, "python3", nil, "main.py", "", true},
}

func init() {
	gob.Register(&SubmittedTest{})
	gob.Register(&SubmittedProgramming{})
}

func GetSubmittedStepScore(step interface{}, submittedStep interface{}) (int, int) {
	var nsteps int
	switch step := step.(type) {
	default:
		panic("invalid step type")
	case *StepTest:
		nsteps = len(step.Questions)
	case *StepProgramming:
		nsteps = len(step.Checks[CheckTypeTest])
	}

	if submittedStep == nil {
		return 0, nsteps
	}

	var scores []int
	switch step := submittedStep.(type) {
	default:
		panic("invalid step type")
	case *SubmittedTest:
		scores = step.Scores
	case *SubmittedProgramming:
		scores = step.Scores[CheckTypeTest]
	}

	var score int
	for i := 0; i < len(scores); i++ {
		score += scores[i]
	}

	return score, nsteps
}

func DisplaySubmissionTotalScore(w *http.Response, submission *Submission) {
	var totalScore, totalMaximum int

	for i := 0; i < len(submission.SubmittedSteps); i++ {
		score, maximum := GetSubmittedStepScore(submission.Steps[i], submission.SubmittedSteps[i])
		totalScore += score
		totalMaximum += maximum
	}
	w.WriteInt(totalScore)
	w.AppendString(`/`)
	w.WriteInt(totalMaximum)
}

func SubmissionMainPageHandler(w *http.Response, r *http.Request, submission *Submission) error {
	teacher := r.Form.Get("Teacher") != ""

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>`)
	w.AppendString(`Submission for `)
	w.WriteHTMLString(submission.LessonName)
	w.AppendString(` by `)
	w.WriteHTMLString(submission.User.LastName)
	w.AppendString(` `)
	w.WriteHTMLString(submission.User.FirstName)
	w.AppendString(`</title></head>`)
	w.AppendString(`<body>`)

	w.AppendString(`<h1>`)
	w.AppendString(`Submission for `)
	w.WriteHTMLString(submission.LessonName)
	w.AppendString(` by `)
	w.WriteHTMLString(submission.User.LastName)
	w.AppendString(` `)
	w.WriteHTMLString(submission.User.FirstName)
	w.AppendString(`</h1>`)

	w.AppendString(`<form style="min-width: 300px; max-width: max-content;" method="POST" action="/submission">`)

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

		if i > 0 {
			w.AppendString(`<br>`)
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

		submittedStep := submission.SubmittedSteps[i]
		if submittedStep == nil {
			score, maximum := GetSubmittedStepScore(step, submittedStep)
			w.AppendString(`<p>Score: `)
			w.WriteInt(score)
			w.AppendString(`/`)
			w.WriteInt(maximum)
			w.AppendString(`</p>`)

			w.AppendString(`<p><i>This step has been skipped.</i></p>`)
		} else {
			var status SubmissionCheckStatus
			var err error

			switch submittedStep := submittedStep.(type) {
			case *SubmittedTest:
				status = submittedStep.Status
			case *SubmittedProgramming:
				status = submittedStep.Status
				err = submittedStep.Error
			}

			switch status {
			case SubmissionCheckPending:
				w.AppendString(`<p><i>Verification is pending...</i></p>`)
			case SubmissionCheckInProgress:
				w.AppendString(`<p><i>Verification is in progress...</i></p>`)
			case SubmissionCheckDone:
				score, maximum := GetSubmittedStepScore(step, submittedStep)
				w.AppendString(`<p>Score: `)
				w.WriteInt(score)
				w.AppendString(`/`)
				w.WriteInt(maximum)
				w.AppendString(`</p>`)

				if err != nil {
					w.AppendString(`<p>Error: `)
					w.WriteHTMLString(err.Error())
					w.AppendString(`</p>`)
				}
				DisplayIndexedCommand(w, i, "Open")
				if teacher {
					DisplayIndexedCommand(w, i, "Re-check")
				}
			}
		}

		w.AppendString(`</fieldset>`)
	}

	switch submission.Status {
	case SubmissionCheckPending:
		w.AppendString(`<p><i>Verification is pending...</i></p>`)
	case SubmissionCheckInProgress:
		w.AppendString(`<p><i>Verification is in progress...</i></p>`)
	case SubmissionCheckDone:
		w.AppendString(`<p>Total score: `)
		DisplaySubmissionTotalScore(w, submission)
		w.AppendString(`</p>`)
		if teacher {
			w.AppendString(`<input type="submit" name="Command" value="Re-check">`)
		}
	}
	w.AppendString(`</form>`)

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil

}

func SubmissionTestPageHandler(w *http.Response, r *http.Request, submittedTest *SubmittedTest) error {
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

			w.AppendString(` disabled> `)

			w.AppendString(`<span>`)
			w.WriteHTMLString(answer)
			w.AppendString(`</span>`)

			if teacher {
				for k := 0; k < len(question.CorrectAnswers); k++ {
					correctAnswer := question.CorrectAnswers[k]
					if j == correctAnswer {
						w.AppendString(` <span>[x]</span>`)
						break
					}
				}
			}

			w.AppendString(`</li>`)
		}
		w.AppendString(`</ol>`)

		w.AppendString(`<p>Score: `)
		w.WriteInt(submittedTest.Scores[i])
		w.AppendString(`/1</p>`)

		w.AppendString(`</fieldset>`)
		w.AppendString(`<br>`)
	}

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func SubmissionProgrammingDisplayChecks(w *http.Response, submittedTask *SubmittedProgramming, checkType CheckType) {
	task := submittedTask.Task
	scores := submittedTask.Scores[checkType]
	messages := submittedTask.Messages[checkType]

	w.AppendString(`<ol>`)
	for i := 0; i < len(task.Checks[checkType]); i++ {
		check := &task.Checks[checkType][i]
		score := scores[i]
		message := messages[i]

		w.AppendString(`<li>`)

		w.AppendString(`<label>Input: `)

		w.AppendString(`<textarea rows="1" readonly>`)
		w.WriteHTMLString(check.Input)
		w.AppendString(`</textarea>`)

		w.AppendString(`</label> `)

		w.AppendString(`<label>output: `)

		w.AppendString(`<textarea rows="1" readonly>`)
		w.WriteHTMLString(check.Output)
		w.AppendString(`</textarea>`)

		w.AppendString(`</label> `)

		w.AppendString(`<span>score: `)
		w.WriteInt(score)
		w.AppendString(`/1</span>`)

		if message != "" {
			w.AppendString(` <span>`)
			w.WriteHTMLString(message)
			w.AppendString(`</span>`)
		}

		w.AppendString(`</li>`)
	}
	w.AppendString(`</ol>`)
}

func SubmissionProgrammingPageHandler(w *http.Response, r *http.Request, submittedTask *SubmittedProgramming) error {
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
	w.AppendString(`<select name="LanguageID" disabled>`)
	for i := 0; i < len(ProgrammingLanguages); i++ {
		lang := &ProgrammingLanguages[i]

		w.AppendString(`<option value="`)
		w.WriteInt(i)
		w.AppendString(`"`)
		if lang == submittedTask.Language {
			w.AppendString(` selected`)
		}
		w.AppendString(`>`)
		w.AppendString(lang.Name)
		w.AppendString(`</option>`)
	}
	w.AppendString(`</select>`)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<textarea cols="80" rows="24" name="Solution" readonly>`)
	w.WriteHTMLString(submittedTask.Solution)
	w.AppendString(`</textarea>`)

	if teacher {
		w.AppendString(`<h2>Tests</h2>`)
		SubmissionProgrammingDisplayChecks(w, submittedTask, CheckTypeTest)
	}

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func SubmissionHandleCommand(w *http.Response, r *http.Request, submission *Submission, currentPage, k, command string) error {
	pindex, spindex, _, _, err := GetIndicies(k[len("Command"):])
	if err != nil {
		return http.ClientError(err)
	}

	switch currentPage {
	default:
		return http.ClientError(nil)
	case "Main":
		switch command {
		default:
			return http.ClientError(nil)
		case "Open":
			if (pindex < 0) || (pindex >= len(submission.Steps)) {
				return http.ClientError(nil)
			}

			switch submittedStep := submission.SubmittedSteps[pindex].(type) {
			default:
				panic("invalid step type")
			case *SubmittedTest:
				return SubmissionTestPageHandler(w, r, submittedStep)
			case *SubmittedProgramming:
				return SubmissionProgrammingPageHandler(w, r, submittedStep)
			}
		case "Re-check":
			if spindex != "" {
				if (pindex < 0) || (pindex >= len(submission.Steps)) {
					return http.ClientError(nil)
				}
				submission.Status = SubmissionCheckPending

				switch submittedStep := submission.SubmittedSteps[pindex].(type) {
				default:
					panic("invaid step type")
				case *SubmittedTest:
					submittedStep.Status = SubmissionCheckPending
				case *SubmittedProgramming:
					submittedStep.Status = SubmissionCheckPending
				}
			} else {
				submission.Status = SubmissionCheckPending

				for i := 0; i < len(submission.SubmittedSteps); i++ {
					submittedStep := submission.SubmittedSteps[i]
					if submittedStep != nil {
						switch submittedStep := submittedStep.(type) {
						default:
							panic("invaid step type")
						case *SubmittedTest:
							submittedStep.Status = SubmissionCheckPending
						case *SubmittedProgramming:
							submittedStep.Status = SubmissionCheckPending
						}
					}
				}
			}

			SubmissionVerifyChannel <- submission

			return SubmissionMainPageHandler(w, r, submission)
		}
	}
}

func SubmissionPageHandler(w *http.Response, r *http.Request) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}

	if err := r.ParseForm(); err != nil {
		return http.ClientError(err)
	}

	subjectID, err := GetValidIndex(r.Form.Get("ID"), DB.Subjects)
	if err != nil {
		return http.ClientError(err)
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
		return http.ClientError(err)
	}
	lesson := subject.Lessons[li]

	si, err := GetValidIndex(r.Form.Get("SubmissionIndex"), lesson.Submissions)
	if err != nil {
		return http.ClientError(err)
	}
	submission := lesson.Submissions[si]

	currentPage := r.Form.Get("CurrentPage")
	nextPage := r.Form.Get("NextPage")

	for i := 0; i < len(r.Form); i++ {
		k := r.Form[i].Key
		v := r.Form[i].Values[0]

		/* 'command' is button, which modifies content of a current page. */
		if strings.StartsWith(k, "Command") {
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

func SubmissionNewTestVerifyRequest(vs url.Values, submittedTest *SubmittedTest) error {
	test := submittedTest.Test

	selectedAnswerKey := make([]byte, 30)
	copy(selectedAnswerKey, "SelectedAnswer")

	for i := 0; i < len(test.Questions); i++ {
		question := &test.Questions[i]
		submittedQuestion := &submittedTest.SubmittedQuestions[i]

		n := slices.PutInt(selectedAnswerKey[len("SelectedAnswer"):], i)
		selectedAnswers := vs.GetMany(unsafe.String(unsafe.SliceData(selectedAnswerKey), len("SelectedAnswer")+n))
		if len(selectedAnswers) == 0 {
			return http.BadRequest("question %d: select at least one answer", i+1)
		}
		if (len(question.CorrectAnswers) == 1) && (len(selectedAnswers) > 1) {
			return http.ClientError(nil)
		}
		for j := 0; j < len(selectedAnswers); j++ {
			if j >= len(submittedQuestion.SelectedAnswers) {
				submittedQuestion.SelectedAnswers = append(submittedQuestion.SelectedAnswers, 0)
			}

			var err error
			submittedQuestion.SelectedAnswers[j], err = GetValidIndex(selectedAnswers[j], question.Answers)
			if err != nil {
				return http.ClientError(err)
			}
		}
	}

	return nil
}

func SubmissionNewProgrammingVerifyRequest(vs url.Values, submittedTask *SubmittedProgramming) error {
	id, err := GetValidIndex(vs.Get("LanguageID"), ProgrammingLanguages)
	if err != nil {
		return http.ClientError(err)
	}
	submittedTask.Language = &ProgrammingLanguages[id]
	if !submittedTask.Language.Available {
		return http.BadRequest("selected language is not available")
	}

	submittedTask.Solution = vs.Get("Solution")
	if !strings.LengthInRange(submittedTask.Solution, MinSolutionLen, MaxSolutionLen) {
		return http.BadRequest("solution length must be between %d and %d characters long", MinSolutionLen, MaxSolutionLen)
	}

	return nil
}

func SubmissionNewMainPageHandler(w *http.Response, r *http.Request, submission *Submission) error {
	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>`)
	w.AppendString(`Evaluation for `)
	w.WriteHTMLString(submission.LessonName)
	w.AppendString(`</title></head>`)
	w.AppendString(`<body>`)

	w.AppendString(`<h1>`)
	w.AppendString(`Evaluation for `)
	w.WriteHTMLString(submission.LessonName)
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
		var draft bool

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

		submittedStep := submission.SubmittedSteps[i]
		if submittedStep != nil {
			switch submittedStep := submittedStep.(type) {
			case *SubmittedTest:
				draft = submittedStep.Draft
			case *SubmittedProgramming:
				draft = submittedStep.Draft
			}
		}

		w.AppendString(`<fieldset>`)

		w.AppendString(`<legend>Step #`)
		w.WriteInt(i + 1)
		if draft {
			w.AppendString(` (draft)`)
		}
		w.AppendString(`</legend>`)

		w.AppendString(`<p>Name: `)
		w.WriteHTMLString(name)
		w.AppendString(`</p>`)

		w.AppendString(`<p>Type: `)
		w.AppendString(stepType)
		w.AppendString(`</p>`)

		if submission.SubmittedSteps[i] == nil {
			DisplayIndexedCommand(w, i, "Pass")
		} else {
			DisplayIndexedCommand(w, i, "Edit")
		}

		w.AppendString(`</fieldset>`)
		w.AppendString(`<br>`)
	}

	w.AppendString(`<input type="submit" name="NextPage" value="Finish">`)

	w.AppendString(`</form>`)

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil

}

func SubmissionNewTestPageHandler(w *http.Response, r *http.Request, submittedTest *SubmittedTest) error {
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

			w.AppendString(`> `)

			w.AppendString(`<span>`)
			w.WriteHTMLString(answer)
			w.AppendString(`</span>`)

			w.AppendString(`</li>`)
		}
		w.AppendString(`</ol>`)

		w.AppendString(`</fieldset>`)
		w.AppendString(`<br>`)
	}

	w.AppendString(`<input type="submit" name="NextPage" value="Save"> `)
	w.AppendString(`<input type="submit" name="NextPage" value="Discard">`)

	w.AppendString(`</form>`)

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func SubmissionNewProgrammingDisplayChecks(w *http.Response, task *StepProgramming, checkType CheckType) {
	w.AppendString(`<ol>`)
	for i := 0; i < len(task.Checks[checkType]); i++ {
		check := &task.Checks[checkType][i]

		w.AppendString(`<li>`)

		w.AppendString(`<label>Input: `)

		w.AppendString(`<textarea rows="1" readonly>`)
		w.WriteHTMLString(check.Input)
		w.AppendString(`</textarea>`)

		w.AppendString(`</label> `)

		w.AppendString(`<label>output: `)

		w.AppendString(`<textarea rows="1" readonly>`)
		w.WriteHTMLString(check.Output)
		w.AppendString(`</textarea>`)

		w.AppendString(`</label>`)

		w.AppendString(`</li>`)
	}
	w.AppendString(`</ol>`)
}

func SubmissionNewProgrammingPageHandler(w *http.Response, r *http.Request, submittedTask *SubmittedProgramming) error {
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
	w.AppendString(`<select name="LanguageID">`)
	for i := 0; i < len(ProgrammingLanguages); i++ {
		lang := &ProgrammingLanguages[i]

		w.AppendString(`<option value="`)
		w.WriteInt(i)
		w.AppendString(`"`)
		if lang == submittedTask.Language {
			w.AppendString(` selected`)
		}
		w.AppendString(`>`)
		w.AppendString(lang.Name)
		w.AppendString(`</option>`)
	}
	w.AppendString(`</select>`)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<textarea cols="80" rows="24" name="Solution">`)
	w.WriteHTMLString(submittedTask.Solution)
	w.AppendString(`</textarea>`)

	w.AppendString(`<br><br>`)
	w.AppendString(`<input type="submit" name="NextPage" value="Save"> `)
	w.AppendString(`<input type="submit" name="NextPage" value="Discard">`)

	w.AppendString(`</form>`)
	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func SubmissionNewHandleCommand(w *http.Response, r *http.Request, submission *Submission, currentPage, k, command string) error {
	pindex, spindex, _, _, err := GetIndicies(k[len("Command"):])
	if err != nil {
		return http.ClientError(err)
	}

	switch currentPage {
	default:
		return http.ClientError(nil)
	case "Main":
		switch command {
		default:
			return http.ClientError(nil)
		case "Pass", "Edit":
			if (pindex < 0) || (pindex >= len(submission.Steps)) {
				return http.ClientError(nil)
			}

			switch step := submission.Steps[pindex].(type) {
			default:
				panic("invalid step type")
			case *StepTest:
				submittedStep, ok := submission.SubmittedSteps[pindex].(*SubmittedTest)
				if !ok {
					submittedStep = new(SubmittedTest)
					submittedStep.Test = step
					submittedStep.Draft = true
					submission.SubmittedSteps[pindex] = submittedStep
				}

				r.Form.Set("StepIndex", spindex)
				return SubmissionNewTestPageHandler(w, r, submittedStep)
			case *StepProgramming:
				submittedStep, ok := submission.SubmittedSteps[pindex].(*SubmittedProgramming)
				if !ok {
					submittedStep = new(SubmittedProgramming)
					submittedStep.Task = step
					submittedStep.Draft = true
					submission.SubmittedSteps[pindex] = submittedStep
				}

				r.Form.Set("StepIndex", spindex)
				return SubmissionNewProgrammingPageHandler(w, r, submittedStep)
			}
		}
	}
}

func SubmissionNewPageHandler(w *http.Response, r *http.Request) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}

	if err := r.ParseForm(); err != nil {
		return http.ClientError(err)
	}

	subjectID, err := GetValidIndex(r.Form.Get("ID"), DB.Subjects)
	if err != nil {
		return http.ClientError(err)
	}
	subject := &DB.Subjects[subjectID]

	li, err := GetValidIndex(r.Form.Get("LessonIndex"), subject.Lessons)
	if err != nil {
		return http.ClientError(err)
	}
	lesson := subject.Lessons[li]

	if WhoIsUserInSubject(session.ID, subject) != SubjectUserStudent {
		return http.ForbiddenError
	}

	submissionIndex := r.Form.Get("SubmissionIndex")
	var submission *Submission
	if submissionIndex == "" {
		submission = new(Submission)
		submission.Draft = true
		submission.LessonName = lesson.Name
		submission.StartedAt = time.Now()
		submission.User = &DB.Users[session.ID]
		StepsDeepCopy(&submission.Steps, lesson.Steps)
		submission.SubmittedSteps = make([]interface{}, len(lesson.Steps))

		lesson.Submissions = append(lesson.Submissions, submission)
		r.Form.SetInt("SubmissionIndex", len(lesson.Submissions)-1)
	} else {
		si, err := GetValidIndex(submissionIndex, lesson.Submissions)
		if err != nil {
			return http.ClientError(err)
		}
		submission = lesson.Submissions[si]
	}

	currentPage := r.Form.Get("CurrentPage")
	nextPage := r.Form.Get("NextPage")

	for i := 0; i < len(r.Form); i++ {
		k := r.Form[i].Key
		v := r.Form[i].Values[0]

		/* 'command' is button, which modifies content of a current page. */
		if strings.StartsWith(k, "Command") {
			/* NOTE(anton2920): after command is executed, function must return. */
			return SubmissionNewHandleCommand(w, r, submission, currentPage, k, v)
		}
	}

	stepIndex := r.Form.Get("StepIndex")
	if stepIndex != "" {
		si, err := GetValidIndex(r.Form.Get("StepIndex"), lesson.Steps)
		if err != nil {
			return http.ClientError(err)
		}
		if nextPage != "Discard" {
			switch currentPage {
			case "Test":
				submittedTest, ok := submission.SubmittedSteps[si].(*SubmittedTest)
				if !ok {
					return http.ClientError(nil)
				}

				if err := SubmissionNewTestVerifyRequest(r.Form, submittedTest); err != nil {
					return WritePageEx(w, r, SubmissionNewTestPageHandler, submittedTest, err)
				}

				submittedTest.Draft = false
			case "Programming":
				submittedTask, ok := submission.SubmittedSteps[si].(*SubmittedProgramming)
				if !ok {
					return http.ClientError(nil)
				}

				if err := SubmissionNewProgrammingVerifyRequest(r.Form, submittedTask); err != nil {
					return WritePageEx(w, r, SubmissionNewProgrammingPageHandler, submittedTask, err)
				}

				if err := SubmissionVerifyProgramming(submittedTask, CheckTypeExample); err != nil {
					return WritePageEx(w, r, SubmissionNewProgrammingPageHandler, submittedTask, http.BadRequest(err.Error()))
				}

				scores := submittedTask.Scores[CheckTypeExample]
				messages := submittedTask.Messages[CheckTypeExample]
				for i := 0; i < len(scores); i++ {
					if scores[i] == 0 {
						return WritePageEx(w, r, SubmissionNewProgrammingPageHandler, submittedTask, http.BadRequest("example %d: %s", i+1, messages[i]))
					}
				}

				submittedTask.Draft = false
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

func SubmissionDiscardHandler(w *http.Response, r *http.Request) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}

	if err := r.ParseForm(); err != nil {
		return http.ClientError(err)
	}

	subjectID, err := GetValidIndex(r.Form.Get("ID"), DB.Subjects)
	if err != nil {
		return http.ClientError(err)
	}
	subject := &DB.Subjects[subjectID]
	if (session.ID != AdminID) && (session.ID != subject.Teacher.ID) {
		return http.ForbiddenError
	}

	li, err := GetValidIndex(r.Form.Get("LessonIndex"), subject.Lessons)
	if err != nil {
		return http.ClientError(err)
	}
	lesson := subject.Lessons[li]

	si, err := GetValidIndex(r.Form.Get("SubmissionIndex"), lesson.Submissions)
	if err != nil {
		return http.ClientError(err)
	}
	lesson.Submissions = RemoveAtIndex(lesson.Submissions, si)

	w.RedirectID("/subject/", subjectID, http.StatusSeeOther)
	return nil
}

func SubmissionNewHandler(w *http.Response, r *http.Request) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}

	if err := r.ParseForm(); err != nil {
		return http.ClientError(err)
	}

	subjectID, err := GetValidIndex(r.Form.Get("ID"), DB.Subjects)
	if err != nil {
		return http.ClientError(err)
	}
	subject := &DB.Subjects[subjectID]
	if WhoIsUserInSubject(session.ID, subject) != SubjectUserStudent {
		return http.ForbiddenError
	}

	li, err := GetValidIndex(r.Form.Get("LessonIndex"), subject.Lessons)
	if err != nil {
		return http.ClientError(err)
	}
	lesson := subject.Lessons[li]

	si, err := GetValidIndex(r.Form.Get("SubmissionIndex"), lesson.Submissions)
	if err != nil {
		return http.ClientError(err)
	}
	submission := lesson.Submissions[si]

	empty := true
	for i := 0; i < len(submission.SubmittedSteps); i++ {
		step := submission.SubmittedSteps[i]
		if step != nil {
			empty = false

			var draft bool
			switch step := step.(type) {
			case *SubmittedTest:
				draft = step.Draft
			case *SubmittedProgramming:
				draft = step.Draft
			}
			if draft {
				return WritePageEx(w, r, SubmissionNewMainPageHandler, submission, http.BadRequest("step %d is still a draft", i+1))
			}
		}
	}
	if empty {
		return WritePageEx(w, r, SubmissionNewMainPageHandler, submission, http.BadRequest("you have to pass at least one step"))
	}

	submission.Draft = false
	submission.FinishedAt = time.Now()

	SubmissionVerifyChannel <- submission

	w.RedirectID("/subject/", subjectID, http.StatusSeeOther)
	return nil
}
