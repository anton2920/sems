package main

import (
	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/net/url"
	"github.com/anton2920/gofa/strings"
)

type Course struct {
	ID    int32
	Flags int32

	Name    string
	Lessons []int32

	Data [1024]byte
}

const (
	CourseActive  int32 = 0
	CourseDeleted       = 1
	CourseDraft         = 2
)

const (
	MinNameLen = 1
	MaxNameLen = 45
)

func DisplayCourseLink(w *http.Response, index int, course *Course) {
	w.AppendString(`<a href="/course/`)
	w.WriteInt(index)
	w.AppendString(`">`)
	w.WriteHTMLString(course.Name)
	if course.Flags == CourseDraft {
		w.AppendString(` (draft)`)
	}
	w.AppendString(`</a>`)
}

func CoursePageHandler(w *http.Response, r *http.Request) error {
	var user User

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}
	if err := GetUserByID(DB2, session.ID, &user); err != nil {
		return http.ServerError(err)
	}

	id, err := GetIDFromURL(r.URL, "/course/")
	if err != nil {
		return http.ClientError(err)
	}
	if (id < 0) || (id > len(DB.Courses)) {
		return http.NotFound("course with this ID does not exist")
	}
	if !UserOwnsCourse(&user, int32(id)) {
		return http.ForbiddenError
	}
	course := &DB.Courses[id]

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>`)
	w.WriteHTMLString(course.Name)
	if course.Flags == CourseDraft {
		w.AppendString(` (draft)`)
	}
	w.AppendString(`</title></head>`)
	w.AppendString(`<body>`)

	w.AppendString(`<h1>`)
	w.WriteHTMLString(course.Name)
	if course.Flags == CourseDraft {
		w.AppendString(` (draft)`)
	}
	w.AppendString(`</h1>`)

	w.AppendString(`<h2>Lessons</h2>`)
	for i := 0; i < len(course.Lessons); i++ {
		lesson := &DB.Lessons[course.Lessons[i]]

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

		w.AppendString(`<form method="POST" action="/course/lesson">`)

		w.AppendString(`<input type="hidden" name="ID" value="`)
		w.WriteString(r.URL.Path[len("/course/"):])
		w.AppendString(`">`)

		w.AppendString(`<input type="hidden" name="LessonIndex" value="`)
		w.WriteInt(i)
		w.AppendString(`">`)

		w.AppendString(`<input type="submit" value="Open">`)

		w.AppendString(`</form>`)

		w.AppendString(`</fieldset>`)
		w.AppendString(`<br>`)
	}

	w.AppendString(`<div>`)

	w.AppendString(`<form style="display:inline" method="POST" action="/course/edit">`)
	w.AppendString(`<input type="hidden" name="ID" value="`)
	w.WriteHTMLString(r.URL.Path[len("/course/"):])
	w.AppendString(`">`)
	w.AppendString(`<input type="submit" value="Edit">`)
	w.AppendString(`</form> `)

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

