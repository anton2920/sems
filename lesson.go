package main

import (
	"fmt"
	"unsafe"

	"github.com/anton2920/gofa/database"
	"github.com/anton2920/gofa/errors"
	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/net/url"
	"github.com/anton2920/gofa/slices"
	"github.com/anton2920/gofa/strings"
	"github.com/anton2920/gofa/trace"
	"github.com/anton2920/gofa/util"
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

		/* TODO(anton2920): I don't like this. */
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
		ContainerType LessonContainerType

		Name        string
		Theory      string
		Steps       []Step
		Submissions []database.ID

		Data [16384]byte
	}

	LessonContainer struct {
		ID    database.ID
		Flags int32

		Name    string
		Lessons []database.ID
	}
)

type LessonContainerType int32

const (
	LessonContainerCourse LessonContainerType = iota
	LessonContainerSubject
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
	_
	LessonDraft
)

const (
	MinTheoryLen = 1
	MaxTheoryLen = 4096

	MinStepNameLen = 1
	MaxStepNameLen = 128
	MinQuestionLen = 1
	MaxQuestionLen = 128
	MinAnswerLen   = 1
	MaxAnswerLen   = 128

	MinDescriptionLen = 1
	MaxDescriptionLen = 1024
	MinCheckLen       = 1
	MaxCheckLen       = 512
)

const LessonTheoryMaxDisplayLen = 30

const (
	CheckKeyDisplay = iota
	CheckKeyInput
	CheckKeyOutput
)

var CheckKeys = [2][3]string{
	CheckTypeExample: {CheckKeyDisplay: "example", CheckKeyInput: "ExampleInput", CheckKeyOutput: "ExampleOutput"},
	CheckTypeTest:    {CheckKeyDisplay: "test", CheckKeyInput: "TestInput", CheckKeyOutput: "TestOutput"},
}

/* NOTE(anton2920): for sizeof. */
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

func MoveLessonDown(vs []database.ID, i int) {
	if (i >= 0) && (i < len(vs)-1) {
		vs[i], vs[i+1] = vs[i+1], vs[i]
	}
}

func MoveStepDown(vs []Step, i int) {
	if (i >= 0) && (i < len(vs)-1) {
		vs[i], vs[i+1] = vs[i+1], vs[i]
	}
}

func MoveQuestionDown(vs []Question, i int) {
	if (i >= 0) && (i < len(vs)-1) {
		vs[i], vs[i+1] = vs[i+1], vs[i]
	}
}

func MoveCheckDown(vs []Check, i int) {
	if (i >= 0) && (i < len(vs)-1) {
		vs[i], vs[i+1] = vs[i+1], vs[i]
	}
}

func MoveLessonUp(vs []database.ID, i int) {
	if (i > 0) && (i <= len(vs)-1) {
		vs[i-1], vs[i] = vs[i], vs[i-1]
	}
}

func MoveStepUp(vs []Step, i int) {
	if (i > 0) && (i <= len(vs)-1) {
		vs[i-1], vs[i] = vs[i], vs[i-1]
	}
}

func MoveQuestionUp(vs []Question, i int) {
	if (i > 0) && (i <= len(vs)-1) {
		vs[i-1], vs[i] = vs[i], vs[i-1]
	}
}

func MoveCheckUp(vs []Check, i int) {
	if (i > 0) && (i <= len(vs)-1) {
		vs[i-1], vs[i] = vs[i], vs[i-1]
	}
}

func RemoveLessonAtIndex(vs []database.ID, i int) []database.ID {
	if (len(vs) == 0) || (i < 0) || (i >= len(vs)) {
		return vs
	}
	if i < len(vs)-1 {
		copy(vs[i:], vs[i+1:])
	}
	return vs[:len(vs)-1]
}

func RemoveStepAtIndex(vs []Step, i int) []Step {
	if (len(vs) == 0) || (i < 0) || (i >= len(vs)) {
		return vs
	}
	if i < len(vs)-1 {
		copy(vs[i:], vs[i+1:])
	}
	return vs[:len(vs)-1]
}

func RemoveQuestionAtIndex(vs []Question, i int) []Question {
	if (len(vs) == 0) || (i < 0) || (i >= len(vs)) {
		return vs
	}
	if i < len(vs)-1 {
		copy(vs[i:], vs[i+1:])
	}
	return vs[:len(vs)-1]
}

func RemoveCheckAtIndex(vs []Check, i int) []Check {
	if (len(vs) == 0) || (i < 0) || (i >= len(vs)) {
		return vs
	}
	if i < len(vs)-1 {
		copy(vs[i:], vs[i+1:])
	}
	return vs[:len(vs)-1]
}

func CreateLesson(lesson *Lesson) error {
	defer trace.End(trace.Begin(""))

	var err error

	lesson.ID, err = database.IncrementNextID(LessonsDB)
	if err != nil {
		return fmt.Errorf("failed to increment lesson ID: %w", err)
	}

	return SaveLesson(lesson)
}

func DBStep2Step(step *Step, data *byte) {
	defer trace.End(trace.Begin(""))

	step.Name = database.Offset2String(step.Name, data)

	switch step.Type {
	default:
		panic("invalid step type")
	case StepTypeTest:
		test, _ := Step2Test(step)

		slice := database.Offset2Slice(*(*[]byte)(unsafe.Pointer(&test.Questions)), data)
		test.Questions = *(*[]Question)(unsafe.Pointer(&slice))

		for i := 0; i < len(test.Questions); i++ {
			question := &test.Questions[i]
			question.Name = database.Offset2String(question.Name, data)

			slice = database.Offset2Slice(*(*[]byte)(unsafe.Pointer(&question.Answers)), data)
			question.Answers = *(*[]string)(unsafe.Pointer(&slice))

			slice = database.Offset2Slice(*(*[]byte)(unsafe.Pointer(&question.CorrectAnswers)), data)
			question.CorrectAnswers = *(*[]int)(unsafe.Pointer(&slice))

			for j := 0; j < len(question.Answers); j++ {
				question.Answers[j] = database.Offset2String(question.Answers[j], data)
			}
		}
	case StepTypeProgramming:
		task, _ := Step2Programming(step)

		task.Description = database.Offset2String(task.Description, data)

		for i := 0; i < len(task.Checks); i++ {
			slice := database.Offset2Slice(*(*[]byte)(unsafe.Pointer(&task.Checks[i])), data)
			task.Checks[i] = *(*[]Check)(unsafe.Pointer(&slice))

			for j := 0; j < len(task.Checks[i]); j++ {
				check := &task.Checks[i][j]
				check.Input = database.Offset2String(check.Input, data)
				check.Output = database.Offset2String(check.Output, data)
			}
		}
	}
}

