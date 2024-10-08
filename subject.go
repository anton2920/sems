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

type Subject struct {
	LessonContainer

	TeacherID database.ID
	GroupID   database.ID
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
	defer trace.End(trace.Begin(""))

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
	defer trace.End(trace.Begin(""))

	var err error

	subject.ID, err = database.IncrementNextID(SubjectsDB)
	if err != nil {
		return fmt.Errorf("failed to increment subject ID: %w", err)
	}

	return SaveSubject(subject)
}

func DBSubject2Subject(subject *Subject) {
	defer trace.End(trace.Begin(""))

	data := &subject.Data[0]

	subject.Name = database.Offset2String(subject.Name, data)

	slice := database.Offset2Slice(*(*[]byte)(unsafe.Pointer(&subject.Lessons)), data)
	subject.Lessons = *(*[]database.ID)(unsafe.Pointer(&slice))
}

func GetSubjectByID(id database.ID, subject *Subject) error {
	defer trace.End(trace.Begin(""))

	if err := database.Read(SubjectsDB, id, unsafe.Pointer(subject), int(unsafe.Sizeof(*subject))); err != nil {
		return err
	}

	DBSubject2Subject(subject)
	return nil
}

func GetSubjects(pos *int64, subjects []Subject) (int, error) {
	defer trace.End(trace.Begin(""))

	n, err := database.ReadMany(SubjectsDB, pos, *(*[]byte)(unsafe.Pointer(&subjects)), int(unsafe.Sizeof(subjects[0])))
	if err != nil {
		return 0, err
	}

	for i := 0; i < n; i++ {
		DBSubject2Subject(&subjects[i])
	}
	return n, nil
}

func DeleteSubjectByID(id database.ID) error {
	defer trace.End(trace.Begin(""))

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
	defer trace.End(trace.Begin(""))

	var subjectDB Subject
	var n int

	subjectDB.ID = subject.ID
	subjectDB.Flags = subject.Flags
	subjectDB.TeacherID = subject.TeacherID
	subjectDB.GroupID = subject.GroupID

	/* TODO(anton2920): save up to a sizeof(subject.Data). */
	data := unsafe.Slice(&subjectDB.Data[0], len(subjectDB.Data))
	n += database.String2DBString(&subjectDB.Name, subject.Name, data, n)
	n += database.Slice2DBSlice((*[]byte)(unsafe.Pointer(&subjectDB.Lessons)), *(*[]byte)(unsafe.Pointer(&subject.Lessons)), int(unsafe.Sizeof(subject.Lessons[0])), int(unsafe.Alignof(subject.Lessons[0])), data, n)

	subjectDB.CreatedOn = subject.CreatedOn

	return database.Write(SubjectsDB, subjectDB.ID, unsafe.Pointer(&subjectDB), int(unsafe.Sizeof(subjectDB)))
}

func DisplaySubjectCoursesSelect(w *http.Response, l Language, subject *Subject, teacher *User) {
	w.WriteString(`<form method="POST" action="/subject/lessons">`)
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
					w.WriteString(`<label>`)
					w.WriteString(Ls(l, "Courses"))
					w.WriteString(`: `)
					w.WriteString(`<select name="CourseID">`)
					displayed = true
				}
				w.WriteString(`<option value="`)
				w.WriteInt(i)
				w.WriteString(`">`)
				w.WriteHTMLString(course.Name)
				w.WriteString(`</option>`)
			}
		}
		if displayed {
			w.WriteString(`</select>`)
			w.WriteString(`</label> `)

			DisplayButton(w, l, "Action", "create from")
			w.WriteString(`, `)
			DisplayButton(w, l, "Action", "give as is")
			w.WriteString(` `)
			w.WriteString(Ls(l, "or"))
			w.WriteString(` `)
		}
		DisplayButton(w, l, "Action", "create new from scratch")
	} else {
		DisplayButton(w, l, "", "Edit")
	}

	w.WriteString(`</form>`)
}

func DisplaySubjectTitle(w *http.Response, l Language, subject *Subject, teacher *User) {
	w.WriteHTMLString(subject.Name)
	w.WriteString(` `)
	w.WriteString(Ls(l, "with"))
	w.WriteString(` `)
	w.WriteHTMLString(teacher.LastName)
	w.WriteString(` `)
	w.WriteHTMLString(teacher.FirstName)
	w.WriteString(` (ID: `)
	w.WriteID(subject.ID)
	w.WriteString(`)`)
	if subject.Flags == SubjectDeleted {
		w.WriteString(` [deleted]`)
	}
}

