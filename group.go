package main

import (
	"fmt"
	"time"
)

const (
	MinGroupNameLen = 5
	MaxGroupNameLen = 15
)

func UserInGroup(userID int, group *Group) bool {
	if userID == AdminID {
		return true
	}
	for i := 0; i < len(group.Users); i++ {
		user := group.Users[i]

		if userID == user.ID {
			return true
		}
	}

	return false
}

func DisplayGroupLink(w *HTTPResponse, group *Group) {
	w.AppendString(`<a href="/group/`)
	w.WriteInt(group.ID)
	w.AppendString(`">`)
	w.WriteHTMLString(group.Name)
	w.AppendString(` (ID: `)
	w.WriteInt(group.ID)
	w.AppendString(`)`)
	w.AppendString(`</a>`)
}

func GroupPageHandler(w *HTTPResponse, r *HTTPRequest) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}

	id, err := GetIDFromURL(r.URL, "/group/")
	if err != nil {
		return ClientError(err)
	}
	if (id < 0) || (id >= len(DB.Groups)) {
		return NotFound("group with this ID does not exist")
	}
	group := &DB.Groups[id]
	if !UserInGroup(session.ID, group) {
		return ForbiddenError
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
	w.Write(group.CreatedOn.AppendFormat(make([]byte, 0, 20), "2006/01/02 15:04:05"))
	w.AppendString(`</p>`)

	w.AppendString(`<h2>Students</h2>`)
	w.AppendString(`<ul>`)
	for i := 0; i < len(group.Users); i++ {
		user := group.Users[i]

		w.AppendString(`<li>`)
		DisplayUserLink(w, user)
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

		for i := 0; i < len(group.Users); i++ {
			user := group.Users[i]
			w.AppendString(`<input type="hidden" name="UserID" value="`)
			w.WriteInt(user.ID)
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

func GroupCreatePageHandler(w *HTTPResponse, r *HTTPRequest) error {
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
	w.AppendString(`<head><title>Create group</title></head>`)
	w.AppendString(`<body>`)
	w.AppendString(`<h1>Group</h1>`)
	w.AppendString(`<h2>Create</h2>`)

	DisplayErrorMessage(w, r.Form.Get("Error"))

	w.AppendString(`<form method="POST" action="/api/group/create">`)

	/* TODO(anton2920): insert length constraints parametrically. */
	w.AppendString(`<label>Name: `)
	w.AppendString(`<input type="text" minlength="5" maxlength="15" name="Name" value="`)
	w.WriteHTMLString(r.Form.Get("Name"))
	w.AppendString(`" required>`)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	/* TODO(anton2920): think about bulk add. */
	w.AppendString(`<label>Users:<br>`)
	w.AppendString(`<select name="UserID" multiple>`)
	ids := r.Form.GetMany("UserID")
	for i := AdminID + 1; i < len(DB.Users); i++ {
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

	w.AppendString(`<input type="submit" value="Create">`)

	w.AppendString(`</form>`)
	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func GroupEditPageHandler(w *HTTPResponse, r *HTTPRequest) error {
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
	w.AppendString(`<head><title>Create group</title></head>`)
	w.AppendString(`<body>`)
	w.AppendString(`<h1>Group</h1>`)
	w.AppendString(`<h2>Edit</h2>`)

	DisplayErrorMessage(w, r.Form.Get("Error"))

	w.AppendString(`<form method="POST" action="/api/group/edit">`)

	w.AppendString(`<input type="hidden" name="ID" value="`)
	w.WriteHTMLString(r.Form.Get("ID"))
	w.AppendString(`">`)

	/* TODO(anton2920): insert length constraints parametrically. */
	w.AppendString(`<label>Name: `)
	w.AppendString(`<input type="text" minlength="5" maxlength="15" name="Name" value="`)
	w.WriteHTMLString(r.Form.Get("Name"))
	w.AppendString(`" required>`)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	/* TODO(anton2920): think about bulk add. */
	w.AppendString(`<label>Users:<br>`)
	w.AppendString(`<select name="UserID" multiple>`)
	ids := r.Form.GetMany("UserID")
	for i := AdminID + 1; i < len(DB.Users); i++ {
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

	w.AppendString(`<input type="submit" value="Save">`)

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
	if session.ID != AdminID {
		return ForbiddenError
	}

	if err := r.ParseForm(); err != nil {
		return ClientError(err)
	}

	name := r.Form.Get("Name")
	if !StringLengthInRange(name, MinGroupNameLen, MaxGroupNameLen) {
		return WritePage(w, r, GroupCreatePageHandler, BadRequest(fmt.Sprintf("group name length must be between %d and %d characters long", MinGroupNameLen, MaxGroupNameLen)))
	}

	sids := r.Form.GetMany("UserID")
	users := make([]*User, len(sids))
	for i := 0; i < len(sids); i++ {
		id, err := GetValidIndex(sids[i], DB.Users)
		if (err != nil) || (id == AdminID) {
			return ClientError(err)
		}
		users[i] = &DB.Users[id]
	}
	DB.Groups = append(DB.Groups, Group{ID: len(DB.Groups), Name: name, Users: users, CreatedOn: time.Now()})

	w.Redirect("/", HTTPStatusSeeOther)
	return nil
}

func GroupEditHandler(w *HTTPResponse, r *HTTPRequest) error {
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

	groupID, err := GetValidIndex(r.Form.Get("ID"), DB.Groups)
	if err != nil {
		return ClientError(err)
	}
	group := &DB.Groups[groupID]

	name := r.Form.Get("Name")
	if !StringLengthInRange(name, MinGroupNameLen, MaxGroupNameLen) {
		return WritePage(w, r, GroupEditPageHandler, BadRequest(fmt.Sprintf("group name length must be between %d and %d characters long", MinGroupNameLen, MaxGroupNameLen)))
	}

	sids := r.Form.GetMany("UserID")
	users := make([]*User, len(sids))
	for i := 0; i < len(sids); i++ {
		id, err := GetValidIndex(sids[i], DB.Users)
		if (err != nil) || (id == AdminID) {
			return ClientError(err)
		}
		users[i] = &DB.Users[id]
	}
	group.Name = name
	group.Users = users

	w.RedirectID("/group/", groupID, HTTPStatusSeeOther)
	return nil
}
