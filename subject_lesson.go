package main

import (
	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/strings"
)

func SubjectLessonPageHandler(w *http.Response, r *http.Request) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}

	if err := r.ParseForm(); err != nil {
		return http.ClientError(err)
	}

	subjectID, err := GetValidIndex(r.Form.Get("ID"), len(DB.Subjects))
	if err != nil {
		return http.ClientError(err)
	}
	subject := &DB.Subjects[subjectID]

	li, err := GetValidIndex(r.Form.Get("LessonIndex"), len(subject.Lessons))
	if err != nil {
		return http.ClientError(err)
	}
	lesson := &DB.Lessons[subject.Lessons[li]]

	who, err := WhoIsUserInSubject(session.ID, subject)
	if err != nil {
		return http.ServerError(err)
	}
	if who == SubjectUserNone {
		return http.ForbiddenError
	}

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>`)
	w.WriteHTMLString(subject.Name)
	w.AppendString(`: `)
	w.WriteHTMLString(lesson.Name)
	w.AppendString(`</title></head>`)
	w.AppendString(`<body>`)

	w.AppendString(`<h1>`)
	w.WriteHTMLString(subject.Name)
	w.AppendString(`: `)
	w.WriteHTMLString(lesson.Name)
	w.AppendString(`</h1>`)

	w.AppendString(`<h2>Theory</h2>`)
	w.AppendString(`<p>`)
	w.WriteHTMLString(lesson.Theory)
	w.AppendString(`</p>`)

	w.AppendString(`<h2>Evaluation</h2>`)

	w.AppendString(`<div style="max-width: max-content">`)
	for i := 0; i < len(lesson.Steps); i++ {
		step := &lesson.Steps[i]
		name := step.Name
		stepType := GetStepStringType(step)

		w.AppendString(`<fieldset>`)

		w.AppendString(`<legend>Step #`)
		w.WriteInt(i + 1)
		w.AppendString(`</legend>`)

		w.AppendString(`<p>Name: `)
		w.WriteHTMLString(name)
		w.AppendString(`</p>`)

		w.AppendString(`<p>Type: `)
		w.AppendString(stepType)
		w.AppendString(`</p>`)

		w.AppendString(`</fieldset>`)
		w.AppendString(`<br>`)
	}
	w.AppendString(`</div>`)

	if who == SubjectUserStudent {
		var submission *Submission
		var i int

		for i = 0; i < len(lesson.Submissions); i++ {
			if session.ID == lesson.Submissions[i].UserID {
				submission = lesson.Submissions[i]
				break
			}
		}

		if (submission != nil) && (!submission.Draft) {
			w.AppendString(`<form method="POST" action="/submission">`)
		} else {
			w.AppendString(`<form method="POST" action="/submission/new">`)
		}

		w.AppendString(`<input type="hidden" name="ID" value="`)
		w.WriteHTMLString(r.Form.Get("ID"))
		w.AppendString(`">`)

		w.AppendString(`<input type="hidden" name="LessonIndex" value="`)
		w.WriteHTMLString(r.Form.Get("LessonIndex"))
		w.AppendString(`">`)

		if submission == nil {
			w.AppendString(`<input type="submit" value="Pass">`)
		} else {
			w.AppendString(`<input type="hidden" name="SubmissionIndex" value="`)
			w.WriteInt(i)
			w.AppendString(`">`)

			if submission.Draft {
				w.AppendString(`<input type="submit" value="Edit">`)
			} else {
				w.AppendString(`<input type="submit" value="See submission">`)
			}
		}

		w.AppendString(`</form>`)
	}

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

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
			lesson := &DB.Lessons[subject.Lessons[pindex]]
			lesson.Flags = LessonDraft

			r.Form.Set("LessonIndex", spindex)
			return LessonAddPageHandler(w, r, lesson)
		case "↑", "^|":
			MoveUp(subject.Lessons, pindex)
		case "↓", "|v":
			MoveDown(subject.Lessons, pindex)
		}

		return SubjectLessonEditMainPageHandler(w, r, subject)
	}
}

