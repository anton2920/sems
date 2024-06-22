package main

import (
	"fmt"
	"unsafe"

	"github.com/anton2920/gofa/database"
	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/net/url"
	"github.com/anton2920/gofa/strings"
	"github.com/anton2920/gofa/syscall"
)

type Course struct {
	ID    database.ID
	Flags int32

	Name    string
	Lessons []database.ID

	Data [1024]byte
}

const (
	CourseActive int32 = iota
	CourseDraft
	CourseDeleted
)

const (
	MinNameLen = 1
	MaxNameLen = 45
)

func CreateCourse(course *Course) error {
	var err error

	course.ID, err = database.IncrementNextID(CoursesDB)
	if err != nil {
		return fmt.Errorf("failed to increment course ID: %w", err)
	}

	return SaveCourse(course)
}

func DBCourse2Course(course *Course) {
	data := &course.Data[0]

	course.Name = database.Offset2String(course.Name, data)
	course.Lessons = database.Offset2Slice(course.Lessons, data)
}

func GetCourseByID(id database.ID, course *Course) error {
	if err := database.Read(CoursesDB, id, course); err != nil {
		return err
	}

	DBCourse2Course(course)
	return nil
}

func GetCourses(pos *int64, courses []Course) (int, error) {
	n, err := database.ReadMany(CoursesDB, pos, courses)
	if err != nil {
		return 0, err
	}

	for i := 0; i < n; i++ {
		DBCourse2Course(&courses[i])
	}
	return n, nil
}

func DeleteCourseByID(id database.ID) error {
	flags := CourseDeleted
	var course Course

	offset := int64(int(id)*int(unsafe.Sizeof(course))) + database.DataOffset + int64(unsafe.Offsetof(course.Flags))
	_, err := syscall.Pwrite(CoursesDB.FD, unsafe.Slice((*byte)(unsafe.Pointer(&flags)), unsafe.Sizeof(flags)), offset)
	if err != nil {
		return fmt.Errorf("failed to delete course from DB: %w", err)
	}

	return nil
}

func SaveCourse(course *Course) error {
	var courseDB Course
	var n int

	courseDB.ID = course.ID
	courseDB.Flags = course.Flags

	/* TODO(anton2920): save up to a sizeof(course.Data). */
	data := unsafe.Slice(&courseDB.Data[0], len(courseDB.Data))
	n += database.String2DBString(&courseDB.Name, course.Name, data, n)
	n += database.Slice2DBSlice(&courseDB.Lessons, course.Lessons, data, n)

	return database.Write(CoursesDB, courseDB.ID, &courseDB)
}

func DisplayCourseTitle(w *http.Response, l Language, course *Course, italics bool) {
	if len(course.Name) == 0 {
		if italics {
			w.AppendString(`<i>`)
		}
		w.AppendString(Ls(l, "Unnamed"))
		if italics {
			w.AppendString(`</i>`)
		}
	} else {
		w.WriteHTMLString(course.Name)
	}
	DisplayDraft(w, l, course.Flags == CourseDraft)
	DisplayDeleted(w, l, course.Flags == CourseDeleted)
}

func DisplayCourseLink(w *http.Response, l Language, course *Course) {
	w.AppendString(`<a href="/course/`)
	w.WriteID(course.ID)
	w.AppendString(`">`)
	DisplayCourseTitle(w, l, course, true)
	w.AppendString(`</a>`)
}

