package main

import (
	"fmt"
	"strconv"
	"time"
)

type SubjectUserType int

const (
	SubjectUserNone SubjectUserType = iota
	SubjectUserAdmin
	SubjectUserTeacher
	SubjectUserStudent
)

const (
	MinSubjectNameLen = 1
	MaxSubjectNameLen = 45
)

func WhoIsUserInSubject(userID int, subject *Subject) SubjectUserType {
	if userID == AdminID {
		return SubjectUserAdmin
	}

	if userID == subject.Teacher.ID {
		return SubjectUserTeacher
	}

	for i := 0; i < len(subject.Group.Users); i++ {
		student := subject.Group.Users[i]
		if userID == student.ID {
			return SubjectUserStudent
		}
	}

	return SubjectUserNone
}

func SubjectPageHandler(w *HTTPResponse, r *HTTPRequest) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}

	id, err := GetIDFromURL(r.URL, "/subject/")
	if err != nil {
		return err
	}
	if (id < 0) || (id >= len(DB.Subjects)) {
		return NotFoundError
	}
	subject := &DB.Subjects[id]

	if WhoIsUserInSubject(session.ID, subject) == SubjectUserNone {
		return ForbiddenError
	}

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>`)
	w.WriteHTMLString(subject.Name)
	w.AppendString(` with `)
	w.WriteHTMLString(subject.Teacher.LastName)
	w.AppendString(` `)
	w.WriteHTMLString(subject.Teacher.FirstName)
	w.AppendString(`</title></head>`)
	w.AppendString(`<body>`)
	w.AppendString(`<h1>`)
	w.WriteHTMLString(subject.Name)
	w.AppendString(` with `)
	w.WriteHTMLString(subject.Teacher.LastName)
	w.AppendString(` `)
	w.WriteHTMLString(subject.Teacher.FirstName)
	w.AppendString(`</h1>`)

	w.AppendString(`<p>ID: `)
	w.WriteInt(subject.ID)
	w.AppendString(`</p>`)

	w.AppendString(`<p>Teacher: `)
	w.AppendString(`<a href="/user/`)
	w.WriteInt(subject.Teacher.ID)
	w.AppendString(`">`)
	w.WriteHTMLString(subject.Teacher.LastName)
	w.AppendString(` `)
	w.WriteHTMLString(subject.Teacher.FirstName)
	w.AppendString(` (ID: `)
	w.WriteInt(subject.Teacher.ID)
	w.AppendString(`)`)
	w.AppendString(`</a>`)
	w.AppendString(`</p>`)

	w.AppendString(`<p>Group: `)
	w.AppendString(`<a href="/group/`)
	w.WriteInt(subject.Group.ID)
	w.AppendString(`">`)
	w.WriteHTMLString(subject.Group.Name)
	w.AppendString(` (ID: `)
	w.WriteInt(subject.Group.ID)
	w.AppendString(`)`)
	w.AppendString(`</a>`)
	w.AppendString(`</p>`)

	w.AppendString(`<p>Created on: `)
	w.Write(subject.CreatedOn.AppendFormat(make([]byte, 0, 20), "2006/01/02 15:04:05"))
	w.AppendString(`</p>`)

	if session.ID == AdminID {
		w.AppendString(`<form method="POST" action="/subject/edit">`)

		w.AppendString(`<input type="hidden" name="ID" value="`)
		w.WriteString(r.URL.Path[len("/subject/"):])
		w.AppendString(`">`)

		w.AppendString(`<input type="hidden" name="Name" value="`)
		w.WriteHTMLString(subject.Name)
		w.AppendString(`">`)

		w.AppendString(`<input type="hidden" name="TeacherID" value="`)
		w.WriteInt(subject.Teacher.ID)
		w.AppendString(`">`)

		w.AppendString(`<input type="hidden" name="GroupID" value="`)
		w.WriteInt(subject.Group.ID)
		w.AppendString(`">`)

		w.AppendString(`<input type="submit" value="Edit">`)

		w.AppendString(`</form>`)
	}

	w.AppendString(`<h2>Lessons</h2>`)

	for i := 0; i < len(subject.Lessons); i++ {
		lesson := subject.Lessons[i]

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

		/*
			w.AppendString(`<a href="/subject/`)
			w.WriteString(r.URL.Path[len("/subject/"):])
			w.AppendString(`/lesson/`)
			w.WriteInt(i)
			w.AppendString(`">Open</a>`)
		*/

		w.AppendString(`<form method="POST" action="/subject/lesson">`)

		w.AppendString(`<input type="hidden" name="ID" value="`)
		w.WriteString(r.URL.Path[len("/subject/"):])
		w.AppendString(`">`)

		w.AppendString(`<input type="hidden" name="LessonIndex" value="`)
		w.WriteInt(i)
		w.AppendString(`">`)

		w.AppendString(`<input type="submit" value="Open">`)

		w.AppendString(`</form>`)

		w.AppendString(`</fieldset>`)
		w.AppendString(`<br>`)
	}

	if (session.ID == AdminID) || (session.ID == subject.Teacher.ID) {
		w.AppendString(`<form method="POST" action="/subject/lessons/edit">`)
		w.AppendString(`<input type="hidden" name="ID" value="`)
		w.WriteString(r.URL.Path[len("/subject/"):])
		w.AppendString(`">`)

		if len(subject.Lessons) == 0 {
			/* TODO(anton2920): calculate length on non-draft slice. */
			if len(subject.Teacher.Courses) != 0 {
				w.AppendString(`<label>Courses: `)
				w.AppendString(`<select name="CourseID">`)
				for i := 0; i < len(subject.Teacher.Courses); i++ {
					course := subject.Teacher.Courses[i]

					if !course.Draft {
						w.AppendString(`<option value="`)
						w.WriteInt(i)
						w.AppendString(`">`)
						w.WriteHTMLString(course.Name)
						w.AppendString(`</option>`)
					}
				}
				w.AppendString(`</select>`)

				w.AppendString("\r\n")
				w.AppendString(`<input type="submit" name="Action" value="create from">`)
				w.AppendString("\r\n")
				w.AppendString(`, `)
				w.AppendString(`<input type="submit" name="Action" value="give as is">`)
				w.AppendString("\r\n")
				w.AppendString(`or `)
			}

			w.AppendString(`<input type="submit" value="create new from scratch">`)

			w.AppendString(`</label>`)
		} else {
			w.AppendString(`<input type="submit" value="Edit">`)
		}

		w.AppendString(`</form>`)

	}

	return nil
}

