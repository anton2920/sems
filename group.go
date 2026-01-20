package main

import (
	"fmt"
	"time"
	"unsafe"

	"github.com/anton2920/gofa/database"
	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/strings"
	"github.com/anton2920/gofa/syscall"
	"github.com/anton2920/gofa/trace"
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
	defer trace.End(trace.Begin(""))

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
	defer trace.End(trace.Begin(""))

	var err error

	group.ID, err = database.IncrementNextID(GroupsDB)
	if err != nil {
		return fmt.Errorf("failed to increment group ID: %w", err)
	}

	return SaveGroup(group)
}

func DBGroup2Group(group *Group) {
	defer trace.End(trace.Begin(""))

	data := &group.Data[0]

	group.Name = database.Offset2String(group.Name, data)

	slice := database.Offset2Slice(*(*[]byte)(unsafe.Pointer(&group.Students)), data)
	group.Students = *(*[]database.ID)(unsafe.Pointer(&slice))
}

func GetGroupByID(id database.ID, group *Group) error {
	defer trace.End(trace.Begin(""))

	if err := database.Read(GroupsDB, id, unsafe.Pointer(group), int(unsafe.Sizeof(*group))); err != nil {
		return err
	}

	DBGroup2Group(group)
	return nil
}

func GetGroups(pos *int64, groups []Group) (int, error) {
	defer trace.End(trace.Begin(""))

	n, err := database.ReadMany(GroupsDB, pos, *(*[]byte)(unsafe.Pointer(&groups)), int(unsafe.Sizeof(groups[0])))
	if err != nil {
		return 0, err
	}

	for i := 0; i < n; i++ {
		DBGroup2Group(&groups[i])
	}
	return n, nil
}

func DeleteGroupByID(id database.ID) error {
	defer trace.End(trace.Begin(""))

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
	defer trace.End(trace.Begin(""))

	var groupDB Group
	var n int

	groupDB.ID = group.ID
	groupDB.Flags = group.Flags

	/* TODO(anton2920): save up to a sizeof(group.Data). */
	data := unsafe.Slice(&groupDB.Data[0], len(groupDB.Data))
	n += database.String2DBString(&groupDB.Name, group.Name, data, n)
	n += database.Slice2DBSlice((*[]byte)(unsafe.Pointer(&groupDB.Students)), *(*[]byte)(unsafe.Pointer(&group.Students)), int(unsafe.Sizeof(group.Students[0])), int(unsafe.Alignof(group.Students[0])), data, n)

	groupDB.CreatedOn = group.CreatedOn

	return database.Write(GroupsDB, groupDB.ID, unsafe.Pointer(&groupDB), int(unsafe.Sizeof(groupDB)))
}

func DisplayGroupStudents(w *http.Response, l Language, group *Group) {
	w.WriteString(`<h3>`)
	w.WriteString(Ls(GL, "Students"))
	w.WriteString(`</h3>`)
	w.WriteString(`<ul>`)
	for i := 0; i < len(group.Students); i++ {
		var user User
		if err := GetUserByID(group.Students[i], &user); err != nil {
			/* TODO(anton2920): report error. */
		}

		w.WriteString(`<li>`)
		DisplayUserLink(w, GL, &user)
		w.WriteString(`</li>`)
	}
	w.WriteString(`</ul>`)
	w.WriteString(`<br>`)
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
				w.WriteString(`<h3>`)
				w.WriteString(Ls(l, "Subjects"))
				w.WriteString(`</h3>`)
				w.WriteString(`<ul>`)
				displayed = true
			}

			w.WriteString(`<li>`)
			DisplaySubjectLink(w, l, subject)
			w.WriteString(`</li>`)
		}
	}
	if displayed {
		w.WriteString(`</ul>`)
	}
}

func DisplayGroupTitle(w *http.Response, l Language, group *Group) {
	w.WriteHTMLString(group.Name)
	w.WriteString(` (ID: `)
	w.WriteInt(int(group.ID))
	w.WriteString(`)`)
	DisplayDeleted(w, l, group.Flags == GroupDeleted)
}

