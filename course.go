package main

import (
	"fmt"
	"strconv"
)

type Course struct {
	Name    string
	Lessons []*Lesson

	/* TODO(anton2920): I don't like this. Replace with 'pointer|1'. */
	Draft bool
}

const (
	MinNameLen = 1
	MaxNameLen = 45
)

func CoursePageHandler(w *HTTPResponse, r *HTTPRequest) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}
	user := &DB.Users[session.ID]

	id, err := GetIDFromURL(r.URL, "/course/")
	if err != nil {
		return err
	}
	if (id < 0) || (id > len(user.Courses)) {
		return NotFoundError
	}
	course := user.Courses[id]

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>`)
	w.WriteHTMLString(course.Name)
	w.AppendString(`</title></head>`)

	w.AppendString(`<body>`)

	w.AppendString(`<h1>`)
	w.WriteHTMLString(course.Name)
	if course.Draft {
		w.AppendString(` (draft)`)
	}
	w.AppendString(`</h1>`)

	w.AppendString(`<h2>Lessons</h2>`)
	for i := 0; i < len(course.Lessons); i++ {
		lesson := course.Lessons[i]

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
		LessonDisplayTheory(w, lesson.Theory)
		w.AppendString(`</p>`)

		w.AppendString(`</fieldset>`)
		w.AppendString(`<br>`)
	}

	w.AppendString(`<div>`)

	w.AppendString(`<form style="display:inline" method="POST" action="/course/edit">`)
	w.AppendString(`<input type="hidden" name="ID" value="`)
	w.WriteHTMLString(r.URL.Path[len("/course/"):])
	w.AppendString(`">`)
	w.AppendString(`<input type="submit" value="Edit">`)
	w.AppendString(`</form>`)
	w.AppendString("\r\n")
	w.AppendString(`<form style="display:inline" method="POST" action="/api/course/delete">`)
	w.AppendString(`<input type="hidden" name="ID" value="`)
	w.WriteHTMLString(r.URL.Path[len("/course/"):])
	w.AppendString(`">`)
	w.AppendString(`<input type="submit" value="Delete">`)
	w.AppendString(`</form>`)

	w.AppendString(`</div>`)

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func CourseCreateEditCourseVerifyRequest(vs URLValues, course *Course) error {
	course.Name = vs.Get("Name")
	if !StringLengthInRange(course.Name, MinNameLen, MaxNameLen) {
		return NewHTTPError(HTTPStatusBadRequest, fmt.Sprintf("course name length must be between %d and %d characters long", MinNameLen, MaxNameLen))
	}

	return nil
}