func DisplaySubjectLink(w *http.Response, l Language, subject *Subject) {
	var teacher User
	if err := GetUserByID(subject.TeacherID, &teacher); err != nil {
		/* TODO(anton2920): report error. */
		return
	}

	w.WriteString(`<a href="/subject/`)
	w.WriteID(subject.ID)
	w.WriteString(`">`)
	DisplaySubjectTitle(w, l, subject, &teacher)
	w.WriteString(`</a>`)
}

func SubjectsPageHandler(w *http.Response, r *http.Request) error {
	defer trace.End(trace.Begin(""))

	const width = WidthLarge

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}

	DisplayHTMLStart(w)

	DisplayHeadStart(w)
	{
		w.WriteString(`<title>`)
		w.WriteString(Ls(GL, "Subjects"))
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
			DisplayCrumbsItem(w, GL, "Subjects")
		}
		DisplayCrumbsEnd(w)

		DisplayPageStart(w, width)
		{
			w.WriteString(`<h2 class="text-center">`)
			w.WriteString(Ls(GL, "Subjects"))
			w.WriteString(`</h2>`)
			w.WriteString(`<br>`)

			DisplayTableStart(w, GL, []string{"ID", "Name", "Teacher", "Group", "Lessons", "Status"})
			{
				subjects := make([]Subject, 32)
				var pos int64

				for {
					n, err := GetSubjects(&pos, subjects)
					if err != nil {
						return http.ServerError(err)
					}
					if n == 0 {
						break
					}

					for i := 0; i < n; i++ {
						subject := &subjects[i]
						who, err := WhoIsUserInSubject(session.ID, subject)
						if err != nil {
							return http.ServerError(err)
						}
						if who == SubjectUserNone {
							continue
						}

						var teacher User
						if err := GetUserByID(subject.TeacherID, &teacher); err != nil {
							return http.ServerError(err)
						}

						var group Group
						if err := GetGroupByID(subject.GroupID, &group); err != nil {
							return http.ServerError(err)
						}

						DisplayTableRowLinkIDStart(w, "/subject", subject.ID)

						DisplayTableItemString(w, subject.Name)

						DisplayTableItemStart(w)
						DisplayUserTitle(w, GL, &teacher)
						DisplayTableItemEnd(w)

						DisplayTableItemStart(w)
						DisplayGroupTitle(w, GL, &group)
						DisplayTableItemEnd(w)

						DisplayTableItemInt(w, len(subject.Lessons))
						DisplayTableItemFlags(w, GL, subject.Flags)

						DisplayTableRowEnd(w)
					}
				}
			}
			DisplayTableEnd(w)

			if session.ID == AdminID {
				w.WriteString(`<br>`)
				w.WriteString(`<form method="POST" action="/subject/create">`)
				DisplaySubmit(w, GL, "", "Create subject", true)
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

func SubjectPageHandler(w *http.Response, r *http.Request) error {
	defer trace.End(trace.Begin(""))

	const width = WidthLarge

	var subject Subject

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}

	id, err := GetIDFromURL(GL, r.URL, "/subject/")
	if err != nil {
		return err
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
		w.WriteString(`<title>`)
		DisplaySubjectTitle(w, GL, &subject, &teacher)
		w.WriteString(`</title>`)
	}
	DisplayHeadEnd(w)

	DisplayBodyStart(w)
	{
		DisplayHeader(w, GL)
		DisplaySidebarWithLessons(w, GL, session, subject.Lessons)

		DisplayMainStart(w)

		DisplayCrumbsStart(w, width)
		{
			DisplayCrumbsItemRaw(w, subject.Name)
		}
		DisplayCrumbsEnd(w)

		DisplayPageStart(w, width)
		{
			w.WriteString(`<h2>`)
			DisplaySubjectTitle(w, GL, &subject, &teacher)
			w.WriteString(`</h2>`)
			w.WriteString(`<br>`)

			w.WriteString(`<p>`)
			w.WriteString(Ls(GL, "Teacher"))
			w.WriteString(`: `)
			DisplayUserLink(w, GL, &teacher)
			w.WriteString(`</p>`)

			w.WriteString(`<p>`)
			w.WriteString(Ls(GL, "Group"))
			w.WriteString(`: `)
			DisplayGroupLink(w, GL, &group)
			w.WriteString(`</p>`)

			w.WriteString(`<p>`)
			w.WriteString(Ls(GL, "Created on"))
			w.WriteString(`: `)
			DisplayFormattedTime(w, subject.CreatedOn)
			w.WriteString(`</p>`)

			if session.ID == AdminID {
				w.WriteString(`<div>`)
				w.WriteString(`<form style="display:inline" method="POST" action="/subject/edit">`)
				DisplayHiddenID(w, "ID", subject.ID)
				DisplayHiddenID(w, "TeacherID", subject.TeacherID)
				DisplayHiddenID(w, "GroupID", subject.GroupID)
				DisplayHiddenString(w, "Name", subject.Name)
				DisplayButton(w, GL, "", "Edit")
				w.WriteString(`</form>`)

				w.WriteString(` <form style="display:inline" method="POST" action="/api/subject/delete">`)
				DisplayHiddenID(w, "ID", subject.ID)
				DisplayButton(w, GL, "", "Delete")
				w.WriteString(`</form>`)
				w.WriteString(`</div>`)
				w.WriteString(`<br>`)
			}

			if (len(subject.Lessons) != 0) || (who != SubjectUserStudent) {
				w.WriteString(`<br>`)
				w.WriteString(`<h3>`)
				w.WriteString(Ls(GL, "Lessons"))
				w.WriteString(`</h3>`)
				DisplayLessons(w, GL, subject.Lessons)
			}

			if (session.ID == AdminID) || (session.ID == subject.TeacherID) {
				DisplaySubjectCoursesSelect(w, GL, &subject, &teacher)
			}
		}
		DisplayPageEnd(w)
		DisplayMainEnd(w)
	}
	DisplayBodyEnd(w)

	DisplayHTMLEnd(w)
	return nil
}

func DisplayTeacherSelect(w *http.Response, ids []string) {
	users := make([]User, 32)
	var pos int64

	w.WriteString(`<select class="form-select" name="TeacherID">`)
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

			w.WriteString(`<option value="`)
			w.WriteID(user.ID)
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

func DisplayGroupSelect(w *http.Response, ids []string) {
	groups := make([]Group, 32)
	var pos int64

	w.WriteString(`<select class="form-select" name="GroupID">`)
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

			w.WriteString(`<option value="`)
			w.WriteID(group.ID)
			w.WriteString(`"`)
			for j := 0; j < len(ids); j++ {
				id, err := GetValidID(ids[j], database.MaxValidID)
				if err != nil {
					continue
				}
				if id == group.ID {
					w.WriteString(` selected`)
				}
			}
			w.WriteString(`>`)
			w.WriteHTMLString(group.Name)
			w.WriteString(`</option>`)
		}
	}
	w.WriteString(`</select>`)
}

