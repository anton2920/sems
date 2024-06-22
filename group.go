package main

import (
	"fmt"
	"time"
	"unsafe"

	"github.com/anton2920/gofa/database"
	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/strings"
	"github.com/anton2920/gofa/syscall"
)

type Group struct {
	ID    database.ID
	Flags int32

	Name      string
	Students  []database.ID
	CreatedOn int64

	Data [1024]byte
}

const (
	GroupActive int32 = iota
	GroupDeleted
)

const (
	MinGroupNameLen = 5
	MaxGroupNameLen = 15
)

func UserInGroup(userID database.ID, group *Group) bool {
	if userID == AdminID {
		return true
	}
	for i := 0; i < len(group.Students); i++ {
		if userID == group.Students[i] {
			return true
		}
	}
	return false
}

func CreateGroup(group *Group) error {
	var err error

	group.ID, err = database.IncrementNextID(GroupsDB)
	if err != nil {
		return fmt.Errorf("failed to increment group ID: %w", err)
	}

	return SaveGroup(group)
}

func DBGroup2Group(group *Group) {
	data := &group.Data[0]

	group.Name = database.Offset2String(group.Name, data)
	group.Students = database.Offset2Slice(group.Students, data)
}

func GetGroupByID(id database.ID, group *Group) error {
	if err := database.Read(GroupsDB, id, group); err != nil {
		return err
	}

	DBGroup2Group(group)
	return nil
}

func GetGroups(pos *int64, groups []Group) (int, error) {
	n, err := database.ReadMany(GroupsDB, pos, groups)
	if err != nil {
		return 0, err
	}

	for i := 0; i < n; i++ {
		DBGroup2Group(&groups[i])
	}
	return n, nil
}

func DeleteGroupByID(id database.ID) error {
	flags := GroupDeleted
	var group Group

	offset := int64(int(id)*int(unsafe.Sizeof(group))) + database.DataOffset + int64(unsafe.Offsetof(group.Flags))
	_, err := syscall.Pwrite(GroupsDB.FD, unsafe.Slice((*byte)(unsafe.Pointer(&flags)), unsafe.Sizeof(flags)), offset)
	if err != nil {
		return fmt.Errorf("failed to delete group from DB: %w", err)
	}

	return nil
}

func SaveGroup(group *Group) error {
	var groupDB Group
	var n int

	groupDB.ID = group.ID
	groupDB.Flags = group.Flags

	/* TODO(anton2920): save up to a sizeof(group.Data). */
	data := unsafe.Slice(&groupDB.Data[0], len(groupDB.Data))
	n += database.String2DBString(&groupDB.Name, group.Name, data, n)
	n += database.Slice2DBSlice(&groupDB.Students, group.Students, data, n)

	groupDB.CreatedOn = group.CreatedOn

	return database.Write(GroupsDB, groupDB.ID, &groupDB)
}

func DisplayGroupStudents(w *http.Response, l Language, group *Group) {
	w.AppendString(`<h3>`)
	w.AppendString(Ls(GL, "Students"))
	w.AppendString(`</h3>`)
	w.AppendString(`<ul>`)
	for i := 0; i < len(group.Students); i++ {
		var user User
		if err := GetUserByID(group.Students[i], &user); err != nil {
			/* TODO(anton2920): report error. */
		}

		w.AppendString(`<li>`)
		DisplayUserLink(w, GL, &user)
		w.AppendString(`</li>`)
	}
	w.AppendString(`</ul>`)
	w.AppendString(`<br>`)
}

func DisplayGroupSubjects(w *http.Response, l Language, group *Group) {
	subjects := make([]Subject, 32)
	var displayed bool
	var pos int64

	for {
		n, err := GetSubjects(&pos, subjects)
		if err != nil {
			/* TODO(anton2920): report error. */
		}
		if n == 0 {
			break
		}
		for i := 0; i < n; i++ {
			subject := &subjects[i]
			if (subject.Flags == SubjectDeleted) || (group.ID != subject.GroupID) {
				continue
			}

			if !displayed {
				w.AppendString(`<h3>`)
				w.AppendString(Ls(l, "Subjects"))
				w.AppendString(`</h3>`)
				w.AppendString(`<ul>`)
				displayed = true
			}

			w.AppendString(`<li>`)
			DisplaySubjectLink(w, l, subject)
			w.AppendString(`</li>`)
		}
	}
	if displayed {
		w.AppendString(`</ul>`)
	}
}

func DisplayGroupTitle(w *http.Response, l Language, group *Group) {
	w.WriteHTMLString(group.Name)
	w.AppendString(` (ID: `)
	w.WriteID(group.ID)
	w.AppendString(`)`)
	DisplayDeleted(w, l, group.Flags == GroupDeleted)
}