func DisplayGroupLink(w *http.Response, l Language, group *Group) {
	w.WriteString(`<a href="/group/`)
	w.WriteInt(int(group.ID))
	w.WriteString(`">`)
	DisplayGroupTitle(w, l, group)
	w.WriteString(`</a>`)
}

func GroupsPageHandler(w *http.Response, r *http.Request) error {
	defer trace.End(trace.Begin(""))

	const width = WidthMedium

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}

	DisplayHTMLStart(w)

	DisplayHeadStart(w)
	{
		w.WriteString(`<title>`)
		w.WriteString(Ls(GL, "Groups"))
		w.WriteString(`</title>`)
	}
	DisplayHeadEnd(w)

	DisplayBodyStart(w)
	{
		DisplayHeader(w, GL)
		DisplaySidebar(w, GL, session)

		DisplayMainStart(w)

		DisplayCrumbsStart(w, width)
		{
			DisplayCrumbsItem(w, GL, "Groups")
		}
		DisplayCrumbsEnd(w)

		DisplayPageStart(w, width)
		{
			w.WriteString(`<h2 class="text-center">`)
			w.WriteString(Ls(GL, "Groups"))
			w.WriteString(`</h2>`)
			w.WriteString(`<br>`)

			DisplayTableStart(w, GL, []string{"ID", "Name", "Created on", "Status"})
			{
				groups := make([]Group, 32)
				var pos int64

				for {
					n, err := GetGroups(&pos, groups)
					if err != nil {
						return http.ServerError(err)
					}
					if n == 0 {
						break
					}

					for i := 0; i < n; i++ {
						group := &groups[i]
						if !UserInGroup(session.ID, group) {
							continue
						}

						DisplayTableRowLinkIDStart(w, "/group", group.ID)

						DisplayTableItemString(w, group.Name)
						DisplayTableItemTime(w, group.CreatedOn)
						DisplayTableItemFlags(w, GL, group.Flags)

						DisplayTableRowEnd(w)
					}
				}
			}
			DisplayTableEnd(w)

			if session.ID == AdminID {
				w.WriteString(`<br>`)
				w.WriteString(`<form method="POST" action="/group/create">`)
				DisplaySubmit(w, GL, "", "Create group", true)
				w.WriteString(`</form>`)
			}
		}
		DisplayPageEnd(w)
		DisplayMainEnd(w)
	}
	DisplayBodyEnd(w)

	DisplayHTMLEnd(w)
	return nil
}

