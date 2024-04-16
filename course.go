package main

import (
	"encoding/gob"
	"fmt"
	"strconv"
	"unicode/utf8"
	"unsafe"
)

type CheckType int

const (
	CheckTypeExample CheckType = iota
	CheckTypeTest
)

const (
	CheckKeyDisplay = iota
	CheckKeyInput
	CheckKeyOutput
)

type (
	Question struct {
		Name           string
		Answers        []string
		CorrectAnswers []int
	}
	Check struct {
		Input  string
		Output string
	}

	StepTest struct {
		Name      string
		Questions []Question

		/* TODO(anton2920): I don't like this. Replace with 'pointer|1'. */
		Draft bool
	}
	StepProgramming struct {
		Name        string
		Description string
		Checks      [2][]Check

		/* TODO(anton2920): I don't like this. Replace with 'pointer|1'. */
		Draft bool
	}

	Lesson struct {
		Name   string
		Theory string

		/* TODO(anton2920): using 'interface{}' so 'encoding/gob' does what it supposed to do. */
		Steps []interface{}

		Draft bool
	}

	Course struct {
		Name    string
		Lessons []*Lesson

		/* TODO(anton2920): I don't like this. Replace with 'pointer|1'. */
		Draft bool
	}
)

const (
	MinNameLen = 1
	MaxNameLen = 45

	MinTheoryLen = 1
	MaxTheoryLen = 1024

	MinStepNameLen = 1
	MaxStepNameLen = 128
	MinQuestionLen = 1
	MaxQuestionLen = 128
	MinAnswerLen   = 1
	MaxAnswerLen   = 128

	MinDescriptionLen = 1
	MaxDescriptionLen = 1024
	MinCheckLen       = 1
	MaxCheckLen       = 512
)

var CheckKeys = [2][3]string{
	CheckTypeExample: {CheckKeyDisplay: "example", CheckKeyInput: "ExampleInput", CheckKeyOutput: "ExampleOutput"},
	CheckTypeTest:    {CheckKeyDisplay: "test", CheckKeyInput: "TestInput", CheckKeyOutput: "TestOutput"},
}

func init() {
	gob.Register(&StepTest{})
	gob.Register(&StepProgramming{})
}

func CourseLessonDisplayTheory(w *HTTPResponse, theory string) {
	const maxVisibleLen = 30
	if utf8.RuneCountInString(theory) < maxVisibleLen {
		w.WriteHTMLString(theory)
	} else {
		space := FindChar(theory[maxVisibleLen:], ' ')
		if space == -1 {
			w.WriteHTMLString(theory[:maxVisibleLen])
		} else {
			w.WriteHTMLString(theory[:space+maxVisibleLen])
		}
		w.AppendString(`...`)
	}
}

func CoursePageHandler(w *HTTPResponse, r *HTTPRequest) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}
	user := &DB.Users[session.ID]

	id, err := GetIDFromURL(r.URL, "/course/")
	if err != nil {
		return err
	}
	if (id < 0) || (id > len(user.Courses)) {
		return NotFoundError
	}
	course := user.Courses[id]

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>`)
	w.WriteHTMLString(course.Name)
	w.AppendString(`</title></head>`)

	w.AppendString(`<body>`)

	w.AppendString(`<h1>`)
	w.WriteHTMLString(course.Name)
	if course.Draft {
		w.AppendString(` (draft)`)
	}
	w.AppendString(`</h1>`)

	w.AppendString(`<h2>Lessons</h2>`)
	for i := 0; i < len(course.Lessons); i++ {
		lesson := course.Lessons[i]

		buffer := make([]byte, 20)
		n := SlicePutInt(buffer, i+1)

		w.AppendString(`<fieldset>`)

		w.AppendString(`<legend>Lesson #`)
		w.Write(buffer[:n])
		if lesson.Draft {
			w.AppendString(` (draft)`)
		}
		w.AppendString(`</legend>`)

		w.AppendString(`<p>Name: `)
		w.WriteHTMLString(lesson.Name)
		w.AppendString(`</p>`)

		w.AppendString(`<p>Theory: `)
		CourseLessonDisplayTheory(w, lesson.Theory)
		w.AppendString(`</p>`)

		w.AppendString(`</fieldset>`)
		w.AppendString(`<br>`)
	}

	w.AppendString(`<div>`)

	w.AppendString(`<form style="display:inline" method="POST" action="/course/edit">`)
	w.AppendString(`<input type="hidden" name="CourseIndex" value="`)
	w.WriteHTMLString(r.URL.Path[len("/course/"):])
	w.AppendString(`">`)
	w.AppendString(`<input type="submit" value="Edit">`)
	w.AppendString(`</form>`)
	w.AppendString("\r\n")
	w.AppendString(`<form style="display:inline" method="POST" action="/api/course/delete">`)
	w.AppendString(`<input type="hidden" name="CourseIndex" value="`)
	w.WriteHTMLString(r.URL.Path[len("/course/"):])
	w.AppendString(`">`)
	w.AppendString(`<input type="submit" value="Delete">`)
	w.AppendString(`</form>`)

	w.AppendString(`</div>`)

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func CourseCreateCourseVerifyRequest(vs URLValues, course *Course) error {
	course.Name = vs.Get("Name")
	if !StringLengthInRange(course.Name, MinNameLen, MaxNameLen) {
		return NewHTTPError(HTTPStatusBadRequest, fmt.Sprintf("course name length must be between %d and %d characters long", MinNameLen, MaxNameLen))
	}

	return nil
}

