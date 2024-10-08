package main

import (
	"fmt"
	"unsafe"

	"github.com/anton2920/gofa/database"
	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/net/url"
	"github.com/anton2920/gofa/strings"
	"github.com/anton2920/gofa/syscall"
	"github.com/anton2920/gofa/trace"
	"github.com/anton2920/gofa/util"
)

type Course struct {
	LessonContainer

	Data [1024]byte
}

const (
	CourseActive int32 = iota
	CourseDeleted
	CourseDraft
)

const (
	MinNameLen = 1
	MaxNameLen = 45
)

func CreateCourse(course *Course) error {
	defer trace.End(trace.Begin(""))

	var err error

	course.ID, err = database.IncrementNextID(CoursesDB)
	if err != nil {
		return fmt.Errorf("failed to increment course ID: %w", err)
	}

	return SaveCourse(course)
}

func DBCourse2Course(course *Course) {
	defer trace.End(trace.Begin(""))

	data := &course.Data[0]

	course.Name = database.Offset2String(course.Name, data)

	slice := database.Offset2Slice(*(*[]byte)(unsafe.Pointer(&course.Lessons)), data)
	course.Lessons = *(*[]database.ID)(unsafe.Pointer(&slice))
}

func GetCourseByID(id database.ID, course *Course) error {
	defer trace.End(trace.Begin(""))

	if err := database.Read(CoursesDB, id, unsafe.Pointer(course), int(unsafe.Sizeof(*course))); err != nil {
		return err
	}

	DBCourse2Course(course)
	return nil
}

func GetCourses(pos *int64, courses []Course) (int, error) {
	defer trace.End(trace.Begin(""))

	n, err := database.ReadMany(CoursesDB, pos, *(*[]byte)(unsafe.Pointer(&courses)), int(unsafe.Sizeof(courses[0])))
	if err != nil {
		return 0, err
	}

	for i := 0; i < n; i++ {
		DBCourse2Course(&courses[i])
	}
	return n, nil
}

func DeleteCourseByID(id database.ID) error {
	defer trace.End(trace.Begin(""))

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
	defer trace.End(trace.Begin(""))

	var courseDB Course
	var n int

	courseDB.ID = course.ID
	courseDB.Flags = course.Flags

	/* TODO(anton2920): save up to a sizeof(course.Data). */
	data := unsafe.Slice(&courseDB.Data[0], len(courseDB.Data))
	n += database.String2DBString(&courseDB.Name, course.Name, data, n)
	n += database.Slice2DBSlice((*[]byte)(unsafe.Pointer(&courseDB.Lessons)), *(*[]byte)(unsafe.Pointer(&course.Lessons)), int(unsafe.Sizeof(course.Lessons[0])), int(unsafe.Alignof(course.Lessons[0])), data, n)

	return database.Write(CoursesDB, courseDB.ID, unsafe.Pointer(&courseDB), int(unsafe.Sizeof(courseDB)))
}

func DisplayCourseTitle(w *http.Response, l Language, course *Course, italics bool) {
	if len(course.Name) == 0 {
		if italics {
			w.WriteString(`<i>`)
		}
		w.WriteString(Ls(l, "Unnamed"))
		if italics {
			w.WriteString(`</i>`)
		}
	} else {
		w.WriteHTMLString(course.Name)
	}
	DisplayDraft(w, l, course.Flags == CourseDraft)
	DisplayDeleted(w, l, course.Flags == CourseDeleted)
}

func DisplayCourseLink(w *http.Response, l Language, course *Course) {
	w.WriteString(`<a href="/course/`)
	w.WriteID(course.ID)
	w.WriteString(`">`)
	DisplayCourseTitle(w, l, course, true)
	w.WriteString(`</a>`)
}