func CoursePageHandler(w *http.Response, r *http.Request) error {
	var course Course
	var user User

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}
	if err := GetUserByID(session.ID, &user); err != nil {
		return http.ServerError(err)
	}

	id, err := GetIDFromURL(GL, r.URL, "/course/")
	if err != nil {
		return err
	}
	if !UserOwnsCourse(&user, id) {
		return http.ForbiddenError
	}
	if err := GetCourseByID(id, &course); err != nil {
		if err == database.NotFound {
			return http.NotFound(Ls(GL, "course with this ID does not exist"))
		}
		return http.ServerError(err)
	}

	DisplayHTMLStart(w)

	DisplayHeadStart(w)
	{
		w.AppendString(`<title>`)
		DisplayCourseTitle(w, GL, &course, false)
		w.AppendString(`</title>`)
	}
	DisplayHeadEnd(w)

	DisplayBodyStart(w)
	{
		DisplayHeader(w, GL)
		DisplaySidebar(w, GL, session.ID)

		DisplayPageStart(w)
		{
			w.AppendString(`<h2>`)
			DisplayCourseTitle(w, GL, &course, true)
			w.AppendString(`</h2>`)
			w.AppendString(`<br>`)

			w.AppendString(`<h3>`)
			w.AppendString(Ls(GL, "Lessons"))
			w.AppendString(`</h3>`)
			DisplayLessons(w, GL, course.Lessons)

			w.AppendString(`<div>`)
			w.AppendString(`<form style="display:inline" method="POST" action="/course/edit">`)
			DisplayHiddenID(w, "ID", course.ID)
			DisplayButton(w, GL, "", "Edit")
			w.AppendString(`</form> `)

			w.AppendString(`<form style="display:inline" method="POST" action="/api/course/delete">`)
			DisplayHiddenID(w, "ID", course.ID)
			DisplayButton(w, GL, "", "Delete")
			w.AppendString(`</form>`)
			w.AppendString(`</div>`)
		}
		DisplayPageEnd(w)
	}
	DisplayBodyEnd(w)

	DisplayHTMLEnd(w)
	return nil
}

func CourseFillFromRequest(vs url.Values, course *Course) {
	course.Name = vs.Get("Name")
}

func CourseVerify(l Language, course *Course) error {
	var lesson Lesson

	if !strings.LengthInRange(course.Name, MinNameLen, MaxNameLen) {
		return http.BadRequest(Ls(l, "course name length must be between %d and %d characters long"), MinNameLen, MaxNameLen)
	}

	if len(course.Lessons) == 0 {
		return http.BadRequest(Ls(l, "create at least one lesson"))
	}
	for li := 0; li < len(course.Lessons); li++ {
		if err := GetLessonByID(course.Lessons[li], &lesson); err != nil {
			return http.ServerError(err)
		}
		if lesson.Flags == LessonDraft {
			return http.BadRequest(Ls(l, "lesson %d is a draft"), li+1)
		}
	}

	return nil
}

func CourseCreateEditCoursePageHandler(w *http.Response, r *http.Request, session *Session, course *Course) error {
	DisplayHTMLStart(w)

	DisplayHeadStart(w)
	{
		w.AppendString(`<title>`)
		w.AppendString(Ls(GL, "Course"))
		w.AppendString(`</title>`)
	}
	DisplayHeadEnd(w)

	DisplayBodyStart(w)
	{
		DisplayHeader(w, GL)
		DisplaySidebar(w, GL, session.ID)

		DisplayFormStart(w, r, GL, "Course", r.URL.Path, 4)
		{
			DisplayHiddenString(w, "CurrentPage", "Course")

			DisplayInputLabel(w, GL, "Name")
			DisplayConstraintInput(w, "text", MinNameLen, MaxNameLen, "Name", course.Name, true)
			w.AppendString(`<br>`)

			DisplayLessonsEditableList(w, GL, course.Lessons)

			DisplayNextPage(w, GL, "Add lesson")
			w.AppendString(`<br><br>`)

			DisplaySubmit(w, GL, "NextPage", "Save", true)
		}
		DisplayFormEnd(w)
	}
	DisplayBodyEnd(w)

	DisplayHTMLEnd(w)
	return nil
}