func SubjectCreatePageHandler(w *HTTPResponse, r *HTTPRequest) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}
	if session.ID != AdminID {
		return ForbiddenError
	}

	if err := r.ParseForm(); err != nil {
		return ReloadPageError
	}

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>Create subject</title></head>`)
	w.AppendString(`<body>`)
	w.AppendString(`<h1>Subject</h1>`)
	w.AppendString(`<h2>Create</h2>`)

	ErrorDiv(w, r.Form.Get("Error"))

	w.AppendString(`<form method="POST" action="/api/subject/create">`)

	/* TODO(anton2920): insert length constraints parametrically. */
	w.AppendString(`<label>Name: `)
	w.AppendString(`<input type="text" minlength="1" maxlength="45" name="Name" value="`)
	w.WriteHTMLString(r.Form.Get("Name"))
	w.AppendString(`" required>`)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>Teacher: `)
	w.AppendString(`<select name="TeacherID">`)
	ids := r.Form.GetMany("TeacherID")
	for i := 0; i < len(DB.Users); i++ {
		user := &DB.Users[i]

		w.AppendString(`<option value="`)
		w.WriteInt(user.ID)
		w.AppendString(`"`)
		for j := 0; j < len(ids); j++ {
			id, err := strconv.Atoi(ids[j])
			if err != nil {
				return ReloadPageError
			}
			if id == user.ID {
				w.AppendString(` selected`)
			}
		}
		w.AppendString(`>`)
		w.WriteHTMLString(user.LastName)
		w.AppendString(` `)
		w.WriteHTMLString(user.FirstName)
		w.AppendString(`</option>`)
	}
	w.AppendString(`</select>`)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>Group: `)
	w.AppendString(`<select name="GroupID">`)
	ids = r.Form.GetMany("GroupID")
	for i := 0; i < len(DB.Groups); i++ {
		group := &DB.Groups[i]

		w.AppendString(`<option value="`)
		w.WriteInt(group.ID)
		w.AppendString(`"`)
		for j := 0; j < len(ids); j++ {
			id, err := strconv.Atoi(ids[j])
			if err != nil {
				return ReloadPageError
			}
			if id == group.ID {
				w.AppendString(` selected`)
			}
		}
		w.AppendString(`>`)
		w.WriteHTMLString(group.Name)
		w.AppendString(`</option>`)
	}
	w.AppendString(`</select>`)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<input type="submit" value="Create">`)

	w.AppendString(`</form>`)
	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func SubjectEditPageHandler(w *HTTPResponse, r *HTTPRequest) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}
	if session.ID != AdminID {
		return ForbiddenError
	}

	if err := r.ParseForm(); err != nil {
		return ReloadPageError
	}

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>Edit subject</title></head>`)
	w.AppendString(`<body>`)
	w.AppendString(`<h1>Subject</h1>`)
	w.AppendString(`<h2>Edit</h2>`)

	ErrorDiv(w, r.Form.Get("Error"))

	w.AppendString(`<form method="POST" action="/api/subject/edit">`)

	w.AppendString(`<input type="hidden" name="ID" value="`)
	w.WriteHTMLString(r.Form.Get("ID"))
	w.AppendString(`">`)

	/* TODO(anton2920): insert length constraints parametrically. */
	w.AppendString(`<label>Name: `)
	w.AppendString(`<input type="text" minlength="1" maxlength="45" name="Name" value="`)
	w.WriteHTMLString(r.Form.Get("Name"))
	w.AppendString(`" required>`)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>Teacher: `)
	w.AppendString(`<select name="TeacherID">`)
	ids := r.Form.GetMany("TeacherID")
	for i := 0; i < len(DB.Users); i++ {
		user := &DB.Users[i]

		w.AppendString(`<option value="`)
		w.WriteInt(user.ID)
		w.AppendString(`"`)
		for j := 0; j < len(ids); j++ {
			id, err := strconv.Atoi(ids[j])
			if err != nil {
				return ReloadPageError
			}
			if id == user.ID {
				w.AppendString(` selected`)
			}
		}
		w.AppendString(`>`)
		w.WriteHTMLString(user.LastName)
		w.AppendString(` `)
		w.WriteHTMLString(user.FirstName)
		w.AppendString(`</option>`)
	}
	w.AppendString(`</select>`)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>Group: `)
	w.AppendString(`<select name="GroupID">`)
	ids = r.Form.GetMany("GroupID")
	for i := 0; i < len(DB.Groups); i++ {
		group := &DB.Groups[i]

		w.AppendString(`<option value="`)
		w.WriteInt(group.ID)
		w.AppendString(`"`)
		for j := 0; j < len(ids); j++ {
			id, err := strconv.Atoi(ids[j])
			if err != nil {
				return ReloadPageError
			}
			if id == group.ID {
				w.AppendString(` selected`)
			}
		}
		w.AppendString(`>`)
		w.WriteHTMLString(group.Name)
		w.AppendString(`</option>`)
	}
	w.AppendString(`</select>`)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<input type="submit" value="Save">`)

	w.AppendString(`</form>`)
	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func SubjectLessonsEditMainPageHandler(w *HTTPResponse, r *HTTPRequest, subject *Subject) error {
	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>Edit subject lessons</title></head>`)
	w.AppendString(`<body>`)
	w.AppendString(`<h1>Subject</h1>`)
	w.AppendString(`<h2>Lessons</h2>`)

	ErrorDiv(w, r.Form.Get("Error"))

	w.AppendString(`<form method="POST" action="`)
	w.WriteString(r.URL.Path)
	w.AppendString(`">`)

	w.AppendString(`<input type="hidden" name="CurrentPage" value="Main">`)

	w.AppendString(`<input type="hidden" name="ID" value="`)
	w.WriteHTMLString(r.Form.Get("ID"))
	w.AppendString(`">`)

	for i := 0; i < len(subject.Lessons); i++ {
		lesson := subject.Lessons[i]

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
		if len(subject.Lessons) > 1 {
			if i > 0 {
				w.AppendString("\r\n")
				w.AppendString(`<input type="submit" name="Command`)
				w.WriteInt(i)
				w.AppendString(`" value="↑" formnovalidate>`)
			}
			if i < len(subject.Lessons)-1 {
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

func SubjectLessonsEditHandleCommand(w *HTTPResponse, r *HTTPRequest, subject *Subject, currentPage, k, command string) error {
	pindex, spindex, _, _, err := GetIndicies(k[len("Command"):])
	if err != nil {
		return ReloadPageError
	}

	switch currentPage {
	default:
		return LessonAddHandleCommand(w, r, subject.Lessons, currentPage, k, command)
	case "Main":
		switch command {
		case "Delete":
			subject.Lessons = RemoveAtIndex(subject.Lessons, pindex)
		case "Edit":
			if (pindex < 0) || (pindex >= len(subject.Lessons)) {
				return ReloadPageError
			}
			lesson := subject.Lessons[pindex]
			lesson.Draft = true

			r.Form.Set("LessonIndex", spindex)
			return LessonAddPageHandler(w, r, lesson)
		case "↑", "^|":
			MoveUp(subject.Lessons, pindex)
		case "↓", "|v":
			MoveDown(subject.Lessons, pindex)
		}

		return SubjectLessonsEditMainPageHandler(w, r, subject)
	}
}

func SubjectLessonsEditPageHandler(w *HTTPResponse, r *HTTPRequest) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}
	user := &DB.Users[session.ID]

	if err := r.ParseForm(); err != nil {
		return ReloadPageError
	}

	currentPage := r.Form.Get("CurrentPage")
	nextPage := r.Form.Get("NextPage")

	id := r.Form.Get("ID")
	subjectID, err := strconv.Atoi(id)
	if (err != nil) || (subjectID < 0) || (subjectID >= len(DB.Subjects)) {
		return ReloadPageError
	}
	subject := &DB.Subjects[subjectID]
	if (session.ID != AdminID) && (session.ID != subject.Teacher.ID) {
		return ForbiddenError
	}

	switch r.Form.Get("Action") {
	case "create from":
		courseID, err := strconv.Atoi(r.Form.Get("CourseID"))
		if (err != nil) || (courseID < 0) || (courseID >= len(user.Courses)) {
			return ReloadPageError
		}

		LessonsDeepCopy(&subject.Lessons, user.Courses[courseID].Lessons)
	case "give as is":
		courseID, err := strconv.Atoi(r.Form.Get("CourseID"))
		if (err != nil) || (courseID < 0) || (courseID >= len(user.Courses)) {
			return ReloadPageError
		}

		LessonsDeepCopy(&subject.Lessons, user.Courses[courseID].Lessons)

		w.Redirect(fmt.Appendf(make([]byte, 0, 20), "/subject/%s", id), HTTPStatusSeeOther)
		return nil
	}

	for i := 0; i < len(r.Form); i++ {
		k := r.Form[i].Key
		v := r.Form[i].Values[0]

		/* 'command' is button, which modifies content of a current page. */
		if StringStartsWith(k, "Command") {
			/* NOTE(anton2920): after command is executed, function must return. */
			return SubjectLessonsEditHandleCommand(w, r, subject, currentPage, k, v)
		}
	}

	/* 'currentPage' is the page to check before leaving it. */
	switch currentPage {
	case "Lesson":
		li, err := strconv.Atoi(r.Form.Get("LessonIndex"))
		if (err != nil) || (li < 0) || (li >= len(subject.Lessons)) {
			return ReloadPageError
		}
		lesson := subject.Lessons[li]

		if err := LessonAddVerifyRequest(r.Form, lesson); err != nil {
			return WritePageEx(w, r, LessonAddPageHandler, lesson, err)
		}
	case "Test":
		li, err := strconv.Atoi(r.Form.Get("LessonIndex"))
		if (err != nil) || (li < 0) || (li >= len(subject.Lessons)) {
			return ReloadPageError
		}
		lesson := subject.Lessons[li]

		si, err := strconv.Atoi(r.Form.Get("StepIndex"))
		if (err != nil) || (si < 0) || (si >= len(lesson.Steps)) {
			return ReloadPageError
		}
		test, ok := lesson.Steps[si].(*StepTest)
		if !ok {
			return ReloadPageError
		}

		if err := LessonTestAddVerifyRequest(r.Form, test, true); err != nil {
			return WritePageEx(w, r, LessonTestAddPageHandler, test, err)
		}
	case "Programming":
		li, err := strconv.Atoi(r.Form.Get("LessonIndex"))
		if (err != nil) || (li < 0) || (li >= len(subject.Lessons)) {
			return ReloadPageError
		}
		lesson := subject.Lessons[li]

		si, err := strconv.Atoi(r.Form.Get("StepIndex"))
		if (err != nil) || (si < 0) || (si >= len(lesson.Steps)) {
			return ReloadPageError
		}
		task, ok := lesson.Steps[si].(*StepProgramming)
		if !ok {
			return ReloadPageError
		}

		if err := LessonProgrammingAddVerifyRequest(r.Form, task, true); err != nil {
			return WritePageEx(w, r, LessonProgrammingAddPageHandler, task, err)
		}
	}

	switch nextPage {
	default:
		return SubjectLessonsEditMainPageHandler(w, r, subject)
	case "Next":
		li, err := strconv.Atoi(r.Form.Get("LessonIndex"))
		if (err != nil) || (li < 0) || (li >= len(subject.Lessons)) {
			return ReloadPageError
		}
		lesson := subject.Lessons[li]

		for si := 0; si < len(lesson.Steps); si++ {
			switch step := lesson.Steps[si].(type) {
			case *StepTest:
				if step.Draft {
					return WritePageEx(w, r, LessonAddPageHandler, lesson, NewHTTPError(HTTPStatusBadRequest, fmt.Sprintf("test %d is a draft", si+1)))
				}
			case *StepProgramming:
				if step.Draft {
					return WritePageEx(w, r, LessonAddPageHandler, lesson, NewHTTPError(HTTPStatusBadRequest, fmt.Sprintf("programming task %d is a draft", si+1)))
				}
			}
		}
		lesson.Draft = false

		return SubjectLessonsEditMainPageHandler(w, r, subject)
	case "Add lesson":
		lesson := new(Lesson)
		lesson.Draft = true
		subject.Lessons = append(subject.Lessons, lesson)

		r.Form.Set("LessonIndex", strconv.Itoa(len(subject.Lessons)-1))
		return LessonAddPageHandler(w, r, lesson)
	case "Continue":
		li, err := strconv.Atoi(r.Form.Get("LessonIndex"))
		if (err != nil) || (li < 0) || (li >= len(subject.Lessons)) {
			return ReloadPageError
		}
		lesson := subject.Lessons[li]

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

		return LessonAddPageHandler(w, r, lesson)
	case "Add test":
		li, err := strconv.Atoi(r.Form.Get("LessonIndex"))
		if (err != nil) || (li < 0) || (li >= len(subject.Lessons)) {
			return ReloadPageError
		}
		lesson := subject.Lessons[li]
		lesson.Draft = true

		test := new(StepTest)
		test.Draft = true
		lesson.Steps = append(lesson.Steps, test)

		r.Form.Set("StepIndex", strconv.Itoa(len(lesson.Steps)-1))
		return LessonTestAddPageHandler(w, r, test)
	case "Add programming task":
		li, err := strconv.Atoi(r.Form.Get("LessonIndex"))
		if (err != nil) || (li < 0) || (li >= len(subject.Lessons)) {
			return ReloadPageError
		}
		lesson := subject.Lessons[li]
		lesson.Draft = true

		task := new(StepProgramming)
		task.Draft = true
		lesson.Steps = append(lesson.Steps, task)

		r.Form.Set("StepIndex", strconv.Itoa(len(lesson.Steps)-1))
		return LessonProgrammingAddPageHandler(w, r, task)
	case "Save":
		return SubjectLessonsEditHandler(w, r)
	}
}

func SubjectCreateHandler(w *HTTPResponse, r *HTTPRequest) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}
	if session.ID != AdminID {
		return ForbiddenError
	}

	if err := r.ParseForm(); err != nil {
		return ReloadPageError
	}

	name := r.Form.Get("Name")
	if !StringLengthInRange(name, MinSubjectNameLen, MaxSubjectNameLen) {
		return WritePage(w, r, SubjectCreatePageHandler, NewHTTPError(HTTPStatusBadRequest, fmt.Sprintf("subject name length must be between %d and %d characters long", MinSubjectNameLen, MaxSubjectNameLen)))
	}

	teacherID, err := strconv.Atoi(r.Form.Get("TeacherID"))
	if (err != nil) || (teacherID < 0) || (teacherID >= len(DB.Users)) {
		return ReloadPageError
	}
	teacher := &DB.Users[teacherID]

	groupID, err := strconv.Atoi(r.Form.Get("GroupID"))
	if (err != nil) || (groupID < 0) || (groupID >= len(DB.Groups)) {
		return ReloadPageError
	}
	group := &DB.Groups[groupID]

	DB.Subjects = append(DB.Subjects, Subject{ID: len(DB.Subjects), Name: name, Teacher: teacher, Group: group, CreatedOn: time.Now()})

	w.RedirectString("/", HTTPStatusSeeOther)
	return nil
}