func DisplayGroupLink(w *http.Response, l Language, group *Group) {
	w.AppendString(`<a href="/group/`)
	w.WriteID(group.ID)
	w.AppendString(`">`)
	DisplayGroupTitle(w, l, group)
	w.AppendString(`</a>`)
}

func GroupPageHandler(w *http.Response, r *http.Request) error {
	var group Group

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}

	id, err := GetIDFromURL(GL, r.URL, "/group/")
	if err != nil {
		return http.ClientError(err)
	}
	if err := GetGroupByID(id, &group); err != nil {
		if err == database.NotFound {
			return http.NotFound("group with this ID does not exist")
		}
		return http.ServerError(err)
	}
	if !UserInGroup(session.ID, &group) {
		return http.ForbiddenError
	}

	DisplayHTMLStart(w)

	DisplayHeadStart(w)
	{
		w.AppendString(`<title>`)
		DisplayGroupTitle(w, GL, &group)
		w.AppendString(`</title>`)
	}
	DisplayHeadEnd(w)

	DisplayBodyStart(w)
	{
		DisplayHeader(w, GL)
		DisplaySidebar(w, GL, session.ID)

		DisplayPageStart(w)
		{
			w.AppendString(`<h2>`)
			DisplayGroupTitle(w, GL, &group)
			w.AppendString(`</h2>`)

			w.AppendString(`<h3>`)
			w.AppendString(Ls(GL, "Info"))
			w.AppendString(`</h3>`)

			w.AppendString(`<p>`)
			w.AppendString(Ls(GL, "Created on"))
			w.AppendString(`: `)
			DisplayFormattedTime(w, group.CreatedOn)
			w.AppendString(`</p>`)

			if session.ID == AdminID {
				w.AppendString(`<div>`)
				w.AppendString(`<form style="display:inline" method="POST" action="/group/edit">`)
				DisplayHiddenID(w, "ID", group.ID)
				DisplayHiddenString(w, "Name", group.Name)
				for i := 0; i < len(group.Students); i++ {
					DisplayHiddenID(w, "StudentID", group.Students[i])
				}
				DisplayButton(w, GL, "", "Edit")
				w.AppendString(`</form>`)

				w.AppendString(` <form style="display:inline" method="POST" action="/api/group/delete">`)
				DisplayHiddenID(w, "ID", group.ID)
				DisplayButton(w, GL, "", "Delete")
				w.AppendString(`</form>`)
				w.AppendString(`</div>`)
				w.AppendString(`<br>`)
			}

			DisplayGroupStudents(w, GL, &group)
			DisplayGroupSubjects(w, GL, &group)
		}
		DisplayPageEnd(w)
	}
	DisplayBodyEnd(w)

	DisplayHTMLEnd(w)
	return nil
}

