package main

import (
	"fmt"
	"strconv"
)

const (
	MinGroupNameLen = 5
	MaxGroupNameLen = 15
)

func GroupCreatePageHandler(w *HTTPResponse, r *HTTPRequest) error {
	if _, err := GetSessionFromRequest(r); err != nil {
		return UnauthorizedError
	}

	if err := r.ParseForm(); err != nil {
		return ReloadPageError
	}

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>Create group</title></head>`)
	w.AppendString(`<body>`)
	w.AppendString(`<h1>Group</h1>`)
	w.AppendString(`<h2>Create</h2>`)

	ErrorDiv(w, r.Form.Get("Error"))

	w.AppendString(`<form method="POST" action="/api/group/create">`)

	/* TODO(anton2920): insert length constraints parametrically. */
	w.AppendString(`<label>Name: `)
	w.AppendString(`<input type="text" minlength="5" maxlength="15" name="Name" value="`)
	w.WriteHTMLString(r.Form.Get("Name"))
	w.AppendString(`" required>`)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>Teacher: `)
	w.AppendString(`<select name="TeacherID">`)
	id := r.Form.Get("TeacherID")
	for _, user := range DB.Users {
		if user.RoleID == UserRoleTeacher {
			w.AppendString(`<option value="`)
			w.WriteString(user.StringID)
			w.AppendString(`"`)
			if id == user.StringID {
				w.AppendString(` selected`)
			}
			w.AppendString(`>`)
			w.WriteHTMLString(user.LastName)
			w.AppendString(` `)
			w.WriteHTMLString(user.FirstName)
			w.AppendString(`</option>`)
		}
	}
	w.AppendString(`</select>`)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>Students:<br>`)
	w.AppendString(`<select name="StudentID" multiple>`)
	ids := r.Form.GetMany("StudentID")
	for _, user := range DB.Users {
		if user.RoleID == UserRoleStudent {
			w.AppendString(`<option value="`)
			w.WriteString(user.StringID)
			w.AppendString(`"`)
			for _, id := range ids {
				if id == user.StringID {
					w.AppendString(` selected`)
				}
			}
			w.AppendString(`>`)
			w.WriteHTMLString(user.LastName)
			w.AppendString(` `)
			w.WriteHTMLString(user.FirstName)
			w.AppendString(`</option>`)
		}
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

func GroupCreateHandler(w *HTTPResponse, r *HTTPRequest) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}

	admin := DB.Users[session.ID].RoleID == UserRoleAdmin
	if !admin {
		return ForbiddenError
	}

	if err := r.ParseForm(); err != nil {
		return ReloadPageError
	}

	name := r.Form.Get("Name")
	if !StringLengthInRange(name, MinGroupNameLen, MaxGroupNameLen) {
		return WritePage(w, r, GroupCreatePageHandler, NewHTTPError(HTTPStatusBadRequest, fmt.Sprintf("group name lenght must be between %d and %d characters long", MinGroupNameLen, MaxGroupNameLen)))
	}

	teacherID, err := strconv.Atoi(r.Form.Get("TeacherID"))
	if err != nil {
		return WritePage(w, r, GroupCreatePageHandler, ReloadPageError)
	}
	teacher, ok := DB.Users[teacherID]
	if (!ok) || (teacher.RoleID != UserRoleTeacher) {
		return WritePage(w, r, GroupCreatePageHandler, NewHTTPError(HTTPStatusNotFound, "user with this ID does not exist"))
	}

	sids := r.Form.GetMany("StudentID")
	users := make([]*User, len(sids))
	for _, sid := range sids {
		id, err := strconv.Atoi(sid)
		if err != nil {
			return WritePage(w, r, GroupCreatePageHandler, ReloadPageError)
		}

		user, ok := DB.Users[id]
		if (!ok) || (user.RoleID != UserRoleStudent) {
			return WritePage(w, r, GroupCreatePageHandler, NewHTTPError(HTTPStatusNotFound, "student with this ID does not exist"))
		}
		users = append(users, user)
	}

	id := len(DB.Groups) + 1
	DB.Groups[id] = &Group{StringID: strconv.Itoa(id), Name: name, Teacher: teacher, Students: users}

	w.RedirectString("/", HTTPStatusSeeOther)
	return nil
}
