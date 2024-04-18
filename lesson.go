package main

import (
	"encoding/gob"
	"fmt"
	"strconv"
	"unicode/utf8"
	"unsafe"
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

const (
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

func LessonsDeepCopy(dst *[]*Lesson, src []*Lesson) {
	*dst = make([]*Lesson, len(src))

	for l := 0; l < len(src); l++ {
		sl := src[l]

		dl := new(Lesson)
		(*dst)[l] = dl

		dl.Name = sl.Name
		dl.Theory = sl.Theory
		dl.Steps = make([]interface{}, len(sl.Steps))

		for s := 0; s < len(sl.Steps); s++ {
			switch ss := sl.Steps[s].(type) {
			default:
				panic("invalid step type")
			case *StepTest:
				ds := new(StepTest)
				dl.Steps[s] = ds

				ds.Name = ss.Name
				ds.Questions = make([]Question, len(ss.Questions))

				for q := 0; q < len(ss.Questions); q++ {
					sq := &ss.Questions[q]

					dq := &ds.Questions[q]
					dq.Name = sq.Name
					dq.Answers = make([]string, len(sq.Answers))
					copy(dq.Answers, sq.Answers)
					dq.CorrectAnswers = make([]int, len(sq.CorrectAnswers))
					copy(dq.CorrectAnswers, sq.CorrectAnswers)
				}
			case *StepProgramming:
				ds := new(StepProgramming)
				dl.Steps[s] = ds

				ds.Name = ss.Name
				ds.Description = ss.Description
				ds.Checks[CheckTypeExample] = make([]Check, len(ss.Checks[CheckTypeExample]))
				copy(ds.Checks[CheckTypeExample], ss.Checks[CheckTypeExample])
				ds.Checks[CheckTypeTest] = make([]Check, len(ss.Checks[CheckTypeTest]))
				copy(ds.Checks[CheckTypeTest], ss.Checks[CheckTypeTest])
			}
		}
	}
}

func LessonVerifyRequest(vs URLValues, lesson *Lesson) error {
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

func TestVerifyRequest(vs URLValues, test *StepTest, shouldCheck bool) error {
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

func ProgrammingVerifyRequest(vs URLValues, task *StepProgramming, shouldCheck bool) error {
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

func LessonDisplayTheory(w *HTTPResponse, theory string) {
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

func TestPageHandler(w *HTTPResponse, r *HTTPRequest, test *StepTest) error {
	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>Test</title></head>`)
	w.AppendString(`<body>`)
	w.AppendString(`<h1>Lesson</h1>`)
	w.AppendString(`<h2>Test</h2>`)

	ErrorDiv(w, r.Form.Get("Error"))

	w.AppendString(`<form method="POST" action="`)
	w.WriteString(r.URL.Path)
	w.AppendString(`">`)

	w.AppendString(`<input type="hidden" name="CurrentPage" value="Test">`)

	w.AppendString(`<input type="hidden" name="ID" value="`)
	w.WriteHTMLString(r.Form.Get("ID"))
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

		w.AppendString(`<fieldset>`)
		w.AppendString(`<legend>Question #`)
		w.WriteInt(i + 1)
		w.AppendString(`</legend>`)
		w.AppendString(`<label>Title: `)
		w.AppendString(`<input type="text" minlength="1" maxlength="128" name="Question" value="`)
		w.WriteHTMLString(question.Name)
		w.AppendString(`" required>`)
		w.AppendString(`</label>`)
		w.AppendString(`<br>`)

		w.AppendString(`<p>Answers (mark the correct ones):</p>`)
		w.AppendString(`<ol>`)

		if len(question.Answers) == 0 {
			question.Answers = append(question.Answers, "")
		}
		for j := 0; j < len(question.Answers); j++ {
			answer := question.Answers[j]

			if j > 0 {
				w.AppendString(`<br>`)
			}

			w.AppendString(`<li>`)

			w.AppendString(`<input type="checkbox" name="CorrectAnswer`)
			w.WriteInt(i)
			w.AppendString(`" value="`)
			w.WriteInt(j)
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
			w.WriteInt(i)
			w.AppendString(`" value="`)
			w.WriteHTMLString(answer)
			w.AppendString(`" required>`)

			if len(question.Answers) > 1 {
				w.AppendString("\r\n")
				w.AppendString(`<input type="submit" name="Command`)
				w.WriteInt(i)
				w.AppendString(`.`)
				w.WriteInt(j)
				w.AppendString(`" value="-" formnovalidate>`)
				if j > 0 {
					w.AppendString("\r\n")
					w.AppendString(`<input type="submit" name="Command`)
					w.WriteInt(i)
					w.AppendString(`.`)
					w.WriteInt(j)
					w.AppendString(`" value="↑" formnovalidate>`)
				}
				if j < len(question.Answers)-1 {
					w.AppendString("\r\n")
					w.AppendString(`<input type="submit" name="Command`)
					w.WriteInt(i)
					w.AppendString(`.`)
					w.WriteInt(j)
					w.AppendString(`" value="↓" formnovalidate>`)
				}
			}

			w.AppendString(`</li>`)
		}
		w.AppendString(`</ol>`)

		w.AppendString(`<input type="submit" name="Command`)
		w.WriteInt(i)
		w.AppendString(`" value="Add another answer" formnovalidate>`)

		if len(test.Questions) > 1 {
			w.AppendString(`<br><br>`)
			w.AppendString("\r\n")
			w.AppendString(`<input type="submit" name="Command`)
			w.WriteInt(i)
			w.AppendString(`" value="Delete" formnovalidate>`)
			if i > 0 {
				w.AppendString("\r\n")
				w.AppendString(`<input type="submit" name="Command`)
				w.WriteInt(i)
				w.AppendString(`" value="↑" formnovalidate>`)
			}
			if i < len(test.Questions)-1 {
				w.AppendString("\r\n")
				w.AppendString(`<input type="submit" name="Command`)
				w.WriteInt(i)
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

func ProgrammingDisplayChecks(w *HTTPResponse, task *StepProgramming, checkType CheckType) {
	checks := task.Checks[checkType]

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

		w.AppendString("\r\n")
		w.AppendString(`<input type="submit" name="Command`)
		w.WriteInt(i)
		w.AppendString(`.`)
		w.WriteInt(int(checkType))
		w.AppendString(`" value="-" formnovalidate>`)

		if len(checks) > 1 {
			if i > 0 {
				w.AppendString("\r\n")
				w.AppendString(`<input type="submit" name="Command`)
				w.WriteInt(i)
				w.AppendString(`.`)
				w.WriteInt(int(checkType))
				w.AppendString(`" value="↑" formnovalidate>`)
			}
			if i < len(checks)-1 {
				w.AppendString("\r\n")
				w.AppendString(`<input type="submit" name="Command`)
				w.WriteInt(i)
				w.AppendString(`.`)
				w.WriteInt(int(checkType))
				w.AppendString(`" value="↓" formnovalidate>`)
			}
		}

		w.AppendString(`</li>`)
	}
	w.AppendString(`</ol>`)
}

func ProgrammingPageHandler(w *HTTPResponse, r *HTTPRequest, task *StepProgramming) error {
	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>Programming task</title></head>`)
	w.AppendString(`<body>`)
	w.AppendString(`<h1>Lesson</h1>`)
	w.AppendString(`<h2>Programming task</h2>`)

	ErrorDiv(w, r.Form.Get("Error"))

	w.AppendString(`<form method="POST" action="`)
	w.WriteString(r.URL.Path)
	w.AppendString(`">`)

	w.AppendString(`<input type="hidden" name="CurrentPage" value="Programming">`)

	w.AppendString(`<input type="hidden" name="ID" value="`)
	w.WriteHTMLString(r.Form.Get("ID"))
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
	ProgrammingDisplayChecks(w, task, CheckTypeExample)
	w.AppendString(`<input type="submit" name="Command" value="Add example">`)

	w.AppendString(`<h3>Tests</h3>`)
	ProgrammingDisplayChecks(w, task, CheckTypeTest)
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

func LessonPageHandler(w *HTTPResponse, r *HTTPRequest, lesson *Lesson) error {
	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>Create lesson</title></head>`)
	w.AppendString(`<body>`)
	w.AppendString(`<h1>Lesson</h1>`)

	ErrorDiv(w, r.Form.Get("Error"))

	w.AppendString(`<form method="POST" action="`)
	w.WriteString(r.URL.Path)
	w.AppendString(`">`)

	w.AppendString(`<input type="hidden" name="CurrentPage" value="Lesson">`)

	w.AppendString(`<input type="hidden" name="ID" value="`)
	w.WriteHTMLString(r.Form.Get("ID"))
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

		w.AppendString(`<fieldset>`)

		w.AppendString(`<legend>Step #`)
		w.WriteInt(i + 1)
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

		w.AppendString(`<input type="submit" name="Command`)
		w.WriteInt(i)
		w.AppendString(`" value="Edit" formnovalidate>`)
		w.AppendString("\r\n")
		w.AppendString(`<input type="submit" name="Command`)
		w.WriteInt(i)
		w.AppendString(`" value="Delete" formnovalidate>`)
		if len(lesson.Steps) > 1 {
			if i > 0 {
				w.AppendString("\r\n")
				w.AppendString(`<input type="submit" name="Command`)
				w.WriteInt(i)
				w.AppendString(`" value="↑" formnovalidate>`)
			}
			if i < len(lesson.Steps)-1 {
				w.AppendString("\r\n")
				w.AppendString(`<input type="submit" name="Command`)
				w.WriteInt(i)
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

func LessonHandleCommand(w *HTTPResponse, r *HTTPRequest, lessons []*Lesson, currentPage, k, command string) error {
	/* TODO(anton2920): pass as params. */
	pindex, spindex, sindex, ssindex, err := GetIndicies(k[len("Command"):])
	if err != nil {
		return ReloadPageError
	}

	switch currentPage {
	default:
		return ReloadPageError
	case "Lesson":
		li, err := strconv.Atoi(r.Form.Get("LessonIndex"))
		if (err != nil) || (li < 0) || (li >= len(lessons)) {
			return ReloadPageError
		}
		lesson := lessons[li]

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
				return TestPageHandler(w, r, step)
			case *StepProgramming:
				step.Draft = true
				return ProgrammingPageHandler(w, r, step)
			}
		case "↑", "^|":
			MoveUp(lesson.Steps, pindex)
		case "↓", "|v":
			MoveDown(lesson.Steps, pindex)
		}

		return LessonPageHandler(w, r, lesson)
	case "Test":
		li, err := strconv.Atoi(r.Form.Get("LessonIndex"))
		if (err != nil) || (li < 0) || (li >= len(lessons)) {
			return ReloadPageError
		}
		lesson := lessons[li]

		si, err := strconv.Atoi(r.Form.Get("StepIndex"))
		if (err != nil) || (si < 0) || (si >= len(lesson.Steps)) {
			return ReloadPageError
		}
		test, ok := lesson.Steps[si].(*StepTest)
		if !ok {
			return ReloadPageError
		}

		if err := TestVerifyRequest(r.Form, test, false); err != nil {
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

		return TestPageHandler(w, r, test)

	case "Programming":
		li, err := strconv.Atoi(r.Form.Get("LessonIndex"))
		if (err != nil) || (li < 0) || (li >= len(lessons)) {
			return ReloadPageError
		}
		lesson := lessons[li]

		si, err := strconv.Atoi(r.Form.Get("StepIndex"))
		if (err != nil) || (si < 0) || (si >= len(lesson.Steps)) {
			return ReloadPageError
		}
		task, ok := lesson.Steps[si].(*StepProgramming)
		if !ok {
			return ReloadPageError
		}

		if err := ProgrammingVerifyRequest(r.Form, task, false); err != nil {
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

		return ProgrammingPageHandler(w, r, task)
	}
}
