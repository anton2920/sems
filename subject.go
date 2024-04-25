package main

import (
	"fmt"
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

func DisplaySubjectLink(w *HTTPResponse, subject *Subject) {
	w.AppendString(`<a href="/subject/`)
	w.WriteInt(subject.ID)
	w.AppendString(`">`)
	w.WriteHTMLString(subject.Name)
	w.AppendString(` with `)
	w.WriteHTMLString(subject.Teacher.LastName)
	w.AppendString(` `)
	w.WriteHTMLString(subject.Teacher.FirstName)
	w.AppendString(` (ID: `)
	w.WriteInt(subject.ID)
	w.AppendString(`)`)
	w.AppendString(`</a>`)
}

func SubjectPageHandler(w *HTTPResponse, r *HTTPRequest) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}

	id, err := GetIDFromURL(r.URL, "/subject/")
	if err != nil {
		return ClientError(err)
	}
	if (id < 0) || (id >= len(DB.Subjects)) {
		return NotFound("subject with this ID does not exist")
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

		if ((session.ID == AdminID) || (session.ID == subject.Teacher.ID)) && (len(lesson.Submissions) > 0) {
			w.AppendString(`<form method="POST" action="/submission">`)

			w.AppendString(`<input type="hidden" name="ID" value="`)
			w.WriteString(r.URL.Path[len("/subject/"):])
			w.AppendString(`">`)

			w.AppendString(`<input type="hidden" name="LessonIndex" value="`)
			w.WriteInt(i)
			w.AppendString(`">`)

			/* TODO(anton2920): do not display if no submissions are ready. */
			w.AppendString(`<label>Submissions: `)
			w.AppendString(`<select name="SubmissionIndex">`)
			for j := 0; j < len(lesson.Submissions); j++ {
				submission := lesson.Submissions[j]
				if submission.Draft {
					continue
				}

				w.AppendString(`<option value="`)
				w.WriteInt(j)
				w.AppendString(`">`)
				w.WriteHTMLString(submission.User.LastName)
				w.AppendString(` `)
				w.WriteHTMLString(submission.User.FirstName)
				w.AppendString(`</option>`)
			}
			w.AppendString(`</select>`)
			w.AppendString(`</label>`)

			w.AppendString("\r\n")
			w.AppendString(`<input type="submit" value="See results">`)
			w.AppendString("\r\n")
			w.AppendString(`<input type="submit" name="NextPage" value="Discard">`)

			w.AppendString(`</form>`)

			w.AppendString(`<br>`)
		}

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
		w.AppendString(`<form method="POST" action="/subject/lesson/edit">`)
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
		return ClientError(err)
	}

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>Create subject</title></head>`)
	w.AppendString(`<body>`)
	w.AppendString(`<h1>Subject</h1>`)
	w.AppendString(`<h2>Create</h2>`)

	DisplayErrorMessage(w, r.Form.Get("Error"))

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
			id, err := GetValidIndex(ids[j], DB.Users)
			if err != nil {
				return ClientError(err)
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
			id, err := GetValidIndex(ids[j], DB.Groups)
			if err != nil {
				return ClientError(err)
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
		return ClientError(err)
	}

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>Edit subject</title></head>`)
	w.AppendString(`<body>`)
	w.AppendString(`<h1>Subject</h1>`)
	w.AppendString(`<h2>Edit</h2>`)

	DisplayErrorMessage(w, r.Form.Get("Error"))

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
			id, err := GetValidIndex(ids[j], DB.Users)
			if err != nil {
				return ClientError(err)
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
			id, err := GetValidIndex(ids[j], DB.Groups)
			if err != nil {
				return ClientError(err)
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

func SubjectCreateHandler(w *HTTPResponse, r *HTTPRequest) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}
	if session.ID != AdminID {
		return ForbiddenError
	}

	if err := r.ParseForm(); err != nil {
		return ClientError(err)
	}

	name := r.Form.Get("Name")
	if !StringLengthInRange(name, MinSubjectNameLen, MaxSubjectNameLen) {
		return WritePage(w, r, SubjectCreatePageHandler, BadRequest(fmt.Sprintf("subject name length must be between %d and %d characters long", MinSubjectNameLen, MaxSubjectNameLen)))
	}

	teacherID, err := GetValidIndex(r.Form.Get("TeacherID"), DB.Users)
	if err != nil {
		return ClientError(err)
	}
	teacher := &DB.Users[teacherID]

	groupID, err := GetValidIndex(r.Form.Get("GroupID"), DB.Groups)
	if err != nil {
		return ClientError(err)
	}
	group := &DB.Groups[groupID]

	DB.Subjects = append(DB.Subjects, Subject{ID: len(DB.Subjects), Name: name, Teacher: teacher, Group: group, CreatedOn: time.Now()})

	w.Redirect("/", HTTPStatusSeeOther)
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
		return ClientError(err)
	}

	subjectID, err := GetValidIndex(r.Form.Get("ID"), DB.Subjects)
	if err != nil {
		return ClientError(err)
	}
	subject := &DB.Subjects[subjectID]

	name := r.Form.Get("Name")
	if !StringLengthInRange(name, MinSubjectNameLen, MaxSubjectNameLen) {
		return WritePage(w, r, SubjectCreatePageHandler, BadRequest(fmt.Sprintf("subject name length must be between %d and %d characters long", MinSubjectNameLen, MaxSubjectNameLen)))
	}

	teacherID, err := GetValidIndex(r.Form.Get("TeacherID"), DB.Users)
	if err != nil {
		return ClientError(err)
	}
	teacher := &DB.Users[teacherID]

	groupID, err := GetValidIndex(r.Form.Get("GroupID"), DB.Groups)
	if err != nil {
		return ClientError(err)
	}
	group := &DB.Groups[groupID]

	subject.Name = name
	subject.Teacher = teacher
	subject.Group = group

	w.RedirectID("/subject/", subjectID, HTTPStatusSeeOther)
	return nil
}
