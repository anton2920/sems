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

type Subject struct {
	ID    int32
	Flags int32

	TeacherID int32
	GroupID   int32
	Name      string
	Lessons   []int32
	CreatedOn int64

	Data [1024]byte
}

const (
	SubjectActive int32 = iota
	SubjectDeleted
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

func WhoIsUserInSubject(userID int32, subject *Subject) (SubjectUserType, error) {
	if userID == AdminID {
		return SubjectUserAdmin, nil
	}

	if userID == subject.TeacherID {
		return SubjectUserTeacher, nil
	}

	var group Group
	if err := GetGroupByID(DB2, subject.GroupID, &group); err != nil {
		return SubjectUserNone, err
	}
	if UserInGroup(userID, &group) {
		return SubjectUserStudent, nil
	}

	return SubjectUserNone, nil
}

func CreateSubject(db *Database, subject *Subject) error {
	var err error

	subject.ID, err = IncrementNextID(db.SubjectsFile)
	if err != nil {
		return fmt.Errorf("failed to increment subject ID: %w", err)
	}

	return SaveSubject(db, subject)
}

func DBSubject2Subject(subject *Subject) {
	data := &subject.Data[0]

	subject.Name = Offset2String(subject.Name, data)
	subject.Lessons = Offset2Slice(subject.Lessons, data)
}

func GetSubjectByID(db *Database, id int32, subject *Subject) error {
	size := int(unsafe.Sizeof(*subject))
	offset := int64(int(id)*size) + DataOffset

	n, err := syscall.Pread(db.SubjectsFile, unsafe.Slice((*byte)(unsafe.Pointer(subject)), size), offset)
	if err != nil {
		return fmt.Errorf("failed to read subject from DB: %w", err)
	}
	if n < size {
		return DBNotFound
	}

	DBSubject2Subject(subject)
	return nil
}

func GetSubjects(db *Database, pos *int64, subjects []Subject) (int, error) {
	if *pos < DataOffset {
		*pos = DataOffset
	}
	size := int(unsafe.Sizeof(subjects[0]))

	n, err := syscall.Pread(db.SubjectsFile, unsafe.Slice((*byte)(unsafe.Pointer(unsafe.SliceData(subjects))), len(subjects)*size), *pos)
	if err != nil {
		return 0, fmt.Errorf("failed to read subject from DB: %w", err)
	}
	*pos += int64(n)

	n /= size
	for i := 0; i < n; i++ {
		DBSubject2Subject(&subjects[i])
	}

	return n, nil
}

func DeleteSubjectByID(db *Database, id int32) error {
	flags := SubjectDeleted
	var subject Subject

	offset := int64(int(id)*int(unsafe.Sizeof(subject))) + DataOffset + int64(unsafe.Offsetof(subject.Flags))
	_, err := syscall.Pwrite(db.SubjectsFile, unsafe.Slice((*byte)(unsafe.Pointer(&flags)), unsafe.Sizeof(flags)), offset)
	if err != nil {
		return fmt.Errorf("failed to delete subject from DB: %w", err)
	}

	return nil
}

func SaveSubject(db *Database, subject *Subject) error {
	var subjectDB Subject
	var n int

	size := int(unsafe.Sizeof(*subject))
	offset := int64(int(subject.ID)*size) + DataOffset
	data := unsafe.Slice(&subjectDB.Data[0], len(subjectDB.Data))

	subjectDB.ID = subject.ID
	subjectDB.Flags = subject.Flags
	subjectDB.TeacherID = subject.TeacherID
	subjectDB.GroupID = subject.GroupID

	/* TODO(anton2920): save up to a sizeof(subject.Data). */
	n += String2DBString(&subjectDB.Name, subject.Name, data, n)
	n += Slice2DBSlice(&subjectDB.Lessons, subject.Lessons, data, n)

	subjectDB.CreatedOn = subject.CreatedOn

	_, err := syscall.Pwrite(db.SubjectsFile, unsafe.Slice((*byte)(unsafe.Pointer(&subjectDB)), size), offset)
	if err != nil {
		return fmt.Errorf("failed to write subject to DB: %w", err)
	}

	return nil
}

func DisplaySubjectLink(w *http.Response, subject *Subject) {
	var teacher User
	if err := GetUserByID(DB2, subject.TeacherID, &teacher); err != nil {
		/* TODO(anton2920): report error. */
		return
	}

	w.AppendString(`<a href="/subject/`)
	w.WriteInt(int(subject.ID))
	w.AppendString(`">`)
	w.WriteHTMLString(subject.Name)
	w.AppendString(` with `)
	w.WriteHTMLString(teacher.LastName)
	w.AppendString(` `)
	w.WriteHTMLString(teacher.FirstName)
	w.AppendString(` (ID: `)
	w.WriteInt(int(subject.ID))
	w.AppendString(`)`)
	w.AppendString(`</a>`)
}

func SubjectPageHandler(w *http.Response, r *http.Request) error {
	var subject Subject
	var lesson Lesson

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}

	id, err := GetIDFromURL(r.URL, "/subject/")
	if err != nil {
		return http.ClientError(err)
	}
	if err := GetSubjectByID(DB2, int32(id), &subject); err != nil {
		if err == DBNotFound {
			return http.NotFound("subject with this ID does not exist")
		}
		return http.ServerError(err)
	}

	who, err := WhoIsUserInSubject(session.ID, &subject)
	if err != nil {
		return http.ServerError(err)
	}
	if who == SubjectUserNone {
		return http.ForbiddenError
	}

	var teacher User
	if err := GetUserByID(DB2, subject.TeacherID, &teacher); err != nil {
		return http.ServerError(err)
	}

	var group Group
	if err := GetGroupByID(DB2, subject.GroupID, &group); err != nil {
		return http.ServerError(err)
	}

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>`)
	w.WriteHTMLString(subject.Name)
	w.AppendString(` with `)
	w.WriteHTMLString(teacher.LastName)
	w.AppendString(` `)
	w.WriteHTMLString(teacher.FirstName)
	w.AppendString(`</title></head>`)

	w.AppendString(`<body>`)

	w.AppendString(`<h1>`)
	w.WriteHTMLString(subject.Name)
	w.AppendString(` with `)
	w.WriteHTMLString(teacher.LastName)
	w.AppendString(` `)
	w.WriteHTMLString(teacher.FirstName)
	w.AppendString(`</h1>`)

	w.AppendString(`<p>ID: `)
	w.WriteInt(int(subject.ID))
	w.AppendString(`</p>`)

	w.AppendString(`<p>Teacher: `)
	DisplayUserLink(w, &teacher)
	w.AppendString(`</p>`)

	w.AppendString(`<p>Group: `)
	DisplayGroupLink(w, &group)
	w.AppendString(`</p>`)

	w.AppendString(`<p>Created on: `)
	DisplayFormattedTime(w, subject.CreatedOn)
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
		w.WriteInt(int(subject.TeacherID))
		w.AppendString(`">`)

		w.AppendString(`<input type="hidden" name="GroupID" value="`)
		w.WriteInt(int(subject.GroupID))
		w.AppendString(`">`)

		w.AppendString(`<input type="submit" value="Edit">`)

		w.AppendString(`</form>`)
	}

	if (len(subject.Lessons) != 0) || (who != SubjectUserStudent) {
		w.AppendString(`<h2>Lessons</h2>`)
	}
	for i := 0; i < len(subject.Lessons); i++ {
		if err := GetLessonByID(DB2, subject.Lessons[i], &lesson); err != nil {
			return http.ServerError(err)
		}

		w.AppendString(`<fieldset>`)

		w.AppendString(`<legend>Lesson #`)
		w.WriteInt(i + 1)
		if lesson.Flags == LessonDraft {
			w.AppendString(` (draft)`)
		}
		w.AppendString(`</legend>`)

		w.AppendString(`<p>Name: `)
		w.WriteHTMLString(lesson.Name)
		w.AppendString(`</p>`)

		w.AppendString(`<p>Theory: `)
		DisplayShortenedString(w, lesson.Theory, LessonTheoryMaxDisplayLen)
		w.AppendString(`</p>`)

		DisplayLessonLink(w, &lesson)

		w.AppendString(`</fieldset>`)
		w.AppendString(`<br>`)
	}

	if (session.ID == AdminID) || (session.ID == subject.TeacherID) {
		w.AppendString(`<form method="POST" action="/subject/lesson/edit">`)
		w.AppendString(`<input type="hidden" name="ID" value="`)
		w.WriteString(r.URL.Path[len("/subject/"):])
		w.AppendString(`">`)

		if len(subject.Lessons) == 0 {
			courses := make([]Course, 32)
			var displayed bool
			var pos int64

			for {
				n, err := GetCourses(DB2, &pos, courses)
				if err != nil {
					return http.ServerError(err)
				}
				if n == 0 {
					break
				}
				for i := 0; i < n; i++ {
					course := &courses[i]
					if (course.Flags != CourseDraft) && (UserOwnsCourse(&teacher, course.ID)) {
						if !displayed {
							w.AppendString(`<label>Courses: `)
							w.AppendString(`<select name="CourseID">`)
							displayed = true
						}
						w.AppendString(`<option value="`)
						w.WriteInt(i)
						w.AppendString(`">`)
						w.WriteHTMLString(course.Name)
						w.AppendString(`</option>`)
					}
				}
			}
			if displayed {
				w.AppendString(`</select>`)
				w.AppendString(`</label> `)

				w.AppendString(`<input type="submit" name="Action" value="create from"> `)
				w.AppendString(`, `)
				w.AppendString(`<input type="submit" name="Action" value="give as is"> `)
				w.AppendString(`or `)
			}
			w.AppendString(`<input type="submit" value="create new from scratch">`)
		} else {
			w.AppendString(`<input type="submit" value="Edit">`)
		}

		w.AppendString(`</form>`)
	}

	return nil
}

