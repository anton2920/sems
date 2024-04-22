package main

import (
	"fmt"
	"strconv"
	"unsafe"
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

func LessonAddVerifyRequest(vs URLValues, lesson *Lesson) error {
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

func LessonTestAddVerifyRequest(vs URLValues, test *StepTest, shouldCheck bool) error {
	test.Name = vs.Get("Name")
	if (shouldCheck) && (!StringLengthInRange(test.Name, MinStepNameLen, MaxStepNameLen)) {
		return NewHTTPError(HTTPStatusBadRequest, fmt.Sprintf("test name length must be between %d and %d characters long", MinStepNameLen, MaxStepNameLen))
	}

	questions := vs.GetMany("Question")

	answerKey := make([]byte, 30)
	copy(answerKey, "Answer")

	correctAnswerKey := make([]byte, 30)
	copy(correctAnswerKey, "CorrectAnswer")

	for i := 0; i < len(questions); i++ {
		if i >= len(test.Questions) {
			test.Questions = append(test.Questions, Question{})
		}
		question := &test.Questions[i]
		question.Name = questions[i]
		if (shouldCheck) && (!StringLengthInRange(question.Name, MinQuestionLen, MaxQuestionLen)) {
			return NewHTTPError(HTTPStatusBadRequest, fmt.Sprintf("question %d: title length must be between %d and %d characters long", i+1, MinQuestionLen, MaxQuestionLen))
		}

		n := SlicePutInt(answerKey[len("Answer"):], i)
		answers := vs.GetMany(unsafe.String(unsafe.SliceData(answerKey), len("Answer")+n))
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

		n = SlicePutInt(correctAnswerKey[len("CorrectAnswer"):], i)
		correctAnswers := vs.GetMany(unsafe.String(unsafe.SliceData(correctAnswerKey), len("CorrectAnswer")+n))
		for j := 0; j < len(correctAnswers); j++ {
			if j >= len(question.CorrectAnswers) {
				question.CorrectAnswers = append(question.CorrectAnswers, 0)
			}

			var err error
			question.CorrectAnswers[j], err = strconv.Atoi(correctAnswers[j])
			if (err != nil) || (question.CorrectAnswers[j] < 0) || (question.CorrectAnswers[j] >= len(question.Answers)) {
				return ClientError(err)
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

func LessonProgrammingAddVerifyRequest(vs URLValues, task *StepProgramming, shouldCheck bool) error {
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
			return ClientError(nil)
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

func LessonTestAddPageHandler(w *HTTPResponse, r *HTTPRequest, test *StepTest) error {
	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>Test</title></head>`)
	w.AppendString(`<body>`)
	w.AppendString(`<h1>Lesson</h1>`)
	w.AppendString(`<h2>Test</h2>`)

	DisplayErrorMessage(w, r.Form.Get("Error"))

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

func LessonProgrammingAddDisplayChecks(w *HTTPResponse, task *StepProgramming, checkType CheckType) {
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

func LessonProgrammingAddPageHandler(w *HTTPResponse, r *HTTPRequest, task *StepProgramming) error {
	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>Programming task</title></head>`)
	w.AppendString(`<body>`)
	w.AppendString(`<h1>Lesson</h1>`)
	w.AppendString(`<h2>Programming task</h2>`)

	DisplayErrorMessage(w, r.Form.Get("Error"))

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
	LessonProgrammingAddDisplayChecks(w, task, CheckTypeExample)
	w.AppendString(`<input type="submit" name="Command" value="Add example">`)

	w.AppendString(`<h3>Tests</h3>`)
	LessonProgrammingAddDisplayChecks(w, task, CheckTypeTest)
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

func LessonAddPageHandler(w *HTTPResponse, r *HTTPRequest, lesson *Lesson) error {
	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>Create lesson</title></head>`)
	w.AppendString(`<body>`)
	w.AppendString(`<h1>Lesson</h1>`)

	DisplayErrorMessage(w, r.Form.Get("Error"))

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

func LessonAddHandleCommand(w *HTTPResponse, r *HTTPRequest, lessons []*Lesson, currentPage, k, command string) error {
	/* TODO(anton2920): pass as params. */
	pindex, spindex, sindex, ssindex, err := GetIndicies(k[len("Command"):])
	if err != nil {
		return ClientError(err)
	}

	switch currentPage {
	default:
		return ClientError(nil)
	case "Lesson":
		li, err := GetValidIndex(r.Form, "LessonIndex", lessons)
		if err != nil {
			return err
		}
		lesson := lessons[li]

		switch command {
		case "Delete":
			lesson.Steps = RemoveAtIndex(lesson.Steps, pindex)
		case "Edit":
			if (pindex < 0) || (pindex >= len(lesson.Steps)) {
				return ClientError(nil)
			}
			step := lesson.Steps[pindex]

			r.Form.Set("StepIndex", spindex)
			switch step := step.(type) {
			default:
				panic("invalid step type")
			case *StepTest:
				step.Draft = true
				return LessonTestAddPageHandler(w, r, step)
			case *StepProgramming:
				step.Draft = true
				return LessonProgrammingAddPageHandler(w, r, step)
			}
		case "↑", "^|":
			MoveUp(lesson.Steps, pindex)
		case "↓", "|v":
			MoveDown(lesson.Steps, pindex)
		}

		return LessonAddPageHandler(w, r, lesson)
	case "Test":
		li, err := GetValidIndex(r.Form, "LessonIndex", lessons)
		if err != nil {
			return err
		}
		lesson := lessons[li]

		si, err := GetValidIndex(r.Form, "StepIndex", lesson.Steps)
		if err != nil {
			return err
		}
		test, ok := lesson.Steps[si].(*StepTest)
		if !ok {
			return ClientError(nil)
		}

		if err := LessonTestAddVerifyRequest(r.Form, test, false); err != nil {
			return err
		}

		switch command {
		case "Add another answer":
			if (pindex < 0) || (pindex >= len(test.Questions)) {
				return ClientError(nil)
			}
			question := &test.Questions[pindex]
			question.Answers = append(question.Answers, "")
		case "-": /* remove answer */
			if (pindex < 0) || (pindex >= len(test.Questions)) {
				return ClientError(nil)
			}
			question := &test.Questions[pindex]

			if (sindex < 0) || (sindex >= len(question.Answers)) {
				return ClientError(nil)
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
					return ClientError(nil)
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
					return ClientError(nil)
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

		return LessonTestAddPageHandler(w, r, test)

	case "Programming":
		li, err := GetValidIndex(r.Form, "LessonIndex", lessons)
		if err != nil {
			return err
		}
		lesson := lessons[li]

		si, err := GetValidIndex(r.Form, "StepIndex", lesson.Steps)
		if err != nil {
			return err
		}
		task, ok := lesson.Steps[si].(*StepProgramming)
		if !ok {
			return ClientError(nil)
		}

		if err := LessonProgrammingAddVerifyRequest(r.Form, task, false); err != nil {
			return err
		}

		switch command {
		case "Add example":
			task.Checks[CheckTypeExample] = append(task.Checks[CheckTypeExample], Check{})
		case "Add test":
			task.Checks[CheckTypeTest] = append(task.Checks[CheckTypeTest], Check{})
		case "-":
			if (sindex < 0) || (sindex >= len(task.Checks)) {
				return ClientError(nil)
			}
			task.Checks[sindex] = RemoveAtIndex(task.Checks[sindex], pindex)
		case "↑", "^|":
			if (sindex < 0) || (sindex >= len(task.Checks)) {
				return ClientError(nil)
			}
			MoveUp(task.Checks[sindex], pindex)
		case "↓", "|v":
			if (sindex < 0) || (sindex >= len(task.Checks)) {
				return ClientError(nil)
			}
			MoveDown(task.Checks[sindex], pindex)
		}

		return LessonProgrammingAddPageHandler(w, r, task)
	}
}