func CourseCreateEditHandleCommand(w *http.Response, r *http.Request, l Language, session *Session, course *Course, currentPage, k, command string) error {
	var lesson Lesson

	pindex, spindex, _, _, err := GetIndicies(k[len("Command"):])
	if err != nil {
		return http.ClientError(err)
	}

	switch currentPage {
	default:
		return LessonAddHandleCommand(w, r, l, session, course.Lessons, currentPage, k, command)
	case "Course":
		switch command {
		case Ls(l, "Delete"):
			course.Lessons = RemoveAtIndex(course.Lessons, pindex)
		case Ls(l, "Edit"):
			if (pindex < 0) || (pindex >= len(course.Lessons)) {
				return http.ClientError(nil)
			}
			if err := GetLessonByID(course.Lessons[pindex], &lesson); err != nil {
				return http.ServerError(err)
			}
			lesson.Flags = LessonDraft
			if err := SaveLesson(&lesson); err != nil {
				return http.ServerError(err)
			}

			r.Form.Set("LessonIndex", spindex)
			return LessonAddPageHandler(w, r, session, &lesson)
		case "↑", "^|":
			MoveUp(course.Lessons, pindex)
		case "↓", "|v":
			MoveDown(course.Lessons, pindex)
		}

		return CourseCreateEditCoursePageHandler(w, r, session, course)
	}
}