func DisplayTeacherSelect(w *http.Response, ids []string) {
	users := make([]User, 32)
	var pos int64

	w.AppendString(`<select name="TeacherID">`)
	for {
		n, err := GetUsers(DB2, &pos, users)
		if err != nil {
			/* TODO(anton2920): report error. */
		}
		if n == 0 {
			break
		}
		for i := 0; i < n; i++ {
			user := &users[i]
			if user.Flags == UserDeleted {
				continue
			}

			w.AppendString(`<option value="`)
			w.WriteInt(int(user.ID))
			w.AppendString(`"`)
			for j := 0; j < len(ids); j++ {
				id, err := strconv.Atoi(ids[j])
				if err != nil {
					continue
				}
				if int32(id) == user.ID {
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

func DisplayGroupSelect(w *http.Response, ids []string) {
	groups := make([]Group, 32)
	var pos int64

	w.AppendString(`<select name="GroupID">`)
	for {
		n, err := GetGroups(DB2, &pos, groups)
		if err != nil {
			/* TODO(anton2920): report error. */
		}
		if n == 0 {
			break
		}
		for i := 0; i < n; i++ {
			group := &groups[i]
			if group.Flags == GroupDeleted {
				continue
			}

			w.AppendString(`<option value="`)
			w.WriteInt(int(group.ID))
			w.AppendString(`"`)
			for j := 0; j < len(ids); j++ {
				id, err := strconv.Atoi(ids[j])
				if err != nil {
					continue
				}
				if int32(id) == group.ID {
					w.AppendString(` selected`)
				}
			}
			w.AppendString(`>`)
			w.WriteHTMLString(group.Name)
			w.AppendString(`</option>`)
		}
	}
	w.AppendString(`</select>`)
}

func SubjectCreatePageHandler(w *http.Response, r *http.Request) error {
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
	w.AppendString(`<head><title>Create subject</title></head>`)
	w.AppendString(`<body>`)

	w.AppendString(`<h1>Subject</h1>`)
	w.AppendString(`<h2>Create</h2>`)

	DisplayErrorMessage(w, r.Form.Get("Error"))

	w.AppendString(`<form method="POST" action="/api/subject/create">`)

	w.AppendString(`<label>Name: `)
	DisplayConstraintInput(w, "text", MinNameLen, MaxNameLen, "Name", r.Form.Get("Name"), true)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>Teacher: `)
	DisplayTeacherSelect(w, r.Form.GetMany("TeacherID"))
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>Group: `)
	DisplayGroupSelect(w, r.Form.GetMany("GroupID"))
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<input type="submit" value="Create">`)

	w.AppendString(`</form>`)
	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func SubjectEditPageHandler(w *http.Response, r *http.Request) error {
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
	w.AppendString(`<head><title>Edit subject</title></head>`)
	w.AppendString(`<body>`)

	w.AppendString(`<h1>Subject</h1>`)
	w.AppendString(`<h2>Edit</h2>`)

	DisplayErrorMessage(w, r.Form.Get("Error"))

	w.AppendString(`<form method="POST" action="/api/subject/edit">`)

	w.AppendString(`<input type="hidden" name="ID" value="`)
	w.WriteHTMLString(r.Form.Get("ID"))
	w.AppendString(`">`)

	w.AppendString(`<label>Name: `)
	DisplayConstraintInput(w, "text", MinNameLen, MaxNameLen, "Name", r.Form.Get("Name"), true)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>Teacher: `)
	DisplayTeacherSelect(w, r.Form.GetMany("TeacherID"))
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>Group: `)
	DisplayGroupSelect(w, r.Form.GetMany("GroupID"))
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<input type="submit" value="Save">`)

	w.AppendString(`</form>`)
	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func SubjectCreateHandler(w *http.Response, r *http.Request) error {
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
	if !strings.LengthInRange(name, MinSubjectNameLen, MaxSubjectNameLen) {
		return WritePage(w, r, SubjectCreatePageHandler, http.BadRequest("subject name length must be between %d and %d characters long", MinSubjectNameLen, MaxSubjectNameLen))
	}

	nextUserID, err := GetNextID(DB2.UsersFile)
	if err != nil {
		return http.ServerError(err)
	}
	teacherID, err := GetValidIndex(r.Form.Get("TeacherID"), int(nextUserID))
	if err != nil {
		return http.ClientError(err)
	}

	nextGroupID, err := GetNextID(DB2.GroupsFile)
	if err != nil {
		return http.ServerError(err)
	}
	groupID, err := GetValidIndex(r.Form.Get("GroupID"), int(nextGroupID))
	if err != nil {
		return http.ClientError(err)
	}

	var subject Subject
	subject.Name = name
	subject.TeacherID = int32(teacherID)
	subject.GroupID = int32(groupID)
	subject.CreatedOn = time.Now().Unix()

	if err := CreateSubject(DB2, &subject); err != nil {
		return http.ServerError(err)
	}

	w.Redirect("/", http.StatusSeeOther)
	return nil
}

func SubjectEditHandler(w *http.Response, r *http.Request) error {
	var subject Subject

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

	subjectID, err := r.Form.GetInt("ID")
	if err != nil {
		return http.ClientError(err)
	}
	if err := GetSubjectByID(DB2, int32(subjectID), &subject); err != nil {
		if err == DBNotFound {
			return http.NotFound("subject with this ID does not exist")
		}
		return http.ServerError(err)
	}

	name := r.Form.Get("Name")
	if !strings.LengthInRange(name, MinSubjectNameLen, MaxSubjectNameLen) {
		return WritePage(w, r, SubjectCreatePageHandler, http.BadRequest("subject name length must be between %d and %d characters long", MinSubjectNameLen, MaxSubjectNameLen))
	}

	nextUserID, err := GetNextID(DB2.UsersFile)
	if err != nil {
		return http.ServerError(err)
	}
	teacherID, err := GetValidIndex(r.Form.Get("TeacherID"), int(nextUserID))
	if err != nil {
		return http.ClientError(err)
	}

	nextGroupID, err := GetNextID(DB2.GroupsFile)
	if err != nil {
		return http.ServerError(err)
	}
	groupID, err := GetValidIndex(r.Form.Get("GroupID"), int(nextGroupID))
	if err != nil {
		return http.ClientError(err)
	}

	subject.Name = name
	subject.TeacherID = int32(teacherID)
	subject.GroupID = int32(groupID)

	if err := SaveSubject(DB2, &subject); err != nil {
		return http.ServerError(err)
	}

	w.RedirectID("/subject/", subjectID, http.StatusSeeOther)
	return nil
}
