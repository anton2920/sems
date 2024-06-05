package main

import (
	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/strings"
)

func SubjectLessonVerify(subject *Subject) error {
	var lesson Lesson

	if len(subject.Lessons) == 0 {
		return http.BadRequest("create at least one lesson")
	}
	for li := 0; li < len(subject.Lessons); li++ {
		if err := GetLessonByID(DB2, subject.Lessons[li], &lesson); err != nil {
			return http.ServerError(err)
		}
		if lesson.Flags == LessonDraft {
			return http.BadRequest("lesson %d is a draft", li+1)
		}
	}
	return nil
}

func SubjectLessonEditMainPageHandler(w *http.Response, r *http.Request, subject *Subject) error {
	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>Edit subject lessons</title></head>`)
	w.AppendString(`<body>`)
	w.AppendString(`<h1>Subject</h1>`)
	w.AppendString(`<h2>Lessons</h2>`)

	DisplayErrorMessage(w, r.Form.Get("Error"))

	w.AppendString(`<form method="POST" action="`)
	w.WriteString(r.URL.Path)
	w.AppendString(`">`)

	w.AppendString(`<input type="hidden" name="CurrentPage" value="Main">`)

	w.AppendString(`<input type="hidden" name="ID" value="`)
	w.WriteHTMLString(r.Form.Get("ID"))
	w.AppendString(`">`)

	DisplayLessonsEditableList(w, subject.Lessons)

	w.AppendString(`<input type="submit" name="NextPage" value="Add lesson">`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<input type="submit" name="NextPage" value="Save">`)

	w.AppendString(`</form>`)

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func SubjectLessonEditHandleCommand(w *http.Response, r *http.Request, subject *Subject, currentPage, k, command string) error {
	var lesson Lesson

	pindex, spindex, _, _, err := GetIndicies(k[len("Command"):])
	if err != nil {
		return http.ClientError(err)
	}

	switch currentPage {
	default:
		return LessonAddHandleCommand(w, r, subject.Lessons, currentPage, k, command)
	case "Main":
		switch command {
		case "Delete":
			subject.Lessons = RemoveAtIndex(subject.Lessons, pindex)
		case "Edit":
			if (pindex < 0) || (pindex >= len(subject.Lessons)) {
				return http.ClientError(nil)
			}
			if err := GetLessonByID(DB2, subject.Lessons[pindex], &lesson); err != nil {
				return http.ServerError(err)
			}
			lesson.Flags = LessonDraft
			if err := SaveLesson(DB2, &lesson); err != nil {
				return http.ServerError(err)
			}

			r.Form.Set("LessonIndex", spindex)
			return LessonAddPageHandler(w, r, &lesson)
		case "↑", "^|":
			MoveUp(subject.Lessons, pindex)
		case "↓", "|v":
			MoveDown(subject.Lessons, pindex)
		}

		return SubjectLessonEditMainPageHandler(w, r, subject)
	}
}

func SubjectLessonEditPageHandler(w *http.Response, r *http.Request) error {
	var subject Subject
	var lesson Lesson
	var user User

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}
	if err := GetUserByID(DB2, session.ID, &user); err != nil {
		return http.ServerError(err)
	}

	if err := r.ParseForm(); err != nil {
		return http.ClientError(err)
	}

	currentPage := r.Form.Get("CurrentPage")
	nextPage := r.Form.Get("NextPage")

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
	defer SaveSubject(DB2, &subject)

	if (session.ID != AdminID) && (session.ID != subject.TeacherID) {
		return http.ForbiddenError
	}

	switch r.Form.Get("Action") {
	case "create from":
		var course Course

		courseID, err := r.Form.GetInt("CourseID")
		if err != nil {
			return http.ClientError(err)
		}
		if !UserOwnsCourse(&user, int32(courseID)) {
			return http.ForbiddenError
		}
		if err := GetCourseByID(DB2, int32(courseID), &course); err != nil {
			if err == DBNotFound {
				return http.NotFound("course with this ID does not exist")
			}
			return http.ServerError(err)
		}

		/* TODO(anton2920): check if it's still a draft. */
		LessonsDeepCopy(&subject.Lessons, course.Lessons, subject.ID, ContainerTypeSubject)
	case "give as is":
		var course Course

		courseID, err := r.Form.GetInt("CourseID")
		if err != nil {
			return http.ClientError(err)
		}
		if !UserOwnsCourse(&user, int32(courseID)) {
			return http.ForbiddenError
		}
		if err := GetCourseByID(DB2, int32(courseID), &course); err != nil {
			if err == DBNotFound {
				return http.NotFound("course with this ID does not exist")
			}
			return http.ServerError(err)
		}

		/* TODO(anton2920): check if it's still a draft. */
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
			return SubjectLessonEditHandleCommand(w, r, &subject, currentPage, k, v)
		}
	}

	/* 'currentPage' is the page to save/check before leaving it. */
	switch currentPage {
	case "Lesson":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), len(subject.Lessons))
		if err != nil {
			return http.ClientError(err)
		}
		if err := GetLessonByID(DB2, subject.Lessons[li], &lesson); err != nil {
			return http.ServerError(err)
		}
		defer SaveLesson(DB2, &lesson)

		LessonFillFromRequest(r.Form, &lesson)
	case "Test":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), len(subject.Lessons))
		if err != nil {
			return http.ClientError(err)
		}
		if err := GetLessonByID(DB2, subject.Lessons[li], &lesson); err != nil {
			return http.ServerError(err)
		}
		defer SaveLesson(DB2, &lesson)

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
		if err := LessonTestVerify(test); err != nil {
			return WritePageEx(w, r, LessonAddTestPageHandler, test, err)
		}
	case "Programming":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), len(subject.Lessons))
		if err != nil {
			return http.ClientError(err)
		}
		if err := GetLessonByID(DB2, subject.Lessons[li], &lesson); err != nil {
			return http.ServerError(err)
		}
		defer SaveLesson(DB2, &lesson)

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
		return SubjectLessonEditMainPageHandler(w, r, &subject)
	case "Next":
		if err := LessonVerify(&lesson); err != nil {
			return WritePageEx(w, r, LessonAddPageHandler, &lesson, err)
		}
		lesson.Flags = LessonActive

		return SubjectLessonEditMainPageHandler(w, r, &subject)
	case "Add lesson":
		lesson.Flags = LessonDraft
		lesson.ContainerID = subject.ID
		lesson.ContainerType = ContainerTypeSubject

		if err := CreateLesson(DB2, &lesson); err != nil {
			return http.ServerError(err)
		}

		subject.Lessons = append(subject.Lessons, lesson.ID)
		r.Form.SetInt("LessonIndex", len(subject.Lessons)-1)

		return LessonAddPageHandler(w, r, &lesson)
	case "Continue":
		si, err := GetValidIndex(r.Form.Get("StepIndex"), len(lesson.Steps))
		if err != nil {
			return http.ClientError(err)
		}
		step := &lesson.Steps[si]
		step.Draft = false

		return LessonAddPageHandler(w, r, &lesson)
	case "Add test":
		lesson.Flags = LessonDraft

		lesson.Steps = append(lesson.Steps, Step{StepCommon: StepCommon{Type: StepTypeTest, Draft: true}})
		test, _ := Step2Test(&lesson.Steps[len(lesson.Steps)-1])

		r.Form.SetInt("StepIndex", len(lesson.Steps)-1)
		return LessonAddTestPageHandler(w, r, test)
	case "Add programming task":
		lesson.Flags = LessonDraft

		lesson.Steps = append(lesson.Steps, Step{StepCommon: StepCommon{Type: StepTypeProgramming, Draft: true}})
		task, _ := Step2Programming(&lesson.Steps[len(lesson.Steps)-1])

		r.Form.SetInt("StepIndex", len(lesson.Steps)-1)
		return LessonAddProgrammingPageHandler(w, r, task)
	case "Save":
		if err := SubjectLessonVerify(&subject); err != nil {
			return WritePageEx(w, r, SubjectLessonEditMainPageHandler, &subject, err)
		}

		w.RedirectID("/subject/", subjectID, http.StatusSeeOther)
		return nil
	}
}