func SubjectEditHandler(w *HTTPResponse, r *HTTPRequest) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}
	if session.ID != AdminID {
		return ForbiddenError
	}

	if err := r.ParseForm(); err != nil {
		return ReloadPageError
	}

	id := r.Form.Get("ID")
	subjectID, err := strconv.Atoi(id)
	if (err != nil) || (subjectID < 0) || (subjectID >= len(DB.Subjects)) {
		return ReloadPageError
	}
	subject := &DB.Subjects[subjectID]

	name := r.Form.Get("Name")
	if !StringLengthInRange(name, MinSubjectNameLen, MaxSubjectNameLen) {
		return WritePage(w, r, SubjectCreatePageHandler, NewHTTPError(HTTPStatusBadRequest, fmt.Sprintf("subject name length must be between %d and %d characters long", MinSubjectNameLen, MaxSubjectNameLen)))
	}

	teacherID, err := strconv.Atoi(r.Form.Get("TeacherID"))
	if (err != nil) || (teacherID < 0) || (teacherID >= len(DB.Users)) {
		return ReloadPageError
	}
	teacher := &DB.Users[teacherID]

	groupID, err := strconv.Atoi(r.Form.Get("GroupID"))
	if (err != nil) || (groupID < 0) || (groupID >= len(DB.Groups)) {
		return ReloadPageError
	}
	group := &DB.Groups[groupID]

	subject.Name = name
	subject.Teacher = teacher
	subject.Group = group

	w.Redirect(fmt.Appendf(make([]byte, 0, 20), "/subject/%s", id), HTTPStatusSeeOther)
	return nil
}

