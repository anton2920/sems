package main

import (
	"strconv"
	"time"
)

type (
	PassedQuestion struct {
		SelectedAnswers []int
	}
	PassedTest struct {
		PassedQuestions []PassedQuestion
	}

	PassedProgramming struct {
		LanguageID int
		Solution   string
	}

	PassContext struct {
		StartedAt, FinishedAt time.Time
		PassedSteps           []interface{}
	}
)

func EvaluationPassMainPageHandler(w *HTTPResponse, r *HTTPRequest, lesson *Lesson) error {
	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>`)
	w.AppendString(`Evaluation for `)
	w.WriteHTMLString(lesson.Name)
	w.AppendString(`</title></head>`)
	w.AppendString(`<body>`)

	w.AppendString(`<h1>`)
	w.AppendString(`Evaluation for `)
	w.WriteHTMLString(lesson.Name)
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

	for i := 0; i < len(lesson.Steps); i++ {
		var name, stepType string

		step := lesson.Steps[i]
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
		w.AppendString(`Pass`)
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

func EvaluationPassTestPageHandler(w *HTTPResponse, r *HTTPRequest, test *StepTest) error {
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

	w.AppendString(`<input type="hidden" name="StepIndex" value="`)
	w.WriteHTMLString(r.Form.Get("StepIndex"))
	w.AppendString(`">`)

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

			/*
				for k := 0; k < len(SelectedAnswers); k++ {
					selectedAnswer := SelectedAnswers[k]
					if j == selectedAnswer {
						w.AppendString(` checked`)
						break
					}
				}
			*/

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

func EvaluationPassProgrammingPageHandler(w *HTTPResponse, r *HTTPRequest, task *StepProgramming) error {
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

	w.AppendString(`<input type="hidden" name="StepIndex" value="`)
	w.WriteHTMLString(r.Form.Get("StepIndex"))
	w.AppendString(`">`)

	w.AppendString(`<h2>Solution</h2>`)

	w.AppendString(`<label>Programming language: `)
	w.AppendString(`<select name="LanguageID"><option>C</option><option>Go</option></select>`)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<textarea cols="80" rows="24" name="Solution">`)
	// w.WriteHTMLString(Solution)
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

func EvaluationPassHandleCommand(w *HTTPResponse, r *HTTPRequest, lesson *Lesson, currentPage, k, command string) error {
	pindex, _, _, _, err := GetIndicies(k[len("Command"):])
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
			if (pindex < 0) || (pindex >= len(lesson.Steps)) {
				return ReloadPageError
			}
			switch step := lesson.Steps[pindex].(type) {
			default:
				panic("invalid step type")
			case *StepTest:
				return EvaluationPassTestPageHandler(w, r, step)
			case *StepProgramming:
				return EvaluationPassProgrammingPageHandler(w, r, step)
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

	currentPage := r.Form.Get("CurrentPage")
	nextPage := r.Form.Get("NextPage")

	for i := 0; i < len(r.Form); i++ {
		k := r.Form[i].Key
		v := r.Form[i].Values[0]

		/* 'command' is button, which modifies content of a current page. */
		if StringStartsWith(k, "Command") {
			/* NOTE(anton2920): after command is executed, function must return. */
			return EvaluationPassHandleCommand(w, r, lesson, currentPage, k, v)
		}
	}

	switch currentPage {
	case "Test":
	case "Programming":
	}

	switch nextPage {
	default:
		return EvaluationPassMainPageHandler(w, r, lesson)
	case "Finish":
		return EvaluationPassHandler(w, r)
	}

	return nil
}

func EvaluationPassHandler(w *HTTPResponse, r *HTTPRequest) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}
	_ = session

	if err := r.ParseForm(); err != nil {
		return ReloadPageError
	}

	return nil
}
