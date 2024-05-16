package main

import (
	"encoding/gob"

	"github.com/anton2920/gofa/net/http"
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

		Submissions []*Submission

		Draft bool
	}
)

type CheckType int

const (
	CheckTypeExample CheckType = iota
	CheckTypeTest
)

const LessonTheoryMaxDisplayLen = 30

func init() {
	gob.Register(&StepTest{})
	gob.Register(&StepProgramming{})
}

func StepsDeepCopy(dst *[]interface{}, src []interface{}) {
	*dst = make([]interface{}, len(src))

	for s := 0; s < len(src); s++ {
		switch ss := src[s].(type) {
		default:
			panic("invalid step type")
		case *StepTest:
			ds := new(StepTest)
			(*dst)[s] = ds

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
			(*dst)[s] = ds

			ds.Name = ss.Name
			ds.Description = ss.Description
			ds.Checks[CheckTypeExample] = make([]Check, len(ss.Checks[CheckTypeExample]))
			copy(ds.Checks[CheckTypeExample], ss.Checks[CheckTypeExample])
			ds.Checks[CheckTypeTest] = make([]Check, len(ss.Checks[CheckTypeTest]))
			copy(ds.Checks[CheckTypeTest], ss.Checks[CheckTypeTest])
		}
	}
}

func LessonsDeepCopy(dst *[]*Lesson, src []*Lesson) {
	*dst = make([]*Lesson, len(src))

	for l := 0; l < len(src); l++ {
		sl := src[l]

		dl := new(Lesson)
		(*dst)[l] = dl

		dl.Name = sl.Name
		dl.Theory = sl.Theory
		StepsDeepCopy(&dl.Steps, sl.Steps)
	}
}

func DisplayLessonsEditableList(w *http.Response, lessons []*Lesson) {
	for i := 0; i < len(lessons); i++ {
		lesson := lessons[i]

		w.AppendString(`<fieldset>`)

		w.AppendString(`<legend>Lesson #`)
		w.WriteInt(i + 1)
		if lesson.Draft {
			w.AppendString(` (draft)`)
		}
		w.AppendString(`</legend>`)

		w.AppendString(`<p>Name: `)
		w.WriteHTMLString(lesson.Name)
		w.AppendString(`</p>`)

		w.AppendString(`<p>Theory: `)
		DisplayShortenedString(w, lesson.Theory, LessonTheoryMaxDisplayLen)
		w.AppendString(`</p>`)

		DisplayIndexedCommand(w, i, "Edit")
		DisplayIndexedCommand(w, i, "Delete")
		if len(lessons) > 1 {
			if i > 0 {
				DisplayIndexedCommand(w, i, "↑")
			}
			if i < len(lessons)-1 {
				DisplayIndexedCommand(w, i, "↓")
			}
		}

		w.AppendString(`</fieldset>`)
		w.AppendString(`<br>`)
	}
}