func CourseCreateEditCoursePageHandler(w *HTTPResponse, r *HTTPRequest, course *Course) error {
	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>Create course</title></head>`)
	w.AppendString(`<body>`)
	w.AppendString(`<h1>Course</h1>`)

	ErrorDiv(w, r.Form.Get("Error"))

	w.AppendString(`<form method="POST" action="`)
	w.WriteString(r.URL.Path)
	w.AppendString(`">`)

	w.AppendString(`<input type="hidden" name="CurrentPage" value="Course">`)

	w.AppendString(`<input type="hidden" name="ID" value="`)
	w.WriteHTMLString(r.Form.Get("ID"))
	w.AppendString(`">`)

	w.AppendString(`<label>Name: `)
	w.AppendString(`<input type="text" minlength="1" maxlength="45" name="Name" value="`)
	w.WriteHTMLString(course.Name)
	w.AppendString(`" required>`)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	for i := 0; i < len(course.Lessons); i++ {
		lesson := course.Lessons[i]

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
		LessonDisplayTheory(w, lesson.Theory)
		w.AppendString(`</p>`)

		w.AppendString(`<input type="submit" name="Command`)
		w.WriteInt(i)
		w.AppendString(`" value="Edit" formnovalidate>`)
		w.AppendString("\r\n")
		w.AppendString(`<input type="submit" name="Command`)
		w.WriteInt(i)
		w.AppendString(`" value="Delete" formnovalidate>`)
		if len(course.Lessons) > 1 {
			if i > 0 {
				w.AppendString("\r\n")
				w.AppendString(`<input type="submit" name="Command`)
				w.WriteInt(i)
				w.AppendString(`" value="↑" formnovalidate>`)
			}
			if i < len(course.Lessons)-1 {
				w.AppendString("\r\n")
				w.AppendString(`<input type="submit" name="Command`)
				w.WriteInt(i)
				w.AppendString(`" value="↓" formnovalidate>`)
			}
		}

		w.AppendString(`</fieldset>`)
		w.AppendString(`<br>`)
	}

	w.AppendString(`<input type="submit" name="NextPage" value="Add lesson">`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<input type="submit" name="NextPage" value="Save">`)

	w.AppendString(`</form>`)

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func CourseCreateEditHandleCommand(w *HTTPResponse, r *HTTPRequest, course *Course, currentPage, k, command string) error {
	pindex, spindex, _, _, err := GetIndicies(k[len("Command"):])
	if err != nil {
		return ReloadPageError
	}

	switch currentPage {
	default:
		return LessonHandleCommand(w, r, course.Lessons, currentPage, k, command)
	case "Course":
		switch command {
		case "Delete":
			course.Lessons = RemoveAtIndex(course.Lessons, pindex)
		case "Edit":
			if (pindex < 0) || (pindex >= len(course.Lessons)) {
				return ReloadPageError
			}
			lesson := course.Lessons[pindex]
			lesson.Draft = true

			r.Form.Set("LessonIndex", spindex)
			return LessonPageHandler(w, r, lesson)
		case "↑", "^|":
			MoveUp(course.Lessons, pindex)
		case "↓", "|v":
			MoveDown(course.Lessons, pindex)
		}

		return CourseCreateEditCoursePageHandler(w, r, course)
	}
}

func CourseCreateEditPageHandler(w *HTTPResponse, r *HTTPRequest) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}
	if session.ID != AdminID {
		return ForbiddenError
	}
	user := &DB.Users[session.ID]

	if err := r.ParseForm(); err != nil {
		return ReloadPageError
	}

	currentPage := r.Form.Get("CurrentPage")
	nextPage := r.Form.Get("NextPage")

	var course *Course
	id := r.Form.Get("ID")
	if id == "" {
		course = new(Course)
		user.Courses = append(user.Courses, course)
		r.Form.Set("ID", strconv.Itoa(len(user.Courses)-1))
	} else {
		ci, err := strconv.Atoi(id)
		if (err != nil) || (ci < 0) || (ci >= len(user.Courses)) {
			return ReloadPageError
		}
		course = user.Courses[ci]
	}
	course.Draft = true

	for i := 0; i < len(r.Form); i++ {
		k := r.Form[i].Key
		v := r.Form[i].Values[0]

		/* 'command' is button, which modifies content of a current page. */
		if StringStartsWith(k, "Command") {
			/* NOTE(anton2920): after command is executed, function must return. */
			return CourseCreateEditHandleCommand(w, r, course, currentPage, k, v)
		}
	}

	/* 'currentPage' is the page to check before leaving it. */
	switch currentPage {
	case "Course":
		if err := CourseCreateEditCourseVerifyRequest(r.Form, course); err != nil {
			return WritePageEx(w, r, CourseCreateEditCoursePageHandler, course, err)
		}
	case "Lesson":
		li, err := strconv.Atoi(r.Form.Get("LessonIndex"))
		if (err != nil) || (li < 0) || (li >= len(course.Lessons)) {
			return ReloadPageError
		}
		lesson := course.Lessons[li]

		if err := LessonVerifyRequest(r.Form, lesson); err != nil {
			return WritePageEx(w, r, LessonPageHandler, lesson, err)
		}
	case "Test":
		li, err := strconv.Atoi(r.Form.Get("LessonIndex"))
		if (err != nil) || (li < 0) || (li >= len(course.Lessons)) {
			return ReloadPageError
		}
		lesson := course.Lessons[li]

		si, err := strconv.Atoi(r.Form.Get("StepIndex"))
		if (err != nil) || (si < 0) || (si >= len(lesson.Steps)) {
			return ReloadPageError
		}
		test, ok := lesson.Steps[si].(*StepTest)
		if !ok {
			return ReloadPageError
		}

		if err := TestVerifyRequest(r.Form, test, true); err != nil {
			return WritePageEx(w, r, TestPageHandler, test, err)
		}
	case "Programming":
		li, err := strconv.Atoi(r.Form.Get("LessonIndex"))
		if (err != nil) || (li < 0) || (li >= len(course.Lessons)) {
			return ReloadPageError
		}
		lesson := course.Lessons[li]

		si, err := strconv.Atoi(r.Form.Get("StepIndex"))
		if (err != nil) || (si < 0) || (si >= len(lesson.Steps)) {
			return ReloadPageError
		}
		task, ok := lesson.Steps[si].(*StepProgramming)
		if !ok {
			return ReloadPageError
		}

		if err := ProgrammingVerifyRequest(r.Form, task, true); err != nil {
			return WritePageEx(w, r, ProgrammingPageHandler, task, err)
		}
	}

	switch nextPage {
	default:
		return CourseCreateEditCoursePageHandler(w, r, course)
	case "Next":
		li, err := strconv.Atoi(r.Form.Get("LessonIndex"))
		if (err != nil) || (li < 0) || (li >= len(course.Lessons)) {
			return ReloadPageError
		}
		lesson := course.Lessons[li]

		for si := 0; si < len(lesson.Steps); si++ {
			switch step := lesson.Steps[si].(type) {
			case *StepTest:
				if step.Draft {
					return WritePageEx(w, r, LessonPageHandler, lesson, NewHTTPError(HTTPStatusBadRequest, fmt.Sprintf("test %d is a draft", si+1)))
				}
			case *StepProgramming:
				if step.Draft {
					return WritePageEx(w, r, LessonPageHandler, lesson, NewHTTPError(HTTPStatusBadRequest, fmt.Sprintf("programming task %d is a draft", si+1)))
				}
			}
		}
		lesson.Draft = false

		return CourseCreateEditCoursePageHandler(w, r, course)
	case "Add lesson":
		lesson := new(Lesson)
		lesson.Draft = true
		course.Lessons = append(course.Lessons, lesson)

		r.Form.Set("LessonIndex", strconv.Itoa(len(course.Lessons)-1))
		return LessonPageHandler(w, r, lesson)
	case "Continue":
		li, err := strconv.Atoi(r.Form.Get("LessonIndex"))
		if (err != nil) || (li < 0) || (li >= len(course.Lessons)) {
			return ReloadPageError
		}
		lesson := course.Lessons[li]

		si, err := strconv.Atoi(r.Form.Get("StepIndex"))
		if (err != nil) || (si < 0) || (si >= len(lesson.Steps)) {
			return ReloadPageError
		}
		switch step := lesson.Steps[si].(type) {
		default:
			panic("invalid step type")
		case *StepTest:
			step.Draft = false
		case *StepProgramming:
			step.Draft = false
		}

		return LessonPageHandler(w, r, lesson)
	case "Add test":
		li, err := strconv.Atoi(r.Form.Get("LessonIndex"))
		if (err != nil) || (li < 0) || (li >= len(course.Lessons)) {
			return ReloadPageError
		}
		lesson := course.Lessons[li]
		lesson.Draft = true

		test := new(StepTest)
		test.Draft = true
		lesson.Steps = append(lesson.Steps, test)

		r.Form.Set("StepIndex", strconv.Itoa(len(lesson.Steps)-1))
		return TestPageHandler(w, r, test)
	case "Add programming task":
		li, err := strconv.Atoi(r.Form.Get("LessonIndex"))
		if (err != nil) || (li < 0) || (li >= len(course.Lessons)) {
			return ReloadPageError
		}
		lesson := course.Lessons[li]
		lesson.Draft = true

		task := new(StepProgramming)
		task.Draft = true
		lesson.Steps = append(lesson.Steps, task)

		r.Form.Set("StepIndex", strconv.Itoa(len(lesson.Steps)-1))
		return ProgrammingPageHandler(w, r, task)
	case "Save":
		return CourseCreateEditHandler(w, r)
	}
}

func CourseCreateEditHandler(w *HTTPResponse, r *HTTPRequest) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}
	user := &DB.Users[session.ID]

	if err := r.ParseForm(); err != nil {
		return ReloadPageError
	}

	id := r.Form.Get("ID")
	courseID, err := strconv.Atoi(id)
	if (err != nil) || (courseID < 0) || (courseID >= len(user.Courses)) {
		return ReloadPageError
	}
	course := user.Courses[courseID]

	if len(course.Lessons) == 0 {
		return WritePageEx(w, r, CourseCreateEditCoursePageHandler, course, NewHTTPError(HTTPStatusBadRequest, "create at least one lesson"))
	}
	for li := 0; li < len(course.Lessons); li++ {
		lesson := course.Lessons[li]
		if lesson.Draft {
			return WritePageEx(w, r, CourseCreateEditCoursePageHandler, course, NewHTTPError(HTTPStatusBadRequest, fmt.Sprintf("lesson %d is a draft", li+1)))
		}
	}
	course.Draft = false

	w.Redirect(fmt.Appendf(make([]byte, 0, 20), "/course/%s", id), HTTPStatusSeeOther)
	return nil
}

func CourseDeleteHandler(w *HTTPResponse, r *HTTPRequest) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}
	user := &DB.Users[session.ID]

	if err := r.ParseForm(); err != nil {
		return ReloadPageError
	}

	courseID, err := strconv.Atoi(r.Form.Get("ID"))
	if (err != nil) || (courseID < 0) || (courseID > len(user.Courses)) {
		return ReloadPageError
	}

	/* TODO(anton2920): this will screw up indicies for courses that are being edited. */
	user.Courses = RemoveAtIndex(user.Courses, courseID)

	w.RedirectString("/", HTTPStatusSeeOther)
	return nil
}