func DisplayStudentsSelect(w *http.Response, ids []string) {
	users := make([]User, 32)
	var pos int64

	w.AppendString(`<select name="StudentID" multiple>`)
	for {
		n, err := GetUsers(&pos, users)
		if err != nil {
			/* TODO(anton2920): report error. */
		}
		if n == 0 {
			break
		}
		for i := 0; i < n; i++ {
			user := &users[i]
			if (user.Flags == UserDeleted) || (user.ID == AdminID) {
				continue
			}

			w.AppendString(`<option value="`)
			w.WriteID(user.ID)
			w.AppendString(`"`)
			for j := 0; j < len(ids); j++ {
				id, err := GetValidID(ids[j], database.MaxValidID)
				if err != nil {
					continue
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
	w.AppendString(`<head><title>`)
	w.AppendString(Ls(GL, "Create group"))
	w.AppendString(`</title></head>`)
	w.AppendString(`<body>`)

	w.AppendString(`<h1>`)
	w.AppendString(Ls(GL, "Group"))
	w.AppendString(`</h1>`)
	w.AppendString(`<h2>`)
	w.AppendString(Ls(GL, "Create"))
	w.AppendString(`</h2>`)

	DisplayErrorMessage(w, GL, r.Form.Get("Error"))

	w.AppendString(`<form method="POST" action="/api/group/create">`)

	w.AppendString(`<label>`)
	w.AppendString(Ls(GL, "Name"))
	w.AppendString(`: `)
	DisplayConstraintInput(w, "text", MinNameLen, MaxNameLen, "Name", r.Form.Get("Name"), true)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>`)
	w.AppendString(Ls(GL, "Students"))
	w.AppendString(`:<br>`)
	DisplayStudentsSelect(w, r.Form.GetMany("StudentID"))
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	DisplaySubmit(w, GL, "", "Create", true)

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
	w.AppendString(`<head><title>`)
	w.AppendString(Ls(GL, "Edit group"))
	w.AppendString(`</title></head>`)

	w.AppendString(`<body>`)

	w.AppendString(`<h1>`)
	w.AppendString(Ls(GL, "Group"))
	w.AppendString(`</h1>`)
	w.AppendString(`<h2>`)
	w.AppendString(Ls(GL, "Edit"))
	w.AppendString(`</h2>`)

	DisplayErrorMessage(w, GL, r.Form.Get("Error"))

	w.AppendString(`<form method="POST" action="/api/group/edit">`)

	DisplayHiddenString(w, "ID", r.Form.Get("ID"))

	w.AppendString(`<label>`)
	w.AppendString(Ls(GL, "Name"))
	w.AppendString(`: `)
	DisplayConstraintInput(w, "text", MinNameLen, MaxNameLen, "Name", r.Form.Get("Name"), true)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>`)
	w.AppendString(Ls(GL, "Students"))
	w.AppendString(`:<br>`)
	DisplayStudentsSelect(w, r.Form.GetMany("StudentID"))
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	DisplaySubmit(w, GL, "", "Save", true)

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
		return WritePage(w, r, GroupCreatePageHandler, http.BadRequest(Ls(GL, "group name length must be between %d and %d characters long"), MinGroupNameLen, MaxGroupNameLen))
	}

	nextUserID, err := database.GetNextID(UsersDB)
	if err != nil {
		return http.ServerError(err)
	}

	sids := r.Form.GetMany("StudentID")
	if len(sids) == 0 {
		return WritePage(w, r, GroupCreatePageHandler, http.BadRequest(Ls(GL, "add at least one student")))
	}
	students := make([]database.ID, len(sids))
	for i := 0; i < len(sids); i++ {
		id, err := GetValidID(sids[i], nextUserID)
		if (err != nil) || (id == AdminID) {
			return http.ClientError(err)
		}
		students[i] = id
	}

	var group Group
	group.Name = name
	group.Students = students
	group.CreatedOn = time.Now().Unix()

	if err := CreateGroup(&group); err != nil {
		return http.ServerError(err)
	}

	w.Redirect("/", http.StatusSeeOther)
	return nil
}

func GroupDeleteHandler(w *http.Response, r *http.Request) error {
	var group Group

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

	groupID, err := r.Form.GetID("ID")
	if err != nil {
		return http.ClientError(err)
	}
	if err := GetGroupByID(groupID, &group); err != nil {
		if err == database.NotFound {
			return WritePage(w, r, GroupEditPageHandler, http.NotFound(Ls(GL, "group with this ID does not exist")))
		}
		return http.ServerError(err)
	}

	if err := DeleteGroupByID(groupID); err != nil {
		return http.ServerError(err)
	}

	w.Redirect("/", http.StatusSeeOther)
	return nil
}

func GroupEditHandler(w *http.Response, r *http.Request) error {
	var group Group

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

	groupID, err := r.Form.GetID("ID")
	if err != nil {
		return http.ClientError(err)
	}
	if err := GetGroupByID(groupID, &group); err != nil {
		if err == database.NotFound {
			return WritePage(w, r, GroupEditPageHandler, http.NotFound(Ls(GL, "group with this ID does not exist")))
		}
		return http.ServerError(err)
	}

	name := r.Form.Get("Name")
	if !strings.LengthInRange(name, MinGroupNameLen, MaxGroupNameLen) {
		return WritePage(w, r, GroupEditPageHandler, http.BadRequest(Ls(GL, "group name length must be between %d and %d characters long"), MinGroupNameLen, MaxGroupNameLen))
	}

	nextUserID, err := database.GetNextID(UsersDB)
	if err != nil {
		return http.ServerError(err)
	}

	sids := r.Form.GetMany("StudentID")
	if len(sids) == 0 {
		return WritePage(w, r, GroupCreatePageHandler, http.BadRequest(Ls(GL, "add at least one student")))
	}
	students := group.Students[:0]
	for i := 0; i < len(sids); i++ {
		id, err := GetValidID(sids[i], nextUserID)
		if (err != nil) || (id == AdminID) {
			return http.ClientError(err)
		}
		students = append(students, id)
	}

	group.Name = name
	group.Students = students

	if err := SaveGroup(&group); err != nil {
		return http.ServerError(err)
	}

	w.RedirectID("/group/", groupID, http.StatusSeeOther)
	return nil
}
