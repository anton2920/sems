package main

import (
	"encoding/gob"
	"strconv"
	"unicode/utf8"
)

type (
	Question struct {
		Name           string
		Answers        []string
		CorrectAnswers []int
	}
	Check struct {
		Input  string
		Output string
	}

	StepTest struct {
		Name      string
		Questions []Question

		/* TODO(anton2920): I don't like this. Replace with 'pointer|1'. */
		Draft bool
	}
	StepProgramming struct {
		Name        string
		Description string
		Checks      [2][]Check

		/* TODO(anton2920): I don't like this. Replace with 'pointer|1'. */
		Draft bool
	}

	Lesson struct {
		Name   string
		Theory string

		/* TODO(anton2920): using 'interface{}' so 'encoding/gob' does what it supposed to do. */
		Steps []interface{}

		Draft bool
	}
)

type CheckType int

const (
	CheckTypeExample CheckType = iota
	CheckTypeTest
)

func init() {
	gob.Register(&StepTest{})
	gob.Register(&StepProgramming{})
}

func LessonsDeepCopy(dst *[]*Lesson, src []*Lesson) {
	*dst = make([]*Lesson, len(src))

	for l := 0; l < len(src); l++ {
		sl := src[l]

		dl := new(Lesson)
		(*dst)[l] = dl

		dl.Name = sl.Name
		dl.Theory = sl.Theory
		dl.Steps = make([]interface{}, len(sl.Steps))

		for s := 0; s < len(sl.Steps); s++ {
			switch ss := sl.Steps[s].(type) {
			default:
				panic("invalid step type")
			case *StepTest:
				ds := new(StepTest)
				dl.Steps[s] = ds

				ds.Name = ss.Name
				ds.Questions = make([]Question, len(ss.Questions))

				for q := 0; q < len(ss.Questions); q++ {
					sq := &ss.Questions[q]

					dq := &ds.Questions[q]
					dq.Name = sq.Name
					dq.Answers = make([]string, len(sq.Answers))
					copy(dq.Answers, sq.Answers)
					dq.CorrectAnswers = make([]int, len(sq.CorrectAnswers))
					copy(dq.CorrectAnswers, sq.CorrectAnswers)
				}
			case *StepProgramming:
				ds := new(StepProgramming)
				dl.Steps[s] = ds

				ds.Name = ss.Name
				ds.Description = ss.Description
				ds.Checks[CheckTypeExample] = make([]Check, len(ss.Checks[CheckTypeExample]))
				copy(ds.Checks[CheckTypeExample], ss.Checks[CheckTypeExample])
				ds.Checks[CheckTypeTest] = make([]Check, len(ss.Checks[CheckTypeTest]))
				copy(ds.Checks[CheckTypeTest], ss.Checks[CheckTypeTest])
			}
		}
	}
}

func LessonDisplayTheory(w *HTTPResponse, theory string) {
	const maxVisibleLen = 30
	if utf8.RuneCountInString(theory) < maxVisibleLen {
		w.WriteHTMLString(theory)
	} else {
		space := FindChar(theory[maxVisibleLen:], ' ')
		if space == -1 {
			w.WriteHTMLString(theory[:maxVisibleLen])
		} else {
			w.WriteHTMLString(theory[:space+maxVisibleLen])
		}
		w.AppendString(`...`)
	}
}

func LessonPageHandler(w *HTTPResponse, r *HTTPRequest) error {
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

	var student bool
	if (session.ID != AdminID) && (session.ID != subject.Teacher.ID) {
		for i := 0; i < len(subject.Group.Users); i++ {
			if session.ID == subject.Group.Users[i].ID {
				student = true
				break
			}
		}

		if !student {
			return ForbiddenError
		}
	}

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>`)
	w.WriteHTMLString(subject.Name)
	w.AppendString(`: `)
	w.WriteHTMLString(lesson.Name)
	w.AppendString(`</title></head>`)
	w.AppendString(`<body>`)

	w.AppendString(`<h1>`)
	w.WriteHTMLString(subject.Name)
	w.AppendString(`: `)
	w.WriteHTMLString(lesson.Name)
	w.AppendString(`</h1>`)

	w.AppendString(`<h2>Theory</h2>`)
	w.AppendString(`<p>`)
	w.WriteHTMLString(lesson.Theory)
	w.AppendString(`</p>`)

	w.AppendString(`<h2>Evaluation steps</h2>`)

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

		if student {
			w.AppendString(`<form method="POST" action="/lesson/pass">`)

			w.AppendString(`<input type="hidden" name="StepIndex" value="`)
			w.WriteInt(i)
			w.AppendString(`">`)

			w.AppendString(`<input type="submit" value="Pass">`)

			w.AppendString(`</form>`)
		}

		w.AppendString(`</fieldset>`)
		w.AppendString(`<br>`)
	}

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}