func SubjectLessonEditPageHandler(w *http.Response, r *http.Request) error {
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

	subjectID, err := GetValidIndex(r.Form.Get("ID"), len(DB.Subjects))
	if err != nil {
		return http.ClientError(err)
	}
	subject := &DB.Subjects[subjectID]
	if (session.ID != AdminID) && (session.ID != subject.TeacherID) {
		return http.ForbiddenError
	}

	switch r.Form.Get("Action") {
	case "create from":
		courseID, err := GetValidIndex(r.Form.Get("CourseID"), len(DB.Courses))
		if err != nil {
			return http.ClientError(err)
		}
		if !UserOwnsCourse(&user, int32(courseID)) {
			return http.ForbiddenError
		}
		/* TODO(anton2920): check if it's still a draft. */
		course := &DB.Courses[courseID]

		LessonsDeepCopy(&subject.Lessons, course.Lessons)
	case "give as is":
		courseID, err := GetValidIndex(r.Form.Get("CourseID"), len(user.Courses))
		if err != nil {
			return http.ClientError(err)
		}
		if !UserOwnsCourse(&user, int32(courseID)) {
			return http.ForbiddenError
		}
		/* TODO(anton2920): check if it's still a draft. */
		course := &DB.Courses[courseID]

		LessonsDeepCopy(&subject.Lessons, course.Lessons)

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
			return SubjectLessonEditHandleCommand(w, r, subject, currentPage, k, v)
		}
	}

	/* 'currentPage' is the page to save/check before leaving it. */
	switch currentPage {
	case "Lesson":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), len(subject.Lessons))
		if err != nil {
			return http.ClientError(err)
		}
		lesson := &DB.Lessons[subject.Lessons[li]]

		LessonFillFromRequest(r.Form, lesson)
	case "Test":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), len(subject.Lessons))
		if err != nil {
			return http.ClientError(err)
		}
		lesson := &DB.Lessons[subject.Lessons[li]]

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
		lesson := &DB.Lessons[subject.Lessons[li]]

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
		return SubjectLessonEditMainPageHandler(w, r, subject)
	case "Next":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), len(subject.Lessons))
		if err != nil {
			return http.ClientError(err)
		}
		lesson := &DB.Lessons[subject.Lessons[li]]

		if err := LessonVerify(lesson); err != nil {
			return WritePageEx(w, r, LessonAddPageHandler, lesson, err)
		}
		lesson.Flags = LessonActive

		return SubjectLessonEditMainPageHandler(w, r, subject)
	case "Add lesson":
		DB.Lessons = append(DB.Lessons, Lesson{ID: int32(len(DB.Lessons)), Flags: LessonDraft})
		lesson := &DB.Lessons[len(DB.Lessons)-1]

		subject.Lessons = append(subject.Lessons, lesson.ID)
		r.Form.SetInt("LessonIndex", int(lesson.ID))

		return LessonAddPageHandler(w, r, lesson)
	case "Continue":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), len(subject.Lessons))
		if err != nil {
			return http.ClientError(err)
		}
		lesson := &DB.Lessons[subject.Lessons[li]]

		si, err := GetValidIndex(r.Form.Get("StepIndex"), len(lesson.Steps))
		if err != nil {
			return http.ClientError(err)
		}
		lesson.Steps[si].Draft = false

		return LessonAddPageHandler(w, r, lesson)
	case "Add test":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), len(subject.Lessons))
		if err != nil {
			return http.ClientError(err)
		}
		lesson := &DB.Lessons[subject.Lessons[li]]
		lesson.Flags = LessonDraft

		lesson.Steps = append(lesson.Steps, Step{StepCommon: StepCommon{Type: StepTypeTest, Draft: true}})
		test, _ := Step2Test(&lesson.Steps[len(lesson.Steps)-1])

		r.Form.SetInt("StepIndex", len(lesson.Steps)-1)
		return LessonAddTestPageHandler(w, r, test)
	case "Add programming task":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), len(subject.Lessons))
		if err != nil {
			return http.ClientError(err)
		}
		lesson := &DB.Lessons[subject.Lessons[li]]
		lesson.Flags = LessonDraft

		lesson.Steps = append(lesson.Steps, Step{StepCommon: StepCommon{Type: StepTypeProgramming, Draft: true}})
		task, _ := Step2Programming(&lesson.Steps[len(lesson.Steps)-1])

		r.Form.SetInt("StepIndex", len(lesson.Steps)-1)
		return LessonAddProgrammingPageHandler(w, r, task)
	case "Save":
		if len(subject.Lessons) == 0 {
			return WritePageEx(w, r, SubjectLessonEditMainPageHandler, subject, http.BadRequest("create at least one lesson"))
		}
		for li := 0; li < len(subject.Lessons); li++ {
			lesson := &DB.Lessons[subject.Lessons[li]]
			if lesson.Flags == LessonDraft {
				return WritePageEx(w, r, SubjectLessonEditMainPageHandler, subject, http.BadRequest("lesson %d is a draft", li+1))
			}
		}

		w.RedirectID("/subject/", subjectID, http.StatusSeeOther)
		return nil
	}
}
