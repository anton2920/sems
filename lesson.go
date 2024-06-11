package main

import (
	"fmt"
	"unsafe"

	"github.com/anton2920/gofa/database"
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
	Step/* union */ struct {
		StepCommon

		_ [max(unsafe.Sizeof(st), unsafe.Sizeof(sp)) - unsafe.Sizeof(sc)]byte
	}

	Lesson struct {
		ID            database.ID
		Flags         int32
		ContainerID   database.ID
		ContainerType ContainerType

		Name        string
		Theory      string
		Steps       []Step
		Submissions []database.ID

		Data [16384]byte
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
	LessonActive int32 = iota
	LessonDraft
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

func CreateLesson(lesson *Lesson) error {
	var err error

	lesson.ID, err = database.IncrementNextID(LessonsDB)
	if err != nil {
		return fmt.Errorf("failed to increment lesson ID: %w", err)
	}

	return SaveLesson(lesson)
}

func DBStep2Step(step *Step, data *byte) {
	step.Name = database.Offset2String(step.Name, data)

	switch step.Type {
	default:
		panic("invalid step type")
	case StepTypeTest:
		test, _ := Step2Test(step)

		test.Questions = database.Offset2Slice(test.Questions, data)
		for i := 0; i < len(test.Questions); i++ {
			question := &test.Questions[i]
			question.Name = database.Offset2String(question.Name, data)
			question.Answers = database.Offset2Slice(question.Answers, data)
			question.CorrectAnswers = database.Offset2Slice(question.CorrectAnswers, data)

			for j := 0; j < len(question.Answers); j++ {
				question.Answers[j] = database.Offset2String(question.Answers[j], data)
			}
		}
	case StepTypeProgramming:
		task, _ := Step2Programming(step)

		task.Description = database.Offset2String(task.Description, data)

		for i := 0; i < len(task.Checks); i++ {
			task.Checks[i] = database.Offset2Slice(task.Checks[i], data)
			for j := 0; j < len(task.Checks[i]); j++ {
				check := &task.Checks[i][j]
				check.Input = database.Offset2String(check.Input, data)
				check.Output = database.Offset2String(check.Output, data)
			}
		}
	}
}

func DBLesson2Lesson(lesson *Lesson) {
	data := &lesson.Data[0]

	lesson.Name = database.Offset2String(lesson.Name, data)
	lesson.Theory = database.Offset2String(lesson.Theory, data)

	lesson.Steps = database.Offset2Slice(lesson.Steps, data)
	for i := 0; i < len(lesson.Steps); i++ {
		DBStep2Step(&lesson.Steps[i], data)
	}
	lesson.Submissions = database.Offset2Slice(lesson.Submissions, data)
}

func GetLessonByID(id database.ID, lesson *Lesson) error {
	if err := database.Read(LessonsDB, id, lesson); err != nil {
		return err
	}

	DBLesson2Lesson(lesson)
	return nil
}

func GetLessons(pos *int64, lessons []Lesson) (int, error) {
	n, err := database.ReadMany(LessonsDB, pos, lessons)
	if err != nil {
		return 0, err
	}

	for i := 0; i < n; i++ {
		DBLesson2Lesson(&lessons[i])
	}
	return n, nil
}

func Step2DBStep(ds *Step, ss *Step, data []byte, n int) int {
	n += database.String2DBString(&ds.Name, ss.Name, data, n)

	switch ss.Type {
	default:
		panic("invalid step type")
	case StepTypeTest:
		st, _ := Step2Test(ss)

		ds.Type = StepTypeTest
		dt, _ := Step2Test(ds)

		dt.Questions = make([]Question, len(st.Questions))
		for i := 0; i < len(st.Questions); i++ {
			sq := &st.Questions[i]
			dq := &dt.Questions[i]

			n += database.String2DBString(&dq.Name, sq.Name, data, n)

			dq.Answers = make([]string, len(sq.Answers))
			for j := 0; j < len(sq.Answers); j++ {
				n += database.String2DBString(&dq.Answers[j], sq.Answers[j], data, n)
			}
			n += database.Slice2DBSlice(&dq.Answers, dq.Answers, data, n)

			n += database.Slice2DBSlice(&dq.CorrectAnswers, sq.CorrectAnswers, data, n)
		}
		n += database.Slice2DBSlice(&dt.Questions, dt.Questions, data, n)
	case StepTypeProgramming:
		st, _ := Step2Programming(ss)

		ds.Type = StepTypeProgramming
		dt, _ := Step2Programming(ds)

		n += database.String2DBString(&dt.Description, st.Description, data, n)

		for i := 0; i < len(st.Checks); i++ {
			dt.Checks[i] = make([]Check, len(st.Checks[i]))
			for j := 0; j < len(st.Checks[i]); j++ {
				sc := &st.Checks[i][j]
				dc := &dt.Checks[i][j]

				n += database.String2DBString(&dc.Input, sc.Input, data, n)
				n += database.String2DBString(&dc.Output, sc.Output, data, n)
			}
			n += database.Slice2DBSlice(&dt.Checks[i], dt.Checks[i], data, n)
		}
	}

	return n
}

func SaveLesson(lesson *Lesson) error {
	var lessonDB Lesson
	var n int

	lessonDB.ID = lesson.ID
	lessonDB.Flags = lesson.Flags
	lessonDB.ContainerID = lesson.ContainerID
	lessonDB.ContainerType = lesson.ContainerType

	/* TODO(anton2920): save up to a sizeof(lesson.Data). */
	data := unsafe.Slice(&lessonDB.Data[0], len(lessonDB.Data))
	lessonDB.Steps = make([]Step, len(lesson.Steps))
	for i := 0; i < len(lesson.Steps); i++ {
		n += Step2DBStep(&lessonDB.Steps[i], &lesson.Steps[i], data, n)
	}

	n += database.String2DBString(&lessonDB.Name, lesson.Name, data, n)
	n += database.String2DBString(&lessonDB.Theory, lesson.Theory, data, n)
	n += database.Slice2DBSlice(&lessonDB.Steps, lessonDB.Steps, data, n)
	n += database.Slice2DBSlice(&lessonDB.Submissions, lesson.Submissions, data, n)

	return database.Write(LessonsDB, lessonDB.ID, &lessonDB)
}

func StepStringType(s *Step) string {
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
	w.WriteID(lesson.ID)
	w.AppendString(`">Open</a>`)
}

func LessonPageHandler(w *http.Response, r *http.Request) error {
	var submission Submission
	var who SubjectUserType
	var container string
	var displayed bool
	var lesson Lesson

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}

	id, err := GetIDFromURL(r.URL, "/lesson/")
	if err != nil {
		return http.ClientError(err)
	}
	if err := GetLessonByID(id, &lesson); err != nil {
		if err == database.NotFound {
			return http.NotFound("lesson with this ID does not exist")
		}
		return http.ServerError(err)
	}

	switch lesson.ContainerType {
	default:
		panic("invalid container type")
	case ContainerTypeCourse:
		var course Course
		var user User

		if err := GetUserByID(session.ID, &user); err != nil {
			return http.ServerError(err)
		}
		if !UserOwnsCourse(&user, lesson.ContainerID) {
			return http.ForbiddenError
		}
		if err := GetCourseByID(lesson.ContainerID, &course); err != nil {
			return http.ServerError(err)
		}
		container = course.Name
	case ContainerTypeSubject:
		var subject Subject
		var err error

		if err := GetSubjectByID(lesson.ContainerID, &subject); err != nil {
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

		if i > 0 {
			w.AppendString(`<br>`)
		}

		w.AppendString(`<fieldset>`)

		w.AppendString(`<legend>Step #`)
		w.WriteInt(i + 1)
		w.AppendString(`</legend>`)

		w.AppendString(`<p>Name: `)
		w.WriteHTMLString(step.Name)
		w.AppendString(`</p>`)

		w.AppendString(`<p>Type: `)
		w.AppendString(StepStringType(step))
		w.AppendString(`</p>`)

		w.AppendString(`</fieldset>`)
	}
	w.AppendString(`</div>`)

	switch who {
	case SubjectUserAdmin, SubjectUserTeacher:
		if len(lesson.Submissions) > 0 {
			for i := 0; i < len(lesson.Submissions); i++ {
				if err := GetSubmissionByID(lesson.Submissions[i], &submission); err != nil {
					return http.ServerError(err)
				}

				if submission.Flags == SubmissionActive {
					if !displayed {
						w.AppendString(`<h2>Submissions</h2>`)
						w.AppendString(`<ul>`)
						displayed = true
					}

					w.AppendString(`<li>`)
					DisplaySubmissionLink(w, &submission)
					w.AppendString(`</li>`)
				}
			}
			if displayed {
				w.AppendString(`</ul>`)
			}
		}
	case SubjectUserStudent:
		si := -1

		for i := 0; i < len(lesson.Submissions); i++ {
			if err := GetSubmissionByID(lesson.Submissions[i], &submission); err != nil {
				return http.ServerError(err)
			}

			if submission.UserID == session.ID {
				if submission.Flags == SubmissionActive {
					if !displayed {
						w.AppendString(`<h2>Submissions</h2>`)
						w.AppendString(`<ul>`)
						displayed = true
					}
					w.AppendString(`<li>`)
					DisplaySubmissionLink(w, &submission)
					w.AppendString(`</li>`)
				} else if submission.Flags == SubmissionDraft {
					si = i
				}
			}
		}
		if displayed {
			w.AppendString(`</ul>`)
		} else {
			w.AppendString(`<br>`)
		}

		w.AppendString(`<form method="POST" action="/submission/new">`)

		w.AppendString(`<input type="hidden" name="ID" value="`)
		w.WriteID(lesson.ID)
		w.AppendString(`">`)

		if si == -1 {
			w.AppendString(`<input type="submit" value="Pass">`)
		} else {
			w.AppendString(`<input type="hidden" name="SubmissionIndex" value="`)
			w.WriteInt(si)
			w.AppendString(`">`)

			w.AppendString(`<input type="submit" value="Edit">`)
		}

		w.AppendString(`</form>`)
	}

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

/* TODO(anton2920): check whether this function is needed. */
func StepDeepCopy(dst *Step, src *Step) {
	switch src.Type {
	default:
		panic("invalid step type")
	case StepTypeTest:
		ss, _ := Step2Test(src)

		dst.Type = StepTypeTest
		ds, _ := Step2Test(dst)

		ds.Name = ss.Name

		ds.Questions = make([]Question, len(ss.Questions))
		for i := 0; i < len(ss.Questions); i++ {
			sq := &ss.Questions[i]
			dq := &ds.Questions[i]

			dq.Name = sq.Name

			dq.Answers = make([]string, len(sq.Answers))
			copy(dq.Answers, sq.Answers)

			dq.CorrectAnswers = make([]int, len(sq.CorrectAnswers))
			copy(dq.CorrectAnswers, sq.CorrectAnswers)
		}
	case StepTypeProgramming:
		ss, _ := Step2Programming(src)

		dst.Type = StepTypeProgramming
		ds, _ := Step2Programming(dst)

		ds.Name = ss.Name
		ds.Description = ss.Description

		ds.Checks[CheckTypeExample] = make([]Check, len(ss.Checks[CheckTypeExample]))
		copy(ds.Checks[CheckTypeExample], ss.Checks[CheckTypeExample])

		ds.Checks[CheckTypeTest] = make([]Check, len(ss.Checks[CheckTypeTest]))
		copy(ds.Checks[CheckTypeTest], ss.Checks[CheckTypeTest])
	}
}

func LessonsDeepCopy(dst *[]database.ID, src []database.ID, containerID database.ID, containerType ContainerType) {
	*dst = make([]database.ID, len(src))

	for i := 0; i < len(src); i++ {
		var sl, dl Lesson

		if err := GetLessonByID(src[i], &sl); err != nil {
			/* TODO(anton2920): report error. */
		}

		dl.Flags = sl.Flags
		dl.ContainerID = containerID
		dl.ContainerType = containerType

		dl.Name = sl.Name
		dl.Theory = sl.Theory
		dl.Steps = make([]Step, len(sl.Steps))
		for j := 0; j < len(sl.Steps); j++ {
			StepDeepCopy(&dl.Steps[j], &sl.Steps[j])
		}

		if err := CreateLesson(&dl); err != nil {
			/* TODO(anton2920): report error. */
		}
		(*dst)[i] = dl.ID
	}
}

func DisplayLessonsEditableList(w *http.Response, lessons []database.ID) {
	var lesson Lesson

	for i := 0; i < len(lessons); i++ {
		if err := GetLessonByID(lessons[i], &lesson); err != nil {
			/* TODO(anton2920): report error. */
		}

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