func CourseCreateLessonVerifyRequest(vs URLValues, lesson *Lesson) error {
	lesson.Name = vs.Get("Name")
	if !StringLengthInRange(lesson.Name, MinNameLen, MaxNameLen) {
		return NewHTTPError(HTTPStatusBadRequest, fmt.Sprintf("lesson name length must be between %d and %d characters long", MinNameLen, MaxNameLen))
	}

	lesson.Theory = vs.Get("Theory")
	if !StringLengthInRange(lesson.Theory, MinTheoryLen, MaxTheoryLen) {
		return NewHTTPError(HTTPStatusBadRequest, fmt.Sprintf("lesson theory length must be between %d and %d characters long", MinTheoryLen, MaxTheoryLen))
	}

	return nil
}

func CourseCreateTestVerifyRequest(vs URLValues, test *StepTest, shouldCheck bool) error {
	test.Name = vs.Get("Name")
	if (shouldCheck) && (!StringLengthInRange(test.Name, MinStepNameLen, MaxStepNameLen)) {
		return NewHTTPError(HTTPStatusBadRequest, fmt.Sprintf("test name length must be between %d and %d characters long", MinStepNameLen, MaxStepNameLen))
	}

	questions := vs.GetMany("Question")
	for i := 0; i < len(questions); i++ {
		if i >= len(test.Questions) {
			test.Questions = append(test.Questions, Question{})
		}
		question := &test.Questions[i]
		question.Name = questions[i]
		if (shouldCheck) && (!StringLengthInRange(question.Name, MinQuestionLen, MaxQuestionLen)) {
			return NewHTTPError(HTTPStatusBadRequest, fmt.Sprintf("question %d: title length must be between %d and %d characters long", i+1, MinQuestionLen, MaxQuestionLen))
		}

		buffer := fmt.Appendf(make([]byte, 0, 30), "Answer%d", i)
		answers := vs.GetMany(unsafe.String(unsafe.SliceData(buffer), len(buffer)))
		for j := 0; j < len(answers); j++ {
			if j >= len(question.Answers) {
				question.Answers = append(question.Answers, "")
			}
			question.Answers[j] = answers[j]

			if (shouldCheck) && (!StringLengthInRange(question.Answers[j], MinAnswerLen, MaxAnswerLen)) {
				return NewHTTPError(HTTPStatusBadRequest, fmt.Sprintf("question %d: answer %d: length must be between %d and %d characters long", i+1, j+1, MinAnswerLen, MaxAnswerLen))
			}
		}
		question.Answers = question.Answers[:len(answers)]

		buffer = fmt.Appendf(make([]byte, 0, 30), "CorrectAnswer%d", i)
		correctAnswers := vs.GetMany(unsafe.String(unsafe.SliceData(buffer), len(buffer)))
		for j := 0; j < len(correctAnswers); j++ {
			if j >= len(question.CorrectAnswers) {
				question.CorrectAnswers = append(question.CorrectAnswers, 0)
			}

			var err error
			question.CorrectAnswers[j], err = strconv.Atoi(correctAnswers[j])
			if (err != nil) || (question.CorrectAnswers[j] < 0) || (question.CorrectAnswers[j] >= len(question.Answers)) {
				return ReloadPageError
			}
		}
		if (shouldCheck) && (len(correctAnswers) == 0) {
			return NewHTTPError(HTTPStatusBadRequest, fmt.Sprintf("question %d: select at least one correct answer", i+1))
		}
		question.CorrectAnswers = question.CorrectAnswers[:len(correctAnswers)]
	}
	test.Questions = test.Questions[:len(questions)]

	return nil
}

func CourseCreateProgrammingVerifyRequest(vs URLValues, task *StepProgramming, shouldCheck bool) error {
	task.Name = vs.Get("Name")
	if (shouldCheck) && (!StringLengthInRange(task.Name, MinStepNameLen, MaxStepNameLen)) {
		return NewHTTPError(HTTPStatusBadRequest, fmt.Sprintf("programming task name length must be between %d and %d characters long", MinStepNameLen, MaxStepNameLen))
	}

	task.Description = vs.Get("Description")
	if (shouldCheck) && (!StringLengthInRange(task.Name, MinDescriptionLen, MaxDescriptionLen)) {
		return NewHTTPError(HTTPStatusBadRequest, fmt.Sprintf("programming task description length must be between %d and %d characters long", MinDescriptionLen, MaxDescriptionLen))
	}

	for i := 0; i < len(CheckKeys); i++ {
		checks := &task.Checks[i]

		inputs := vs.GetMany(CheckKeys[i][CheckKeyInput])
		outputs := vs.GetMany(CheckKeys[i][CheckKeyOutput])

		if len(inputs) != len(outputs) {
			Infof("len(inputs) == %d, len(outputs) == %d", len(inputs), len(outputs))
			return ReloadPageError
		}

		for j := 0; j < len(inputs); j++ {
			if j >= len(*checks) {
				*checks = append(*checks, Check{})
			}
			check := &(*checks)[j]

			check.Input = inputs[j]
			check.Output = outputs[j]

			if (shouldCheck) && (!StringLengthInRange(check.Input, MinCheckLen, MaxCheckLen)) {
				return NewHTTPError(HTTPStatusBadRequest, fmt.Sprintf("%s %d: input length must be between %d and %d characters long", CheckKeys[i][CheckKeyDisplay], j+1, MinCheckLen, MaxCheckLen))
			}

			if (shouldCheck) && (!StringLengthInRange(check.Output, MinCheckLen, MaxCheckLen)) {
				return NewHTTPError(HTTPStatusBadRequest, fmt.Sprintf("%s %d: output length must be between %d and %d characters long", CheckKeys[i][CheckKeyDisplay], j+1, MinCheckLen, MaxCheckLen))
			}
		}
	}

	return nil
}