func GroupPageHandler(w *http.Response, r *http.Request) error {
	defer trace.End(trace.Begin(""))

	const width = WidthLarge

	var group Group

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}

	id, err := GetIDFromURL(GL, r.URL, "/group/")
	if err != nil {
		return err
	}
	if err := GetGroupByID(id, &group); err != nil {
		if err == database.NotFound {
			return http.NotFound("group with this ID does not exist")
		}
		return http.ServerError(err)
	}
	if !UserInGroup(session.ID, &group) {
		return ForbiddenError
	}

	DisplayHTMLStart(w)

	DisplayHeadStart(w)
	{
		w.WriteString(`<title>`)
		DisplayGroupTitle(w, GL, &group)
		w.WriteString(`</title>`)
	}
	DisplayHeadEnd(w)

	DisplayBodyStart(w)
	{
		DisplayHeader(w, GL)
		DisplaySidebar(w, GL, session)

		DisplayMainStart(w)

		DisplayCrumbsStart(w, width)
		{
			DisplayCrumbsItemStart(w)
			DisplayGroupTitle(w, GL, &group)
			DisplayCrumbsItemEnd(w)
		}
		DisplayCrumbsEnd(w)

		DisplayPageStart(w, width)
		{
			w.WriteString(`<h2>`)
			DisplayGroupTitle(w, GL, &group)
			w.WriteString(`</h2>`)
			w.WriteString(`<br>`)

			w.WriteString(`<h3>`)
			w.WriteString(Ls(GL, "Info"))
			w.WriteString(`</h3>`)

			w.WriteString(`<p>`)
			w.WriteString(Ls(GL, "Created on"))
			w.WriteString(`: `)
			DisplayFormattedTime(w, group.CreatedOn)
			w.WriteString(`</p>`)

			if session.ID == AdminID {
				w.WriteString(`<div>`)
				w.WriteString(`<form style="display:inline" method="POST" action="/group/edit">`)
				DisplayHiddenID(w, "ID", group.ID)
				DisplayHiddenString(w, "Name", group.Name)
				for i := 0; i < len(group.Students); i++ {
					DisplayHiddenID(w, "StudentID", group.Students[i])
				}
				DisplayButton(w, GL, "", "Edit")
				w.WriteString(`</form>`)

				w.WriteString(` <form style="display:inline" method="POST" action="/api/group/delete">`)
				DisplayHiddenID(w, "ID", group.ID)
				DisplayButton(w, GL, "", "Delete")
				w.WriteString(`</form>`)
				w.WriteString(`</div>`)
				w.WriteString(`<br>`)
			}

			DisplayGroupStudents(w, GL, &group)
			DisplayGroupSubjects(w, GL, &group)
		}
		DisplayPageEnd(w)
		DisplayMainEnd(w)
	}
	DisplayBodyEnd(w)

	DisplayHTMLEnd(w)
	return nil
}

func DisplayStudentsSelect(w *http.Response, ids []string) {
	users := make([]User, 32)
	var pos int64

	w.WriteString(`<select class="form-select" name="StudentID" multiple>`)
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

			w.WriteString(`<option value="`)
			w.WriteInt(int(user.ID))
			w.WriteString(`"`)
			for j := 0; j < len(ids); j++ {
				id, err := GetValidID(ids[j], database.MaxValidID)
				if err != nil {
					continue
				}
				if id == user.ID {
					w.WriteString(` selected`)
				}
			}
			w.WriteString(`>`)
			w.WriteHTMLString(user.LastName)
			w.WriteString(` `)
			w.WriteHTMLString(user.FirstName)
			w.WriteString(`</option>`)
		}
	}
	w.WriteString(`</select>`)
}

func GroupCreateEditPageHandler(w *http.Response, r *http.Request, session *Session, group *Group, endpoint string, title string, action string, err error) error {
	defer trace.End(trace.Begin(""))

	const width = WidthSmall

	DisplayHTMLStart(w)

	DisplayHeadStart(w)
	{
		w.WriteString(`<title>`)
		w.WriteString(Ls(GL, title))
		w.WriteString(`</title>`)
	}
	DisplayHeadEnd(w)

	DisplayBodyStart(w)
	{
		DisplayHeader(w, GL)
		DisplaySidebar(w, GL, session)

		DisplayMainStart(w)

		DisplayCrumbsStart(w, width)
		{
			switch title {
			case "Create group":
				DisplayCrumbsLink(w, GL, "/groups", "Groups")
			case "Edit group":
				DisplayCrumbsLinkIDStart(w, "/group", group.ID)
				DisplayGroupTitle(w, GL, group)
				DisplayCrumbsLinkEnd(w)
			}
			DisplayCrumbsItem(w, GL, title)
		}
		DisplayCrumbsEnd(w)

		DisplayFormPageStart(w, r, GL, width, title, endpoint, err)
		{
			DisplayLabel(w, GL, "Name")
			DisplayConstraintInput(w, "text", MinGroupNameLen, MaxGroupNameLen, "Name", r.Form.Get("Name"), true)
			w.WriteString(`<br>`)

			DisplayLabel(w, GL, "Students")
			DisplayStudentsSelect(w, r.Form.GetMany("StudentID"))
			w.WriteString(`<br>`)

			DisplaySubmit(w, GL, "", action, true)
		}
		DisplayFormPageEnd(w)
		DisplayMainEnd(w)
	}
	DisplayBodyEnd(w)

	DisplayHTMLEnd(w)
	return nil
}

