package main

import (
	"time"

	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/strings"
)

type Group struct {
	ID        int
	Name      string
	Students  []*User
	CreatedOn int64
}

const (
	MinGroupNameLen = 5
	MaxGroupNameLen = 15
)

func UserInGroup(userID int32, group *Group) bool {
	if userID == AdminID {
		return true
	}
	for i := 0; i < len(group.Students); i++ {
		student := group.Students[i]

		if userID == student.ID {
			return true
		}
	}

	return false
}

func DisplayGroupLink(w *http.Response, group *Group) {
	w.AppendString(`<a href="/group/`)
	w.WriteInt(group.ID)
	w.AppendString(`">`)
	w.WriteHTMLString(group.Name)
	w.AppendString(` (ID: `)
	w.WriteInt(group.ID)
	w.AppendString(`)`)
	w.AppendString(`</a>`)
}

func GroupPageHandler(w *http.Response, r *http.Request) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}

	id, err := GetIDFromURL(r.URL, "/group/")
	if err != nil {
		return http.ClientError(err)
	}
	if (id < 0) || (id >= len(DB.Groups)) {
		return http.NotFound("group with this ID does not exist")
	}
	group := &DB.Groups[id]

	if !UserInGroup(session.ID, group) {
		return http.ForbiddenError
	}

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>`)
	w.WriteHTMLString(group.Name)
	w.AppendString(`</title></head>`)
	w.AppendString(`<body>`)

	w.AppendString(`<h1>`)
	w.WriteHTMLString(group.Name)
	w.AppendString(`</h1>`)

	w.AppendString(`<h2>Info</h2>`)

	w.AppendString(`<p>ID: `)
	w.WriteInt(group.ID)
	w.AppendString(`</p>`)

	w.AppendString(`<p>Created on: `)
	DisplayFormattedTime(w, group.CreatedOn)
	w.AppendString(`</p>`)

	w.AppendString(`<h2>Students</h2>`)
	w.AppendString(`<ul>`)
	for i := 0; i < len(group.Students); i++ {
		student := group.Students[i]

		w.AppendString(`<li>`)
		DisplayUserLink(w, student)
		w.AppendString(`</li>`)
	}
	w.AppendString(`</ul>`)

	if session.ID == AdminID {
		w.AppendString(`<form method="POST" action="/group/edit">`)

		w.AppendString(`<input type="hidden" name="ID" value="`)
		w.WriteString(r.URL.Path[len("/group/"):])
		w.AppendString(`">`)

		w.AppendString(`<input type="hidden" name="Name" value="`)
		w.WriteHTMLString(group.Name)
		w.AppendString(`">`)

		for i := 0; i < len(group.Students); i++ {
			student := group.Students[i]
			w.AppendString(`<input type="hidden" name="StudentID" value="`)
			w.WriteInt(int(student.ID))
			w.AppendString(`">`)
		}

		w.AppendString(`<input type="submit" value="Edit">`)

		w.AppendString(`</form>`)
	}

	var displaySubjects bool
	for i := 0; i < len(DB.Subjects); i++ {
		subject := &DB.Subjects[i]

		if group.ID == subject.Group.ID {
			displaySubjects = true
			break
		}
	}
	if displaySubjects {
		w.AppendString(`<h2>Subjects</h2>`)
		w.AppendString(`<ul>`)
		for i := 0; i < len(DB.Subjects); i++ {
			subject := &DB.Subjects[i]

			if group.ID == subject.Group.ID {
				w.AppendString(`<li>`)
				DisplaySubjectLink(w, subject)
				w.AppendString(`</li>`)
			}
		}
		w.AppendString(`</ul>`)
	}

	return nil
}

func DisplayStudentsSelect(w *http.Response, ids []string) {
	w.AppendString(`<select name="StudentID" multiple>`)
	for i := AdminID + 1; i < len(DB.Users); i++ {
		student := &DB.Users[i]

		w.AppendString(`<option value="`)
		w.WriteInt(int(student.ID))
		w.AppendString(`"`)
		for j := 0; j < len(ids); j++ {
			id, err := GetValidIndex(ids[j], DB.Users)
			if err != nil {
				continue
			}
			if int32(id) == student.ID {
				w.AppendString(` selected`)
			}
		}
		w.AppendString(`>`)
		w.WriteHTMLString(student.LastName)
		w.AppendString(` `)
		w.WriteHTMLString(student.FirstName)
		w.AppendString(`</option>`)
	}
	w.AppendString(`</select>`)
}