func CoursesPageHandler(w *http.Response, r *http.Request) error {
	defer trace.End(trace.Begin(""))

	const width = WidthLarge

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}

	if err := r.URL.ParseQuery(); err != nil {
		return http.ClientError(err)
	}

	var user User
	if err := GetUserByID(session.ID, &user); err != nil {
		return http.ServerError(err)
	}
	ncourses := len(user.Courses)

	var page int
	if r.URL.Query.Has("Page") {
		page, err = r.URL.Query.GetInt("Page")
		if err != nil {
			return http.ClientError(err)
		}
	}

	const coursesPerPage = 10
	npages := ncourses / coursesPerPage
	page = util.Clamp(page, 0, npages)

	DisplayHTMLStart(w)

	DisplayHeadStart(w)
	{
		w.WriteString(`<title>`)
		w.WriteString(Ls(GL, "Courses"))
		w.WriteString(`</title>`)
	}
	DisplayHeadEnd(w)

	DisplayBodyStart(w)
	{
		DisplayHeader(w, GL)
		DisplaySidebar(w, GL, session)

		DisplayMainStart(w)

		DisplayCrumbsStart(w, width)
		{
			DisplayCrumbsItem(w, GL, "Courses")
		}
		DisplayCrumbsEnd(w)

		DisplayPageStart(w, width)
		{
			w.WriteString(`<h2 class="text-center">`)
			w.WriteString(Ls(GL, "Courses"))
			w.WriteString(`</h2>`)
			w.WriteString(`<br>`)

			DisplayTableStart(w, GL, []string{"ID", "Name", "Lessons", "Status"})
			{
				var course Course

				start := page * coursesPerPage
				for i := start; i < min(len(user.Courses), start+coursesPerPage); i++ {
					id := user.Courses[i]
					if err := GetCourseByID(id, &course); err != nil {
						return http.ServerError(err)
					}

					DisplayTableRowLinkIDStart(w, "/course", course.ID)

					DisplayTableItemString(w, strings.Or(course.Name, Ls(GL, "Unnamed")))
					DisplayTableItemInt(w, len(course.Lessons))
					DisplayTableItemFlags(w, GL, course.Flags)

					DisplayTableRowEnd(w)
				}
			}
			DisplayTableEnd(w)

			DisplayPageSelector(w, "/courses", page, npages)
			w.WriteString(`<br>`)

			w.WriteString(`<form method="POST" action="/course/create">`)
			DisplaySubmit(w, GL, "", "Create course", true)
			w.WriteString(`</form>`)
		}
		DisplayPageEnd(w)
		DisplayMainEnd(w)
	}
	DisplayBodyEnd(w)

	DisplayHTMLEnd(w)
	return nil
}

func CoursePageHandler(w *http.Response, r *http.Request) error {
	defer trace.End(trace.Begin(""))

	const width = WidthLarge

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
		w.WriteString(`<title>`)
		DisplayCourseTitle(w, GL, &course, false)
		w.WriteString(`</title>`)
	}
	DisplayHeadEnd(w)

	DisplayBodyStart(w)
	{
		DisplayHeader(w, GL)
		DisplaySidebarWithLessons(w, GL, session, course.Lessons)

		DisplayMainStart(w)

		DisplayCrumbsStart(w, width)
		{
			DisplayCrumbsItemRaw(w, strings.Or(course.Name, Ls(GL, "Unnamed")))
		}
		DisplayCrumbsEnd(w)

		DisplayPageStart(w, width)
		{
			w.WriteString(`<h2>`)
			DisplayCourseTitle(w, GL, &course, true)
			w.WriteString(`</h2>`)
			w.WriteString(`<br>`)

			w.WriteString(`<h3>`)
			w.WriteString(Ls(GL, "Lessons"))
			w.WriteString(`</h3>`)
			DisplayLessons(w, GL, course.Lessons)

			w.WriteString(`<div>`)
			w.WriteString(`<form style="display:inline" method="POST" action="/course/edit">`)
			DisplayHiddenID(w, "ID", course.ID)
			DisplayButton(w, GL, "", "Edit")
			w.WriteString(`</form> `)

			w.WriteString(`<form style="display:inline" method="POST" action="/api/course/delete">`)
			DisplayHiddenID(w, "ID", course.ID)
			DisplayButton(w, GL, "", "Delete")
			w.WriteString(`</form>`)
			w.WriteString(`</div>`)
		}
		DisplayPageEnd(w)
		DisplayMainEnd(w)
	}
	DisplayBodyEnd(w)

	DisplayHTMLEnd(w)
	return nil
}