func GroupCreatePageHandler(w *http.Response, r *http.Request, e error) error {
	defer trace.End(trace.Begin(""))

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}
	if session.ID != AdminID {
		return ForbiddenError
	}

	return GroupCreateEditPageHandler(w, r, session, nil, APIPrefix+"/group/create", "Create group", "Create", e)
}

func GroupEditPageHandler(w *http.Response, r *http.Request, e error) error {
	defer trace.End(trace.Begin(""))

	var group Group

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}
	if session.ID != AdminID {
		return ForbiddenError
	}

	groupID, err := r.Form.GetID("ID")
	if err != nil {
		return http.ClientError(err)
	}
	if err := GetGroupByID(groupID, &group); err != nil {
		if err == database.NotFound {
			return http.NotFound("group with this ID does not exist")
		}
		return http.ServerError(err)
	}

	return GroupCreateEditPageHandler(w, r, session, &group, APIPrefix+"/group/edit", "Edit group", "Save", e)
}

func GroupCreateHandler(w *http.Response, r *http.Request) error {
	defer trace.End(trace.Begin(""))

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}
	if session.ID != AdminID {
		return ForbiddenError
	}

	name := r.Form.Get("Name")
	if !strings.LengthInRange(name, MinGroupNameLen, MaxGroupNameLen) {
		return GroupCreatePageHandler(w, r, http.BadRequest(Ls(GL, "group name length must be between %d and %d characters long"), MinGroupNameLen, MaxGroupNameLen))
	}

	nextUserID, err := database.GetNextID(UsersDB)
	if err != nil {
		return http.ServerError(err)
	}

	sids := r.Form.GetMany("StudentID")
	if len(sids) == 0 {
		return GroupCreatePageHandler(w, r, http.BadRequest("%s", Ls(GL, "add at least one student")))
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

	w.Redirect("/groups", http.StatusSeeOther)
	return nil
}

func GroupDeleteHandler(w *http.Response, r *http.Request) error {
	defer trace.End(trace.Begin(""))

	var group Group

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}
	if session.ID != AdminID {
		return ForbiddenError
	}

	groupID, err := r.Form.GetID("ID")
	if err != nil {
		return http.ClientError(err)
	}
	if err := GetGroupByID(groupID, &group); err != nil {
		if err == database.NotFound {
			return http.NotFound("%s", Ls(GL, "group with this ID does not exist"))
		}
		return http.ServerError(err)
	}

	if err := DeleteGroupByID(groupID); err != nil {
		return http.ServerError(err)
	}

	w.Redirect("/groups", http.StatusSeeOther)
	return nil
}

func GroupEditHandler(w *http.Response, r *http.Request) error {
	defer trace.End(trace.Begin(""))

	var group Group

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}
	if session.ID != AdminID {
		return ForbiddenError
	}

	groupID, err := r.Form.GetID("ID")
	if err != nil {
		return http.ClientError(err)
	}
	if err := GetGroupByID(groupID, &group); err != nil {
		if err == database.NotFound {
			return GroupEditPageHandler(w, r, http.NotFound("%s", Ls(GL, "group with this ID does not exist")))
		}
		return http.ServerError(err)
	}

	name := r.Form.Get("Name")
	if !strings.LengthInRange(name, MinGroupNameLen, MaxGroupNameLen) {
		return GroupEditPageHandler(w, r, http.BadRequest(Ls(GL, "group name length must be between %d and %d characters long"), MinGroupNameLen, MaxGroupNameLen))
	}

	nextUserID, err := database.GetNextID(UsersDB)
	if err != nil {
		return http.ServerError(err)
	}

	sids := r.Form.GetMany("StudentID")
	if len(sids) == 0 {
		return GroupEditPageHandler(w, r, http.BadRequest("%s", Ls(GL, "add at least one student")))
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

	w.Redirect(w.PathID("/group/", groupID), http.StatusSeeOther)
	return nil
}
