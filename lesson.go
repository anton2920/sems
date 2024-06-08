package main

import (
	"fmt"
	"unsafe"

	"github.com/anton2920/gofa/errors"
	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/syscall"
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

		/* TODO(anton2920): garbage collector cannot see pointers inside. */
		_ [max(unsafe.Sizeof(st), unsafe.Sizeof(sp)) - unsafe.Sizeof(sc)]byte
	}

	Lesson struct {
		ID            int32
		Flags         int32
		ContainerID   int32
		ContainerType ContainerType

		Name        string
		Theory      string
		Steps       []Step
		Submissions []int32

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

func CreateLesson(db *Database, lesson *Lesson) error {
	var err error

	lesson.ID, err = IncrementNextID(db.LessonsFile)
	if err != nil {
		return fmt.Errorf("failed to increment lesson ID: %w", err)
	}

	return SaveLesson(db, lesson)
}

func DBLesson2Lesson(lesson *Lesson) {
	lesson.Name = Offset2String(lesson.Name, &lesson.Data[0])
	lesson.Theory = Offset2String(lesson.Theory, &lesson.Data[0])
	lesson.Steps = Offset2Slice(lesson.Steps, &lesson.Data[0])

	for s := 0; s < len(lesson.Steps); s++ {
		step := &lesson.Steps[s]
		step.Name = Offset2String(step.Name, &lesson.Data[0])

		switch step.Type {
		case StepTypeTest:
			test, _ := Step2Test(step)
			test.Questions = Offset2Slice(test.Questions, &lesson.Data[0])

			for q := 0; q < len(test.Questions); q++ {
				question := &test.Questions[q]
				question.Name = Offset2String(question.Name, &lesson.Data[0])
				question.Answers = Offset2Slice(question.Answers, &lesson.Data[0])
				question.CorrectAnswers = Offset2Slice(question.CorrectAnswers, &lesson.Data[0])

				for a := 0; a < len(question.Answers); a++ {
					answer := &question.Answers[a]
					*answer = Offset2String(*answer, &lesson.Data[0])
				}
			}
		case StepTypeProgramming:
			task, _ := Step2Programming(step)
			task.Description = Offset2String(task.Description, &lesson.Data[0])

			for i := 0; i < len(task.Checks); i++ {
				task.Checks[i] = Offset2Slice(task.Checks[i], &lesson.Data[0])

				for c := 0; c < len(task.Checks[i]); c++ {
					check := &task.Checks[i][c]
					check.Input = Offset2String(check.Input, &lesson.Data[0])
					check.Output = Offset2String(check.Output, &lesson.Data[0])
				}
			}
		}
	}

	lesson.Submissions = Offset2Slice(lesson.Submissions, &lesson.Data[0])
}

func GetLessonByID(db *Database, id int32, lesson *Lesson) error {
	size := int(unsafe.Sizeof(*lesson))
	offset := int64(int(id)*size) + DataOffset

	n, err := syscall.Pread(db.LessonsFile, unsafe.Slice((*byte)(unsafe.Pointer(lesson)), size), offset)
	if err != nil {
		return fmt.Errorf("failed to read lesson from DB: %w", err)
	}
	if n < size {
		return DBNotFound
	}

	DBLesson2Lesson(lesson)
	return nil
}

func GetLessons(db *Database, pos *int64, lessons []Lesson) (int, error) {
	if *pos < DataOffset {
		*pos = DataOffset
	}
	size := int(unsafe.Sizeof(lessons[0]))

	n, err := syscall.Pread(db.LessonsFile, unsafe.Slice((*byte)(unsafe.Pointer(unsafe.SliceData(lessons))), len(lessons)*size), *pos)
	if err != nil {
		return 0, fmt.Errorf("failed to read lesson from DB: %w", err)
	}
	*pos += int64(n)

	n /= size
	for i := 0; i < n; i++ {
		DBLesson2Lesson(&lessons[i])
	}

	return n, nil
}

func DeleteLessonByID(db *Database, id int32) error {
	flags := LessonDeleted
	var lesson Lesson

	offset := int64(int(id)*int(unsafe.Sizeof(lesson))) + DataOffset + int64(unsafe.Offsetof(lesson.Flags))
	_, err := syscall.Pwrite(db.LessonsFile, unsafe.Slice((*byte)(unsafe.Pointer(&flags)), unsafe.Sizeof(flags)), offset)
	if err != nil {
		return fmt.Errorf("failed to delete lesson from DB: %w", err)
	}

	return nil
}

func SaveLesson(db *Database, lesson *Lesson) error {
	var lessonDB Lesson

	size := int(unsafe.Sizeof(*lesson))
	offset := int64(int(lesson.ID)*size) + DataOffset

	n := DataStartOffset
	var nbytes int

	lessonDB.ID = lesson.ID
	lessonDB.Flags = lesson.Flags
	lessonDB.ContainerID = lesson.ContainerID
	lessonDB.ContainerType = lesson.ContainerType

	/* TODO(anton2920): saving up to a sizeof(lesson.Data). */
	nbytes = copy(lessonDB.Data[n:], lesson.Name)
	lessonDB.Name = String2Offset(lesson.Name, n)
	n += nbytes

	nbytes = copy(lessonDB.Data[n:], lesson.Theory)
	lessonDB.Theory = String2Offset(lesson.Theory, n)
	n += nbytes

	if len(lesson.Steps) > 0 {
		lessonDB.Steps = make([]Step, len(lesson.Steps))
		for s := 0; s < len(lesson.Steps); s++ {
			ss := &lesson.Steps[s]
			ds := &lessonDB.Steps[s]

			nbytes = copy(lessonDB.Data[n:], ss.Name)
			ds.Name = String2Offset(ss.Name, n)
			n += nbytes

			switch ss.Type {
			case StepTypeTest:
				st, _ := Step2Test(ss)

				ds.Type = StepTypeTest
				dt, _ := Step2Test(ds)

				if len(st.Questions) > 0 {
					dt.Questions = make([]Question, len(st.Questions))
					for q := 0; q < len(st.Questions); q++ {
						sq := &st.Questions[q]
						dq := &dt.Questions[q]

						nbytes = copy(lessonDB.Data[n:], sq.Name)
						dq.Name = String2Offset(sq.Name, n)
						n += nbytes

						if len(sq.Answers) > 0 {
							dq.Answers = make([]string, len(sq.Answers))
							for a := 0; a < len(sq.Answers); a++ {
								nbytes = copy(lessonDB.Data[n:], sq.Answers[a])
								dq.Answers[a] = String2Offset(sq.Answers[a], n)
								n += nbytes
							}

							nbytes = copy(lessonDB.Data[n:], unsafe.Slice((*byte)(unsafe.Pointer(&dq.Answers[0])), len(sq.Answers)*int(unsafe.Sizeof(dq.Answers[0]))))
							dq.Answers = Slice2Offset(dq.Answers, n)
							n += nbytes
						}

						if len(sq.CorrectAnswers) > 0 {
							nbytes = copy(lessonDB.Data[n:], unsafe.Slice((*byte)(unsafe.Pointer(&sq.CorrectAnswers[0])), len(sq.CorrectAnswers)*int(unsafe.Sizeof(sq.CorrectAnswers[0]))))
							dq.CorrectAnswers = Slice2Offset(sq.CorrectAnswers, n)
							n += nbytes
						}
					}

					nbytes = copy(lessonDB.Data[n:], unsafe.Slice((*byte)(unsafe.Pointer(&dt.Questions[0])), len(dt.Questions)*int(unsafe.Sizeof(dt.Questions[0]))))
					dt.Questions = Slice2Offset(dt.Questions, n)
					n += nbytes
				}
			case StepTypeProgramming:
				st, _ := Step2Programming(ss)

				ds.Type = StepTypeProgramming
				dt, _ := Step2Programming(ds)

				nbytes = copy(lessonDB.Data[n:], st.Description)
				dt.Description = String2Offset(st.Description, n)
				n += nbytes

				for i := 0; i < len(st.Checks); i++ {
					if len(st.Checks[i]) > 0 {
						dt.Checks[i] = make([]Check, len(st.Checks[i]))
						for c := 0; c < len(st.Checks[i]); c++ {
							sc := &st.Checks[i][c]
							dc := &dt.Checks[i][c]

							nbytes = copy(lessonDB.Data[n:], sc.Input)
							dc.Input = String2Offset(sc.Input, n)
							n += nbytes

							nbytes = copy(lessonDB.Data[n:], sc.Output)
							dc.Output = String2Offset(sc.Output, n)
							n += nbytes
						}

						nbytes = copy(lessonDB.Data[n:], unsafe.Slice((*byte)(unsafe.Pointer(&dt.Checks[i][0])), len(dt.Checks[i])*int(unsafe.Sizeof(dt.Checks[i][0]))))
						dt.Checks[i] = Slice2Offset(dt.Checks[i], n)
						n += nbytes
					}
				}
			}
		}
		nbytes = copy(lessonDB.Data[n:], unsafe.Slice((*byte)(unsafe.Pointer(&lessonDB.Steps[0])), len(lessonDB.Steps)*int(unsafe.Sizeof(lessonDB.Steps[0]))))
		lessonDB.Steps = Slice2Offset(lessonDB.Steps, n)
		n += nbytes
	}

	if len(lesson.Submissions) > 0 {
		nbytes = copy(lessonDB.Data[n:], unsafe.Slice((*byte)(unsafe.Pointer(&lesson.Submissions[0])), len(lesson.Submissions)*int(unsafe.Sizeof(lesson.Submissions[0]))))
		lessonDB.Submissions = Slice2Offset(lesson.Submissions, n)
		n += nbytes
	}

	_, err := syscall.Pwrite(db.LessonsFile, unsafe.Slice((*byte)(unsafe.Pointer(&lessonDB)), size), offset)
	if err != nil {
		return fmt.Errorf("failed to write lesson to DB: %w", err)
	}

	return nil
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
	w.WriteInt(int(lesson.ID))
	w.AppendString(`">Open</a>`)
}

func LessonPageHandler(w *http.Response, r *http.Request) error {
	var who SubjectUserType
	var container string
	var lesson Lesson

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}

	id, err := GetIDFromURL(r.URL, "/lesson/")
	if err != nil {
		return http.ClientError(err)
	}
	if err := GetLessonByID(DB2, int32(id), &lesson); err != nil {
		if err == DBNotFound {
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
			w.AppendString(`<h2>Submissions</h2>`)
			w.AppendString(`<ul>`)
			for i := 0; i < len(lesson.Submissions); i++ {
				submission := &DB.Submissions[lesson.Submissions[i]]

				w.AppendString(`<li>`)
				DisplaySubmissionLink(w, submission)
				w.AppendString(`</li>`)
			}
			w.AppendString(`</ul>`)
		}
	case SubjectUserStudent:
		var submission *Submission
		var displayed bool
		var si int

		for i := 0; i < len(lesson.Submissions); i++ {
			submission = &DB.Submissions[lesson.Submissions[i]]

			if submission.UserID == session.ID {
				if !displayed {
					w.AppendString(`<h2>Submissions</h2>`)
					w.AppendString(`<ul>`)
					displayed = true
				}
				if submission.Flags == SubmissionActive {
					w.AppendString(`<li>`)
					DisplaySubmissionLink(w, submission)
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
		w.WriteInt(int(lesson.ID))
		w.AppendString(`">`)

		if (submission == nil) || (submission.Flags == SubmissionDraft) {
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

func LessonsDeepCopy(dst *[]int32, src []int32, containerID int32, containerType ContainerType) {
	*dst = make([]int32, len(src))

	for l := 0; l < len(src); l++ {
		var sl, dl Lesson

		if err := GetLessonByID(DB2, src[l], &sl); err != nil {
			/* TODO(anton2920): report error. */
		}

		dl.Flags = sl.Flags
		dl.ContainerID = containerID
		dl.ContainerType = containerType

		dl.Name = sl.Name
		dl.Theory = sl.Theory
		StepsDeepCopy(&dl.Steps, sl.Steps)

		if err := CreateLesson(DB2, &dl); err != nil {
			/* TODO(anton2920): report error. */
		}
		(*dst)[l] = dl.ID
	}
}

func DisplayLessonsEditableList(w *http.Response, lessons []int32) {
	var lesson Lesson

	for i := 0; i < len(lessons); i++ {
		if err := GetLessonByID(DB2, lessons[i], &lesson); err != nil {
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
