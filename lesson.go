package main

import (
	"unsafe"

	"github.com/anton2920/gofa/errors"
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

		/* TODO(anton2920): I don't like this. Replace with 'pointer|1'. */
		Draft bool
	}
	StepTest struct {
		StepCommon

		Questions []Question
	}
	StepProgramming struct {
		StepCommon

		Description string
		Checks      [2][]Check
	}
	Step struct {
		StepCommon

		/* TODO(anton2920): garbage collector cannot see pointers inside. */
		_ [max(unsafe.Sizeof(st), unsafe.Sizeof(sp)) - unsafe.Sizeof(sc)]byte
	}

	Lesson struct {
		ID    int32
		Flags int32

		ContainerID   int32
		ContainerType ContainerType

		Name   string
		Theory string

		Steps []Step

		Submissions []int32
	}
)

type ContainerType int32

const (
	ContainerTypeCourse ContainerType = iota
	ContainerTypeSubject
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

func Step2Test(s *Step) (*StepTest, error) {
	if s.Type != StepTypeTest {
		return nil, errors.New("invalid step type for test")
	}
	return (*StepTest)(unsafe.Pointer(s)), nil
}

func Step2Programming(s *Step) (*StepProgramming, error) {
	if s.Type != StepTypeProgramming {
		return nil, errors.New("invalid step type for programming")
	}
	return (*StepProgramming)(unsafe.Pointer(s)), nil
}

func GetStepStringType(s *Step) string {
	switch s.Type {
	default:
		panic("invalid step type")
	case StepTypeTest:
		return "Test"
	case StepTypeProgramming:
		return "Programming task"
	}
}

func DisplayLessonLink(w *http.Response, lesson *Lesson) {
	w.AppendString(`<a href="/lesson/`)
	w.WriteInt(int(lesson.ID))
	w.AppendString(`">Open</a>`)
}

func LessonPageHandler(w *http.Response, r *http.Request) error {
	var who SubjectUserType
	var container string

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}
	_ = session

	id, err := GetIDFromURL(r.URL, "/lesson/")
	if err != nil {
		return http.ClientError(err)
	}
	if (id < 0) || (id >= len(DB.Lessons)) {
		return http.NotFound("lesson with this ID does not exist")
	}
	lesson := &DB.Lessons[id]

	switch lesson.ContainerType {
	default:
		panic("invalid container type")
	case ContainerTypeCourse:
		var course Course
		var user User

		if err := GetUserByID(DB2, session.ID, &user); err != nil {
			return http.ServerError(err)
		}
		if !UserOwnsCourse(&user, lesson.ContainerID) {
			return http.ForbiddenError
		}
		if err := GetCourseByID(DB2, lesson.ContainerID, &course); err != nil {
			return http.ServerError(err)
		}
		container = course.Name
	case ContainerTypeSubject:
		var subject Subject
		var err error

		if err := GetSubjectByID(DB2, lesson.ContainerID, &subject); err != nil {
			return http.ServerError(err)
		}
		who, err = WhoIsUserInSubject(session.ID, &subject)
		if err != nil {
			return http.ServerError(err)
		}
		if who == SubjectUserNone {
			return http.ForbiddenError
		}
		container = subject.Name
	}

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>`)
	w.WriteHTMLString(container)
	w.AppendString(`: `)
	w.WriteHTMLString(lesson.Name)
	w.AppendString(`</title></head>`)
	w.AppendString(`<body>`)

	w.AppendString(`<h1>`)
	w.WriteHTMLString(container)
	w.AppendString(`: `)
	w.WriteHTMLString(lesson.Name)
	w.AppendString(`</h1>`)

	w.AppendString(`<h2>Theory</h2>`)
	w.AppendString(`<p>`)
	w.WriteHTMLString(lesson.Theory)
	w.AppendString(`</p>`)

	w.AppendString(`<h2>Evaluation</h2>`)

	w.AppendString(`<div style="max-width: max-content">`)
	for i := 0; i < len(lesson.Steps); i++ {
		step := &lesson.Steps[i]

		w.AppendString(`<fieldset>`)

		w.AppendString(`<legend>Step #`)
		w.WriteInt(i + 1)
		w.AppendString(`</legend>`)

		w.AppendString(`<p>Name: `)
		w.WriteHTMLString(step.Name)
		w.AppendString(`</p>`)

		w.AppendString(`<p>Type: `)
		w.AppendString(GetStepStringType(step))
		w.AppendString(`</p>`)

		w.AppendString(`</fieldset>`)
		w.AppendString(`<br>`)
	}
	w.AppendString(`</div>`)

	if who == SubjectUserStudent {
		var submission *Submission
		var i int

		for i = 0; i < len(lesson.Submissions); i++ {
			if session.ID == DB.Submissions[lesson.Submissions[i]].UserID {
				submission = &DB.Submissions[lesson.Submissions[i]]
				break
			}
		}

		if (submission != nil) && (!submission.Draft) {
			w.AppendString(`<form method="POST" action="/submission">`)
		} else {
			w.AppendString(`<form method="POST" action="/submission/new">`)
		}

		w.AppendString(`<input type="hidden" name="ID" value="`)
		w.WriteInt(int(lesson.ID))
		w.AppendString(`">`)

		if submission == nil {
			w.AppendString(`<input type="submit" value="Pass">`)
		} else {
			w.AppendString(`<input type="hidden" name="SubmissionIndex" value="`)
			w.WriteInt(i)
			w.AppendString(`">`)

			if submission.Draft {
				w.AppendString(`<input type="submit" value="Edit">`)
			} else {
				w.AppendString(`<input type="submit" value="See submission">`)
			}
		}

		w.AppendString(`</form>`)
	}

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func StepsDeepCopy(dst *[]Step, src []Step) {
	*dst = make([]Step, len(src))

	for s := 0; s < len(src); s++ {
		switch src[s].Type {
		default:
			panic("invalid step type")
		case StepTypeTest:
			ss, _ := Step2Test(&src[s])

			(*dst)[s].Type = StepTypeTest
			ds, _ := Step2Test(&(*dst)[s])

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
		case StepTypeProgramming:
			ss, _ := Step2Programming(&src[s])

			(*dst)[s].Type = StepTypeProgramming
			ds, _ := Step2Programming(&(*dst)[s])

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