func CourseCreateTestPageHandler(w *HTTPResponse, r *HTTPRequest, test *StepTest) error {
	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>Test</title></head>`)
	w.AppendString(`<body>`)
	w.AppendString(`<h1>Lesson</h1>`)
	w.AppendString(`<h2>Test</h2>`)

	ErrorDiv(w, r.Form.Get("Error"))

	w.AppendString(`<form method="POST" action="/course/create">`)

	w.AppendString(`<input type="hidden" name="CurrentPage" value="Test">`)

	w.AppendString(`<input type="hidden" name="CourseIndex" value="`)
	w.WriteHTMLString(r.Form.Get("CourseIndex"))
	w.AppendString(`">`)

	w.AppendString(`<input type="hidden" name="LessonIndex" value="`)
	w.WriteHTMLString(r.Form.Get("LessonIndex"))
	w.AppendString(`">`)

	w.AppendString(`<input type="hidden" name="StepIndex" value="`)
	w.WriteHTMLString(r.Form.Get("StepIndex"))
	w.AppendString(`">`)

	w.AppendString(`<label>Title: `)
	w.AppendString(`<input type="text" minlength="1" maxlength="45" name="Name" value="`)
	w.WriteHTMLString(test.Name)
	w.AppendString(`" required>`)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	if len(test.Questions) == 0 {
		test.Questions = append(test.Questions, Question{})
	}
	for i := 0; i < len(test.Questions); i++ {
		question := &test.Questions[i]

		buffer := make([]byte, 20)

		w.AppendString(`<fieldset>`)
		w.AppendString(`<legend>Question #`)
		n := SlicePutInt(buffer, i+1)
		w.Write(buffer[:n])
		w.AppendString(`</legend>`)
		w.AppendString(`<label>Title: `)
		w.AppendString(`<input type="text" minlength="1" maxlength="128" name="Question" value="`)
		w.WriteHTMLString(question.Name)
		w.AppendString(`" required>`)
		w.AppendString(`</label>`)
		w.AppendString(`<br>`)

		n = SlicePutInt(buffer, i)
		si := unsafe.String(unsafe.SliceData(buffer), n)

		w.AppendString(`<p>Answers (mark the correct ones):</p>`)
		w.AppendString(`<ol>`)

		if len(question.Answers) == 0 {
			question.Answers = append(question.Answers, "")
		}
		for j := 0; j < len(question.Answers); j++ {
			buffer := make([]byte, 20)
			n := SlicePutInt(buffer, j)
			sj := unsafe.String(unsafe.SliceData(buffer), n)

			answer := question.Answers[j]

			if j > 0 {
				w.AppendString(`<br>`)
			}

			w.AppendString(`<li>`)

			w.AppendString(`<input type="checkbox" name="CorrectAnswer`)
			w.WriteString(si)
			w.AppendString(`" value="`)
			w.WriteString(sj)
			w.AppendString(`"`)
			for k := 0; k < len(question.CorrectAnswers); k++ {
				correctAnswer := question.CorrectAnswers[k]
				if j == correctAnswer {
					w.AppendString(` checked`)
					break
				}
			}
			w.AppendString(`>`)
			w.AppendString("\r\n")

			w.AppendString(`<input type="text" minlength="1" maxlength="128" name="Answer`)
			w.WriteString(si)
			w.AppendString(`" value="`)
			w.WriteHTMLString(answer)
			w.AppendString(`" required>`)

			if len(question.Answers) > 1 {
				w.AppendString("\r\n")
				w.AppendString(`<input type="submit" name="Command`)
				w.WriteString(si)
				w.AppendString(`.`)
				w.WriteString(sj)
				w.AppendString(`" value="-" formnovalidate>`)
				if j > 0 {
					w.AppendString("\r\n")
					w.AppendString(`<input type="submit" name="Command`)
					w.WriteString(si)
					w.AppendString(`.`)
					w.WriteString(sj)
					w.AppendString(`" value="↑" formnovalidate>`)
				}
				if j < len(question.Answers)-1 {
					w.AppendString("\r\n")
					w.AppendString(`<input type="submit" name="Command`)
					w.WriteString(si)
					w.AppendString(`.`)
					w.WriteString(sj)
					w.AppendString(`" value="↓" formnovalidate>`)
				}
			}

			w.AppendString(`</li>`)
		}
		w.AppendString(`</ol>`)

		w.AppendString(`<input type="submit" name="Command`)
		w.WriteString(si)
		w.AppendString(`" value="Add another answer" formnovalidate>`)

		if len(test.Questions) > 1 {
			w.AppendString(`<br><br>`)
			w.AppendString("\r\n")
			w.AppendString(`<input type="submit" name="Command`)
			w.WriteString(si)
			w.AppendString(`" value="Delete" formnovalidate>`)
			if i > 0 {
				w.AppendString("\r\n")
				w.AppendString(`<input type="submit" name="Command`)
				w.WriteString(si)
				w.AppendString(`" value="↑" formnovalidate>`)
			}
			if i < len(test.Questions)-1 {
				w.AppendString("\r\n")
				w.AppendString(`<input type="submit" name="Command`)
				w.WriteString(si)
				w.AppendString(`" value="↓" formnovalidate>`)
			}
		}

		w.AppendString(`</fieldset>`)
		w.AppendString(`<br>`)
	}

	w.AppendString(`<input type="submit" name="Command" value="Add another question" formnovalidate>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<input type="submit" name="NextPage" value="Continue">`)

	w.AppendString(`</form>`)

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func CourseCreateProgrammingDisplayChecks(w *HTTPResponse, task *StepProgramming, checkType CheckType) {
	buffer := make([]byte, 20)
	n := SlicePutInt(buffer, int(checkType))

	checks := task.Checks[checkType]
	ssindex := unsafe.String(unsafe.SliceData(buffer), n)

	w.AppendString(`<ol>`)
	for i := 0; i < len(checks); i++ {
		check := &checks[i]

		w.AppendString(`<li>`)

		w.AppendString(`<label>Input: `)

		w.AppendString(`<textarea rows="1" minlength ="1" maxlength="512" name="`)
		w.AppendString(CheckKeys[checkType][CheckKeyInput])
		w.AppendString(`">`)
		w.WriteHTMLString(check.Input)
		w.AppendString(`</textarea>`)

		w.AppendString(`</label>`)
		w.AppendString("\r\n")
		w.AppendString(`<label>output: `)

		w.AppendString(`<textarea rows="1" minlength ="1" maxlength="512" name="`)
		w.AppendString(CheckKeys[checkType][CheckKeyOutput])
		w.AppendString(`">`)
		w.WriteHTMLString(check.Output)
		w.AppendString(`</textarea>`)

		w.AppendString(`</label>`)

		buffer := make([]byte, 20)
		n := SlicePutInt(buffer, i)
		si := unsafe.String(unsafe.SliceData(buffer), n)

		w.AppendString("\r\n")
		w.AppendString(`<input type="submit" name="Command`)
		w.WriteString(si)
		w.AppendString(`.`)
		w.WriteString(ssindex)
		w.AppendString(`" value="-" formnovalidate>`)

		if len(checks) > 1 {
			if i > 0 {
				w.AppendString("\r\n")
				w.AppendString(`<input type="submit" name="Command`)
				w.WriteString(si)
				w.AppendString(`.`)
				w.WriteString(ssindex)
				w.AppendString(`" value="↑" formnovalidate>`)
			}
			if i < len(checks)-1 {
				w.AppendString("\r\n")
				w.AppendString(`<input type="submit" name="Command`)
				w.WriteString(si)
				w.AppendString(`.`)
				w.WriteString(ssindex)
				w.AppendString(`" value="↓" formnovalidate>`)
			}
		}

		w.AppendString(`</li>`)
	}
	w.AppendString(`</ol>`)
}

func CourseCreateProgrammingPageHandler(w *HTTPResponse, r *HTTPRequest, task *StepProgramming) error {
	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>Programming task</title></head>`)
	w.AppendString(`<body>`)
	w.AppendString(`<h1>Lesson</h1>`)
	w.AppendString(`<h2>Programming task</h2>`)

	ErrorDiv(w, r.Form.Get("Error"))

	w.AppendString(`<form method="POST" action="/course/create">`)

	w.AppendString(`<input type="hidden" name="CurrentPage" value="Programming">`)

	w.AppendString(`<input type="hidden" name="CourseIndex" value="`)
	w.WriteHTMLString(r.Form.Get("CourseIndex"))
	w.AppendString(`">`)

	w.AppendString(`<input type="hidden" name="LessonIndex" value="`)
	w.WriteHTMLString(r.Form.Get("LessonIndex"))
	w.AppendString(`">`)

	w.AppendString(`<input type="hidden" name="StepIndex" value="`)
	w.WriteHTMLString(r.Form.Get("StepIndex"))
	w.AppendString(`">`)

	w.AppendString(`<label>Name: `)
	w.AppendString(`<input type="text" minlength="1" maxlength="128" name="Name" value="`)
	w.WriteHTMLString(task.Name)
	w.AppendString(`" required>`)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>Description:<br>`)
	w.AppendString(`<textarea cols="80" rows="24" minlength="1" maxlength="1024" name="Description" required>`)
	w.WriteHTMLString(task.Description)
	w.AppendString(`</textarea>`)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<h3>Examples</h3>`)
	CourseCreateProgrammingDisplayChecks(w, task, CheckTypeExample)
	w.AppendString(`<input type="submit" name="Command" value="Add example">`)

	w.AppendString(`<h3>Tests</h3>`)
	CourseCreateProgrammingDisplayChecks(w, task, CheckTypeTest)
	w.AppendString(`<input type="submit" name="Command" value="Add test">`)

	w.AppendString(`<br><br>`)
	w.AppendString(`<input type="submit" name="NextPage" value="Continue">`)

	w.AppendString(`</form>`)

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	w.AppendString(`</form>`)
	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func CourseCreateLessonPageHandler(w *HTTPResponse, r *HTTPRequest, lesson *Lesson) error {
	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>Create lesson</title></head>`)
	w.AppendString(`<body>`)
	w.AppendString(`<h1>Course</h1>`)
	w.AppendString(`<h2>Lesson</h2>`)

	ErrorDiv(w, r.Form.Get("Error"))

	w.AppendString(`<form method="POST" action="/course/create">`)

	w.AppendString(`<input type="hidden" name="CurrentPage" value="Lesson">`)

	w.AppendString(`<input type="hidden" name="CourseIndex" value="`)
	w.WriteHTMLString(r.Form.Get("CourseIndex"))
	w.AppendString(`">`)

	w.AppendString(`<input type="hidden" name="LessonIndex" value="`)
	w.WriteHTMLString(r.Form.Get("LessonIndex"))
	w.AppendString(`">`)

	w.AppendString(`<label>Name: `)
	w.AppendString(`<input type="text" minlength="1" maxlength="45" name="Name" value="`)
	w.WriteHTMLString(lesson.Name)
	w.AppendString(`" required>`)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>Theory:<br>`)
	w.AppendString(`<textarea cols="80" rows="24" minlength="1" maxlength="1024" name="Theory" required>`)
	w.WriteHTMLString(lesson.Theory)
	w.AppendString(`</textarea>`)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	for i := 0; i < len(lesson.Steps); i++ {
		var name, stepType string
		var draft bool

		step := lesson.Steps[i]
		switch step := step.(type) {
		default:
			panic("invalid step type")
		case *StepTest:
			name = step.Name
			draft = step.Draft
			stepType = "Test"
		case *StepProgramming:
			name = step.Name
			draft = step.Draft
			stepType = "Programming task"
		}

		buffer := make([]byte, 20)
		n := SlicePutInt(buffer, i+1)

		w.AppendString(`<fieldset>`)

		w.AppendString(`<legend>Step #`)
		w.Write(buffer[:n])
		if draft {
			w.AppendString(` (draft)`)
		}
		w.AppendString(`</legend>`)

		w.AppendString(`<p>Name: `)
		w.WriteHTMLString(name)
		w.AppendString(`</p>`)

		w.AppendString(`<p>Type: `)
		w.AppendString(stepType)
		w.AppendString(`</p>`)

		n = SlicePutInt(buffer, i)
		si := unsafe.String(unsafe.SliceData(buffer), n)

		w.AppendString(`<input type="submit" name="Command`)
		w.WriteString(si)
		w.AppendString(`" value="Edit" formnovalidate>`)
		w.AppendString("\r\n")
		w.AppendString(`<input type="submit" name="Command`)
		w.WriteString(si)
		w.AppendString(`" value="Delete" formnovalidate>`)
		if len(lesson.Steps) > 1 {
			if i > 0 {
				w.AppendString("\r\n")
				w.AppendString(`<input type="submit" name="Command`)
				w.WriteString(si)
				w.AppendString(`" value="↑", "^|" formnovalidate>`)
			}
			if i < len(lesson.Steps)-1 {
				w.AppendString("\r\n")
				w.AppendString(`<input type="submit" name="Command`)
				w.WriteString(si)
				w.AppendString(`" value="↓" formnovalidate>`)
			}
		}

		w.AppendString(`</fieldset>`)
		w.AppendString(`<br>`)
	}

	w.AppendString(`<input type="submit" name="NextPage" value="Add test">`)
	w.AppendString("\r\n")
	w.AppendString(`<input type="submit" name="NextPage" value="Add programming task">`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<input type="submit" name="NextPage" value="Next">`)

	w.AppendString(`</form>`)

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func CourseCreateCoursePageHandler(w *HTTPResponse, r *HTTPRequest, course *Course) error {
	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>Create course</title></head>`)
	w.AppendString(`<body>`)
	w.AppendString(`<h1>Course</h1>`)
	w.AppendString(`<h2>Create</h2>`)

	ErrorDiv(w, r.Form.Get("Error"))

	w.AppendString(`<form method="POST" action="/course/create">`)

	w.AppendString(`<input type="hidden" name="CurrentPage" value="Course">`)

	w.AppendString(`<input type="hidden" name="CourseIndex" value="`)
	w.WriteHTMLString(r.Form.Get("CourseIndex"))
	w.AppendString(`">`)

	w.AppendString(`<label>Name: `)
	w.AppendString(`<input type="text" minlength="1" maxlength="45" name="Name" value="`)
	w.WriteHTMLString(course.Name)
	w.AppendString(`" required>`)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	for i := 0; i < len(course.Lessons); i++ {
		lesson := course.Lessons[i]

		buffer := make([]byte, 20)
		n := SlicePutInt(buffer, i+1)

		w.AppendString(`<fieldset>`)

		w.AppendString(`<legend>Lesson #`)
		w.Write(buffer[:n])
		if lesson.Draft {
			w.AppendString(` (draft)`)
		}
		w.AppendString(`</legend>`)

		w.AppendString(`<p>Name: `)
		w.WriteHTMLString(lesson.Name)
		w.AppendString(`</p>`)

		w.AppendString(`<p>Theory: `)
		CourseLessonDisplayTheory(w, lesson.Theory)
		w.AppendString(`</p>`)

		n = SlicePutInt(buffer, i)
		si := unsafe.String(unsafe.SliceData(buffer), n)

		w.AppendString(`<input type="submit" name="Command`)
		w.WriteString(si)
		w.AppendString(`" value="Edit" formnovalidate>`)
		w.AppendString("\r\n")
		w.AppendString(`<input type="submit" name="Command`)
		w.WriteString(si)
		w.AppendString(`" value="Delete" formnovalidate>`)
		if len(course.Lessons) > 1 {
			if i > 0 {
				w.AppendString("\r\n")
				w.AppendString(`<input type="submit" name="Command`)
				w.WriteString(si)
				w.AppendString(`" value="↑", "^|" formnovalidate>`)
			}
			if i < len(course.Lessons)-1 {
				w.AppendString("\r\n")
				w.AppendString(`<input type="submit" name="Command`)
				w.WriteString(si)
				w.AppendString(`" value="↓" formnovalidate>`)
			}
		}

		w.AppendString(`</fieldset>`)
		w.AppendString(`<br>`)
	}

	w.AppendString(`<input type="submit" name="NextPage" value="Add lesson">`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<input type="submit" name="NextPage" value="Create">`)

	w.AppendString(`</form>`)

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func CourseCreateGetIndicies(indicies string) (pindex int, spindex string, sindex int, ssindex string, err error) {
	if len(indicies) == 0 {
		return
	}

	spindex = indicies
	if i := FindChar(indicies, '.'); i != -1 {
		ssindex = indicies[i+1:]
		sindex, err = strconv.Atoi(ssindex)
		if err != nil {
			return
		}
		spindex = indicies[:i]
	}
	pindex, err = strconv.Atoi(spindex)
	return
}

func CourseCreateHandleCommand(w *HTTPResponse, r *HTTPRequest, course *Course, currentPage, k, command string) error {
	pindex, spindex, sindex, ssindex, err := CourseCreateGetIndicies(k[len("Command"):])
	if err != nil {
		return ReloadPageError
	}

	switch currentPage {
	default:
		return ReloadPageError
	case "Course":
		switch command {
		case "Delete":
			course.Lessons = RemoveAtIndex(course.Lessons, pindex)
		case "Edit":
			if (pindex < 0) || (pindex >= len(course.Lessons)) {
				return ReloadPageError
			}
			lesson := course.Lessons[pindex]
			lesson.Draft = true

			r.Form.Set("LessonIndex", spindex)
			return CourseCreateLessonPageHandler(w, r, lesson)
		case "↑", "^|":
			MoveUp(course.Lessons, pindex)
		case "↓", "|v":
			MoveDown(course.Lessons, pindex)
		}

		return CourseCreateCoursePageHandler(w, r, course)
	case "Lesson":
		li, err := strconv.Atoi(r.Form.Get("LessonIndex"))
		if (err != nil) || (li < 0) || (li >= len(course.Lessons)) {
			return ReloadPageError
		}
		lesson := course.Lessons[li]

		switch command {
		case "Delete":
			lesson.Steps = RemoveAtIndex(lesson.Steps, pindex)
		case "Edit":
			if (pindex < 0) || (pindex >= len(lesson.Steps)) {
				return ReloadPageError
			}
			step := lesson.Steps[pindex]

			r.Form.Set("StepIndex", spindex)
			switch step := step.(type) {
			default:
				panic("invalid step type")
			case *StepTest:
				step.Draft = true
				return CourseCreateTestPageHandler(w, r, step)
			case *StepProgramming:
				step.Draft = true
				return CourseCreateProgrammingPageHandler(w, r, step)
			}
		case "↑", "^|":
			MoveUp(lesson.Steps, pindex)
		case "↓", "|v":
			MoveDown(lesson.Steps, pindex)
		}

		return CourseCreateLessonPageHandler(w, r, lesson)
	case "Test":
		li, err := strconv.Atoi(r.Form.Get("LessonIndex"))
		if (err != nil) || (li < 0) || (li >= len(course.Lessons)) {
			return ReloadPageError
		}
		lesson := course.Lessons[li]

		si, err := strconv.Atoi(r.Form.Get("StepIndex"))
		if (err != nil) || (si < 0) || (si >= len(lesson.Steps)) {
			return ReloadPageError
		}
		test, ok := lesson.Steps[si].(*StepTest)
		if !ok {
			return ReloadPageError
		}

		if err := CourseCreateTestVerifyRequest(r.Form, test, false); err != nil {
			return err
		}

		switch command {
		case "Add another answer":
			if (pindex < 0) || (pindex >= len(test.Questions)) {
				return ReloadPageError
			}
			question := &test.Questions[pindex]
			question.Answers = append(question.Answers, "")
		case "-": /* remove answer */
			if (pindex < 0) || (pindex >= len(test.Questions)) {
				return ReloadPageError
			}
			question := &test.Questions[pindex]

			if (sindex < 0) || (sindex >= len(question.Answers)) {
				return ReloadPageError
			}
			question.Answers = RemoveAtIndex(question.Answers, sindex)

			for i := 0; i < len(question.CorrectAnswers); i++ {
				if question.CorrectAnswers[i] == sindex {
					question.CorrectAnswers = RemoveAtIndex(question.CorrectAnswers, i)
					i--
				} else if question.CorrectAnswers[i] > sindex {
					question.CorrectAnswers[i]--
				}
			}
		case "Add another question":
			test.Questions = append(test.Questions, Question{})
		case "Delete":
			test.Questions = RemoveAtIndex(test.Questions, pindex)
		case "↑", "^|":
			if ssindex == "" {
				MoveUp(test.Questions, pindex)
			} else {
				if (pindex < 0) || (pindex >= len(test.Questions)) {
					return ReloadPageError
				}
				question := &test.Questions[pindex]

				MoveUp(question.Answers, sindex)
				for i := 0; i < len(question.CorrectAnswers); i++ {
					if question.CorrectAnswers[i] == sindex-1 {
						/* If previous answer is correct and current is not. */
						if (i == len(question.CorrectAnswers)-1) || (question.CorrectAnswers[i+1] != sindex) {
							question.CorrectAnswers[i] = sindex
						}
						break
					} else if question.CorrectAnswers[i] == sindex {
						/* If current answer is correct and previous is not. */
						if (i == 0) || (question.CorrectAnswers[i-1] != sindex-1) {
							question.CorrectAnswers[i] = sindex - 1
						}
						break
					}
				}
			}
		case "↓", "|v":
			if ssindex == "" {
				MoveDown(test.Questions, pindex)
			} else {
				if (pindex < 0) || (pindex >= len(test.Questions)) {
					return ReloadPageError
				}
				question := &test.Questions[pindex]

				MoveDown(question.Answers, sindex)
				for i := 0; i < len(question.CorrectAnswers); i++ {
					if question.CorrectAnswers[i] == sindex {
						/* If current answer is correct and next is not. */
						if (i == len(question.CorrectAnswers)-1) || (question.CorrectAnswers[i+1] != sindex+1) {
							question.CorrectAnswers[i] = sindex + 1
						}
						break
					} else if question.CorrectAnswers[i] == sindex+1 {
						/* If next answer is correct and current is not. */
						if (i == 0) || (question.CorrectAnswers[i-1] != sindex) {
							question.CorrectAnswers[i] = sindex
						}
						break
					}
				}
			}
		}

		return CourseCreateTestPageHandler(w, r, test)

	case "Programming":
		li, err := strconv.Atoi(r.Form.Get("LessonIndex"))
		if (err != nil) || (li < 0) || (li >= len(course.Lessons)) {
			return ReloadPageError
		}
		lesson := course.Lessons[li]

		si, err := strconv.Atoi(r.Form.Get("StepIndex"))
		if (err != nil) || (si < 0) || (si >= len(lesson.Steps)) {
			return ReloadPageError
		}
		task, ok := lesson.Steps[si].(*StepProgramming)
		if !ok {
			return ReloadPageError
		}

		if err := CourseCreateProgrammingVerifyRequest(r.Form, task, false); err != nil {
			return err
		}

		switch command {
		case "Add example":
			task.Checks[CheckTypeExample] = append(task.Checks[CheckTypeExample], Check{})
		case "Add test":
			task.Checks[CheckTypeTest] = append(task.Checks[CheckTypeTest], Check{})
		case "-":
			if (sindex < 0) || (sindex >= len(task.Checks)) {
				return ReloadPageError
			}
			task.Checks[sindex] = RemoveAtIndex(task.Checks[sindex], pindex)
		case "↑", "^|":
			if (sindex < 0) || (sindex >= len(task.Checks)) {
				return ReloadPageError
			}
			MoveUp(task.Checks[sindex], pindex)
		case "↓", "|v":
			if (sindex < 0) || (sindex >= len(task.Checks)) {
				return ReloadPageError
			}
			MoveDown(task.Checks[sindex], pindex)
		}

		return CourseCreateProgrammingPageHandler(w, r, task)
	}
	return nil
}

func CourseCreatePageHandler(w *HTTPResponse, r *HTTPRequest) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}
	if session.ID != AdminID {
		return ForbiddenError
	}
	user := &DB.Users[session.ID]

	if err := r.ParseForm(); err != nil {
		return ReloadPageError
	}

	currentPage := r.Form.Get("CurrentPage")
	nextPage := r.Form.Get("NextPage")

	var course *Course
	courseIndex := r.Form.Get("CourseIndex")
	if courseIndex == "" {
		course = new(Course)
		user.Courses = append(user.Courses, course)
		r.Form.Set("CourseIndex", strconv.Itoa(len(user.Courses)-1))
	} else {
		ci, err := strconv.Atoi(courseIndex)
		if (err != nil) || (ci < 0) || (ci >= len(user.Courses)) {
			return ReloadPageError
		}
		course = user.Courses[ci]
	}
	course.Draft = true

	for i := 0; i < len(r.Form); i++ {
		k := r.Form[i].Key
		v := r.Form[i].Values[0]

		/* 'command' is button, which modifies content of a current page. */
		if StringStartsWith(k, "Command") {
			/* NOTE(anton2920): after command is executed, function must return. */
			return CourseCreateHandleCommand(w, r, course, currentPage, k, v)
		}
	}

	/* 'currentPage' is the page to check before leaving it. */
	switch currentPage {
	case "Course":
		if err := CourseCreateCourseVerifyRequest(r.Form, course); err != nil {
			return WritePageEx(w, r, CourseCreateCoursePageHandler, course, err)
		}
	case "Lesson":
		li, err := strconv.Atoi(r.Form.Get("LessonIndex"))
		if (err != nil) || (li < 0) || (li >= len(course.Lessons)) {
			return ReloadPageError
		}
		lesson := course.Lessons[li]

		if err := CourseCreateLessonVerifyRequest(r.Form, lesson); err != nil {
			return WritePageEx(w, r, CourseCreateLessonPageHandler, lesson, err)
		}
	case "Test":
		li, err := strconv.Atoi(r.Form.Get("LessonIndex"))
		if (err != nil) || (li < 0) || (li >= len(course.Lessons)) {
			return ReloadPageError
		}
		lesson := course.Lessons[li]

		si, err := strconv.Atoi(r.Form.Get("StepIndex"))
		if (err != nil) || (si < 0) || (si >= len(lesson.Steps)) {
			return ReloadPageError
		}
		test, ok := lesson.Steps[si].(*StepTest)
		if !ok {
			return ReloadPageError
		}

		if err := CourseCreateTestVerifyRequest(r.Form, test, true); err != nil {
			return WritePageEx(w, r, CourseCreateTestPageHandler, test, err)
		}
	case "Programming":
		li, err := strconv.Atoi(r.Form.Get("LessonIndex"))
		if (err != nil) || (li < 0) || (li >= len(course.Lessons)) {
			return ReloadPageError
		}
		lesson := course.Lessons[li]

		si, err := strconv.Atoi(r.Form.Get("StepIndex"))
		if (err != nil) || (si < 0) || (si >= len(lesson.Steps)) {
			return ReloadPageError
		}
		task, ok := lesson.Steps[si].(*StepProgramming)
		if !ok {
			return ReloadPageError
		}

		if err := CourseCreateProgrammingVerifyRequest(r.Form, task, true); err != nil {
			return WritePageEx(w, r, CourseCreateProgrammingPageHandler, task, err)
		}
	}

	switch nextPage {
	default:
		return CourseCreateCoursePageHandler(w, r, course)
	case "Next":
		li, err := strconv.Atoi(r.Form.Get("LessonIndex"))
		if (err != nil) || (li < 0) || (li >= len(course.Lessons)) {
			return ReloadPageError
		}
		lesson := course.Lessons[li]
		lesson.Draft = false

		return CourseCreateCoursePageHandler(w, r, course)
	case "Add lesson":
		lesson := new(Lesson)
		lesson.Draft = true
		course.Lessons = append(course.Lessons, lesson)

		r.Form.Set("LessonIndex", strconv.Itoa(len(course.Lessons)-1))
		return CourseCreateLessonPageHandler(w, r, lesson)
	case "Continue":
		li, err := strconv.Atoi(r.Form.Get("LessonIndex"))
		if (err != nil) || (li < 0) || (li >= len(course.Lessons)) {
			return ReloadPageError
		}
		lesson := course.Lessons[li]

		si, err := strconv.Atoi(r.Form.Get("StepIndex"))
		if (err != nil) || (si < 0) || (si >= len(lesson.Steps)) {
			return ReloadPageError
		}
		switch step := lesson.Steps[si].(type) {
		default:
			panic("invalid step type")
		case *StepTest:
			step.Draft = false
		case *StepProgramming:
			step.Draft = false
		}

		return CourseCreateLessonPageHandler(w, r, lesson)
	case "Add test":
		li, err := strconv.Atoi(r.Form.Get("LessonIndex"))
		if (err != nil) || (li < 0) || (li >= len(course.Lessons)) {
			return ReloadPageError
		}
		lesson := course.Lessons[li]

		test := new(StepTest)
		test.Draft = true
		lesson.Steps = append(lesson.Steps, test)

		r.Form.Set("StepIndex", strconv.Itoa(len(lesson.Steps)-1))
		return CourseCreateTestPageHandler(w, r, test)
	case "Add programming task":
		li, err := strconv.Atoi(r.Form.Get("LessonIndex"))
		if (err != nil) || (li < 0) || (li >= len(course.Lessons)) {
			return ReloadPageError
		}
		lesson := course.Lessons[li]

		task := new(StepProgramming)
		task.Draft = true
		lesson.Steps = append(lesson.Steps, task)

		r.Form.Set("StepIndex", strconv.Itoa(len(lesson.Steps)-1))
		return CourseCreateProgrammingPageHandler(w, r, task)
	}
}

func CourseDeleteHandler(w *HTTPResponse, r *HTTPRequest) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}
	user := &DB.Users[session.ID]

	if err := r.ParseForm(); err != nil {
		return ReloadPageError
	}

	id, err := strconv.Atoi(r.Form.Get("CourseIndex"))
	if (err != nil) || (id < 0) || (id > len(user.Courses)) {
		return ReloadPageError
	}

	/* TODO(anton2920): this will screw up indicies for courses that are being edited. */
	user.Courses = RemoveAtIndex(user.Courses, id)

	w.RedirectString("/", HTTPStatusSeeOther)
	return nil
}
