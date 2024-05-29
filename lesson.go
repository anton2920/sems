package main

import (
	"encoding/gob"
	"unsafe"

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

	StepCommon struct {
		Name string
		Type StepType
	}
	StepTest struct {
		StepCommon
		Questions []Question

		/* TODO(anton2920): I don't like this. Replace with 'pointer|1'. */
		Draft bool
	}
	StepProgramming struct {
		StepCommon
		Description string
		Checks      [2][]Check

		/* TODO(anton2920): I don't like this. Replace with 'pointer|1'. */
		Draft bool
	}
	Step struct {
		StepCommon
		_ [max(unsafe.Sizeof(st), unsafe.Sizeof(sp)) - unsafe.Sizeof(sc)]byte
	}

	Lesson struct {
		ID    int32
		Flags int32

		Name   string
		Theory string

		/* TODO(anton2920): using 'interface{}' so 'encoding/gob' does what it supposed to do. */
		Steps []interface{}

		Submissions []*Submission
	}
)

type CheckType int

const (
	CheckTypeExample CheckType = iota
	CheckTypeTest
)

type StepType byte

const (
	StepTypeTest StepType = iota
	StepTypeProgramming
)

const (
	LessonActive  int32 = 0
	LessonDeleted       = 1
	LessonDraft         = 2
)

const LessonTheoryMaxDisplayLen = 30

var (
	sc StepCommon
	st StepTest
	sp StepProgramming
)

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

func LessonsDeepCopy(dst *[]int32, src []int32) {
	*dst = make([]int32, len(src))

	for l := 0; l < len(src); l++ {
		sl := &DB.Lessons[src[l]]

		DB.Lessons = append(DB.Lessons, Lesson{ID: int32(len(DB.Lessons))})
		dl := &DB.Lessons[len(DB.Lessons)-1]
		(*dst)[l] = dl.ID

		dl.Flags = sl.Flags
		dl.Name = sl.Name
		dl.Theory = sl.Theory
		StepsDeepCopy(&dl.Steps, sl.Steps)
	}
}

func DisplayLessonsEditableList(w *http.Response, lessons []int32) {
	for i := 0; i < len(lessons); i++ {
		lesson := &DB.Lessons[lessons[i]]

		w.AppendString(`<fieldset>`)

		w.AppendString(`<legend>Lesson #`)
		w.WriteInt(i + 1)
		if lesson.Flags == LessonDraft {
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