func DBLesson2Lesson(lesson *Lesson) {
	defer trace.End(trace.Begin(""))

	data := &lesson.Data[0]

	lesson.Name = database.Offset2String(lesson.Name, data)
	lesson.Theory = database.Offset2String(lesson.Theory, data)

	slice := database.Offset2Slice(*(*[]byte)(unsafe.Pointer(&lesson.Steps)), data)
	lesson.Steps = *(*[]Step)(unsafe.Pointer(&slice))

	for i := 0; i < len(lesson.Steps); i++ {
		DBStep2Step(&lesson.Steps[i], data)
	}

	slice = database.Offset2Slice(*(*[]byte)(unsafe.Pointer(&lesson.Submissions)), data)
	lesson.Submissions = *(*[]database.ID)(unsafe.Pointer(&slice))
}

func GetLessonByID(id database.ID, lesson *Lesson) error {
	defer trace.End(trace.Begin(""))

	if err := database.Read(LessonsDB, id, unsafe.Pointer(lesson), int(unsafe.Sizeof(*lesson))); err != nil {
		return err
	}

	DBLesson2Lesson(lesson)
	return nil
}

func GetLessons(pos *int64, lessons []Lesson) (int, error) {
	defer trace.End(trace.Begin(""))

	n, err := database.ReadMany(LessonsDB, pos, *(*[]byte)(unsafe.Pointer(&lessons)), int(unsafe.Sizeof(lessons[0])))
	if err != nil {
		return 0, err
	}

	for i := 0; i < n; i++ {
		DBLesson2Lesson(&lessons[i])
	}
	return n, nil
}

func Step2DBStep(ds *Step, ss *Step, data []byte, n int) int {
	defer trace.End(trace.Begin(""))

	ds.Draft = ss.Draft

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
			n += database.Slice2DBSlice((*[]byte)(unsafe.Pointer(&dq.Answers)), *(*[]byte)(unsafe.Pointer(&dq.Answers)), int(unsafe.Sizeof(dq.Answers[0])), int(unsafe.Alignof(dq.Answers[0])), data, n)

			n += database.Slice2DBSlice((*[]byte)(unsafe.Pointer(&dq.CorrectAnswers)), *(*[]byte)(unsafe.Pointer(&sq.CorrectAnswers)), int(unsafe.Sizeof(sq.CorrectAnswers[0])), int(unsafe.Alignof(sq.CorrectAnswers[0])), data, n)
		}
		n += database.Slice2DBSlice((*[]byte)(unsafe.Pointer(&dt.Questions)), *(*[]byte)(unsafe.Pointer(&dt.Questions)), int(unsafe.Sizeof(dt.Questions[0])), int(unsafe.Alignof(dt.Questions[0])), data, n)
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
			n += database.Slice2DBSlice((*[]byte)(unsafe.Pointer(&dt.Checks[i])), *(*[]byte)(unsafe.Pointer(&dt.Checks[i])), int(unsafe.Sizeof(dt.Checks[i][0])), int(unsafe.Alignof(dt.Checks[i][0])), data, n)
		}
	}

	return n
}

func SaveLesson(lesson *Lesson) error {
	defer trace.End(trace.Begin(""))

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
	n += database.Slice2DBSlice((*[]byte)(unsafe.Pointer(&lessonDB.Steps)), *(*[]byte)(unsafe.Pointer(&lessonDB.Steps)), int(unsafe.Sizeof(lessonDB.Steps[0])), int(unsafe.Alignof(lessonDB.Steps[0])), data, n)
	n += database.Slice2DBSlice((*[]byte)(unsafe.Pointer(&lessonDB.Submissions)), *(*[]byte)(unsafe.Pointer(&lesson.Submissions)), int(unsafe.Sizeof(lesson.Submissions[0])), int(unsafe.Alignof(lesson.Submissions[0])), data, n)

	return database.Write(LessonsDB, lessonDB.ID, unsafe.Pointer(&lessonDB), int(unsafe.Sizeof(lessonDB)))
}

func StepStringType(l Language, s *Step) string {
	defer trace.End(trace.Begin(""))

	switch s.Type {
	default:
		panic("invalid step type")
	case StepTypeTest:
		return Ls(l, "Test")
	case StepTypeProgramming:
		return Ls(l, "Programming task")
	}
}

func LessonContainerLink(containerType LessonContainerType) string {
	defer trace.End(trace.Begin(""))

	switch containerType {
	default:
		panic("invalid lesson container type")
	case LessonContainerCourse:
		return "/course"
	case LessonContainerSubject:
		return "/subject"
	}
}

func LessonContainerName(l Language, containerType LessonContainerType) string {
	defer trace.End(trace.Begin(""))

	switch containerType {
	default:
		panic("invalid lesson container type")
	case LessonContainerCourse:
		return Ls(l, "Course")
	case LessonContainerSubject:
		return Ls(l, "Subject")
	}
}

