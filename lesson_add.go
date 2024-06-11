package main

import (
	"unsafe"

	"github.com/anton2920/gofa/database"
	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/net/url"
	"github.com/anton2920/gofa/slices"
	"github.com/anton2920/gofa/strings"
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

func LessonFillFromRequest(vs url.Values, lesson *Lesson) {
	lesson.Name = vs.Get("Name")
	lesson.Theory = vs.Get("Theory")
}

func LessonVerify(lesson *Lesson) error {
	if !strings.LengthInRange(lesson.Name, MinNameLen, MaxNameLen) {
		return http.BadRequest("lesson name length must be between %d and %d characters long", MinNameLen, MaxNameLen)
	}

	if !strings.LengthInRange(lesson.Theory, MinTheoryLen, MaxTheoryLen) {
		return http.BadRequest("lesson theory length must be between %d and %d characters long", MinTheoryLen, MaxTheoryLen)
	}

	for si := 0; si < len(lesson.Steps); si++ {
		step := &lesson.Steps[si]

		switch step.Type {
		case StepTypeTest:
			if step.Draft {
				return http.BadRequest("test %d is a draft", si+1)
			}
		case StepTypeProgramming:
			if step.Draft {
				return http.BadRequest("programming task %d is a draft", si+1)
			}
		}
	}

	return nil
}

func LessonTestFillFromRequest(vs url.Values, test *StepTest) error {
	test.Name = vs.Get("Name")

	answerKey := make([]byte, 30)
	copy(answerKey, "Answer")

	correctAnswerKey := make([]byte, 30)
	copy(correctAnswerKey, "CorrectAnswer")

	questions := vs.GetMany("Question")
	for i := 0; i < len(questions); i++ {
		if i >= len(test.Questions) {
			test.Questions = append(test.Questions, Question{})
		}
		question := &test.Questions[i]
		question.Name = questions[i]

		n := slices.PutInt(answerKey[len("Answer"):], i)
		answers := vs.GetMany(unsafe.String(unsafe.SliceData(answerKey), len("Answer")+n))
		for j := 0; j < len(answers); j++ {
			if j >= len(question.Answers) {
				question.Answers = append(question.Answers, "")
			}
			question.Answers[j] = answers[j]
		}
		question.Answers = question.Answers[:len(answers)]

		n = slices.PutInt(correctAnswerKey[len("CorrectAnswer"):], i)
		correctAnswers := vs.GetMany(unsafe.String(unsafe.SliceData(correctAnswerKey), len("CorrectAnswer")+n))
		for j := 0; j < len(correctAnswers); j++ {
			if j >= len(question.CorrectAnswers) {
				question.CorrectAnswers = append(question.CorrectAnswers, 0)
			}

			var err error
			question.CorrectAnswers[j], err = GetValidIndex(correctAnswers[j], len(question.Answers))
			if err != nil {
				return http.ClientError(err)
			}
		}
		question.CorrectAnswers = question.CorrectAnswers[:len(correctAnswers)]
	}
	test.Questions = test.Questions[:len(questions)]

	return nil
}

func LessonTestVerify(test *StepTest) error {
	if !strings.LengthInRange(test.Name, MinStepNameLen, MaxStepNameLen) {
		return http.BadRequest("test name length must be between %d and %d characters long", MinStepNameLen, MaxStepNameLen)
	}

	for i := 0; i < len(test.Questions); i++ {
		question := &test.Questions[i]

		if !strings.LengthInRange(question.Name, MinQuestionLen, MaxQuestionLen) {
			return http.BadRequest("question %d: title length must be between %d and %d characters long", i+1, MinQuestionLen, MaxQuestionLen)
		}

		for j := 0; j < len(question.Answers); j++ {
			if !strings.LengthInRange(question.Answers[j], MinAnswerLen, MaxAnswerLen) {
				return http.BadRequest("question %d: answer %d: length must be between %d and %d characters long", i+1, j+1, MinAnswerLen, MaxAnswerLen)
			}
		}

		if len(question.CorrectAnswers) == 0 {
			return http.BadRequest("question %d: select at least one correct answer", i+1)
		}
	}

	return nil
}

func LessonProgrammingFillFromRequest(vs url.Values, task *StepProgramming) error {
	task.Name = vs.Get("Name")
	task.Description = vs.Get("Description")

	for i := 0; i < len(CheckKeys); i++ {
		checks := &task.Checks[i]

		inputs := vs.GetMany(CheckKeys[i][CheckKeyInput])
		outputs := vs.GetMany(CheckKeys[i][CheckKeyOutput])

		if len(inputs) != len(outputs) {
			return http.ClientError(nil)
		}

		for j := 0; j < len(inputs); j++ {
			if j >= len(*checks) {
				*checks = append(*checks, Check{})
			}
			check := &(*checks)[j]

			check.Input = inputs[j]
			check.Output = outputs[j]
		}
	}

	return nil
}

func LessonProgrammingVerify(task *StepProgramming) error {
	if !strings.LengthInRange(task.Name, MinStepNameLen, MaxStepNameLen) {
		return http.BadRequest("programming task name length must be between %d and %d characters long", MinStepNameLen, MaxStepNameLen)
	}

	if !strings.LengthInRange(task.Description, MinDescriptionLen, MaxDescriptionLen) {
		return http.BadRequest("programming task description length must be between %d and %d characters long", MinDescriptionLen, MaxDescriptionLen)
	}

	for i := 0; i < len(task.Checks); i++ {
		checks := task.Checks[i]

		for j := 0; j < len(checks); j++ {
			check := &checks[j]

			if !strings.LengthInRange(check.Input, MinCheckLen, MaxCheckLen) {
				return http.BadRequest("%s %d: input length must be between %d and %d characters long", CheckKeys[i][CheckKeyDisplay], j+1, MinCheckLen, MaxCheckLen)
			}

			if !strings.LengthInRange(check.Output, MinCheckLen, MaxCheckLen) {
				return http.BadRequest("%s %d: output length must be between %d and %d characters long", CheckKeys[i][CheckKeyDisplay], j+1, MinCheckLen, MaxCheckLen)
			}
		}
	}

	return nil
}

func LessonAddPageHandler(w *http.Response, r *http.Request, lesson *Lesson) error {
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
	DisplayConstraintInput(w, "text", MinNameLen, MaxNameLen, "Name", lesson.Name, true)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>Theory:<br>`)
	DisplayConstraintTextarea(w, "80", "24", MinTheoryLen, MaxTheoryLen, "Theory", lesson.Theory, true)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	for i := 0; i < len(lesson.Steps); i++ {
		step := &lesson.Steps[i]
		name := step.Name
		draft := step.Draft
		stepType := StepStringType(step)

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

		DisplayIndexedCommand(w, i, "Edit")
		DisplayIndexedCommand(w, i, "Delete")
		if len(lesson.Steps) > 1 {
			if i > 0 {
				DisplayIndexedCommand(w, i, "↑")
			}
			if i < len(lesson.Steps)-1 {
				DisplayIndexedCommand(w, i, "↓")
			}
		}

		w.AppendString(`</fieldset>`)
		w.AppendString(`<br>`)
	}

	w.AppendString(`<input type="submit" name="NextPage" value="Add test" formnovalidate> `)
	w.AppendString(`<input type="submit" name="NextPage" value="Add programming task" formnovalidate>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<input type="submit" name="NextPage" value="Next">`)

	w.AppendString(`</form>`)

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func LessonAddTestPageHandler(w *http.Response, r *http.Request, test *StepTest) error {
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
	DisplayConstraintInput(w, "text", MinStepNameLen, MaxStepNameLen, "Name", test.Name, true)
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
		DisplayConstraintInput(w, "text", MinQuestionLen, MaxQuestionLen, "Question", question.Name, true)
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

			DisplayConstraintIndexedInput(w, "text", MinAnswerLen, MaxAnswerLen, "Answer", i, answer, true)

			if len(question.Answers) > 1 {
				DisplayDoublyIndexedCommand(w, i, j, "-")
				if j > 0 {
					DisplayDoublyIndexedCommand(w, i, j, "↑")
				}
				if j < len(question.Answers)-1 {
					DisplayDoublyIndexedCommand(w, i, j, "↓")
				}
			}

			w.AppendString(`</li>`)
		}
		w.AppendString(`</ol>`)

		DisplayIndexedCommand(w, i, "Add another answer")
		if len(test.Questions) > 1 {
			w.AppendString(`<br><br>`)
			DisplayIndexedCommand(w, i, "Delete")
			if i > 0 {
				DisplayIndexedCommand(w, i, "↑")
			}
			if i < len(test.Questions)-1 {
				DisplayIndexedCommand(w, i, "↓")
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

func LessonAddProgrammingDisplayChecks(w *http.Response, task *StepProgramming, checkType CheckType) {
	checks := task.Checks[checkType]

	w.AppendString(`<ol>`)
	for i := 0; i < len(checks); i++ {
		check := &checks[i]

		w.AppendString(`<li>`)

		w.AppendString(`<label>Input: `)
		DisplayConstraintTextarea(w, "", "1", MinCheckLen, MaxCheckLen, CheckKeys[checkType][CheckKeyInput], check.Input, true)
		w.AppendString(`</label> `)

		w.AppendString(`<label>output: `)
		DisplayConstraintTextarea(w, "", "1", MinCheckLen, MaxCheckLen, CheckKeys[checkType][CheckKeyOutput], check.Output, true)
		w.AppendString(`</label>`)

		DisplayDoublyIndexedCommand(w, i, int(checkType), "-")
		if len(checks) > 1 {
			if i > 0 {
				DisplayDoublyIndexedCommand(w, i, int(checkType), "↑")
			}
			if i < len(checks)-1 {
				DisplayDoublyIndexedCommand(w, i, int(checkType), "↓")
			}
		}

		w.AppendString(`</li>`)
	}
	w.AppendString(`</ol>`)
}

func LessonAddProgrammingPageHandler(w *http.Response, r *http.Request, task *StepProgramming) error {
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
	DisplayConstraintInput(w, "text", MinStepNameLen, MaxStepNameLen, "Name", task.Name, true)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>Description:<br>`)
	DisplayConstraintTextarea(w, "80", "24", MinDescriptionLen, MaxDescriptionLen, "Description", task.Description, true)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<h3>Examples</h3>`)
	LessonAddProgrammingDisplayChecks(w, task, CheckTypeExample)
	w.AppendString(`<input type="submit" name="Command" value="Add example" formnovalidate>`)

	w.AppendString(`<h3>Tests</h3>`)
	LessonAddProgrammingDisplayChecks(w, task, CheckTypeTest)
	w.AppendString(`<input type="submit" name="Command" value="Add test" formnovalidate>`)

	w.AppendString(`<br><br>`)
	w.AppendString(`<input type="submit" name="NextPage" value="Continue">`)

	w.AppendString(`</form>`)

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func LessonAddStepPageHandler(w *http.Response, r *http.Request, step *Step) error {
	switch step.Type {
	default:
		panic("invalid step type")
	case StepTypeTest:
		test, _ := Step2Test(step)
		return LessonAddTestPageHandler(w, r, test)
	case StepTypeProgramming:
		task, _ := Step2Programming(step)
		return LessonAddProgrammingPageHandler(w, r, task)
	}
}

func LessonAddHandleCommand(w *http.Response, r *http.Request, lessons []database.ID, currentPage, k, command string) error {
	var lesson Lesson

	/* TODO(anton2920): pass these as parameters. */
	pindex, spindex, sindex, ssindex, err := GetIndicies(k[len("Command"):])
	if err != nil {
		return http.ClientError(err)
	}

	switch currentPage {
	default:
		return http.ClientError(nil)
	case "Lesson":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), len(lessons))
		if err != nil {
			return http.ClientError(err)
		}
		if err := GetLessonByID(lessons[li], &lesson); err != nil {
			return http.ServerError(err)
		}
		defer SaveLesson(&lesson)

		switch command {
		case "Delete":
			lesson.Steps = RemoveAtIndex(lesson.Steps, pindex)
		case "Edit":
			if (pindex < 0) || (pindex >= len(lesson.Steps)) {
				return http.ClientError(nil)
			}
			step := &lesson.Steps[pindex]
			step.Draft = true

			r.Form.Set("StepIndex", spindex)
			return LessonAddStepPageHandler(w, r, step)
		case "↑", "^|":
			MoveUp(lesson.Steps, pindex)
		case "↓", "|v":
			MoveDown(lesson.Steps, pindex)
		}

		return LessonAddPageHandler(w, r, &lesson)
	case "Test":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), len(lessons))
		if err != nil {
			return http.ClientError(err)
		}
		if err := GetLessonByID(lessons[li], &lesson); err != nil {
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
			return http.ClientError(err)
		}

		switch command {
		case "Add another answer":
			if (pindex < 0) || (pindex >= len(test.Questions)) {
				return http.ClientError(nil)
			}
			question := &test.Questions[pindex]
			question.Answers = append(question.Answers, "")
		case "-": /* remove answer */
			if (pindex < 0) || (pindex >= len(test.Questions)) {
				return http.ClientError(nil)
			}
			question := &test.Questions[pindex]

			if (sindex < 0) || (sindex >= len(question.Answers)) {
				return http.ClientError(nil)
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
					return http.ClientError(nil)
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
					return http.ClientError(nil)
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

		return LessonAddTestPageHandler(w, r, test)

	case "Programming":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), len(lessons))
		if err != nil {
			return http.ClientError(err)
		}
		if err := GetLessonByID(lessons[li], &lesson); err != nil {
			return http.ServerError(err)
		}
		defer SaveLesson(&lesson)

		si, err := GetValidIndex(r.Form.Get("StepIndex"), len(lesson.Steps))
		if err != nil {
			return http.ClientError(err)
		}
		task, err := Step2Programming(&lesson.Steps[si])
		if err != nil {
			return http.ClientError(nil)
		}

		if err := LessonProgrammingFillFromRequest(r.Form, task); err != nil {
			return http.ClientError(err)
		}

		switch command {
		case "Add example":
			task.Checks[CheckTypeExample] = append(task.Checks[CheckTypeExample], Check{})
		case "Add test":
			task.Checks[CheckTypeTest] = append(task.Checks[CheckTypeTest], Check{})
		case "-":
			if (sindex < 0) || (sindex >= len(task.Checks)) {
				return http.ClientError(nil)
			}
			task.Checks[sindex] = RemoveAtIndex(task.Checks[sindex], pindex)
		case "↑", "^|":
			if (sindex < 0) || (sindex >= len(task.Checks)) {
				return http.ClientError(nil)
			}
			MoveUp(task.Checks[sindex], pindex)
		case "↓", "|v":
			if (sindex < 0) || (sindex >= len(task.Checks)) {
				return http.ClientError(nil)
			}
			MoveDown(task.Checks[sindex], pindex)
		}

		return LessonAddProgrammingPageHandler(w, r, task)
	}
}
