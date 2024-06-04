package main

import (
	"fmt"
	"strconv"
	"time"
	"unsafe"

	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/strings"
	"github.com/anton2920/gofa/syscall"
)

type Group struct {
	ID    int32
	Flags int32

	Name      string
	Students  []int32
	CreatedOn int64

	Data [1024]byte
}

const (
	GroupActive  int32 = 0
	GroupDeleted       = 1
)

const (
	MinGroupNameLen = 5
	MaxGroupNameLen = 15
)

func UserInGroup(userID int32, group *Group) bool {
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

func CreateGroup(db *Database, group *Group) error {
	var err error

	group.ID, err = IncrementNextID(db.GroupsFile)
	if err != nil {
		return fmt.Errorf("failed to increment group ID: %w", err)
	}

	return SaveGroup(db, group)
}

func DBGroup2Group(group *Group) {
	group.Name = Offset2String(group.Name, &group.Data[0])
	group.Students = Offset2Slice(group.Students, &group.Data[0])
}

func GetGroupByID(db *Database, id int32, group *Group) error {
	size := int(unsafe.Sizeof(*group))
	offset := int64(int(id)*size) + DataOffset

	n, err := syscall.Pread(db.GroupsFile, unsafe.Slice((*byte)(unsafe.Pointer(group)), size), offset)
	if err != nil {
		return fmt.Errorf("failed to read group from DB: %w", err)
	}
	if n < size {
		return DBNotFound
	}

	DBGroup2Group(group)
	return nil
}

func GetGroups(db *Database, pos *int64, groups []Group) (int, error) {
	if *pos < DataOffset {
		*pos = DataOffset
	}
	size := int(unsafe.Sizeof(groups[0]))

	n, err := syscall.Pread(db.GroupsFile, unsafe.Slice((*byte)(unsafe.Pointer(unsafe.SliceData(groups))), len(groups)*size), *pos)
	if err != nil {
		return 0, fmt.Errorf("failed to read group from DB: %w", err)
	}
	*pos += int64(n)

	n /= size
	for i := 0; i < n; i++ {
		DBGroup2Group(&groups[i])
	}

	return n, nil
}

func DeleteGroupByID(db *Database, id int32) error {
	flags := GroupDeleted
	var group Group

	offset := int64(int(id)*int(unsafe.Sizeof(group))) + DataOffset + int64(unsafe.Offsetof(group.Flags))
	_, err := syscall.Pwrite(db.GroupsFile, unsafe.Slice((*byte)(unsafe.Pointer(&flags)), unsafe.Sizeof(flags)), offset)
	if err != nil {
		return fmt.Errorf("failed to delete group from DB: %w", err)
	}

	return nil
}

func SaveGroup(db *Database, group *Group) error {
	var groupDB Group

	size := int(unsafe.Sizeof(*group))
	offset := int64(int(group.ID)*size) + DataOffset

	n := DataStartOffset
	var nbytes int

	groupDB.ID = group.ID
	groupDB.Flags = group.Flags

	/* TODO(anton2920): saving up to a sizeof(group.Data). */
	nbytes = copy(groupDB.Data[n:], group.Name)
	groupDB.Name = String2Offset(group.Name, n)
	n += nbytes

	nbytes = copy(groupDB.Data[n:], unsafe.Slice((*byte)(unsafe.Pointer(&group.Students[0])), len(group.Students)*int(unsafe.Sizeof(group.Students[0]))))
	groupDB.Students = Slice2Offset(group.Students, n)
	nbytes += n

	groupDB.CreatedOn = group.CreatedOn

	_, err := syscall.Pwrite(db.GroupsFile, unsafe.Slice((*byte)(unsafe.Pointer(&groupDB)), size), offset)
	if err != nil {
		return fmt.Errorf("failed to write group to DB: %w", err)
	}

	return nil
}

func DisplayGroupLink(w *http.Response, group *Group) {
	w.AppendString(`<a href="/group/`)
	w.WriteInt(int(group.ID))
	w.AppendString(`">`)
	w.WriteHTMLString(group.Name)
	w.AppendString(` (ID: `)
	w.WriteInt(int(group.ID))
	w.AppendString(`)`)
	w.AppendString(`</a>`)
}

func GroupPageHandler(w *http.Response, r *http.Request) error {
	var group Group

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}

	id, err := GetIDFromURL(r.URL, "/group/")
	if err != nil {
		return http.ClientError(err)
	}
	if err := GetGroupByID(DB2, int32(id), &group); err != nil {
		if err == DBNotFound {
			return http.NotFound("group with this ID does not exist")
		}
		return http.ServerError(err)
	}

	if !UserInGroup(session.ID, &group) {
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
	w.WriteInt(int(group.ID))
	w.AppendString(`</p>`)

	w.AppendString(`<p>Created on: `)
	DisplayFormattedTime(w, group.CreatedOn)
	w.AppendString(`</p>`)

	w.AppendString(`<h2>Students</h2>`)
	w.AppendString(`<ul>`)
	for i := 0; i < len(group.Students); i++ {
		var user User
		if err := GetUserByID(DB2, group.Students[i], &user); err != nil {
			return http.ServerError(err)
		}

		w.AppendString(`<li>`)
		DisplayUserLink(w, &user)
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
			studentID := group.Students[i]
			w.AppendString(`<input type="hidden" name="StudentID" value="`)
			w.WriteInt(int(studentID))
			w.AppendString(`">`)
		}

		w.AppendString(`<input type="submit" value="Edit">`)

		w.AppendString(`</form>`)
	}

	var displaySubjects bool
	for i := 0; i < len(DB.Subjects); i++ {
		subject := &DB.Subjects[i]

		if group.ID == subject.GroupID {
			displaySubjects = true
			break
		}
	}
	if displaySubjects {
		w.AppendString(`<h2>Subjects</h2>`)
		w.AppendString(`<ul>`)
		for i := 0; i < len(DB.Subjects); i++ {
			subject := &DB.Subjects[i]

			if group.ID == subject.GroupID {
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
	students := make([]User, 32)
	var pos int64

	w.AppendString(`<select name="StudentID" multiple>`)
	for {
		n, err := GetUsers(DB2, &pos, students)
		if err != nil {
			/* TODO(anton2920): report error. */
		}
		if n == 0 {
			break
		}
		for i := 0; i < n; i++ {
			student := &students[i]
			if student.ID == AdminID {
				continue
			}

			w.AppendString(`<option value="`)
			w.WriteInt(int(student.ID))
			w.AppendString(`"`)
			for j := 0; j < len(ids); j++ {
				id, err := strconv.Atoi(ids[j])
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

	nextUserID, err := GetNextID(DB2.UsersFile)
	if err != nil {
		return http.ServerError(err)
	}

	sids := r.Form.GetMany("StudentID")
	if len(sids) == 0 {
		return WritePage(w, r, GroupCreatePageHandler, http.BadRequest("add at least one student"))
	}

	students := make([]int32, len(sids))
	for i := 0; i < len(sids); i++ {
		id, err := GetValidIndex(sids[i], int(nextUserID))
		if (err != nil) || (id == AdminID) {
			return http.ClientError(err)
		}
		students[i] = int32(id)
	}

	var group Group
	group.Name = name
	group.Students = students
	group.CreatedOn = time.Now().Unix()

	if err := CreateGroup(DB2, &group); err != nil {
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

	groupID, err := r.Form.GetInt("ID")
	if err != nil {
		return http.ClientError(err)
	}
	if err := GetGroupByID(DB2, int32(groupID), &group); err != nil {
		if err == DBNotFound {
			return WritePage(w, r, GroupEditPageHandler, http.NotFound("group with this ID is not found"))
		}
		return http.ServerError(err)
	}

	name := r.Form.Get("Name")
	if !strings.LengthInRange(name, MinGroupNameLen, MaxGroupNameLen) {
		return WritePage(w, r, GroupEditPageHandler, http.BadRequest("group name length must be between %d and %d characters long", MinGroupNameLen, MaxGroupNameLen))
	}

	nextUserID, err := GetNextID(DB2.UsersFile)
	if err != nil {
		return http.ServerError(err)
	}

	sids := r.Form.GetMany("StudentID")
	if len(sids) == 0 {
		return WritePage(w, r, GroupCreatePageHandler, http.BadRequest("add at least one student"))
	}

	students := group.Students[:0]
	for i := 0; i < len(sids); i++ {
		id, err := GetValidIndex(sids[i], int(nextUserID))
		if (err != nil) || (id == AdminID) {
			return http.ClientError(err)
		}
		students = append(students, int32(id))
	}

	group.Name = name
	group.Students = students

	if err := SaveGroup(DB2, &group); err != nil {
		return http.ServerError(err)
	}

	w.RedirectID("/group/", groupID, http.StatusSeeOther)
	return nil
}
