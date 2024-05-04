package main

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

func DisplayCourseLink(w *HTTPResponse, index int, course *Course) {
	w.AppendString(`<a href="/course/`)
	w.WriteInt(index)
	w.AppendString(`">`)
	w.WriteHTMLString(course.Name)
	if course.Draft {
		w.AppendString(` (draft)`)
	}
	w.AppendString(`</a>`)
}

func CoursePageHandler(w *HTTPResponse, r *HTTPRequest) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}
	user := &DB.Users[session.ID]

	id, err := GetIDFromURL(r.URL, "/course/")
	if err != nil {
		return ClientError(err)
	}
	if (id < 0) || (id > len(user.Courses)) {
		return NotFound("course with this ID does not exist")
	}
	course := user.Courses[id]

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>`)
	w.WriteHTMLString(course.Name)
	if course.Draft {
		w.AppendString(` (draft)`)
	}
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

func CourseLessonPageHandler(w *HTTPResponse, r *HTTPRequest) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}

	if err := r.ParseForm(); err != nil {
		return ClientError(err)
	}

	user := &DB.Users[session.ID]
	courseID, err := GetValidIndex(r.Form.Get("ID"), user.Courses)
	if err != nil {
		return ClientError(err)
	}
	course := user.Courses[courseID]

	li, err := GetValidIndex(r.Form.Get("LessonIndex"), course.Lessons)
	if err != nil {
		return ClientError(err)
	}
	lesson := course.Lessons[li]

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
		var name, stepType string

		step := lesson.Steps[i]
		switch step := step.(type) {
		default:
			panic("invalid step type")
		case *StepTest:
			name = step.Name
			stepType = "Test"
		case *StepProgramming:
			name = step.Name
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

func CourseFillFromRequest(vs URLValues, course *Course) {
	course.Name = vs.Get("Name")
}

func CourseVerify(course *Course) error {
	if !StringLengthInRange(course.Name, MinNameLen, MaxNameLen) {
		return BadRequest("course name length must be between %d and %d characters long", MinNameLen, MaxNameLen)
	}

	if len(course.Lessons) == 0 {
		return BadRequest("create at least one lesson")
	}
	for li := 0; li < len(course.Lessons); li++ {
		lesson := course.Lessons[li]
		if lesson.Draft {
			return BadRequest("lesson %d is a draft", li+1)
		}
	}

	return nil
}

func CourseCreateEditCoursePageHandler(w *HTTPResponse, r *HTTPRequest, course *Course) error {
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

func CourseCreateEditHandleCommand(w *HTTPResponse, r *HTTPRequest, course *Course, currentPage, k, command string) error {
	pindex, spindex, _, _, err := GetIndicies(k[len("Command"):])
	if err != nil {
		return ClientError(err)
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
				return ClientError(nil)
			}
			lesson := course.Lessons[pindex]
			lesson.Draft = true

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

func CourseCreateEditPageHandler(w *HTTPResponse, r *HTTPRequest) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}
	user := &DB.Users[session.ID]

	if err := r.ParseForm(); err != nil {
		return ClientError(err)
	}

	currentPage := r.Form.Get("CurrentPage")
	nextPage := r.Form.Get("NextPage")

	id := r.Form.Get("ID")
	var course *Course
	if id == "" {
		course = new(Course)
		user.Courses = append(user.Courses, course)
		r.Form.SetInt("ID", len(user.Courses)-1)
	} else {
		ci, err := GetValidIndex(id, user.Courses)
		if err != nil {
			return ClientError(err)
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

	/* 'currentPage' is the page to save/check before leaving it. */
	switch currentPage {
	case "Course":
		CourseFillFromRequest(r.Form, course)
	case "Lesson":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), course.Lessons)
		if err != nil {
			return ClientError(err)
		}
		lesson := course.Lessons[li]

		LessonFillFromRequest(r.Form, lesson)
	case "Test":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), course.Lessons)
		if err != nil {
			return ClientError(err)
		}
		lesson := course.Lessons[li]

		si, err := GetValidIndex(r.Form.Get("StepIndex"), lesson.Steps)
		if err != nil {
			return ClientError(err)
		}
		test, ok := lesson.Steps[si].(*StepTest)
		if !ok {
			return ClientError(nil)
		}

		if err := LessonTestFillFromRequest(r.Form, test); err != nil {
			return WritePageEx(w, r, LessonAddTestPageHandler, test, err)
		}
		if err := LessonTestVerify(test); err != nil {
			return WritePageEx(w, r, LessonAddTestPageHandler, test, err)
		}
	case "Programming":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), course.Lessons)
		if err != nil {
			return ClientError(err)
		}
		lesson := course.Lessons[li]

		si, err := GetValidIndex(r.Form.Get("StepIndex"), lesson.Steps)
		if err != nil {
			return ClientError(err)
		}
		task, ok := lesson.Steps[si].(*StepProgramming)
		if !ok {
			return ClientError(nil)
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
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), course.Lessons)
		if err != nil {
			return ClientError(err)
		}
		lesson := course.Lessons[li]

		LessonFillFromRequest(r.Form, lesson)
		if err := LessonVerify(lesson); err != nil {
			return WritePageEx(w, r, LessonAddPageHandler, lesson, err)
		}
		lesson.Draft = false

		return CourseCreateEditCoursePageHandler(w, r, course)
	case "Add lesson":
		lesson := new(Lesson)
		lesson.Draft = true
		course.Lessons = append(course.Lessons, lesson)

		r.Form.SetInt("LessonIndex", len(course.Lessons)-1)
		return LessonAddPageHandler(w, r, lesson)
	case "Continue":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), course.Lessons)
		if err != nil {
			return ClientError(err)
		}
		lesson := course.Lessons[li]

		si, err := GetValidIndex(r.Form.Get("StepIndex"), lesson.Steps)
		if err != nil {
			return ClientError(err)
		}
		switch step := lesson.Steps[si].(type) {
		default:
			panic("invalid step type")
		case *StepTest:
			step.Draft = false
		case *StepProgramming:
			step.Draft = false
		}

		return LessonAddPageHandler(w, r, lesson)
	case "Add test":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), course.Lessons)
		if err != nil {
			return ClientError(err)
		}
		lesson := course.Lessons[li]
		lesson.Draft = true

		test := new(StepTest)
		test.Draft = true
		lesson.Steps = append(lesson.Steps, test)

		r.Form.SetInt("StepIndex", len(lesson.Steps)-1)
		return LessonAddTestPageHandler(w, r, test)
	case "Add programming task":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), course.Lessons)
		if err != nil {
			return ClientError(err)
		}
		lesson := course.Lessons[li]
		lesson.Draft = true

		task := new(StepProgramming)
		task.Draft = true
		lesson.Steps = append(lesson.Steps, task)

		r.Form.SetInt("StepIndex", len(lesson.Steps)-1)
		return LessonAddProgrammingPageHandler(w, r, task)
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
		return ClientError(err)
	}

	courseID, err := GetValidIndex(r.Form.Get("ID"), user.Courses)
	if err != nil {
		return ClientError(err)
	}
	course := user.Courses[courseID]

	if err := CourseVerify(course); err != nil {
		return WritePageEx(w, r, CourseCreateEditCoursePageHandler, course, err)
	}
	course.Draft = false

	w.RedirectID("/course/", courseID, HTTPStatusSeeOther)
	return nil
}

func CourseDeleteHandler(w *HTTPResponse, r *HTTPRequest) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}
	user := &DB.Users[session.ID]

	if err := r.ParseForm(); err != nil {
		return ClientError(err)
	}

	courseID, err := GetValidIndex(r.Form.Get("ID"), user.Courses)
	if err != nil {
		return ClientError(err)
	}

	/* TODO(anton2920): this will screw up indicies for courses that are being edited. */
	user.Courses = RemoveAtIndex(user.Courses, courseID)

	w.Redirect("/", HTTPStatusSeeOther)
	return nil
}