func CourseFillFromRequest(vs url.Values, course *Course) {
	defer trace.End(trace.Begin(""))

	course.Name = vs.Get("Name")
}

func CourseVerify(l Language, course *Course) error {
	defer trace.End(trace.Begin(""))

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

func CourseCreateEditCoursePageHandler(w *http.Response, r *http.Request, session *Session, course *Course, err error) error {
	defer trace.End(trace.Begin(""))

	const width = WidthSmall

	DisplayHTMLStart(w)

	DisplayHeadStart(w)
	{
		w.WriteString(`<title>`)
		w.WriteString(Ls(GL, "Course"))
		w.WriteString(`</title>`)
	}
	DisplayHeadEnd(w)

	DisplayBodyStart(w)
	{
		DisplayHeader(w, GL)
		DisplaySidebar(w, GL, session)

		DisplayMainStart(w)

		DisplayCrumbsStart(w, width)
		{
			DisplayCrumbsLinkID(w, "/course", course.ID, strings.Or(course.Name, Ls(GL, "Course")))
			DisplayCrumbsItem(w, GL, "Edit lessons")
		}
		DisplayCrumbsEnd(w)

		DisplayFormPageStart(w, r, GL, width, "Course", r.URL.Path, err)
		{
			DisplayHiddenString(w, "CurrentPage", "Course")

			DisplayLabel(w, GL, "Name")
			DisplayConstraintInput(w, "text", MinNameLen, MaxNameLen, "Name", course.Name, true)
			w.WriteString(`<br>`)

			DisplayLessonsEditableList(w, GL, course.Lessons)

			DisplayNextPage(w, GL, "Add lesson")
			w.WriteString(`<br><br>`)

			DisplaySubmit(w, GL, "NextPage", "Save", true)
		}
		DisplayFormPageEnd(w)
		DisplayMainEnd(w)
	}
	DisplayBodyEnd(w)

	DisplayHTMLEnd(w)
	return nil
}

func CourseCreateEditHandleCommand(w *http.Response, r *http.Request, l Language, session *Session, course *Course, currentPage, k, command string) error {
	defer trace.End(trace.Begin(""))

	var lesson Lesson

	pindex, spindex, _, _, err := GetIndicies(k[len("Command"):])
	if err != nil {
		return http.ClientError(err)
	}

	switch currentPage {
	default:
		return LessonAddHandleCommand(w, r, l, session, &course.LessonContainer, currentPage, k, command)
	case "Course":
		switch command {
		case Ls(l, "Delete"):
			course.Lessons = RemoveLessonAtIndex(course.Lessons, pindex)
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
			return LessonAddPageHandler(w, r, session, &course.LessonContainer, &lesson, nil)
		case "↑", "^|":
			MoveLessonUp(course.Lessons, pindex)
		case "↓", "|v":
			MoveLessonDown(course.Lessons, pindex)
		}

		return CourseCreateEditCoursePageHandler(w, r, session, course, nil)
	}
}

func CourseCreateEditPageHandler(w *http.Response, r *http.Request) error {
	defer trace.End(trace.Begin(""))

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
	course.Flags = CourseDraft
	defer SaveCourse(&course)

	for i := 0; i < len(r.Form.Keys); i++ {
		k := r.Form.Keys[i]

		/* 'command' is button, which modifies content of a current page. */
		if strings.StartsWith(k, "Command") {
			if len(r.Form.Values[i]) == 0 {
				continue
			}
			v := r.Form.Values[i][0]

			/* NOTE(anton2920): after command is executed, function must return. */
			return CourseCreateEditHandleCommand(w, r, GL, session, &course, currentPage, k, v)
		}
	}

	/* 'currentPage' is the page to save before leaving it. */
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
			return LessonAddTestPageHandler(w, r, session, &course.LessonContainer, &lesson, test, err)
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
			return LessonAddProgrammingPageHandler(w, r, session, &course.LessonContainer, &lesson, task, err)
		}
	}

	switch nextPage {
	default:
		return CourseCreateEditCoursePageHandler(w, r, session, &course, nil)
	case Ls(GL, "Back"):
		switch currentPage {
		default:
			return CourseCreateEditCoursePageHandler(w, r, session, &course, nil)
		case "Test", "Programming":
			return LessonAddPageHandler(w, r, session, &course.LessonContainer, &lesson, nil)
		}
	case Ls(GL, "Next"):
		if err := LessonVerify(GL, &lesson); err != nil {
			return LessonAddPageHandler(w, r, session, &course.LessonContainer, &lesson, err)
		}
		lesson.Flags = LessonActive
		if err := SaveLesson(&lesson); err != nil {
			return http.ServerError(err)
		}

		return CourseCreateEditCoursePageHandler(w, r, session, &course, nil)
	case Ls(GL, "Add lesson"):
		lesson.Flags = LessonDraft
		lesson.ContainerID = course.ID
		lesson.ContainerType = LessonContainerCourse

		if err := CreateLesson(&lesson); err != nil {
			return http.ServerError(err)
		}

		course.Lessons = append(course.Lessons, lesson.ID)
		r.Form.SetInt("LessonIndex", len(course.Lessons)-1)

		return LessonAddPageHandler(w, r, session, &course.LessonContainer, &lesson, nil)
	case Ls(GL, "Continue"):
		si, err := GetValidIndex(r.Form.Get("StepIndex"), len(lesson.Steps))
		if err != nil {
			return http.ClientError(err)
		}
		step := &lesson.Steps[si]
		if err := LessonStepVerify(GL, step); err != nil {
			return LessonAddStepPageHandler(w, r, session, &course.LessonContainer, &lesson, step, err)
		}
		step.Draft = false

		return LessonAddPageHandler(w, r, session, &course.LessonContainer, &lesson, nil)
	case Ls(GL, "Add test"):
		lesson.Flags = LessonDraft

		lesson.Steps = append(lesson.Steps, Step{StepCommon: StepCommon{Type: StepTypeTest, Draft: true}})
		test, _ := Step2Test(&lesson.Steps[len(lesson.Steps)-1])

		r.Form.SetInt("StepIndex", len(lesson.Steps)-1)
		return LessonAddTestPageHandler(w, r, session, &course.LessonContainer, &lesson, test, nil)
	case Ls(GL, "Add programming task"):
		lesson.Flags = LessonDraft

		lesson.Steps = append(lesson.Steps, Step{StepCommon: StepCommon{Type: StepTypeProgramming, Draft: true}})
		task, _ := Step2Programming(&lesson.Steps[len(lesson.Steps)-1])

		r.Form.SetInt("StepIndex", len(lesson.Steps)-1)
		return LessonAddProgrammingPageHandler(w, r, session, &course.LessonContainer, &lesson, task, nil)
	case Ls(GL, "Save"):
		if err := CourseVerify(GL, &course); err != nil {
			return CourseCreateEditCoursePageHandler(w, r, session, &course, err)
		}
		course.Flags = CourseActive

		w.RedirectID("/course/", course.ID, http.StatusSeeOther)
		return nil
	}
}

func CourseDeleteHandler(w *http.Response, r *http.Request) error {
	defer trace.End(trace.Begin(""))

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

	w.Redirect("/courses", http.StatusSeeOther)
	return nil
}