func CourseLessonPageHandler(w *http.Response, r *http.Request) error {
	var user User

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}
	if err := GetUserByID(DB2, session.ID, &user); err != nil {
		return http.ServerError(err)
	}

	if err := r.ParseForm(); err != nil {
		return http.ClientError(err)
	}

	courseID, err := GetValidIndex(r.Form.Get("ID"), len(DB.Courses))
	if err != nil {
		return http.ClientError(err)
	}
	if !UserOwnsCourse(&user, int32(courseID)) {
		return http.ForbiddenError
	}
	course := &DB.Courses[courseID]

	li, err := GetValidIndex(r.Form.Get("LessonIndex"), len(course.Lessons))
	if err != nil {
		return http.ClientError(err)
	}
	lesson := &DB.Lessons[course.Lessons[li]]

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>`)
	w.WriteHTMLString(course.Name)
	w.AppendString(`: `)
	w.WriteHTMLString(lesson.Name)
	w.AppendString(`</title></head>`)
	w.AppendString(`<body>`)

	w.AppendString(`<h1>`)
	w.WriteHTMLString(course.Name)
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
		name := step.Name

		var stepType string
		if step.Type == StepTypeTest {
			stepType = "Test"
		} else {
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

		w.AppendString(`</fieldset>`)
		w.AppendString(`<br>`)
	}
	w.AppendString(`</div>`)

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func CourseFillFromRequest(vs url.Values, course *Course) {
	course.Name = vs.Get("Name")
}

func CourseVerify(course *Course) error {
	if !strings.LengthInRange(course.Name, MinNameLen, MaxNameLen) {
		return http.BadRequest("course name length must be between %d and %d characters long", MinNameLen, MaxNameLen)
	}

	if len(course.Lessons) == 0 {
		return http.BadRequest("create at least one lesson")
	}
	for li := 0; li < len(course.Lessons); li++ {
		lesson := &DB.Lessons[course.Lessons[li]]
		if lesson.Flags == LessonDraft {
			return http.BadRequest("lesson %d is a draft", li+1)
		}
	}

	return nil
}

func CourseCreateEditCoursePageHandler(w *http.Response, r *http.Request, course *Course) error {
	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>Course</title></head>`)
	w.AppendString(`<body>`)

	w.AppendString(`<h1>Course</h1>`)

	DisplayErrorMessage(w, r.Form.Get("Error"))

	w.AppendString(`<form method="POST" action="`)
	w.WriteString(r.URL.Path)
	w.AppendString(`">`)

	w.AppendString(`<input type="hidden" name="CurrentPage" value="Course">`)

	w.AppendString(`<input type="hidden" name="ID" value="`)
	w.WriteHTMLString(r.Form.Get("ID"))
	w.AppendString(`">`)

	w.AppendString(`<label>Name: `)
	DisplayConstraintInput(w, "text", MinNameLen, MaxNameLen, "Name", course.Name, true)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	DisplayLessonsEditableList(w, course.Lessons)

	w.AppendString(`<input type="submit" name="NextPage" value="Add lesson" formnovalidate>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<input type="submit" name="NextPage" value="Save">`)

	w.AppendString(`</form>`)

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func CourseCreateEditHandleCommand(w *http.Response, r *http.Request, course *Course, currentPage, k, command string) error {
	pindex, spindex, _, _, err := GetIndicies(k[len("Command"):])
	if err != nil {
		return http.ClientError(err)
	}

	switch currentPage {
	default:
		return LessonAddHandleCommand(w, r, course.Lessons, currentPage, k, command)
	case "Course":
		switch command {
		case "Delete":
			course.Lessons = RemoveAtIndex(course.Lessons, pindex)
		case "Edit":
			if (pindex < 0) || (pindex >= len(course.Lessons)) {
				return http.ClientError(nil)
			}
			lesson := &DB.Lessons[course.Lessons[pindex]]
			lesson.Flags = LessonDraft

			r.Form.Set("LessonIndex", spindex)
			return LessonAddPageHandler(w, r, lesson)
		case "↑", "^|":
			MoveUp(course.Lessons, pindex)
		case "↓", "|v":
			MoveDown(course.Lessons, pindex)
		}

		return CourseCreateEditCoursePageHandler(w, r, course)
	}
}