func GroupCreatePageHandler(w *http.Response, r *http.Request) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}
	if session.ID != AdminID {
		return http.ForbiddenError
	}

	if err := r.ParseForm(); err != nil {
		return http.ClientError(err)
	}

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>Create group</title></head>`)
	w.AppendString(`<body>`)

	w.AppendString(`<h1>Group</h1>`)
	w.AppendString(`<h2>Create</h2>`)

	DisplayErrorMessage(w, r.Form.Get("Error"))

	w.AppendString(`<form method="POST" action="/api/group/create">`)

	w.AppendString(`<label>Name: `)
	DisplayConstraintInput(w, "text", MinNameLen, MaxNameLen, "Name", r.Form.Get("Name"), true)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>Students:<br>`)
	DisplayStudentsSelect(w, r.Form.GetMany("StudentID"))
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<input type="submit" value="Create">`)

	w.AppendString(`</form>`)
	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func GroupEditPageHandler(w *http.Response, r *http.Request) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}
	if session.ID != AdminID {
		return http.ForbiddenError
	}

	if err := r.ParseForm(); err != nil {
		return http.ClientError(err)
	}

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>Create group</title></head>`)
	w.AppendString(`<body>`)
	w.AppendString(`<h1>Group</h1>`)
	w.AppendString(`<h2>Edit</h2>`)

	DisplayErrorMessage(w, r.Form.Get("Error"))

	w.AppendString(`<form method="POST" action="/api/group/edit">`)

	w.AppendString(`<input type="hidden" name="ID" value="`)
	w.WriteHTMLString(r.Form.Get("ID"))
	w.AppendString(`">`)

	w.AppendString(`<label>Name: `)
	DisplayConstraintInput(w, "text", MinNameLen, MaxNameLen, "Name", r.Form.Get("Name"), true)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>Students:<br>`)
	DisplayStudentsSelect(w, r.Form.GetMany("StudentID"))
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<input type="submit" value="Save">`)

	w.AppendString(`</form>`)
	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func GroupCreateHandler(w *http.Response, r *http.Request) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}
	if session.ID != AdminID {
		return http.ForbiddenError
	}

	if err := r.ParseForm(); err != nil {
		return http.ClientError(err)
	}

	name := r.Form.Get("Name")
	if !strings.LengthInRange(name, MinGroupNameLen, MaxGroupNameLen) {
		return WritePage(w, r, GroupCreatePageHandler, http.BadRequest("group name length must be between %d and %d characters long", MinGroupNameLen, MaxGroupNameLen))
	}

	sids := r.Form.GetMany("StudentID")
	students := make([]*User, len(sids))
	for i := 0; i < len(sids); i++ {
		id, err := GetValidIndex(sids[i], DB.Users)
		if (err != nil) || (id == AdminID) {
			return http.ClientError(err)
		}
		students[i] = &DB.Users[id]
	}
	DB.Groups = append(DB.Groups, Group{ID: len(DB.Groups), Name: name, Students: students, CreatedOn: time.Now().Unix()})

	w.Redirect("/", http.StatusSeeOther)
	return nil
}

func GroupEditHandler(w *http.Response, r *http.Request) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}
	if session.ID != AdminID {
		return http.ForbiddenError
	}

	if err := r.ParseForm(); err != nil {
		return http.ClientError(err)
	}

	groupID, err := GetValidIndex(r.Form.Get("ID"), DB.Groups)
	if err != nil {
		return http.ClientError(err)
	}
	group := &DB.Groups[groupID]

	name := r.Form.Get("Name")
	if !strings.LengthInRange(name, MinGroupNameLen, MaxGroupNameLen) {
		return WritePage(w, r, GroupEditPageHandler, http.BadRequest("group name length must be between %d and %d characters long", MinGroupNameLen, MaxGroupNameLen))
	}

	sids := r.Form.GetMany("StudentID")
	students := group.Students[:0]
	for i := 0; i < len(sids); i++ {
		id, err := GetValidIndex(sids[i], DB.Users)
		if (err != nil) || (id == AdminID) {
			return http.ClientError(err)
		}
		students = append(students, &DB.Users[id])
	}
	group.Name = name
	group.Students = students

	w.RedirectID("/group/", groupID, http.StatusSeeOther)
	return nil
}
