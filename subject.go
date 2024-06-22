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

type Subject struct {
	ID    database.ID
	Flags int32

	TeacherID database.ID
	GroupID   database.ID
	Name      string
	Lessons   []database.ID
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

func WhoIsUserInSubject(userID database.ID, subject *Subject) (SubjectUserType, error) {
	if userID == AdminID {
		return SubjectUserAdmin, nil
	}

	if userID == subject.TeacherID {
		return SubjectUserTeacher, nil
	}

	var group Group
	if err := GetGroupByID(subject.GroupID, &group); err != nil {
		return SubjectUserNone, err
	}
	if UserInGroup(userID, &group) {
		return SubjectUserStudent, nil
	}

	return SubjectUserNone, nil
}

func CreateSubject(subject *Subject) error {
	var err error

	subject.ID, err = database.IncrementNextID(SubjectsDB)
	if err != nil {
		return fmt.Errorf("failed to increment subject ID: %w", err)
	}

	return SaveSubject(subject)
}

func DBSubject2Subject(subject *Subject) {
	data := &subject.Data[0]

	subject.Name = database.Offset2String(subject.Name, data)
	subject.Lessons = database.Offset2Slice(subject.Lessons, data)
}

func GetSubjectByID(id database.ID, subject *Subject) error {
	if err := database.Read(SubjectsDB, id, subject); err != nil {
		return err
	}

	DBSubject2Subject(subject)
	return nil
}

func GetSubjects(pos *int64, subjects []Subject) (int, error) {
	n, err := database.ReadMany(SubjectsDB, pos, subjects)
	if err != nil {
		return 0, err
	}

	for i := 0; i < n; i++ {
		DBSubject2Subject(&subjects[i])
	}
	return n, nil
}

func DeleteSubjectByID(id database.ID) error {
	flags := SubjectDeleted
	var subject Subject

	offset := int64(int(id)*int(unsafe.Sizeof(subject))) + database.DataOffset + int64(unsafe.Offsetof(subject.Flags))
	_, err := syscall.Pwrite(SubjectsDB.FD, unsafe.Slice((*byte)(unsafe.Pointer(&flags)), unsafe.Sizeof(flags)), offset)
	if err != nil {
		return fmt.Errorf("failed to delete subject from DB: %w", err)
	}

	return nil
}

func SaveSubject(subject *Subject) error {
	var subjectDB Subject
	var n int

	subjectDB.ID = subject.ID
	subjectDB.Flags = subject.Flags
	subjectDB.TeacherID = subject.TeacherID
	subjectDB.GroupID = subject.GroupID

	/* TODO(anton2920): save up to a sizeof(subject.Data). */
	data := unsafe.Slice(&subjectDB.Data[0], len(subjectDB.Data))
	n += database.String2DBString(&subjectDB.Name, subject.Name, data, n)
	n += database.Slice2DBSlice(&subjectDB.Lessons, subject.Lessons, data, n)

	subjectDB.CreatedOn = subject.CreatedOn

	return database.Write(SubjectsDB, subjectDB.ID, &subjectDB)
}

func DisplaySubjectCoursesSelect(w *http.Response, l Language, subject *Subject, teacher *User) {
	w.AppendString(`<form method="POST" action="/subject/lessons">`)
	DisplayHiddenID(w, "ID", subject.ID)

	if len(subject.Lessons) == 0 {
		courses := make([]Course, 32)
		var displayed bool
		var pos int64

		for {
			n, err := GetCourses(&pos, courses)
			if err != nil {
				/* TODO(anton2920): report error. */
			}
			if n == 0 {
				break
			}
			for i := 0; i < n; i++ {
				course := &courses[i]
				if (course.Flags != CourseActive) || (!UserOwnsCourse(teacher, course.ID)) {
					continue
				}

				if !displayed {
					w.AppendString(`<label>`)
					w.AppendString(Ls(l, "Courses"))
					w.AppendString(`: `)
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
		if displayed {
			w.AppendString(`</select>`)
			w.AppendString(`</label> `)

			DisplayButton(w, l, "Action", "create from")
			w.AppendString(`, `)
			DisplayButton(w, l, "Action", "give as is")
			w.AppendString(` `)
			w.AppendString(Ls(l, "or"))
			w.AppendString(` `)
		}
		DisplayButton(w, l, "Action", "create new from scratch")
	} else {
		DisplayButton(w, l, "", "Edit")
	}

	w.AppendString(`</form>`)
}

func DisplaySubjectTitle(w *http.Response, l Language, subject *Subject, teacher *User) {
	w.WriteHTMLString(subject.Name)
	w.AppendString(` `)
	w.AppendString(Ls(l, "with"))
	w.AppendString(` `)
	w.WriteHTMLString(teacher.LastName)
	w.AppendString(` `)
	w.WriteHTMLString(teacher.FirstName)
	w.AppendString(` (ID: `)
	w.WriteID(subject.ID)
	w.AppendString(`)`)
	if subject.Flags == SubjectDeleted {
		w.AppendString(` [deleted]`)
	}
}

func DisplaySubjectLink(w *http.Response, l Language, subject *Subject) {
	var teacher User
	if err := GetUserByID(subject.TeacherID, &teacher); err != nil {
		/* TODO(anton2920): report error. */
		return
	}

	w.AppendString(`<a href="/subject/`)
	w.WriteID(subject.ID)
	w.AppendString(`">`)
	DisplaySubjectTitle(w, l, subject, &teacher)
	w.AppendString(`</a>`)
}

func SubjectPageHandler(w *http.Response, r *http.Request) error {
	var subject Subject

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}

	id, err := GetIDFromURL(GL, r.URL, "/subject/")
	if err != nil {
		return http.ClientError(err)
	}
	if err := GetSubjectByID(id, &subject); err != nil {
		if err == database.NotFound {
			return http.NotFound(Ls(GL, "subject with this ID does not exist"))
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
	if err := GetUserByID(subject.TeacherID, &teacher); err != nil {
		return http.ServerError(err)
	}

	var group Group
	if err := GetGroupByID(subject.GroupID, &group); err != nil {
		return http.ServerError(err)
	}

	DisplayHTMLStart(w)

	DisplayHeadStart(w)
	{
		w.AppendString(`<title>`)
		DisplaySubjectTitle(w, GL, &subject, &teacher)
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
			DisplaySubjectTitle(w, GL, &subject, &teacher)
			w.AppendString(`</h2>`)
			w.AppendString(`<br>`)

			w.AppendString(`<p>`)
			w.AppendString(Ls(GL, "Teacher"))
			w.AppendString(`: `)
			DisplayUserLink(w, GL, &teacher)
			w.AppendString(`</p>`)

			w.AppendString(`<p>`)
			w.AppendString(Ls(GL, "Group"))
			w.AppendString(`: `)
			DisplayGroupLink(w, GL, &group)
			w.AppendString(`</p>`)

			w.AppendString(`<p>`)
			w.AppendString(Ls(GL, "Created on"))
			w.AppendString(`: `)
			DisplayFormattedTime(w, subject.CreatedOn)
			w.AppendString(`</p>`)

			if session.ID == AdminID {
				w.AppendString(`<div>`)
				w.AppendString(`<form style="display:inline" method="POST" action="/subject/edit">`)
				DisplayHiddenID(w, "ID", subject.ID)
				DisplayHiddenID(w, "TeacherID", subject.TeacherID)
				DisplayHiddenID(w, "GroupID", subject.GroupID)
				DisplayHiddenString(w, "Name", subject.Name)
				DisplayButton(w, GL, "", "Edit")
				w.AppendString(`</form>`)

				w.AppendString(` <form style="display:inline" method="POST" action="/api/subject/delete">`)
				DisplayHiddenID(w, "ID", subject.ID)
				DisplayButton(w, GL, "", "Delete")
				w.AppendString(`</form>`)
				w.AppendString(`</div>`)
				w.AppendString(`<br>`)
			}

			if (len(subject.Lessons) != 0) || (who != SubjectUserStudent) {
				w.AppendString(`<h3>`)
				w.AppendString(Ls(GL, "Lessons"))
				w.AppendString(`</h3>`)
				DisplayLessons(w, GL, subject.Lessons)
			}

			if (session.ID == AdminID) || (session.ID == subject.TeacherID) {
				DisplaySubjectCoursesSelect(w, GL, &subject, &teacher)
			}
		}
		DisplayPageEnd(w)
	}
	DisplayBodyEnd(w)

	DisplayHTMLEnd(w)
	return nil
}

func DisplayTeacherSelect(w *http.Response, ids []string) {
	users := make([]User, 32)
	var pos int64

	w.AppendString(`<select class="form-select" name="TeacherID">`)
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
			if user.Flags == UserDeleted {
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

func DisplayGroupSelect(w *http.Response, ids []string) {
	groups := make([]Group, 32)
	var pos int64

	w.AppendString(`<select class="form-select" name="GroupID">`)
	for {
		n, err := GetGroups(&pos, groups)
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
			w.WriteID(group.ID)
			w.AppendString(`"`)
			for j := 0; j < len(ids); j++ {
				id, err := GetValidID(ids[j], database.MaxValidID)
				if err != nil {
					continue
				}
				if id == group.ID {
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

func SubjectCreateEditPageHandler(w *http.Response, r *http.Request, endpoint string, title string, action string) error {
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

	DisplayHTMLStart(w)

	DisplayHeadStart(w)
	{
		w.AppendString(`<title>`)
		w.AppendString(Ls(GL, title))
		w.AppendString(`</title>`)
	}
	DisplayHeadEnd(w)

	DisplayBodyStart(w)
	{
		DisplayHeader(w, GL)
		DisplaySidebar(w, GL, session.ID)

		DisplayFormStart(w, r, GL, title, endpoint)
		{
			DisplayInputLabel(w, GL, "Name")
			DisplayConstraintInput(w, "text", MinNameLen, MaxNameLen, "Name", r.Form.Get("Name"), true)
			w.AppendString(`<br>`)

			DisplayInputLabel(w, GL, "Teacher")
			DisplayTeacherSelect(w, r.Form.GetMany("TeacherID"))
			w.AppendString(`<br>`)

			DisplayInputLabel(w, GL, "Group")
			DisplayGroupSelect(w, r.Form.GetMany("GroupID"))
			w.AppendString(`<br>`)
		}
		DisplayFormEnd(w, GL, action)
	}
	DisplayBodyEnd(w)

	DisplayHTMLEnd(w)
	return nil
}

func SubjectCreatePageHandler(w *http.Response, r *http.Request) error {
	return SubjectCreateEditPageHandler(w, r, APIPrefix+"/subject/create", "Create subject", "Create")
}

func SubjectEditPageHandler(w *http.Response, r *http.Request) error {
	return SubjectCreateEditPageHandler(w, r, APIPrefix+"/subject/edit", "Edit subject", "Save")
}

func SubjectLessonsVerify(l Language, subject *Subject) error {
	var lesson Lesson

	if len(subject.Lessons) == 0 {
		return http.BadRequest(Ls(l, "create at least one lesson"))
	}
	for li := 0; li < len(subject.Lessons); li++ {
		if err := GetLessonByID(subject.Lessons[li], &lesson); err != nil {
			return http.ServerError(err)
		}
		if lesson.Flags == LessonDraft {
			return http.BadRequest(Ls(l, "lesson %d is a draft"), li+1)
		}
	}
	return nil
}

func SubjectLessonsMainPageHandler(w *http.Response, r *http.Request, subject *Subject) error {
	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>`)
	w.AppendString(Ls(GL, "Edit subject lessons"))
	w.AppendString(`</title></head>`)

	w.AppendString(`<body>`)

	w.AppendString(`<h1>`)
	w.AppendString(Ls(GL, "Subject"))
	w.AppendString(`</h1>`)
	w.AppendString(`<h2>`)
	w.AppendString(Ls(GL, "Lessons"))
	w.AppendString(`</h2>`)

	DisplayErrorMessage(w, GL, r.Form.Get("Error"))

	w.AppendString(`<form method="POST" action="`)
	w.WriteString(r.URL.Path)
	w.AppendString(`">`)

	DisplayHiddenID(w, "ID", subject.ID)
	DisplayHiddenString(w, "CurrentPage", "Main")

	DisplayLessonsEditableList(w, GL, subject.Lessons)

	DisplaySubmit(w, GL, "NextPage", "Add lesson", true)
	w.AppendString(`<br><br>`)

	DisplaySubmit(w, GL, "NextPage", "Save", true)
	w.AppendString(`</form>`)

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func SubjectLessonsHandleCommand(w *http.Response, l Language, r *http.Request, subject *Subject, currentPage, k, command string) error {
	var lesson Lesson

	pindex, spindex, _, _, err := GetIndicies(k[len("Command"):])
	if err != nil {
		return http.ClientError(err)
	}

	switch currentPage {
	default:
		return LessonAddHandleCommand(w, l, r, subject.Lessons, currentPage, k, command)
	case "Main":
		switch command {
		case Ls(l, "Delete"):
			subject.Lessons = RemoveAtIndex(subject.Lessons, pindex)
		case Ls(l, "Edit"):
			if (pindex < 0) || (pindex >= len(subject.Lessons)) {
				return http.ClientError(nil)
			}
			if err := GetLessonByID(subject.Lessons[pindex], &lesson); err != nil {
				return http.ServerError(err)
			}
			lesson.Flags = LessonDraft
			if err := SaveLesson(&lesson); err != nil {
				return http.ServerError(err)
			}

			r.Form.Set("LessonIndex", spindex)
			return LessonAddPageHandler(w, r, &lesson)
		case "↑", "^|":
			MoveUp(subject.Lessons, pindex)
		case "↓", "|v":
			MoveDown(subject.Lessons, pindex)
		}

		return SubjectLessonsMainPageHandler(w, r, subject)
	}
}

func SubjectLessonsPageHandler(w *http.Response, r *http.Request) error {
	var subject Subject
	var lesson Lesson
	var user User

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}
	if err := GetUserByID(session.ID, &user); err != nil {
		return http.ServerError(err)
	}

	if err := r.ParseForm(); err != nil {
		return http.ClientError(err)
	}

	currentPage := r.Form.Get("CurrentPage")
	nextPage := r.Form.Get("NextPage")

	subjectID, err := r.Form.GetID("ID")
	if err != nil {
		return http.ClientError(err)
	}
	if err := GetSubjectByID(subjectID, &subject); err != nil {
		if err == database.NotFound {
			return http.NotFound(Ls(GL, "subject with this ID does not exist"))
		}
		return http.ServerError(err)
	}
	defer SaveSubject(&subject)

	if (session.ID != AdminID) && (session.ID != subject.TeacherID) {
		return http.ForbiddenError
	}

	switch r.Form.Get("Action") {
	case Ls(GL, "create from"):
		var course Course

		courseID, err := r.Form.GetID("CourseID")
		if err != nil {
			return http.ClientError(err)
		}
		if !UserOwnsCourse(&user, courseID) {
			return http.ForbiddenError
		}
		if err := GetCourseByID(courseID, &course); err != nil {
			if err == database.NotFound {
				return http.NotFound(Ls(GL, "course with this ID does not exist"))
			}
			return http.ServerError(err)
		}
		if course.Flags != CourseActive {
			return http.ClientError(nil)
		}

		LessonsDeepCopy(&subject.Lessons, course.Lessons, subject.ID, ContainerTypeSubject)
	case Ls(GL, "give as is"):
		var course Course

		courseID, err := r.Form.GetID("CourseID")
		if err != nil {
			return http.ClientError(err)
		}
		if !UserOwnsCourse(&user, courseID) {
			return http.ForbiddenError
		}
		if err := GetCourseByID(courseID, &course); err != nil {
			if err == database.NotFound {
				return http.NotFound(Ls(GL, "course with this ID does not exist"))
			}
			return http.ServerError(err)
		}
		if course.Flags != CourseActive {
			return http.ClientError(nil)
		}

		LessonsDeepCopy(&subject.Lessons, course.Lessons, subject.ID, ContainerTypeSubject)

		w.RedirectID("/subject/", subjectID, http.StatusSeeOther)
		return nil
	}

	for i := 0; i < len(r.Form); i++ {
		k := r.Form[i].Key
		if len(r.Form[i].Values) == 0 {
			continue
		}
		v := r.Form[i].Values[0]

		/* 'command' is button, which modifies content of a current page. */
		if strings.StartsWith(k, "Command") {
			/* NOTE(anton2920): after command is executed, function must return. */
			return SubjectLessonsHandleCommand(w, GL, r, &subject, currentPage, k, v)
		}
	}

	/* 'currentPage' is the page to save/check before leaving it. */
	switch currentPage {
	case "Lesson":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), len(subject.Lessons))
		if err != nil {
			return http.ClientError(err)
		}
		if err := GetLessonByID(subject.Lessons[li], &lesson); err != nil {
			return http.ServerError(err)
		}
		defer SaveLesson(&lesson)

		LessonFillFromRequest(r.Form, &lesson)
	case "Test":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), len(subject.Lessons))
		if err != nil {
			return http.ClientError(err)
		}
		if err := GetLessonByID(subject.Lessons[li], &lesson); err != nil {
			return http.ServerError(err)
		}
		defer SaveLesson(&lesson)

		si, err := GetValidIndex(r.Form.Get("StepIndex"), len(lesson.Steps))
		if err != nil {
			return http.ClientError(err)
		}
		test, err := Step2Test(&lesson.Steps[si])
		if err != nil {
			return http.ClientError(err)
		}

		if err := LessonTestFillFromRequest(r.Form, test); err != nil {
			return WritePageEx(w, r, LessonAddTestPageHandler, test, err)
		}
		if err := LessonTestVerify(GL, test); err != nil {
			return WritePageEx(w, r, LessonAddTestPageHandler, test, err)
		}
	case "Programming":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), len(subject.Lessons))
		if err != nil {
			return http.ClientError(err)
		}
		if err := GetLessonByID(subject.Lessons[li], &lesson); err != nil {
			return http.ServerError(err)
		}
		defer SaveLesson(&lesson)

		si, err := GetValidIndex(r.Form.Get("StepIndex"), len(lesson.Steps))
		if err != nil {
			return http.ClientError(err)
		}
		task, err := Step2Programming(&lesson.Steps[si])
		if err != nil {
			return http.ClientError(err)
		}

		if err := LessonProgrammingFillFromRequest(r.Form, task); err != nil {
			return WritePageEx(w, r, LessonAddProgrammingPageHandler, task, err)
		}
		if err := LessonProgrammingVerify(task); err != nil {
			return WritePageEx(w, r, LessonAddProgrammingPageHandler, task, err)
		}
	}

	switch nextPage {
	default:
		return SubjectLessonsMainPageHandler(w, r, &subject)
	case Ls(GL, "Next"):
		if err := LessonVerify(GL, &lesson); err != nil {
			return WritePageEx(w, r, LessonAddPageHandler, &lesson, err)
		}
		lesson.Flags = LessonActive
		if err := SaveLesson(&lesson); err != nil {
			return http.ServerError(err)
		}

		return SubjectLessonsMainPageHandler(w, r, &subject)
	case Ls(GL, "Add lesson"):
		lesson.Flags = LessonDraft
		lesson.ContainerID = subject.ID
		lesson.ContainerType = ContainerTypeSubject

		if err := CreateLesson(&lesson); err != nil {
			return http.ServerError(err)
		}

		subject.Lessons = append(subject.Lessons, lesson.ID)
		r.Form.SetInt("LessonIndex", len(subject.Lessons)-1)

		return LessonAddPageHandler(w, r, &lesson)
	case Ls(GL, "Continue"):
		si, err := GetValidIndex(r.Form.Get("StepIndex"), len(lesson.Steps))
		if err != nil {
			return http.ClientError(err)
		}
		step := &lesson.Steps[si]
		step.Draft = false

		return LessonAddPageHandler(w, r, &lesson)
	case Ls(GL, "Add test"):
		lesson.Flags = LessonDraft

		lesson.Steps = append(lesson.Steps, Step{StepCommon: StepCommon{Type: StepTypeTest, Draft: true}})
		test, _ := Step2Test(&lesson.Steps[len(lesson.Steps)-1])

		r.Form.SetInt("StepIndex", len(lesson.Steps)-1)
		return LessonAddTestPageHandler(w, r, test)
	case Ls(GL, "Add programming task"):
		lesson.Flags = LessonDraft

		lesson.Steps = append(lesson.Steps, Step{StepCommon: StepCommon{Type: StepTypeProgramming, Draft: true}})
		task, _ := Step2Programming(&lesson.Steps[len(lesson.Steps)-1])

		r.Form.SetInt("StepIndex", len(lesson.Steps)-1)
		return LessonAddProgrammingPageHandler(w, r, task)
	case Ls(GL, "Save"):
		if err := SubjectLessonsVerify(GL, &subject); err != nil {
			return WritePageEx(w, r, SubjectLessonsMainPageHandler, &subject, err)
		}

		w.RedirectID("/subject/", subjectID, http.StatusSeeOther)
		return nil
	}
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
		return WritePage(w, r, SubjectCreatePageHandler, http.BadRequest(Ls(GL, "subject name length must be between %d and %d characters long"), MinSubjectNameLen, MaxSubjectNameLen))
	}

	nextUserID, err := database.GetNextID(UsersDB)
	if err != nil {
		return http.ServerError(err)
	}
	teacherID, err := GetValidID(r.Form.Get("TeacherID"), nextUserID)
	if err != nil {
		return http.ClientError(err)
	}

	nextGroupID, err := database.GetNextID(GroupsDB)
	if err != nil {
		return http.ServerError(err)
	}
	groupID, err := GetValidID(r.Form.Get("GroupID"), nextGroupID)
	if err != nil {
		return http.ClientError(err)
	}

	var subject Subject
	subject.Name = name
	subject.TeacherID = teacherID
	subject.GroupID = groupID
	subject.CreatedOn = time.Now().Unix()

	if err := CreateSubject(&subject); err != nil {
		return http.ServerError(err)
	}

	w.Redirect("/", http.StatusSeeOther)
	return nil
}

func SubjectDeleteHandler(w *http.Response, r *http.Request) error {
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

	subjectID, err := r.Form.GetID("ID")
	if err != nil {
		return http.ClientError(err)
	}
	if err := GetSubjectByID(subjectID, &subject); err != nil {
		if err == database.NotFound {
			return http.NotFound(Ls(GL, "subject with this ID does not exist"))
		}
		return http.ServerError(err)
	}

	if err := DeleteSubjectByID(subjectID); err != nil {
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

	subjectID, err := r.Form.GetID("ID")
	if err != nil {
		return http.ClientError(err)
	}
	if err := GetSubjectByID(subjectID, &subject); err != nil {
		if err == database.NotFound {
			return http.NotFound("subject with this ID does not exist")
		}
		return http.ServerError(err)
	}

	name := r.Form.Get("Name")
	if !strings.LengthInRange(name, MinSubjectNameLen, MaxSubjectNameLen) {
		return WritePage(w, r, SubjectEditPageHandler, http.BadRequest(Ls(GL, "subject name length must be between %d and %d characters long"), MinSubjectNameLen, MaxSubjectNameLen))
	}

	nextUserID, err := database.GetNextID(UsersDB)
	if err != nil {
		return http.ServerError(err)
	}
	teacherID, err := GetValidID(r.Form.Get("TeacherID"), nextUserID)
	if err != nil {
		return http.ClientError(err)
	}

	nextGroupID, err := database.GetNextID(GroupsDB)
	if err != nil {
		return http.ServerError(err)
	}
	groupID, err := GetValidID(r.Form.Get("GroupID"), nextGroupID)
	if err != nil {
		return http.ClientError(err)
	}

	subject.Name = name
	subject.TeacherID = teacherID
	subject.GroupID = groupID

	if err := SaveSubject(&subject); err != nil {
		return http.ServerError(err)
	}

	w.RedirectID("/subject/", subjectID, http.StatusSeeOther)
	return nil
}