func CourseCreateEditPageHandler(w *http.Response, r *http.Request) error {
	var user User

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}
	if err := GetUserByID(DB2, session.ID, &user); err != nil {
		return http.ServerError(err)
	}

	if err := r.ParseForm(); err != nil {
		return http.ClientError(err)
	}

	currentPage := r.Form.Get("CurrentPage")
	nextPage := r.Form.Get("NextPage")

	id := r.Form.Get("ID")
	var course *Course
	if id == "" {
		DB.Courses = append(DB.Courses, Course{ID: int32(len(DB.Courses))})
		course = &DB.Courses[len(DB.Courses)-1]

		user.Courses = append(user.Courses, course.ID)
		if err := SaveUser(DB2, &user); err != nil {
			return http.ServerError(err)
		}

		r.Form.SetInt("ID", int(course.ID))
	} else {
		ci, err := GetValidIndex(id, len(DB.Courses))
		if err != nil {
			return http.ClientError(err)
		}
		if !UserOwnsCourse(&user, int32(ci)) {
			return http.ForbiddenError
		}
		course = &DB.Courses[ci]
	}
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
			return CourseCreateEditHandleCommand(w, r, course, currentPage, k, v)
		}
	}

	/* 'currentPage' is the page to save/check before leaving it. */
	switch currentPage {
	case "Course":
		CourseFillFromRequest(r.Form, course)
	case "Lesson":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), len(course.Lessons))
		if err != nil {
			return http.ClientError(err)
		}
		lesson := &DB.Lessons[course.Lessons[li]]

		LessonFillFromRequest(r.Form, lesson)
	case "Test":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), len(course.Lessons))
		if err != nil {
			return http.ClientError(err)
		}
		lesson := &DB.Lessons[course.Lessons[li]]

		si, err := GetValidIndex(r.Form.Get("StepIndex"), len(lesson.Steps))
		if err != nil {
			return http.ClientError(err)
		}
		test, err := Step2Test(&lesson.Steps[si])
		if err != nil {
			return http.ClientError(err)
		}

		if err := LessonTestFillFromRequest(r.Form, test); err != nil {
			return WritePageEx(w, r, LessonAddTestPageHandler, test, err)
		}
		if err := LessonTestVerify(test); err != nil {
			return WritePageEx(w, r, LessonAddTestPageHandler, test, err)
		}
	case "Programming":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), len(course.Lessons))
		if err != nil {
			return http.ClientError(err)
		}
		lesson := &DB.Lessons[course.Lessons[li]]

		si, err := GetValidIndex(r.Form.Get("StepIndex"), len(lesson.Steps))
		if err != nil {
			return http.ClientError(err)
		}
		task, err := Step2Programming(&lesson.Steps[si])
		if err != nil {
			return http.ClientError(err)
		}

		if err := LessonProgrammingFillFromRequest(r.Form, task); err != nil {
			return WritePageEx(w, r, LessonAddProgrammingPageHandler, task, err)
		}
		if err := LessonProgrammingVerify(task); err != nil {
			return WritePageEx(w, r, LessonAddProgrammingPageHandler, task, err)
		}
	}

	switch nextPage {
	default:
		return CourseCreateEditCoursePageHandler(w, r, course)
	case "Next":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), len(course.Lessons))
		if err != nil {
			return http.ClientError(err)
		}
		lesson := &DB.Lessons[course.Lessons[li]]

		LessonFillFromRequest(r.Form, lesson)
		if err := LessonVerify(lesson); err != nil {
			return WritePageEx(w, r, LessonAddPageHandler, lesson, err)
		}
		lesson.Flags = LessonActive

		return CourseCreateEditCoursePageHandler(w, r, course)
	case "Add lesson":
		DB.Lessons = append(DB.Lessons, Lesson{ID: int32(len(DB.Lessons)), Flags: LessonDraft})
		lesson := &DB.Lessons[len(DB.Lessons)-1]

		course.Lessons = append(course.Lessons, lesson.ID)
		r.Form.SetInt("LessonIndex", int(lesson.ID))

		return LessonAddPageHandler(w, r, lesson)
	case "Continue":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), len(course.Lessons))
		if err != nil {
			return http.ClientError(err)
		}
		lesson := &DB.Lessons[course.Lessons[li]]

		si, err := GetValidIndex(r.Form.Get("StepIndex"), len(lesson.Steps))
		if err != nil {
			return http.ClientError(err)
		}
		step := &lesson.Steps[si]
		step.Draft = false

		return LessonAddPageHandler(w, r, lesson)
	case "Add test":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), len(course.Lessons))
		if err != nil {
			return http.ClientError(err)
		}
		lesson := &DB.Lessons[course.Lessons[li]]
		lesson.Flags = LessonDraft

		lesson.Steps = append(lesson.Steps, Step{StepCommon: StepCommon{Type: StepTypeTest, Draft: true}})
		test, _ := Step2Test(&lesson.Steps[len(lesson.Steps)-1])

		r.Form.SetInt("StepIndex", len(lesson.Steps)-1)
		return LessonAddTestPageHandler(w, r, test)
	case "Add programming task":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), len(course.Lessons))
		if err != nil {
			return http.ClientError(err)
		}
		lesson := &DB.Lessons[course.Lessons[li]]
		lesson.Flags = LessonDraft

		lesson.Steps = append(lesson.Steps, Step{StepCommon: StepCommon{Type: StepTypeProgramming, Draft: true}})
		task, _ := Step2Programming(&lesson.Steps[len(lesson.Steps)-1])

		r.Form.SetInt("StepIndex", len(lesson.Steps)-1)
		return LessonAddProgrammingPageHandler(w, r, task)
	case "Save":
		if err := CourseVerify(course); err != nil {
			return WritePageEx(w, r, CourseCreateEditCoursePageHandler, course, err)
		}
		course.Flags = CourseActive

		w.RedirectID("/course/", int(course.ID), http.StatusSeeOther)
		return nil
	}
}