func SubjectLessonsEditHandler(w *HTTPResponse, r *HTTPRequest) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}

	if err := r.ParseForm(); err != nil {
		return ReloadPageError
	}

	id := r.Form.Get("ID")
	subjectID, err := strconv.Atoi(id)
	if (err != nil) || (subjectID < 0) || (subjectID >= len(DB.Subjects)) {
		return ReloadPageError
	}
	subject := &DB.Subjects[subjectID]
	if (session.ID != AdminID) && (session.ID != subject.Teacher.ID) {
		return WritePageEx(w, r, SubjectLessonsEditMainPageHandler, subject, ForbiddenError)
	}

	if len(subject.Lessons) == 0 {
		return WritePageEx(w, r, SubjectLessonsEditMainPageHandler, subject, NewHTTPError(HTTPStatusBadRequest, "create at least one lesson"))
	}
	for li := 0; li < len(subject.Lessons); li++ {
		lesson := subject.Lessons[li]
		if lesson.Draft {
			return WritePageEx(w, r, SubjectLessonsEditMainPageHandler, subject, NewHTTPError(HTTPStatusBadRequest, fmt.Sprintf("lesson %d is a draft", li+1)))
		}
	}

	w.Redirect(fmt.Appendf(make([]byte, 0, 20), "/subject/%s", id), HTTPStatusSeeOther)
	return nil
}