func DisplayLessons(w *http.Response, l Language, lessons []database.ID) {
	var lesson Lesson

	for i := 0; i < len(lessons); i++ {
		if err := GetLessonByID(lessons[i], &lesson); err != nil {
			/* TODO(anton2920): report error. */
		}

		DisplayFrameStart(w)

		w.WriteString(`<p><b>`)
		w.WriteString(Ls(l, "Lesson"))
		w.WriteString(` #`)
		w.WriteInt(i + 1)
		DisplayDraft(w, l, lesson.Flags == LessonDraft)
		w.WriteString(`</b></p>`)

		w.WriteString(`<p>`)
		w.WriteString(Ls(l, "Name"))
		w.WriteString(`: `)
		w.WriteHTMLString(lesson.Name)
		w.WriteString(`</p>`)

		w.WriteString(`<p>`)
		w.WriteString(Ls(l, "Theory"))
		w.WriteString(`: `)
		DisplayShortenedString(w, lesson.Theory, LessonTheoryMaxDisplayLen)
		w.WriteString(`</p>`)

		DisplayLessonLink(w, l, &lesson)

		DisplayFrameEnd(w)
	}
}

func DisplayLessonSubmissions(w *http.Response, l Language, lesson *Lesson, userID database.ID, who SubjectUserType) {
	var submission Submission
	var displayed bool

	switch who {
	case SubjectUserAdmin, SubjectUserTeacher:
		if len(lesson.Submissions) > 0 {
			for i := 0; i < len(lesson.Submissions); i++ {
				if err := GetSubmissionByID(lesson.Submissions[i], &submission); err != nil {
					/* TODO(anton2920): report error. */
				}

				if submission.Flags == SubmissionActive {
					if !displayed {
						w.WriteString(`<h3>`)
						w.WriteString(Ls(l, "Submissions"))
						w.WriteString(`</h3>`)
						w.WriteString(`<ul>`)
						displayed = true
					}

					w.WriteString(`<li>`)
					DisplaySubmissionLink(w, l, &submission)
					w.WriteString(`</li>`)
				}
			}
			if displayed {
				w.WriteString(`</ul>`)
			}
		}
	case SubjectUserStudent:
		si := -1

		for i := 0; i < len(lesson.Submissions); i++ {
			if err := GetSubmissionByID(lesson.Submissions[i], &submission); err != nil {
				/* TODO(anton2920): report error. */
			}

			if submission.UserID == userID {
				if submission.Flags == SubmissionActive {
					si = -1

					if !displayed {
						w.WriteString(`<h3>`)
						w.WriteString(Ls(l, "Submissions"))
						w.WriteString(`</h3>`)
						w.WriteString(`<ul>`)
						displayed = true
					}

					w.WriteString(`<li>`)
					DisplaySubmissionLink(w, l, &submission)
					w.WriteString(`</li>`)
				} else if submission.Flags == SubmissionDraft {
					si = i
				}
			}
		}
		if displayed {
			w.WriteString(`</ul>`)
		} else {
			w.WriteString(`<br>`)
		}

		if len(lesson.Steps) > 0 {
			w.WriteString(`<form method="POST" action="/submission/new">`)
			DisplayHiddenID(w, "ID", lesson.ID)
			if si == -1 {
				DisplayButton(w, l, "", "Pass")
			} else {
				DisplayHiddenInt(w, "SubmissionIndex", si)
				DisplayButton(w, l, "", "Edit")
			}
			w.WriteString(`</form>`)
		}
	}
}

func DisplayLessonTitle(w *http.Response, l Language, container string, lesson *Lesson) {
	w.WriteHTMLString(container)
	w.WriteString(`: `)
	w.WriteHTMLString(lesson.Name)
	DisplayDraft(w, l, lesson.Flags == LessonDraft)
}

func DisplayLessonLink(w *http.Response, l Language, lesson *Lesson) {
	w.WriteString(`<a href="/lesson/`)
	w.WriteID(lesson.ID)
	w.WriteString(`">`)
	w.WriteString(Ls(l, "Open"))
	w.WriteString(`</a>`)
}

func LessonPageHandler(w *http.Response, r *http.Request) error {
	defer trace.End(trace.Begin(""))

	const width = WidthLarge

	var container *LessonContainer
	var who SubjectUserType
	var lesson Lesson

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}

	id, err := GetIDFromURL(GL, r.URL, "/lesson/")
	if err != nil {
		return http.ClientError(err)
	}
	if err := GetLessonByID(id, &lesson); err != nil {
		if err == database.NotFound {
			return http.NotFound(Ls(GL, "lesson with this ID does not exist"))
		}
		return http.ServerError(err)
	}

	switch lesson.ContainerType {
	default:
		panic("invalid container type")
	case LessonContainerCourse:
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
		container = &course.LessonContainer
	case LessonContainerSubject:
		var subject Subject

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
		container = &subject.LessonContainer
	}

	DisplayHTMLStart(w)

	DisplayHeadStart(w)
	{
		w.WriteString(`<title>`)
		DisplayLessonTitle(w, GL, container.Name, &lesson)
		w.WriteString(`</title>`)
	}
	DisplayHeadEnd(w)

	DisplayBodyStart(w)
	{
		DisplayHeader(w, GL)
		DisplaySidebarWithLessons(w, GL, session, container.Lessons)

		DisplayMainStart(w)

		DisplayCrumbsStart(w, width)
		{
			DisplayCrumbsLinkID(w, LessonContainerLink(lesson.ContainerType), lesson.ContainerID, strings.Or(container.Name, LessonContainerName(GL, lesson.ContainerType)))
			DisplayCrumbsItemRaw(w, lesson.Name)
		}
		DisplayCrumbsEnd(w)

		DisplayPageStart(w, width)
		{
			w.WriteString(`<h2>`)
			DisplayLessonTitle(w, GL, container.Name, &lesson)
			w.WriteString(`</h2>`)
			w.WriteString(`<br>`)

			w.WriteString(`<h3>`)
			w.WriteString(Ls(GL, "Theory"))
			w.WriteString(`</h3>`)

			DisplayFrameStart(w)
			DisplayMarkdown(w, lesson.Theory)
			DisplayFrameEnd(w)

			if len(lesson.Steps) > 0 {
				w.WriteString(`<h3>`)
				w.WriteString(Ls(GL, "Evaluation"))
				w.WriteString(`</h3>`)

				for i := 0; i < len(lesson.Steps); i++ {
					step := &lesson.Steps[i]

					DisplayFrameStart(w)

					w.WriteString(`<p><b>`)
					w.WriteString(Ls(GL, "Step"))
					w.WriteString(` #`)
					w.WriteInt(i + 1)
					DisplayDraft(w, GL, step.Draft)
					w.WriteString(`</b></p>`)

					w.WriteString(`<p>`)
					w.WriteString(Ls(GL, "Name"))
					w.WriteString(`: `)
					w.WriteHTMLString(step.Name)
					w.WriteString(`</p>`)

					w.WriteString(`<span>`)
					w.WriteString(Ls(GL, "Type"))
					w.WriteString(`: `)
					w.WriteString(StepStringType(GL, step))
					w.WriteString(`</span>`)

					DisplayFrameEnd(w)
				}
			}

			DisplayLessonSubmissions(w, GL, &lesson, session.ID, who)
		}
		DisplayPageEnd(w)
		DisplayMainEnd(w)
	}
	DisplayBodyEnd(w)

	DisplayHTMLEnd(w)
	return nil
}