func SubjectCreateEditPageHandler(w *http.Response, r *http.Request, session *Session, subject *Subject, endpoint string, title string, action string, err error) error {
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
			case "Create subject":
				DisplayCrumbsLink(w, GL, "/subjects", "Subjects")
			case "Edit subject":
				DisplayCrumbsLinkID(w, "/subject", subject.ID, subject.Name)
			}
			DisplayCrumbsItem(w, GL, title)
		}
		DisplayCrumbsEnd(w)

		DisplayFormPageStart(w, r, GL, width, title, endpoint, err)
		{
			DisplayLabel(w, GL, "Name")
			DisplayConstraintInput(w, "text", MinNameLen, MaxNameLen, "Name", r.Form.Get("Name"), true)
			w.WriteString(`<br>`)

			DisplayLabel(w, GL, "Teacher")
			DisplayTeacherSelect(w, r.Form.GetMany("TeacherID"))
			w.WriteString(`<br>`)

			DisplayLabel(w, GL, "Group")
			DisplayGroupSelect(w, r.Form.GetMany("GroupID"))
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

func SubjectCreatePageHandler(w *http.Response, r *http.Request, e error) error {
	defer trace.End(trace.Begin(""))

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

	return SubjectCreateEditPageHandler(w, r, session, nil, APIPrefix+"/subject/create", "Create subject", "Create", e)
}

func SubjectEditPageHandler(w *http.Response, r *http.Request, e error) error {
	defer trace.End(trace.Begin(""))

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

	return SubjectCreateEditPageHandler(w, r, session, &subject, APIPrefix+"/subject/edit", "Edit subject", "Save", e)
}

func SubjectLessonsVerify(l Language, subject *Subject) error {
	defer trace.End(trace.Begin(""))

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

func SubjectLessonsMainPageHandler(w *http.Response, r *http.Request, session *Session, subject *Subject, err error) error {
	defer trace.End(trace.Begin(""))

	const width = WidthSmall

	DisplayHTMLStart(w)

	DisplayHeadStart(w)
	{
		w.WriteString(`<title>`)
		w.WriteString(Ls(GL, "Edit lessons"))
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
			DisplayCrumbsLinkID(w, "/subject", subject.ID, subject.Name)
			DisplayCrumbsItem(w, GL, "Edit lessons")
		}
		DisplayCrumbsEnd(w)

		DisplayFormPageStart(w, r, GL, width, "Edit lessons", r.URL.Path, err)
		{
			DisplayHiddenString(w, "CurrentPage", "Main")

			DisplayLessonsEditableList(w, GL, subject.Lessons)

			DisplayNextPage(w, GL, "Add lesson")
			w.WriteString(`<br><br>`)

			DisplaySubmit(w, GL, "NextPage", "Save", true)
		}
		DisplayFormPageEnd(w)
		DisplayMainEnd(w)
	}
	DisplayBodyEnd(w)

	DisplayHTMLEnd(w)
	return nil
}

func SubjectLessonsHandleCommand(w *http.Response, r *http.Request, l Language, session *Session, subject *Subject, currentPage, k, command string) error {
	defer trace.End(trace.Begin(""))

	var lesson Lesson

	pindex, spindex, _, _, err := GetIndicies(k[len("Command"):])
	if err != nil {
		return http.ClientError(err)
	}

	switch currentPage {
	default:
		return LessonAddHandleCommand(w, r, l, session, &subject.LessonContainer, currentPage, k, command)
	case "Main":
		switch command {
		case Ls(l, "Delete"):
			subject.Lessons = RemoveLessonAtIndex(subject.Lessons, pindex)
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
			return LessonAddPageHandler(w, r, session, &subject.LessonContainer, &lesson, nil)
		case "↑", "^|":
			MoveLessonUp(subject.Lessons, pindex)
		case "↓", "|v":
			MoveLessonDown(subject.Lessons, pindex)
		}

		return SubjectLessonsMainPageHandler(w, r, session, subject, nil)
	}
}

func SubjectLessonsPageHandler(w *http.Response, r *http.Request) error {
	defer trace.End(trace.Begin(""))

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
				return http.NotFound("course with this ID does not exist")
			}
			return http.ServerError(err)
		}
		if course.Flags != CourseActive {
			return http.ClientError(nil)
		}

		LessonsDeepCopy(&subject.Lessons, course.Lessons, subject.ID, LessonContainerSubject)
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
				return http.NotFound("course with this ID does not exist")
			}
			return http.ServerError(err)
		}
		if course.Flags != CourseActive {
			return http.ClientError(nil)
		}

		LessonsDeepCopy(&subject.Lessons, course.Lessons, subject.ID, LessonContainerSubject)

		w.RedirectID("/subject/", subjectID, http.StatusSeeOther)
		return nil
	}

	for i := 0; i < len(r.Form.Keys); i++ {
		k := r.Form.Keys[i]

		/* 'command' is button, which modifies content of a current page. */
		if strings.StartsWith(k, "Command") {
			if len(r.Form.Values[i]) == 0 {
				continue
			}
			v := r.Form.Values[i][0]

			/* NOTE(anton2920): after command is executed, function must return. */
			return SubjectLessonsHandleCommand(w, r, GL, session, &subject, currentPage, k, v)
		}
	}

	/* 'currentPage' is the page to save before leaving it. */
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
			return LessonAddTestPageHandler(w, r, session, &subject.LessonContainer, &lesson, test, err)
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
			return LessonAddProgrammingPageHandler(w, r, session, &subject.LessonContainer, &lesson, task, err)
		}
	}

	switch nextPage {
	default:
		return SubjectLessonsMainPageHandler(w, r, session, &subject, nil)
	case Ls(GL, "Back"):
		switch currentPage {
		default:
			return SubjectLessonsMainPageHandler(w, r, session, &subject, nil)
		case "Test", "Programming":
			return LessonAddPageHandler(w, r, session, &subject.LessonContainer, &lesson, nil)
		}
	case Ls(GL, "Next"):
		if err := LessonVerify(GL, &lesson); err != nil {
			return LessonAddPageHandler(w, r, session, &subject.LessonContainer, &lesson, err)
		}
		lesson.Flags = LessonActive
		if err := SaveLesson(&lesson); err != nil {
			return http.ServerError(err)
		}

		return SubjectLessonsMainPageHandler(w, r, session, &subject, nil)
	case Ls(GL, "Add lesson"):
		lesson.Flags = LessonDraft
		lesson.ContainerID = subject.ID
		lesson.ContainerType = LessonContainerSubject

		if err := CreateLesson(&lesson); err != nil {
			return http.ServerError(err)
		}

		subject.Lessons = append(subject.Lessons, lesson.ID)
		r.Form.SetInt("LessonIndex", len(subject.Lessons)-1)

		return LessonAddPageHandler(w, r, session, &subject.LessonContainer, &lesson, nil)
	case Ls(GL, "Continue"):
		si, err := GetValidIndex(r.Form.Get("StepIndex"), len(lesson.Steps))
		if err != nil {
			return http.ClientError(err)
		}
		step := &lesson.Steps[si]
		if err := LessonStepVerify(GL, step); err != nil {
			return LessonAddStepPageHandler(w, r, session, &subject.LessonContainer, &lesson, step, err)
		}
		step.Draft = false

		return LessonAddPageHandler(w, r, session, &subject.LessonContainer, &lesson, nil)
	case Ls(GL, "Add test"):
		lesson.Flags = LessonDraft

		lesson.Steps = append(lesson.Steps, Step{StepCommon: StepCommon{Type: StepTypeTest, Draft: true}})
		test, _ := Step2Test(&lesson.Steps[len(lesson.Steps)-1])

		r.Form.SetInt("StepIndex", len(lesson.Steps)-1)
		return LessonAddTestPageHandler(w, r, session, &subject.LessonContainer, &lesson, test, nil)
	case Ls(GL, "Add programming task"):
		lesson.Flags = LessonDraft

		lesson.Steps = append(lesson.Steps, Step{StepCommon: StepCommon{Type: StepTypeProgramming, Draft: true}})
		task, _ := Step2Programming(&lesson.Steps[len(lesson.Steps)-1])

		r.Form.SetInt("StepIndex", len(lesson.Steps)-1)
		return LessonAddProgrammingPageHandler(w, r, session, &subject.LessonContainer, &lesson, task, nil)
	case Ls(GL, "Save"):
		if err := SubjectLessonsVerify(GL, &subject); err != nil {
			return SubjectLessonsMainPageHandler(w, r, session, &subject, err)
		}

		w.RedirectID("/subject/", subjectID, http.StatusSeeOther)
		return nil
	}
}

func SubjectCreateHandler(w *http.Response, r *http.Request) error {
	defer trace.End(trace.Begin(""))

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
		return SubjectCreatePageHandler(w, r, http.BadRequest(Ls(GL, "subject name length must be between %d and %d characters long"), MinSubjectNameLen, MaxSubjectNameLen))
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

	w.Redirect("/subjects", http.StatusSeeOther)
	return nil
}

func SubjectDeleteHandler(w *http.Response, r *http.Request) error {
	defer trace.End(trace.Begin(""))

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

	w.Redirect("/subjects", http.StatusSeeOther)
	return nil
}

func SubjectEditHandler(w *http.Response, r *http.Request) error {
	defer trace.End(trace.Begin(""))

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
		return SubjectEditPageHandler(w, r, http.BadRequest(Ls(GL, "subject name length must be between %d and %d characters long"), MinSubjectNameLen, MaxSubjectNameLen))
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