func CourseCreateEditPageHandler(w *http.Response, r *http.Request) error {
	var course Course
	var lesson Lesson
	var user User

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}
	if err := GetUserByID(session.ID, &user); err != nil {
		return http.ServerError(err)
	}

	if err := r.ParseForm(); err != nil {
		return http.ClientError(err)
	}

	currentPage := r.Form.Get("CurrentPage")
	nextPage := r.Form.Get("NextPage")

	if r.Form.Get("ID") == "" {
		if err := CreateCourse(&course); err != nil {
			return http.ServerError(err)
		}

		user.Courses = append(user.Courses, course.ID)
		if err := SaveUser(&user); err != nil {
			return http.ServerError(err)
		}

		r.Form.SetInt("ID", int(course.ID))
	} else {
		courseID, err := r.Form.GetID("ID")
		if err != nil {
			return http.ClientError(err)
		}
		if !UserOwnsCourse(&user, courseID) {
			return http.ForbiddenError
		}
		if err := GetCourseByID(courseID, &course); err != nil {
			if err == database.NotFound {
				return http.NotFound("course with this ID does not exist")
			}
			return http.ServerError(err)
		}
	}
	defer SaveCourse(&course)

	course.Flags = CourseDraft

	for i := 0; i < len(r.Form); i++ {
		k := r.Form[i].Key
		if len(r.Form[i].Values) == 0 {
			continue
		}
		v := r.Form[i].Values[0]

		/* 'command' is button, which modifies content of a current page. */
		if strings.StartsWith(k, "Command") {
			/* NOTE(anton2920): after command is executed, function must return. */
			return CourseCreateEditHandleCommand(w, r, GL, session, &course, currentPage, k, v)
		}
	}

	/* 'currentPage' is the page to save/check before leaving it. */
	switch currentPage {
	case "Course":
		CourseFillFromRequest(r.Form, &course)
	case "Lesson":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), len(course.Lessons))
		if err != nil {
			return http.ClientError(err)
		}
		if err := GetLessonByID(course.Lessons[li], &lesson); err != nil {
			return http.ServerError(err)
		}
		defer SaveLesson(&lesson)

		LessonFillFromRequest(r.Form, &lesson)
	case "Test":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), len(course.Lessons))
		if err != nil {
			return http.ClientError(err)
		}
		if err := GetLessonByID(course.Lessons[li], &lesson); err != nil {
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
			return WritePageEx(w, r, session, LessonAddTestPageHandler, test, err)
		}
		if err := LessonTestVerify(GL, test); err != nil {
			return WritePageEx(w, r, session, LessonAddTestPageHandler, test, err)
		}
	case "Programming":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), len(course.Lessons))
		if err != nil {
			return http.ClientError(err)
		}
		if err := GetLessonByID(course.Lessons[li], &lesson); err != nil {
			return http.ServerError(err)
		}
		defer SaveLesson(&lesson)

		si, err := GetValidIndex(r.Form.Get("StepIndex"), len(lesson.Steps))
		if err != nil {
			return http.ClientError(err)
		}
		task, err := Step2Programming(&lesson.Steps[si])
		if err != nil {
			return http.ClientError(err)
		}

		if err := LessonProgrammingFillFromRequest(r.Form, task); err != nil {
			return WritePageEx(w, r, session, LessonAddProgrammingPageHandler, task, err)
		}
		if err := LessonProgrammingVerify(task); err != nil {
			return WritePageEx(w, r, session, LessonAddProgrammingPageHandler, task, err)
		}
	}

	switch nextPage {
	default:
		return CourseCreateEditCoursePageHandler(w, r, session, &course)
	case Ls(GL, "Next"):
		if err := LessonVerify(GL, &lesson); err != nil {
			return WritePageEx(w, r, session, LessonAddPageHandler, &lesson, err)
		}
		lesson.Flags = LessonActive
		if err := SaveLesson(&lesson); err != nil {
			return http.ServerError(err)
		}

		return CourseCreateEditCoursePageHandler(w, r, session, &course)
	case Ls(GL, "Add lesson"):
		lesson.Flags = LessonDraft
		lesson.ContainerID = course.ID
		lesson.ContainerType = ContainerTypeCourse

		if err := CreateLesson(&lesson); err != nil {
			return http.ServerError(err)
		}

		course.Lessons = append(course.Lessons, lesson.ID)
		r.Form.SetInt("LessonIndex", len(course.Lessons)-1)

		return LessonAddPageHandler(w, r, session, &lesson)
	case Ls(GL, "Continue"):
		si, err := GetValidIndex(r.Form.Get("StepIndex"), len(lesson.Steps))
		if err != nil {
			return http.ClientError(err)
		}
		step := &lesson.Steps[si]
		step.Draft = false

		return LessonAddPageHandler(w, r, session, &lesson)
	case Ls(GL, "Add test"):
		lesson.Flags = LessonDraft

		lesson.Steps = append(lesson.Steps, Step{StepCommon: StepCommon{Type: StepTypeTest, Draft: true}})
		test, _ := Step2Test(&lesson.Steps[len(lesson.Steps)-1])

		r.Form.SetInt("StepIndex", len(lesson.Steps)-1)
		return LessonAddTestPageHandler(w, r, session, test)
	case Ls(GL, "Add programming task"):
		lesson.Flags = LessonDraft

		lesson.Steps = append(lesson.Steps, Step{StepCommon: StepCommon{Type: StepTypeProgramming, Draft: true}})
		task, _ := Step2Programming(&lesson.Steps[len(lesson.Steps)-1])

		r.Form.SetInt("StepIndex", len(lesson.Steps)-1)
		return LessonAddProgrammingPageHandler(w, r, session, task)
	case Ls(GL, "Save"):
		if err := CourseVerify(GL, &course); err != nil {
			return WritePageEx(w, r, session, CourseCreateEditCoursePageHandler, &course, err)
		}
		course.Flags = CourseActive

		w.RedirectID("/course/", course.ID, http.StatusSeeOther)
		return nil
	}
}

func CourseDeleteHandler(w *http.Response, r *http.Request) error {
	var user User

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}

	if err := r.ParseForm(); err != nil {
		return http.ClientError(err)
	}

	courseID, err := r.Form.GetID("ID")
	if err != nil {
		return http.ClientError(err)
	}

	if err := GetUserByID(session.ID, &user); err != nil {
		return http.ServerError(err)
	}
	if !UserOwnsCourse(&user, courseID) {
		return http.ForbiddenError
	}

	if err := DeleteCourseByID(courseID); err != nil {
		return http.ServerError(err)
	}

	w.Redirect("/", http.StatusSeeOther)
	return nil
}
