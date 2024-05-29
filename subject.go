package main

import (
	"strconv"
	"time"

	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/strings"
)

type Subject struct {
	ID        int
	Name      string
	TeacherID int32
	GroupID   int32
	CreatedOn int64

	Lessons []Lesson
}

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

func DisplaySubjectLink(w *http.Response, subject *Subject) {
	var teacher User
	if err := GetUserByID(DB2, subject.TeacherID, &teacher); err != nil {
		/* TODO(anton2920): report error. */
		return
	}

	w.AppendString(`<a href="/subject/`)
	w.WriteInt(subject.ID)
	w.AppendString(`">`)
	w.WriteHTMLString(subject.Name)
	w.AppendString(` with `)
	w.WriteHTMLString(teacher.LastName)
	w.AppendString(` `)
	w.WriteHTMLString(teacher.FirstName)
	w.AppendString(` (ID: `)
	w.WriteInt(subject.ID)
	w.AppendString(`)`)
	w.AppendString(`</a>`)
}

func SubjectPageHandler(w *http.Response, r *http.Request) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}

	id, err := GetIDFromURL(r.URL, "/subject/")
	if err != nil {
		return http.ClientError(err)
	}
	if (id < 0) || (id >= len(DB.Subjects)) {
		return http.NotFound("subject with this ID does not exist")
	}
	subject := &DB.Subjects[id]

	who, err := WhoIsUserInSubject(session.ID, subject)
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
	w.WriteInt(subject.ID)
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
		DisplayShortenedString(w, lesson.Theory, LessonTheoryMaxDisplayLen)
		w.AppendString(`</p>`)

		if (session.ID == AdminID) || (session.ID == subject.TeacherID) {
			var displaySubmissions bool
			for j := 0; j < len(lesson.Submissions); j++ {
				submission := lesson.Submissions[j]
				if !submission.Draft {
					displaySubmissions = true
					break
				}
			}
			if displaySubmissions {
				w.AppendString(`<form method="POST" action="/submission">`)

				w.AppendString(`<input type="hidden" name="ID" value="`)
				w.WriteString(r.URL.Path[len("/subject/"):])
				w.AppendString(`">`)

				w.AppendString(`<input type="hidden" name="LessonIndex" value="`)
				w.WriteInt(i)
				w.AppendString(`">`)

				w.AppendString(`<label>Submissions: `)
				w.AppendString(`<select name="SubmissionIndex">`)
				for j := 0; j < len(lesson.Submissions); j++ {
					submission := lesson.Submissions[j]
					if submission.Draft {
						continue
					}

					var user User
					if err := GetUserByID(DB2, submission.UserID, &user); err != nil {
						return http.ServerError(err)
					}

					w.AppendString(`<option value="`)
					w.WriteInt(j)
					w.AppendString(`">`)
					w.WriteHTMLString(user.LastName)
					w.AppendString(` `)
					w.WriteHTMLString(user.FirstName)

					if submission.Status == SubmissionCheckDone {
						w.AppendString(` (`)
						DisplaySubmissionTotalScore(w, submission)
						w.AppendString(`)`)
					}
					w.AppendString(`</option>`)
				}
				w.AppendString(`</select>`)
				w.AppendString(`</label> `)

				w.AppendString(`<input type="submit" value="See results"> `)
				w.AppendString(`<input type="submit" name="NextPage" value="Discard">`)

				w.AppendString(`</form>`)

				w.AppendString(`<br>`)
			}
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

	if (session.ID == AdminID) || (session.ID == subject.TeacherID) {
		w.AppendString(`<form method="POST" action="/subject/lesson/edit">`)
		w.AppendString(`<input type="hidden" name="ID" value="`)
		w.WriteString(r.URL.Path[len("/subject/"):])
		w.AppendString(`">`)

		if len(subject.Lessons) == 0 {
			var displayCourses bool
			for i := 0; i < len(DB.Courses); i++ {
				course := &DB.Courses[i]
				if (course.Flags != CourseDraft) && (UserOwnsCourse(&teacher, course.ID)) {
					displayCourses = true
					break
				}
			}
			if displayCourses {
				w.AppendString(`<label>Courses: `)
				w.AppendString(`<select name="CourseID">`)
				for i := 0; i < len(DB.Courses); i++ {
					course := &DB.Courses[i]
					if (course.Flags == CourseDraft) || (!UserOwnsCourse(&teacher, course.ID)) {
						continue
					}

					w.AppendString(`<option value="`)
					w.WriteInt(i)
					w.AppendString(`">`)
					w.WriteHTMLString(course.Name)
					w.AppendString(`</option>`)
				}
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
	teachers := make([]User, 32)
	var pos int64

	w.AppendString(`<select name="TeacherID">`)
	for {
		n, err := GetUsers(DB2, &pos, teachers)
		if err != nil {
			/* TODO(anton2920): report error. */
		}
		if n == 0 {
			break
		}
		for i := 0; i < n; i++ {
			teacher := &teachers[i]

			w.AppendString(`<option value="`)
			w.WriteInt(int(teacher.ID))
			w.AppendString(`"`)
			for j := 0; j < len(ids); j++ {
				id, err := strconv.Atoi(ids[j])
				if err != nil {
					continue
				}
				if int32(id) == teacher.ID {
					w.AppendString(` selected`)
				}
			}
			w.AppendString(`>`)
			w.WriteHTMLString(teacher.LastName)
			w.AppendString(` `)
			w.WriteHTMLString(teacher.FirstName)
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

	DB.Subjects = append(DB.Subjects, Subject{ID: len(DB.Subjects), Name: name, TeacherID: int32(teacherID), GroupID: int32(groupID), CreatedOn: time.Now().Unix()})

	w.Redirect("/", http.StatusSeeOther)
	return nil
}

func SubjectEditHandler(w *http.Response, r *http.Request) error {
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

	subjectID, err := GetValidIndex(r.Form.Get("ID"), len(DB.Subjects))
	if err != nil {
		return http.ClientError(err)
	}
	subject := &DB.Subjects[subjectID]

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

	w.RedirectID("/subject/", subjectID, http.StatusSeeOther)
	return nil
}
