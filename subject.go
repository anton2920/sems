package main

import (
	"fmt"
	"strconv"
	"time"
)

const (
	MinSubjectNameLen = 1
	MaxSubjectNameLen = 45
)

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

	if (session.ID != AdminID) && (session.ID != subject.Teacher.ID) {
		var allowed bool

		for i := 0; i < len(subject.Group.Users); i++ {
			student := subject.Group.Users[i]
			if session.ID == student.ID {
				allowed = true
				break
			}
		}

		if !allowed {
			return ForbiddenError
		}
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