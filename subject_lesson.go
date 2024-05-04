package main

func SubjectLessonPageHandler(w *HTTPResponse, r *HTTPRequest) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}

	if err := r.ParseForm(); err != nil {
		return ClientError(err)
	}

	subjectID, err := GetValidIndex(r.Form.Get("ID"), DB.Subjects)
	if err != nil {
		return ClientError(err)
	}
	subject := &DB.Subjects[subjectID]

	li, err := GetValidIndex(r.Form.Get("LessonIndex"), subject.Lessons)
	if err != nil {
		return ClientError(err)
	}
	lesson := subject.Lessons[li]

	who := WhoIsUserInSubject(session.ID, subject)
	if who == SubjectUserNone {
		return ForbiddenError
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
		var name, stepType string

		step := lesson.Steps[i]
		switch step := step.(type) {
		default:
			panic("invalid step type")
		case *StepTest:
			name = step.Name
			stepType = "Test"
		case *StepProgramming:
			name = step.Name
			stepType = "Programming task"
		}

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
			if session.ID == lesson.Submissions[i].User.ID {
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

func SubjectLessonEditMainPageHandler(w *HTTPResponse, r *HTTPRequest, subject *Subject) error {
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
		LessonDisplayTheory(w, lesson.Theory)
		w.AppendString(`</p>`)

		w.AppendString(`<input type="submit" name="Command`)
		w.WriteInt(i)
		w.AppendString(`" value="Edit" formnovalidate>`)
		w.AppendString("\r\n")
		w.AppendString(`<input type="submit" name="Command`)
		w.WriteInt(i)
		w.AppendString(`" value="Delete" formnovalidate>`)
		if len(subject.Lessons) > 1 {
			if i > 0 {
				w.AppendString("\r\n")
				w.AppendString(`<input type="submit" name="Command`)
				w.WriteInt(i)
				w.AppendString(`" value="↑" formnovalidate>`)
			}
			if i < len(subject.Lessons)-1 {
				w.AppendString("\r\n")
				w.AppendString(`<input type="submit" name="Command`)
				w.WriteInt(i)
				w.AppendString(`" value="↓" formnovalidate>`)
			}
		}

		w.AppendString(`</fieldset>`)
		w.AppendString(`<br>`)
	}

	w.AppendString(`<input type="submit" name="NextPage" value="Add lesson">`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<input type="submit" name="NextPage" value="Save">`)

	w.AppendString(`</form>`)

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func SubjectLessonEditHandleCommand(w *HTTPResponse, r *HTTPRequest, subject *Subject, currentPage, k, command string) error {
	pindex, spindex, _, _, err := GetIndicies(k[len("Command"):])
	if err != nil {
		return ClientError(err)
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
				return ClientError(nil)
			}
			lesson := subject.Lessons[pindex]
			lesson.Draft = true

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

func SubjectLessonEditPageHandler(w *HTTPResponse, r *HTTPRequest) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}
	user := &DB.Users[session.ID]

	if err := r.ParseForm(); err != nil {
		return ClientError(err)
	}

	currentPage := r.Form.Get("CurrentPage")
	nextPage := r.Form.Get("NextPage")

	subjectID, err := GetValidIndex(r.Form.Get("ID"), DB.Subjects)
	if err != nil {
		return ClientError(err)
	}
	subject := &DB.Subjects[subjectID]
	if (session.ID != AdminID) && (session.ID != subject.Teacher.ID) {
		return ForbiddenError
	}

	switch r.Form.Get("Action") {
	case "create from":
		courseID, err := GetValidIndex(r.Form.Get("CourseID"), user.Courses)
		if err != nil {
			return ClientError(err)
		}

		LessonsDeepCopy(&subject.Lessons, user.Courses[courseID].Lessons)
	case "give as is":
		courseID, err := GetValidIndex(r.Form.Get("CourseID"), user.Courses)
		if err != nil {
			return ClientError(err)
		}

		LessonsDeepCopy(&subject.Lessons, user.Courses[courseID].Lessons)

		w.RedirectID("/subject/", subjectID, HTTPStatusSeeOther)
		return nil
	}

	for i := 0; i < len(r.Form); i++ {
		k := r.Form[i].Key
		v := r.Form[i].Values[0]

		/* 'command' is button, which modifies content of a current page. */
		if StringStartsWith(k, "Command") {
			/* NOTE(anton2920): after command is executed, function must return. */
			return SubjectLessonEditHandleCommand(w, r, subject, currentPage, k, v)
		}
	}

	/* 'currentPage' is the page to check before leaving it. */
	switch currentPage {
	case "Lesson":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), subject.Lessons)
		if err != nil {
			return ClientError(err)
		}
		lesson := subject.Lessons[li]

		if err := LessonAddVerifyRequest(r.Form, lesson); err != nil {
			return WritePageEx(w, r, LessonAddPageHandler, lesson, err)
		}
	case "Test":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), subject.Lessons)
		if err != nil {
			return ClientError(err)
		}
		lesson := subject.Lessons[li]

		si, err := GetValidIndex(r.Form.Get("StepIndex"), lesson.Steps)
		if err != nil {
			return ClientError(err)
		}
		test, ok := lesson.Steps[si].(*StepTest)
		if !ok {
			return ClientError(nil)
		}

		if err := LessonTestAddVerifyRequest(r.Form, test, true); err != nil {
			return WritePageEx(w, r, LessonTestAddPageHandler, test, err)
		}
	case "Programming":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), subject.Lessons)
		if err != nil {
			return ClientError(err)
		}
		lesson := subject.Lessons[li]

		si, err := GetValidIndex(r.Form.Get("StepIndex"), lesson.Steps)
		if err != nil {
			return ClientError(err)
		}
		task, ok := lesson.Steps[si].(*StepProgramming)
		if !ok {
			return ClientError(nil)
		}

		if err := LessonProgrammingAddVerifyRequest(r.Form, task, true); err != nil {
			return WritePageEx(w, r, LessonProgrammingAddPageHandler, task, err)
		}
	}

	switch nextPage {
	default:
		return SubjectLessonEditMainPageHandler(w, r, subject)
	case "Next":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), subject.Lessons)
		if err != nil {
			return ClientError(err)
		}
		lesson := subject.Lessons[li]

		for si := 0; si < len(lesson.Steps); si++ {
			switch step := lesson.Steps[si].(type) {
			case *StepTest:
				if step.Draft {
					return WritePageEx(w, r, LessonAddPageHandler, lesson, BadRequest("test %d is a draft", si+1))
				}
			case *StepProgramming:
				if step.Draft {
					return WritePageEx(w, r, LessonAddPageHandler, lesson, BadRequest("programming task %d is a draft", si+1))
				}
			}
		}
		lesson.Draft = false

		return SubjectLessonEditMainPageHandler(w, r, subject)
	case "Add lesson":
		lesson := new(Lesson)
		lesson.Draft = true
		subject.Lessons = append(subject.Lessons, lesson)

		r.Form.SetInt("LessonIndex", len(subject.Lessons)-1)
		return LessonAddPageHandler(w, r, lesson)
	case "Continue":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), subject.Lessons)
		if err != nil {
			return ClientError(err)
		}
		lesson := subject.Lessons[li]

		si, err := GetValidIndex(r.Form.Get("StepIndex"), lesson.Steps)
		if err != nil {
			return ClientError(err)
		}
		switch step := lesson.Steps[si].(type) {
		default:
			panic("invalid step type")
		case *StepTest:
			step.Draft = false
		case *StepProgramming:
			step.Draft = false
		}

		return LessonAddPageHandler(w, r, lesson)
	case "Add test":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), subject.Lessons)
		if err != nil {
			return ClientError(err)
		}
		lesson := subject.Lessons[li]
		lesson.Draft = true

		test := new(StepTest)
		test.Draft = true
		lesson.Steps = append(lesson.Steps, test)

		r.Form.SetInt("StepIndex", len(lesson.Steps)-1)
		return LessonTestAddPageHandler(w, r, test)
	case "Add programming task":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), subject.Lessons)
		if err != nil {
			return ClientError(err)
		}
		lesson := subject.Lessons[li]
		lesson.Draft = true

		task := new(StepProgramming)
		task.Draft = true
		lesson.Steps = append(lesson.Steps, task)

		r.Form.SetInt("StepIndex", len(lesson.Steps)-1)
		return LessonProgrammingAddPageHandler(w, r, task)
	case "Save":
		return SubjectLessonEditHandler(w, r)
	}
}

func SubjectLessonEditHandler(w *HTTPResponse, r *HTTPRequest) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}

	if err := r.ParseForm(); err != nil {
		return ClientError(err)
	}

	subjectID, err := GetValidIndex(r.Form.Get("ID"), DB.Subjects)
	if err != nil {
		return ClientError(err)
	}
	subject := &DB.Subjects[subjectID]
	if (session.ID != AdminID) && (session.ID != subject.Teacher.ID) {
		return WritePageEx(w, r, SubjectLessonEditMainPageHandler, subject, ForbiddenError)
	}

	if len(subject.Lessons) == 0 {
		return WritePageEx(w, r, SubjectLessonEditMainPageHandler, subject, BadRequest("create at least one lesson"))
	}
	for li := 0; li < len(subject.Lessons); li++ {
		lesson := subject.Lessons[li]
		if lesson.Draft {
			return WritePageEx(w, r, SubjectLessonEditMainPageHandler, subject, BadRequest("lesson %d is a draft", li+1))
		}
	}

	w.RedirectID("/subject/", subjectID, HTTPStatusSeeOther)
	return nil
}