/* TODO(anton2920): check whether this function is needed. */
func StepDeepCopy(dst *Step, src *Step) {
	defer trace.End(trace.Begin(""))

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

func LessonsDeepCopy(dst *[]database.ID, src []database.ID, containerID database.ID, containerType LessonContainerType) {
	defer trace.End(trace.Begin(""))

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

func DisplayLessonsEditableList(w *http.Response, l Language, lessons []database.ID) {
	var lesson Lesson

	for i := 0; i < len(lessons); i++ {
		if err := GetLessonByID(lessons[i], &lesson); err != nil {
			/* TODO(anton2920): report error. */
		}

		DisplayFrameStart(w)

		w.WriteString(`<p><b>`)
		w.WriteString(Ls(GL, "Lesson"))
		w.WriteString(` #`)
		w.WriteInt(i + 1)
		DisplayDraft(w, l, lesson.Flags == LessonDraft)
		w.WriteString(`</b></p>`)

		w.WriteString(`<p>`)
		w.WriteString(Ls(GL, "Name"))
		w.WriteString(`: `)
		w.WriteHTMLString(lesson.Name)
		w.WriteString(`</p>`)

		w.WriteString(`<p>`)
		w.WriteString(Ls(GL, "Theory"))
		w.WriteString(`: `)
		DisplayShortenedString(w, lesson.Theory, LessonTheoryMaxDisplayLen)
		w.WriteString(`</p>`)

		DisplayIndexedCommand(w, l, i, "Edit")
		DisplayIndexedCommand(w, l, i, "Delete")
		if len(lessons) > 1 {
			if i > 0 {
				DisplayIndexedCommand(w, l, i, "↑")
			}
			if i < len(lessons)-1 {
				DisplayIndexedCommand(w, l, i, "↓")
			}
		}

		DisplayFrameEnd(w)
	}
}

func LessonFillFromRequest(vs url.Values, lesson *Lesson) {
	defer trace.End(trace.Begin(""))

	lesson.Name = vs.Get("Name")
	lesson.Theory = vs.Get("Theory")
}

func LessonVerify(l Language, lesson *Lesson) error {
	defer trace.End(trace.Begin(""))

	if !strings.LengthInRange(lesson.Name, MinNameLen, MaxNameLen) {
		return http.BadRequest(Ls(l, "lesson name length must be between %d and %d characters long"), MinNameLen, MaxNameLen)
	}

	if !strings.LengthInRange(lesson.Theory, MinTheoryLen, MaxTheoryLen) {
		return http.BadRequest(Ls(l, "lesson theory length must be between %d and %d characters long"), MinTheoryLen, MaxTheoryLen)
	}

	for si := 0; si < len(lesson.Steps); si++ {
		step := &lesson.Steps[si]

		switch step.Type {
		case StepTypeTest:
			if step.Draft {
				return http.BadRequest(Ls(l, "test %d is a draft"), si+1)
			}
		case StepTypeProgramming:
			if step.Draft {
				return http.BadRequest(Ls(l, "programming task %d is a draft"), si+1)
			}
		}
	}

	return nil
}

func LessonTestFillFromRequest(vs url.Values, test *StepTest) error {
	defer trace.End(trace.Begin(""))

	test.Name = vs.Get("Name")

	answerKey := make([]byte, 30)
	copy(answerKey, "Answer")

	correctAnswerKey := make([]byte, 30)
	copy(correctAnswerKey, "CorrectAnswer")

	questions := vs.GetMany("Question")
	for i := 0; i < len(questions); i++ {
		if i >= len(test.Questions) {
			test.Questions = append(test.Questions, Question{})
		}
		question := &test.Questions[i]
		question.Name = questions[i]

		n := slices.PutInt(answerKey[len("Answer"):], i)
		answers := vs.GetMany(unsafe.String(unsafe.SliceData(answerKey), len("Answer")+n))
		for j := 0; j < len(answers); j++ {
			if j >= len(question.Answers) {
				question.Answers = append(question.Answers, "")
			}
			question.Answers[j] = answers[j]
		}
		question.Answers = question.Answers[:len(answers)]

		n = slices.PutInt(correctAnswerKey[len("CorrectAnswer"):], i)
		correctAnswers := vs.GetMany(unsafe.String(unsafe.SliceData(correctAnswerKey), len("CorrectAnswer")+n))
		for j := 0; j < len(correctAnswers); j++ {
			if j >= len(question.CorrectAnswers) {
				question.CorrectAnswers = append(question.CorrectAnswers, 0)
			}

			var err error
			question.CorrectAnswers[j], err = GetValidIndex(correctAnswers[j], len(question.Answers))
			if err != nil {
				return http.ClientError(err)
			}
		}
		question.CorrectAnswers = question.CorrectAnswers[:len(correctAnswers)]
	}
	test.Questions = test.Questions[:len(questions)]

	return nil
}

func LessonTestVerify(l Language, test *StepTest) error {
	defer trace.End(trace.Begin(""))

	if !strings.LengthInRange(test.Name, MinStepNameLen, MaxStepNameLen) {
		return http.BadRequest(Ls(l, "test name length must be between %d and %d characters long"), MinStepNameLen, MaxStepNameLen)
	}

	for i := 0; i < len(test.Questions); i++ {
		question := &test.Questions[i]

		if !strings.LengthInRange(question.Name, MinQuestionLen, MaxQuestionLen) {
			return http.BadRequest(Ls(l, "question %d: title length must be between %d and %d characters long"), i+1, MinQuestionLen, MaxQuestionLen)
		}

		for j := 0; j < len(question.Answers); j++ {
			if !strings.LengthInRange(question.Answers[j], MinAnswerLen, MaxAnswerLen) {
				return http.BadRequest(Ls(l, "question %d: answer %d: length must be between %d and %d characters long"), i+1, j+1, MinAnswerLen, MaxAnswerLen)
			}
		}

		if len(question.CorrectAnswers) == 0 {
			return http.BadRequest(Ls(l, "question %d: select at least one correct answer"), i+1)
		}
	}

	return nil
}

func LessonAddTestPageHandler(w *http.Response, r *http.Request, session *Session, container *LessonContainer, lesson *Lesson, test *StepTest, err error) error {
	defer trace.End(trace.Begin(""))

	const width = WidthMedium + 1

	DisplayHTMLStart(w)

	DisplayHeadStart(w)
	{
		w.WriteString(`<title>`)
		w.WriteString(Ls(GL, "Test"))
		w.WriteString(`</title>`)

		if CSSEnabled {
			w.WriteString(`<style>.input-field{ padding: .375rem .75rem; width: 70%; font-size: 1rem; font-weight: 400; line-height: 1.5; color: var(--bs-body-color); -webkit-appearance: none; -moz-appearance: none; appearance: none; background-color: var(--bs-body-bg); background-clip: padding-box; border: var(--bs-border-width) solid var(--bs-border-color); border-radius: var(--bs-border-radius); transition: border-color .15s ease-in-out, box-shadow .15s ease-in-out; margin-left: 5px; } </style>`)
		}
	}
	DisplayHeadEnd(w)

	DisplayBodyStart(w)
	{
		DisplayHeader(w, GL)
		DisplaySidebar(w, GL, session)

		DisplayMainStart(w)

		DisplayFormStart(w, r, r.URL.Path)
		DisplayHiddenString(w, "CurrentPage", "Test")
		DisplayHiddenString(w, "LessonIndex", r.Form.Get("LessonIndex"))
		DisplayHiddenString(w, "StepIndex", r.Form.Get("StepIndex"))

		DisplayCrumbsStart(w, width)
		{
			DisplayCrumbsLinkID(w, LessonContainerLink(lesson.ContainerType), lesson.ContainerID, strings.Or(container.Name, LessonContainerName(GL, lesson.ContainerType)))
			DisplayCrumbsSubmit(w, GL, "Back2", "Edit lessons")
			DisplayCrumbsSubmitRaw(w, GL, "Back", strings.Or(lesson.Name, Ls(GL, "Lesson")))
			DisplayCrumbsItemRaw(w, strings.Or(test.Name, Ls(GL, "Test")))
		}
		DisplayCrumbsEnd(w)

		DisplayPageStart(w, width)
		{
			DisplayFormTitle(w, GL, "Test", err)

			DisplayLabel(w, GL, "Title")
			DisplayConstraintInput(w, "text", MinStepNameLen, MaxStepNameLen, "Name", test.Name, true)
			w.WriteString(`<br>`)

			if len(test.Questions) == 0 {
				test.Questions = append(test.Questions, Question{})
			}
			for i := 0; i < len(test.Questions); i++ {
				question := &test.Questions[i]

				DisplayFrameStart(w)

				w.WriteString(`<p><b>`)
				w.WriteString(Ls(GL, "Question"))
				w.WriteString(` #`)
				w.WriteInt(i + 1)
				w.WriteString(`</b></p>`)

				DisplayLabel(w, GL, "Title")
				DisplayConstraintInput(w, "text", MinQuestionLen, MaxQuestionLen, "Question", question.Name, true)
				w.WriteString(`<br>`)

				w.WriteString(`<p>`)
				w.WriteString(Ls(GL, "Answers (mark the correct ones)"))
				w.WriteString(`:</p>`)
				w.WriteString(`<ol>`)

				if len(question.Answers) == 0 {
					question.Answers = append(question.Answers, "")
				}
				for j := 0; j < len(question.Answers); j++ {
					answer := question.Answers[j]

					w.WriteString(`<li>`)

					w.WriteString(`<input type="checkbox" name="CorrectAnswer`)
					w.WriteInt(i)
					w.WriteString(`" value="`)
					w.WriteInt(j)
					w.WriteString(`"`)
					for k := 0; k < len(question.CorrectAnswers); k++ {
						correctAnswer := question.CorrectAnswers[k]
						if j == correctAnswer {
							w.WriteString(` checked`)
							break
						}
					}
					w.WriteString(`>`)

					DisplayConstraintIndexedInput(w, "text", MinAnswerLen, MaxAnswerLen, "Answer", i, answer, true)

					if len(question.Answers) > 1 {
						DisplayDoublyIndexedCommand(w, GL, i, j, "-")
						if j > 0 {
							DisplayDoublyIndexedCommand(w, GL, i, j, "↑")
						}
						if j < len(question.Answers)-1 {
							DisplayDoublyIndexedCommand(w, GL, i, j, "↓")
						}
					}

					w.WriteString(`</li>`)
				}
				w.WriteString(`</ol>`)

				DisplayIndexedCommand(w, GL, i, "Add another answer")
				if len(test.Questions) > 1 {
					w.WriteString(`<br><br>`)
					DisplayIndexedCommand(w, GL, i, "Delete")
					if i > 0 {
						DisplayIndexedCommand(w, GL, i, "↑")
					}
					if i < len(test.Questions)-1 {
						DisplayIndexedCommand(w, GL, i, "↓")
					}
				}

				DisplayFrameEnd(w)
			}

			DisplayCommand(w, GL, "Add another question")
			w.WriteString(`<br><br>`)

			DisplaySubmit(w, GL, "NextPage", "Continue", true)
		}
		DisplayPageEnd(w)
		DisplayFormEnd(w)
		DisplayMainEnd(w)
	}
	DisplayBodyEnd(w)

	DisplayHTMLEnd(w)
	return nil
}

func LessonProgrammingFillFromRequest(vs url.Values, task *StepProgramming) error {
	defer trace.End(trace.Begin(""))

	task.Name = vs.Get("Name")
	task.Description = vs.Get("Description")

	for i := 0; i < len(CheckKeys); i++ {
		checks := &task.Checks[i]

		inputs := vs.GetMany(CheckKeys[i][CheckKeyInput])
		outputs := vs.GetMany(CheckKeys[i][CheckKeyOutput])

		if len(inputs) != len(outputs) {
			return http.ClientError(nil)
		}

		for j := 0; j < len(inputs); j++ {
			if j >= len(*checks) {
				*checks = append(*checks, Check{})
			}
			check := &(*checks)[j]

			check.Input = inputs[j]
			check.Output = outputs[j]
		}
	}

	return nil
}

func LessonProgrammingVerify(task *StepProgramming) error {
	defer trace.End(trace.Begin(""))

	if !strings.LengthInRange(task.Name, MinStepNameLen, MaxStepNameLen) {
		return http.BadRequest("programming task name length must be between %d and %d characters long", MinStepNameLen, MaxStepNameLen)
	}

	if !strings.LengthInRange(task.Description, MinDescriptionLen, MaxDescriptionLen) {
		return http.BadRequest("programming task description length must be between %d and %d characters long", MinDescriptionLen, MaxDescriptionLen)
	}

	for i := 0; i < len(task.Checks); i++ {
		checks := task.Checks[i]

		for j := 0; j < len(checks); j++ {
			check := &checks[j]

			if !strings.LengthInRange(check.Input, MinCheckLen, MaxCheckLen) {
				return http.BadRequest("%s %d: input length must be between %d and %d characters long", CheckKeys[i][CheckKeyDisplay], j+1, MinCheckLen, MaxCheckLen)
			}

			if !strings.LengthInRange(check.Output, MinCheckLen, MaxCheckLen) {
				return http.BadRequest("%s %d: output length must be between %d and %d characters long", CheckKeys[i][CheckKeyDisplay], j+1, MinCheckLen, MaxCheckLen)
			}
		}
	}

	return nil
}

func LessonAddProgrammingDisplayChecks(w *http.Response, l Language, task *StepProgramming, checkType CheckType) {
	defer trace.End(trace.Begin(""))

	checks := task.Checks[checkType]

	w.WriteString(`<ol>`)
	for i := 0; i < len(checks); i++ {
		check := &checks[i]

		w.WriteString(`<li class="mt-3">`)

		w.WriteString(`<label>`)
		w.WriteString(Ls(l, "Input"))
		w.WriteString(`: `)
		DisplayConstraintInlineTextarea(w, MinCheckLen, MaxCheckLen, CheckKeys[checkType][CheckKeyInput], check.Input, true)
		w.WriteString(`</label> `)

		w.WriteString(`<label>`)
		w.WriteString(Ls(l, "output"))
		w.WriteString(`: `)
		DisplayConstraintInlineTextarea(w, MinCheckLen, MaxCheckLen, CheckKeys[checkType][CheckKeyOutput], check.Output, true)
		w.WriteString(`</label>`)

		DisplayDoublyIndexedCommand(w, l, i, int(checkType), "-")
		if len(checks) > 1 {
			if i > 0 {
				DisplayDoublyIndexedCommand(w, l, i, int(checkType), "↑")
			}
			if i < len(checks)-1 {
				DisplayDoublyIndexedCommand(w, l, i, int(checkType), "↓")
			}
		}

		w.WriteString(`</li>`)
	}
	w.WriteString(`</ol>`)
}

func LessonAddProgrammingPageHandler(w *http.Response, r *http.Request, session *Session, container *LessonContainer, lesson *Lesson, task *StepProgramming, err error) error {
	defer trace.End(trace.Begin(""))

	const width = WidthLarge

	DisplayHTMLStart(w)

	DisplayHeadStart(w)
	{
		w.WriteString(`<title>`)
		w.WriteString(Ls(GL, "Programming task"))
		w.WriteString(`</title>`)
	}
	DisplayHeadEnd(w)

	DisplayBodyStart(w)
	{
		DisplayHeader(w, GL)
		DisplaySidebar(w, GL, session)

		DisplayMainStart(w)

		DisplayFormStart(w, r, r.URL.Path)
		DisplayHiddenString(w, "CurrentPage", "Programming")
		DisplayHiddenString(w, "LessonIndex", r.Form.Get("LessonIndex"))
		DisplayHiddenString(w, "StepIndex", r.Form.Get("StepIndex"))

		DisplayCrumbsStart(w, width)
		{
			DisplayCrumbsLinkID(w, LessonContainerLink(lesson.ContainerType), lesson.ContainerID, strings.Or(container.Name, LessonContainerName(GL, lesson.ContainerType)))
			DisplayCrumbsSubmit(w, GL, "Back2", "Edit lessons")
			DisplayCrumbsSubmitRaw(w, GL, "Back", strings.Or(lesson.Name, Ls(GL, "Lesson")))
			DisplayCrumbsItemRaw(w, strings.Or(task.Name, Ls(GL, "Programming task")))
		}
		DisplayCrumbsEnd(w)

		DisplayPageStart(w, width)
		{
			DisplayFormTitle(w, GL, "Programming task", err)

			DisplayLabel(w, GL, "Name")
			DisplayConstraintInput(w, "text", MinStepNameLen, MaxStepNameLen, "Name", task.Name, true)
			w.WriteString(`<br>`)

			DisplayLabel(w, GL, "Description")
			DisplayConstraintTextarea(w, MinDescriptionLen, MaxDescriptionLen, "Description", task.Description, true)
			w.WriteString(`<br>`)

			w.WriteString(`<h4>`)
			w.WriteString(Ls(GL, "Examples"))
			w.WriteString(`</h4>`)
			LessonAddProgrammingDisplayChecks(w, GL, task, CheckTypeExample)
			DisplayCommand(w, GL, "Add example")
			w.WriteString(`<br><br>`)

			w.WriteString(`<h4>`)
			w.WriteString(Ls(GL, "Tests"))
			w.WriteString(`</h4>`)
			LessonAddProgrammingDisplayChecks(w, GL, task, CheckTypeTest)
			DisplayCommand(w, GL, "Add test")
			w.WriteString(`<br><br>`)

			DisplaySubmit(w, GL, "NextPage", "Continue", true)
		}
		DisplayPageEnd(w)
		DisplayFormEnd(w)
		DisplayMainEnd(w)
	}
	DisplayBodyEnd(w)

	DisplayHTMLEnd(w)
	return nil
}

func LessonStepVerify(l Language, step *Step) error {
	defer trace.End(trace.Begin(""))

	switch step.Type {
	default:
		panic("invalid step type")
	case StepTypeTest:
		test, _ := Step2Test(step)
		return LessonTestVerify(l, test)
	case StepTypeProgramming:
		task, _ := Step2Programming(step)
		return LessonProgrammingVerify(task)
	}
}

func LessonAddStepPageHandler(w *http.Response, r *http.Request, session *Session, container *LessonContainer, lesson *Lesson, step *Step, err error) error {
	defer trace.End(trace.Begin(""))

	switch step.Type {
	default:
		panic("invalid step type")
	case StepTypeTest:
		test, _ := Step2Test(step)
		return LessonAddTestPageHandler(w, r, session, container, lesson, test, err)
	case StepTypeProgramming:
		task, _ := Step2Programming(step)
		return LessonAddProgrammingPageHandler(w, r, session, container, lesson, task, err)
	}
}

func LessonAddPageHandler(w *http.Response, r *http.Request, session *Session, container *LessonContainer, lesson *Lesson, err error) error {
	defer trace.End(trace.Begin(""))

	const width = WidthMedium

	DisplayHTMLStart(w)

	DisplayHeadStart(w)
	{
		w.WriteString(`<title>`)
		w.WriteString(Ls(GL, "Lesson"))
		w.WriteString(`</title>`)
	}
	DisplayHeadEnd(w)

	DisplayBodyStart(w)
	{
		DisplayHeader(w, GL)
		DisplaySidebar(w, GL, session)

		DisplayMainStart(w)

		DisplayFormStart(w, r, r.URL.Path)
		DisplayHiddenString(w, "CurrentPage", "Lesson")
		DisplayHiddenString(w, "LessonIndex", r.Form.Get("LessonIndex"))

		DisplayCrumbsStart(w, width)
		{
			DisplayCrumbsLinkID(w, LessonContainerLink(lesson.ContainerType), lesson.ContainerID, strings.Or(container.Name, LessonContainerName(GL, lesson.ContainerType)))
			DisplayCrumbsSubmit(w, GL, "Back", "Edit lessons")
			DisplayCrumbsItemRaw(w, strings.Or(lesson.Name, Ls(GL, "Lesson")))
		}
		DisplayCrumbsEnd(w)

		DisplayPageStart(w, width)
		{
			DisplayFormTitle(w, GL, "Lesson", err)

			DisplayLabel(w, GL, "Name")
			DisplayConstraintInput(w, "text", MinNameLen, MaxNameLen, "Name", lesson.Name, true)
			w.WriteString(`<br>`)

			DisplayLabel(w, GL, "Theory")
			DisplayConstraintTextarea(w, MinTheoryLen, MaxTheoryLen, "Theory", lesson.Theory, true)
			w.WriteString(`<br>`)

			for i := 0; i < len(lesson.Steps); i++ {
				step := &lesson.Steps[i]

				DisplayFrameStart(w)

				w.WriteString(`<p><b>`)
				w.WriteString(Ls(GL, "Step"))
				w.WriteString(` #`)
				w.WriteInt(i + 1)
				DisplayDraft(w, GL, step.Draft)
				w.WriteString(`</b></p>`)

				w.WriteString(`<p>`)
				w.WriteString(Ls(GL, "Name"))
				w.WriteString(`: `)
				w.WriteHTMLString(step.Name)
				w.WriteString(`</p>`)

				w.WriteString(`<p>`)
				w.WriteString(Ls(GL, "Type"))
				w.WriteString(`: `)
				w.WriteString(StepStringType(GL, step))
				w.WriteString(`</p>`)

				DisplayIndexedCommand(w, GL, i, "Edit")
				DisplayIndexedCommand(w, GL, i, "Delete")
				if len(lesson.Steps) > 1 {
					if i > 0 {
						DisplayIndexedCommand(w, GL, i, "↑")
					}
					if i < len(lesson.Steps)-1 {
						DisplayIndexedCommand(w, GL, i, "↓")
					}
				}

				DisplayFrameEnd(w)
			}

			DisplayNextPage(w, GL, "Add test")
			DisplayNextPage(w, GL, "Add programming task")
			w.WriteString(`<br><br>`)

			DisplaySubmit(w, GL, "NextPage", "Next", true)
		}
		DisplayPageEnd(w)
		DisplayFormEnd(w)
		DisplayMainEnd(w)
	}
	DisplayBodyEnd(w)

	DisplayHTMLEnd(w)
	return nil
}

func LessonAddHandleCommand(w *http.Response, r *http.Request, l Language, session *Session, container *LessonContainer, currentPage, k, command string) error {
	defer trace.End(trace.Begin(""))

	var lesson Lesson

	/* TODO(anton2920): pass these as parameters. */
	pindex, spindex, sindex, ssindex, err := GetIndicies(k[len("Command"):])
	if err != nil {
		return http.ClientError(err)
	}

	switch currentPage {
	default:
		return http.ClientError(nil)
	case "Lesson":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), len(container.Lessons))
		if err != nil {
			return http.ClientError(err)
		}
		if err := GetLessonByID(container.Lessons[li], &lesson); err != nil {
			return http.ServerError(err)
		}
		defer SaveLesson(&lesson)

		switch command {
		case Ls(l, "Delete"):
			lesson.Steps = RemoveStepAtIndex(lesson.Steps, pindex)
		case Ls(l, "Edit"):
			if (pindex < 0) || (pindex >= len(lesson.Steps)) {
				return http.ClientError(nil)
			}
			step := &lesson.Steps[pindex]
			step.Draft = true

			r.Form.Set("StepIndex", spindex)
			return LessonAddStepPageHandler(w, r, session, container, &lesson, step, nil)
		case "↑", "^|":
			MoveStepUp(lesson.Steps, pindex)
		case "↓", "|v":
			MoveStepDown(lesson.Steps, pindex)
		}

		return LessonAddPageHandler(w, r, session, container, &lesson, nil)
	case "Test":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), len(container.Lessons))
		if err != nil {
			return http.ClientError(err)
		}
		if err := GetLessonByID(container.Lessons[li], &lesson); err != nil {
			return http.ServerError(err)
		}
		defer SaveLesson(&lesson)

		si, err := GetValidIndex(r.Form.Get("StepIndex"), len(lesson.Steps))
		if err != nil {
			return http.ClientError(err)
		}
		test, err := Step2Test(&lesson.Steps[si])
		if err != nil {
			return http.ClientError(err)
		}

		if err := LessonTestFillFromRequest(r.Form, test); err != nil {
			return http.ClientError(err)
		}

		switch command {
		case Ls(l, "Add another answer"):
			if (pindex < 0) || (pindex >= len(test.Questions)) {
				return http.ClientError(nil)
			}
			question := &test.Questions[pindex]
			question.Answers = append(question.Answers, "")
		case "-": /* remove answer */
			if (pindex < 0) || (pindex >= len(test.Questions)) {
				return http.ClientError(nil)
			}
			question := &test.Questions[pindex]

			if (sindex < 0) || (sindex >= len(question.Answers)) {
				return http.ClientError(nil)
			}
			question.Answers = util.RemoveStringAtIndex(question.Answers, sindex)

			for i := 0; i < len(question.CorrectAnswers); i++ {
				if question.CorrectAnswers[i] == sindex {
					question.CorrectAnswers = util.RemoveIntAtIndex(question.CorrectAnswers, i)
					i--
				} else if question.CorrectAnswers[i] > sindex {
					question.CorrectAnswers[i]--
				}
			}
		case Ls(l, "Add another question"):
			test.Questions = append(test.Questions, Question{})
		case Ls(l, "Delete"):
			test.Questions = RemoveQuestionAtIndex(test.Questions, pindex)
		case "↑", "^|":
			if ssindex == "" {
				MoveQuestionUp(test.Questions, pindex)
			} else {
				if (pindex < 0) || (pindex >= len(test.Questions)) {
					return http.ClientError(nil)
				}
				question := &test.Questions[pindex]

				util.MoveStringUp(question.Answers, sindex)
				for i := 0; i < len(question.CorrectAnswers); i++ {
					if question.CorrectAnswers[i] == sindex-1 {
						/* If previous answer is correct and current is not. */
						if (i == len(question.CorrectAnswers)-1) || (question.CorrectAnswers[i+1] != sindex) {
							question.CorrectAnswers[i] = sindex
						}
						break
					} else if question.CorrectAnswers[i] == sindex {
						/* If current answer is correct and previous is not. */
						if (i == 0) || (question.CorrectAnswers[i-1] != sindex-1) {
							question.CorrectAnswers[i] = sindex - 1
						}
						break
					}
				}
			}
		case "↓", "|v":
			if ssindex == "" {
				MoveQuestionDown(test.Questions, pindex)
			} else {
				if (pindex < 0) || (pindex >= len(test.Questions)) {
					return http.ClientError(nil)
				}
				question := &test.Questions[pindex]

				util.MoveStringDown(question.Answers, sindex)
				for i := 0; i < len(question.CorrectAnswers); i++ {
					if question.CorrectAnswers[i] == sindex {
						/* If current answer is correct and next is not. */
						if (i == len(question.CorrectAnswers)-1) || (question.CorrectAnswers[i+1] != sindex+1) {
							question.CorrectAnswers[i] = sindex + 1
						}
						break
					} else if question.CorrectAnswers[i] == sindex+1 {
						/* If next answer is correct and current is not. */
						if (i == 0) || (question.CorrectAnswers[i-1] != sindex) {
							question.CorrectAnswers[i] = sindex
						}
						break
					}
				}
			}
		}

		return LessonAddTestPageHandler(w, r, session, container, &lesson, test, nil)
	case "Programming":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), len(container.Lessons))
		if err != nil {
			return http.ClientError(err)
		}
		if err := GetLessonByID(container.Lessons[li], &lesson); err != nil {
			return http.ServerError(err)
		}
		defer SaveLesson(&lesson)

		si, err := GetValidIndex(r.Form.Get("StepIndex"), len(lesson.Steps))
		if err != nil {
			return http.ClientError(err)
		}
		task, err := Step2Programming(&lesson.Steps[si])
		if err != nil {
			return http.ClientError(nil)
		}

		if err := LessonProgrammingFillFromRequest(r.Form, task); err != nil {
			return http.ClientError(err)
		}

		switch command {
		case Ls(l, "Add example"):
			task.Checks[CheckTypeExample] = append(task.Checks[CheckTypeExample], Check{})
		case Ls(l, "Add test"):
			task.Checks[CheckTypeTest] = append(task.Checks[CheckTypeTest], Check{})
		case "-":
			if (sindex < 0) || (sindex >= len(task.Checks)) {
				return http.ClientError(nil)
			}
			task.Checks[sindex] = RemoveCheckAtIndex(task.Checks[sindex], pindex)
		case "↑", "^|":
			if (sindex < 0) || (sindex >= len(task.Checks)) {
				return http.ClientError(nil)
			}
			MoveCheckUp(task.Checks[sindex], pindex)
		case "↓", "|v":
			if (sindex < 0) || (sindex >= len(task.Checks)) {
				return http.ClientError(nil)
			}
			MoveCheckDown(task.Checks[sindex], pindex)
		}

		return LessonAddProgrammingPageHandler(w, r, session, container, &lesson, task, nil)
	}
}
