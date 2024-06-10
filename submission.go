package main

import (
	"fmt"
	"time"
	"unsafe"

	"github.com/anton2920/gofa/errors"
	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/net/url"
	"github.com/anton2920/gofa/slices"
	"github.com/anton2920/gofa/strings"
	"github.com/anton2920/gofa/syscall"
)

type (
	SubmittedQuestion struct {
		SelectedAnswers []int
	}

	ProgrammingLanguage struct {
		Name         string
		Compiler     string
		CompilerArgs []string
		Runner       string
		RunnerArgs   []string
		SourceFile   string
		Executable   string
		Available    bool
	}

	SubmittedCommon struct {
		Type   SubmittedType
		Flags  SubmittedFlag
		Status SubmissionCheckStatus
		Error  string

		Step Step
	}
	SubmittedTest struct {
		SubmittedCommon
		SubmittedQuestions []SubmittedQuestion

		Scores []int
	}
	SubmittedProgramming struct {
		SubmittedCommon
		LanguageID int32
		Solution   string

		Scores   [2][]int
		Messages [2][]string
	}
	SubmittedStep/* union */ struct {
		SubmittedCommon

		/* TODO(anton2920): garbage collector cannot see pointers inside. */
		Data [max(unsafe.Sizeof(stdt), unsafe.Sizeof(stdp)) - unsafe.Sizeof(stdc)]byte
	}

	Submission struct {
		ID       int32
		Flags    int32
		UserID   int32
		LessonID int32

		Status SubmissionCheckStatus

		StartedAt      int64
		FinishedAt     int64
		SubmittedSteps []SubmittedStep

		Data [16384]byte
	}
)

type SubmittedType int32

const (
	SubmittedTypeTest        SubmittedType = SubmittedType(StepTypeTest)
	SubmittedTypeProgramming               = SubmittedType(StepTypeProgramming)
)

type SubmittedFlag int32

const (
	SubmittedStepSkipped SubmittedFlag = iota
	SubmittedStepDraft
	SubmittedStepPassed
)

const (
	SubmissionActive int32 = iota
	SubmissionDraft
)

const (
	MinSolutionLen = 1
	MaxSolutionLen = 1024
)

var (
	stdc SubmittedCommon
	stdt SubmittedTest
	stdp SubmittedProgramming
)

var ProgrammingLanguages = []ProgrammingLanguage{
	{"c", "cc", nil, "", nil, "main.c", "./a.out", true},
	{"c++", "c++", nil, "", nil, "main.cpp", "./a.out", true},
	{"go", "sh", []string{"-c", "/usr/local/bin/go-build"}, "", nil, "main.go", "./main", true},
	{"php", "php", []string{"-l"}, "php", nil, "main.php", "", true},
	{"python3", "python3", []string{"-c", `import ast; ast.parse(open("main.py").read())`}, "python3", nil, "main.py", "", true},
}

func Submitted2Test(submittedStep *SubmittedStep) (*SubmittedTest, error) {
	if submittedStep.Type != SubmittedTypeTest {
		return nil, errors.New("invalid submitted type for test")
	}
	return (*SubmittedTest)(unsafe.Pointer(submittedStep)), nil
}

func Submitted2Programming(submittedStep *SubmittedStep) (*SubmittedProgramming, error) {
	if submittedStep.Type != SubmittedTypeProgramming {
		return nil, errors.New("invalid submitted type for programming")
	}
	return (*SubmittedProgramming)(unsafe.Pointer(submittedStep)), nil
}

func CreateSubmission(db *Database, submission *Submission) error {
	var err error

	submission.ID, err = IncrementNextID(db.SubmissionsFile)
	if err != nil {
		return fmt.Errorf("failed to increment submission ID: %w", err)
	}

	return SaveSubmission(db, submission)
}

func DBSubmitted2Submitted(submittedStep *SubmittedStep, data *byte) {
	submittedStep.Error = Offset2String(submittedStep.Error, data)
	DBStep2Step(&submittedStep.Step, data)

	switch submittedStep.Type {
	default:
		panic("invalid step type")
	case SubmittedTypeTest:
		submittedTest, _ := Submitted2Test(submittedStep)

		submittedTest.SubmittedQuestions = Offset2Slice(submittedTest.SubmittedQuestions, data)
		for i := 0; i < len(submittedTest.SubmittedQuestions); i++ {
			submittedQuestion := &submittedTest.SubmittedQuestions[i]
			submittedQuestion.SelectedAnswers = Offset2Slice(submittedQuestion.SelectedAnswers, data)
		}
		submittedTest.Scores = Offset2Slice(submittedTest.Scores, data)
	case SubmittedTypeProgramming:
		submittedTask, _ := Submitted2Programming(submittedStep)

		submittedTask.Solution = Offset2String(submittedTask.Solution, data)

		for i := 0; i < 2; i++ {
			submittedTask.Scores[i] = Offset2Slice(submittedTask.Scores[i], data)

			submittedTask.Messages[i] = Offset2Slice(submittedTask.Messages[i], data)
			for j := 0; j < len(submittedTask.Messages[i]); j++ {
				submittedTask.Messages[i][j] = Offset2String(submittedTask.Messages[i][j], data)
			}
		}
	}

}

func DBSubmission2Submission(submission *Submission) {
	data := &submission.Data[0]

	submission.SubmittedSteps = Offset2Slice(submission.SubmittedSteps, data)
	for i := 0; i < len(submission.SubmittedSteps); i++ {
		DBSubmitted2Submitted(&submission.SubmittedSteps[i], data)
	}
}

func GetSubmissionByID(db *Database, id int32, submission *Submission) error {
	size := int(unsafe.Sizeof(*submission))
	offset := int64(int(id)*size) + DataOffset

	n, err := syscall.Pread(db.SubmissionsFile, unsafe.Slice((*byte)(unsafe.Pointer(submission)), size), offset)
	if err != nil {
		return fmt.Errorf("failed to read submission from DB: %w", err)
	}
	if n < size {
		return DBNotFound
	}

	DBSubmission2Submission(submission)
	return nil
}

func GetSubmissions(db *Database, pos *int64, submissions []Submission) (int, error) {
	if *pos < DataOffset {
		*pos = DataOffset
	}
	size := int(unsafe.Sizeof(submissions[0]))

	n, err := syscall.Pread(db.SubmissionsFile, unsafe.Slice((*byte)(unsafe.Pointer(unsafe.SliceData(submissions))), len(submissions)*size), *pos)
	if err != nil {
		return 0, fmt.Errorf("failed to read submission from DB: %w", err)
	}
	*pos += int64(n)

	n /= size
	for i := 0; i < n; i++ {
		DBSubmission2Submission(&submissions[i])
	}

	return n, nil
}

func Submitted2DBSubmitted(ds *SubmittedStep, ss *SubmittedStep, data []byte, n int) int {
	ds.Flags = ss.Flags
	ds.Status = ss.Status

	n += Step2DBStep(&ds.Step, &ss.Step, data, n)
	n += String2DBString(&ds.Error, ss.Error, data, n)

	switch ss.Type {
	default:
		panic("invalid step type")
	case SubmittedTypeTest:
		st, _ := Submitted2Test(ss)

		ds.Type = SubmittedTypeTest
		dt, _ := Submitted2Test(ds)

		dt.SubmittedQuestions = make([]SubmittedQuestion, len(st.SubmittedQuestions))
		for i := 0; i < len(st.SubmittedQuestions); i++ {
			sq := &st.SubmittedQuestions[i]
			dq := &dt.SubmittedQuestions[i]
			n += Slice2DBSlice(&dq.SelectedAnswers, sq.SelectedAnswers, data, n)
		}
		n += Slice2DBSlice(&dt.SubmittedQuestions, dt.SubmittedQuestions, data, n)

		n += Slice2DBSlice(&dt.Scores, st.Scores, data, n)
	case SubmittedTypeProgramming:
		st, _ := Submitted2Programming(ss)

		ds.Type = SubmittedTypeProgramming
		dt, _ := Submitted2Programming(ds)

		dt.LanguageID = st.LanguageID

		n += String2DBString(&dt.Solution, st.Solution, data, n)

		for i := 0; i < 2; i++ {
			n += Slice2DBSlice(&dt.Scores[i], st.Scores[i], data, n)

			dt.Messages[i] = make([]string, len(st.Messages[i]))
			for j := 0; j < len(st.Messages[i]); j++ {
				n += String2DBString(&dt.Messages[i][j], st.Messages[i][j], data, n)
			}
			n += Slice2DBSlice(&dt.Messages[i], dt.Messages[i], data, n)
		}
	}

	return n
}

func SaveSubmission(db *Database, submission *Submission) error {
	var submissionDB Submission
	var n int

	size := int(unsafe.Sizeof(*submission))
	offset := int64(int(submission.ID)*size) + DataOffset
	data := unsafe.Slice(&submissionDB.Data[0], len(submissionDB.Data))

	submissionDB.ID = submission.ID
	submissionDB.Flags = submission.Flags
	submissionDB.Status = submission.Status
	submissionDB.UserID = submission.UserID
	submissionDB.LessonID = submission.LessonID
	submissionDB.StartedAt = submission.StartedAt
	submissionDB.FinishedAt = submission.FinishedAt

	/* TODO(anton2920): save up to a sizeof(lesson.Data). */
	submissionDB.SubmittedSteps = make([]SubmittedStep, len(submission.SubmittedSteps))
	for i := 0; i < len(submission.SubmittedSteps); i++ {
		n += Submitted2DBSubmitted(&submissionDB.SubmittedSteps[i], &submission.SubmittedSteps[i], data, n)
	}
	n += Slice2DBSlice(&submissionDB.SubmittedSteps, submissionDB.SubmittedSteps, data, n)

	_, err := syscall.Pwrite(db.SubmissionsFile, unsafe.Slice((*byte)(unsafe.Pointer(&submissionDB)), size), offset)
	if err != nil {
		return fmt.Errorf("failed to write submission to DB: %w", err)
	}

	return nil
}

func GetSubmittedStepScore(submittedStep *SubmittedStep) int {
	if submittedStep.Flags == SubmittedStepSkipped {
		return 0
	}

	var scores []int
	switch submittedStep.Type {
	default:
		panic("invalid step type")
	case SubmittedTypeTest:
		submittedTest, _ := Submitted2Test(submittedStep)
		scores = submittedTest.Scores
	case SubmittedTypeProgramming:
		submittedTask, _ := Submitted2Programming(submittedStep)
		scores = submittedTask.Scores[CheckTypeTest]
	}

	var score int
	for i := 0; i < len(scores); i++ {
		score += scores[i]
	}
	return score
}

func GetStepMaximumScore(step *Step) int {
	var maximum int
	switch step.Type {
	default:
		panic("invalid step type")
	case StepTypeTest:
		test, _ := Step2Test(step)
		maximum = len(test.Questions)
	case StepTypeProgramming:
		task, _ := Step2Programming(step)
		maximum = len(task.Checks[CheckTypeTest])
	}
	return maximum
}

func DisplaySubmittedStepScore(w *http.Response, submittedStep *SubmittedStep) {
	w.AppendString(`<p>Score: `)
	w.WriteInt(GetSubmittedStepScore(submittedStep))
	w.AppendString(`/`)
	w.WriteInt(GetStepMaximumScore(&submittedStep.Step))
	w.AppendString(`</p>`)
}

func DisplaySubmissionTotalScore(w *http.Response, submission *Submission) {
	var score, maximum int
	for i := 0; i < len(submission.SubmittedSteps); i++ {
		score += GetSubmittedStepScore(&submission.SubmittedSteps[i])
		maximum += GetStepMaximumScore(&submission.SubmittedSteps[i].Step)
	}

	w.WriteInt(score)
	w.AppendString(`/`)
	w.WriteInt(maximum)
}

func DisplaySubmissionLink(w *http.Response, submission *Submission) {
	var user User

	if err := GetUserByID(DB2, submission.UserID, &user); err != nil {
		/* TODO(anton2920): report error. */
	}

	w.AppendString(`<a href="/submission/`)
	w.WriteInt(int(submission.ID))
	w.AppendString(`">`)
	w.WriteHTMLString(user.LastName)
	w.AppendString(` `)
	w.WriteHTMLString(user.FirstName)
	w.AppendString(` (`)
	switch submission.Status {
	case SubmissionCheckPending:
		w.AppendString(`<i>pending</i>`)
	case SubmissionCheckInProgress:
		w.AppendString(`<i>in progress</i>`)
	case SubmissionCheckDone:
		DisplaySubmissionTotalScore(w, submission)
	}
	w.AppendString(`)`)
	w.AppendString(`</a>`)
}

func SubmissionDisplayLanguageSelect(w *http.Response, submittedTask *SubmittedProgramming, enabled bool) {
	w.AppendString(`<select name="LanguageID"`)
	if !enabled {
		w.AppendString(` disabled`)
	}
	w.AppendString(`>`)
	for i := int32(0); i < int32(len(ProgrammingLanguages)); i++ {
		lang := &ProgrammingLanguages[i]

		w.AppendString(`<option value="`)
		w.WriteInt(int(i))
		w.AppendString(`"`)
		if i == submittedTask.LanguageID {
			w.AppendString(` selected`)
		}
		w.AppendString(`>`)
		w.AppendString(lang.Name)
		w.AppendString(`</option>`)
	}
	w.AppendString(`</select>`)
}

func SubmissionPageHandler(w *http.Response, r *http.Request) error {
	var submission Submission
	var subject Subject
	var lesson Lesson
	var user User

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}

	id, err := GetIDFromURL(r.URL, "/submission/")
	if err != nil {
		return http.ClientError(err)
	}
	if err := GetSubmissionByID(DB2, int32(id), &submission); err != nil {
		if err == DBNotFound {
			return http.NotFound("lesson with this ID does not exist")
		}
		return http.ServerError(err)
	}

	if err := GetLessonByID(DB2, submission.LessonID, &lesson); err != nil {
		return http.ServerError(err)
	}

	if err := GetSubjectByID(DB2, lesson.ContainerID, &subject); err != nil {
		return http.ServerError(err)
	}
	who, err := WhoIsUserInSubject(session.ID, &subject)
	if err != nil {
		return http.ServerError(err)
	}
	if who == SubjectUserNone {
		return http.ForbiddenError
	}
	teacher := (who == SubjectUserAdmin) || (who == SubjectUserTeacher)

	if err := GetUserByID(DB2, submission.UserID, &user); err != nil {
		return http.ServerError(err)
	}

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>`)
	w.AppendString(`Submission for `)
	w.WriteHTMLString(subject.Name)
	w.AppendString(`: `)
	w.WriteHTMLString(lesson.Name)
	w.AppendString(` by `)
	w.WriteHTMLString(user.LastName)
	w.AppendString(` `)
	w.WriteHTMLString(user.FirstName)
	w.AppendString(`</title></head>`)
	w.AppendString(`<body>`)

	w.AppendString(`<h1>`)
	w.AppendString(`Submission for `)
	w.WriteHTMLString(subject.Name)
	w.AppendString(`: `)
	w.WriteHTMLString(lesson.Name)
	w.AppendString(` by `)
	w.WriteHTMLString(user.LastName)
	w.AppendString(` `)
	w.WriteHTMLString(user.FirstName)
	w.AppendString(`</h1>`)

	w.AppendString(`<form style="min-width: 300px; max-width: max-content;" method="POST" action="/submission/results">`)

	w.AppendString(`<input type="hidden" name="ID" value="`)
	w.WriteInt(int(submission.ID))
	w.AppendString(`">`)

	for i := 0; i < len(submission.SubmittedSteps); i++ {
		submittedStep := &submission.SubmittedSteps[i]

		if i > 0 {
			w.AppendString(`<br>`)
		}

		w.AppendString(`<fieldset>`)

		w.AppendString(`<legend>Step #`)
		w.WriteInt(i + 1)
		w.AppendString(`</legend>`)

		w.AppendString(`<p>Name: `)
		w.WriteHTMLString(submittedStep.Step.Name)
		w.AppendString(`</p>`)

		w.AppendString(`<p>Type: `)
		w.AppendString(StepStringType(&submittedStep.Step))
		w.AppendString(`</p>`)

		if submittedStep.Flags == SubmittedStepSkipped {
			DisplaySubmittedStepScore(w, submittedStep)

			w.AppendString(`<p><i>This step has been skipped.</i></p>`)
		} else {
			switch submittedStep.Status {
			case SubmissionCheckPending:
				w.AppendString(`<p><i>Verification is pending...</i></p>`)
			case SubmissionCheckInProgress:
				w.AppendString(`<p><i>Verification is in progress...</i></p>`)
			case SubmissionCheckDone:
				DisplaySubmittedStepScore(w, submittedStep)
				DisplayErrorMessage(w, submittedStep.Error)

				DisplayIndexedCommand(w, i, "Open")
				if teacher {
					DisplayIndexedCommand(w, i, "Re-check")
				}
			}
		}

		w.AppendString(`</fieldset>`)
	}

	switch submission.Status {
	case SubmissionCheckPending:
		w.AppendString(`<p><i>Verification is pending...</i></p>`)
	case SubmissionCheckInProgress:
		w.AppendString(`<p><i>Verification is in progress...</i></p>`)
	case SubmissionCheckDone:
		w.AppendString(`<p>Total score: `)
		DisplaySubmissionTotalScore(w, &submission)
		w.AppendString(`</p>`)
		if teacher {
			w.AppendString(`<input type="submit" name="Command" value="Re-check">`)
		}
	}
	w.AppendString(`</form>`)

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil

}

func SubmissionResultsTestPageHandler(w *http.Response, r *http.Request, submittedTest *SubmittedTest) error {
	test, _ := Step2Test(&submittedTest.Step)
	teacher := r.Form.Get("Teacher") != ""

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>`)
	w.AppendString(`Submitted test: `)
	w.WriteHTMLString(test.Name)
	w.AppendString(`</title></head>`)
	w.AppendString(`<body>`)

	w.AppendString(`<h1>`)
	w.AppendString(`Submitted test: `)
	w.WriteHTMLString(test.Name)
	w.AppendString(`</h1>`)

	if teacher {
		w.AppendString(`<p><i>Note: answers marked with [x] are correct.</i></p>`)
	}

	for i := 0; i < len(test.Questions); i++ {
		question := &test.Questions[i]

		w.AppendString(`<fieldset>`)
		w.AppendString(`<legend>Question #`)
		w.WriteInt(i + 1)
		w.AppendString(`</legend>`)
		w.AppendString(`<p>`)
		w.WriteHTMLString(question.Name)
		w.AppendString(`</p>`)

		w.AppendString(`<ol>`)
		for j := 0; j < len(question.Answers); j++ {
			answer := question.Answers[j]

			if j > 0 {
				w.AppendString(`<br>`)
			}

			w.AppendString(`<li>`)

			w.AppendString(`<input type="`)
			if len(question.CorrectAnswers) > 1 {
				w.AppendString(`checkbox`)
			} else {
				w.AppendString(`radio`)
			}
			w.AppendString(`" name="SelectedAnswer`)
			w.WriteInt(i)
			w.AppendString(`" value="`)
			w.WriteInt(j)
			w.AppendString(`"`)

			for k := 0; k < len(submittedTest.SubmittedQuestions[i].SelectedAnswers); k++ {
				selectedAnswer := submittedTest.SubmittedQuestions[i].SelectedAnswers[k]
				if j == selectedAnswer {
					w.AppendString(` checked`)
					break
				}
			}

			w.AppendString(` disabled> `)

			w.AppendString(`<span>`)
			w.WriteHTMLString(answer)
			w.AppendString(`</span>`)

			if teacher {
				for k := 0; k < len(question.CorrectAnswers); k++ {
					correctAnswer := question.CorrectAnswers[k]
					if j == correctAnswer {
						w.AppendString(` <span>[x]</span>`)
						break
					}
				}
			}

			w.AppendString(`</li>`)
		}
		w.AppendString(`</ol>`)

		w.AppendString(`<p>Score: `)
		w.WriteInt(submittedTest.Scores[i])
		w.AppendString(`/1</p>`)

		w.AppendString(`</fieldset>`)
		w.AppendString(`<br>`)
	}

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func SubmissionResultsProgrammingDisplayChecks(w *http.Response, submittedTask *SubmittedProgramming, checkType CheckType) {
	task, _ := Step2Programming(&submittedTask.Step)
	scores := submittedTask.Scores[checkType]
	messages := submittedTask.Messages[checkType]

	w.AppendString(`<ol>`)
	for i := 0; i < len(task.Checks[checkType]); i++ {
		check := &task.Checks[checkType][i]
		score := scores[i]
		message := messages[i]

		w.AppendString(`<li>`)

		w.AppendString(`<label>Input: `)

		w.AppendString(`<textarea rows="1" readonly>`)
		w.WriteHTMLString(check.Input)
		w.AppendString(`</textarea>`)

		w.AppendString(`</label> `)

		w.AppendString(`<label>output: `)

		w.AppendString(`<textarea rows="1" readonly>`)
		w.WriteHTMLString(check.Output)
		w.AppendString(`</textarea>`)

		w.AppendString(`</label> `)

		w.AppendString(`<span>score: `)
		w.WriteInt(score)
		w.AppendString(`/1</span>`)

		if message != "" {
			w.AppendString(` <span>`)
			w.WriteHTMLString(message)
			w.AppendString(`</span>`)
		}

		w.AppendString(`</li>`)
	}
	w.AppendString(`</ol>`)
}

func SubmissionResultsProgrammingPageHandler(w *http.Response, r *http.Request, submittedTask *SubmittedProgramming) error {
	task, _ := Step2Programming(&submittedTask.Step)
	teacher := r.Form.Get("Teacher") != ""

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>`)
	w.AppendString(`Submitted programming task: `)
	w.WriteHTMLString(task.Name)
	w.AppendString(`</title></head>`)
	w.AppendString(`<body>`)

	w.AppendString(`<h1>`)
	w.AppendString(`Submitted programming task: `)
	w.WriteHTMLString(task.Name)
	w.AppendString(`</h1>`)

	DisplayErrorMessage(w, r.Form.Get("Error"))

	w.AppendString(`<h2>Description</h2>`)
	w.AppendString(`<p>`)
	w.WriteHTMLString(task.Description)
	w.AppendString(`</p>`)

	w.AppendString(`<h2>Examples</h2>`)
	SubmissionNewDisplayProgrammingChecks(w, task, CheckTypeExample)

	w.AppendString(`<h2>Solution</h2>`)

	w.AppendString(`<label>Programming language: `)
	SubmissionDisplayLanguageSelect(w, submittedTask, false)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<textarea cols="80" rows="24" name="Solution" readonly>`)
	w.WriteHTMLString(submittedTask.Solution)
	w.AppendString(`</textarea>`)

	if teacher {
		w.AppendString(`<h2>Tests</h2>`)
		SubmissionResultsProgrammingDisplayChecks(w, submittedTask, CheckTypeTest)
	}

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func SubmissionResultsStepPageHandler(w *http.Response, r *http.Request, submittedStep *SubmittedStep) error {
	switch submittedStep.Type {
	default:
		panic("invalid step type")
	case SubmittedTypeTest:
		submittedTest, _ := Submitted2Test(submittedStep)
		return SubmissionResultsTestPageHandler(w, r, submittedTest)
	case SubmittedTypeProgramming:
		submittedTask, _ := Submitted2Programming(submittedStep)
		return SubmissionResultsProgrammingPageHandler(w, r, submittedTask)
	}
}

func SubmissionResultsHandleCommand(w *http.Response, r *http.Request, submission *Submission, k, command string) error {
	pindex, spindex, _, _, err := GetIndicies(k[len("Command"):])
	if err != nil {
		return http.ClientError(err)
	}

	switch command {
	default:
		return http.ClientError(nil)
	case "Open":
		if (pindex < 0) || (pindex >= len(submission.SubmittedSteps)) {
			return http.ClientError(nil)
		}
		submittedStep := &submission.SubmittedSteps[pindex]

		return SubmissionResultsStepPageHandler(w, r, submittedStep)
	case "Re-check":
		if spindex != "" {
			if (pindex < 0) || (pindex >= len(submission.SubmittedSteps)) {
				return http.ClientError(nil)
			}
			submittedStep := &submission.SubmittedSteps[pindex]

			submission.Status = SubmissionCheckPending
			submittedStep.Status = SubmissionCheckPending
		} else {
			submission.Status = SubmissionCheckPending
			for i := 0; i < len(submission.SubmittedSteps); i++ {
				submission.SubmittedSteps[i].Status = SubmissionCheckPending
			}
		}
		if err := SaveSubmission(DB2, submission); err != nil {
			return http.ServerError(err)
		}

		SubmissionVerifyChannel <- submission.ID

		w.RedirectID("/submission/", int(submission.ID), http.StatusSeeOther)
		return nil
	}
}

func SubmissionResultsPageHandler(w *http.Response, r *http.Request) error {
	var submission Submission
	var subject Subject
	var lesson Lesson

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}

	if err := r.ParseForm(); err != nil {
		return http.ClientError(err)
	}

	submissionID, err := r.Form.GetInt("ID")
	if err != nil {
		return http.ClientError(err)
	}
	if err := GetSubmissionByID(DB2, int32(submissionID), &submission); err != nil {
		if err == DBNotFound {
			return http.NotFound("submission with this ID does not exist")
		}
		return http.ServerError(err)
	}

	if err := GetLessonByID(DB2, submission.LessonID, &lesson); err != nil {
		return http.ServerError(err)
	}

	if err := GetSubjectByID(DB2, lesson.ContainerID, &subject); err != nil {
		return http.ServerError(err)
	}
	who, err := WhoIsUserInSubject(session.ID, &subject)
	if err != nil {
		return http.ServerError(err)
	}
	switch who {
	default:
		return http.ForbiddenError
	case SubjectUserAdmin, SubjectUserTeacher:
		r.Form.Set("Teacher", "yay")
	case SubjectUserStudent:
		r.Form.Set("Teacher", "")
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
			return SubmissionResultsHandleCommand(w, r, &submission, k, v)
		}
	}

	return http.ClientError(nil)
}

func SubmissionNewVerify(submission *Submission) error {
	empty := true
	for i := 0; i < len(submission.SubmittedSteps); i++ {
		submittedStep := &submission.SubmittedSteps[i]
		if submittedStep.Flags != SubmittedStepSkipped {
			empty = false

			if submittedStep.Flags == SubmittedStepDraft {
				return http.BadRequest("step %d is still a draft", i+1)
			}
		}
	}
	if empty {
		return http.BadRequest("you have to pass at least one step")
	}
	return nil
}

func SubmissionNewTestFillFromRequest(vs url.Values, submittedTest *SubmittedTest) error {
	test, _ := Step2Test(&submittedTest.Step)

	selectedAnswerKey := make([]byte, 30)
	copy(selectedAnswerKey, "SelectedAnswer")

	for i := 0; i < len(submittedTest.SubmittedQuestions); i++ {
		submittedQuestion := &submittedTest.SubmittedQuestions[i]
		question := &test.Questions[i]

		n := slices.PutInt(selectedAnswerKey[len("SelectedAnswer"):], i)
		selectedAnswers := vs.GetMany(unsafe.String(unsafe.SliceData(selectedAnswerKey), len("SelectedAnswer")+n))

		for j := 0; j < len(selectedAnswers); j++ {
			if j >= len(submittedQuestion.SelectedAnswers) {
				submittedQuestion.SelectedAnswers = append(submittedQuestion.SelectedAnswers, 0)
			}

			var err error
			submittedQuestion.SelectedAnswers[j], err = GetValidIndex(selectedAnswers[j], len(question.Answers))
			if err != nil {
				return http.ClientError(err)
			}
		}
		submittedQuestion.SelectedAnswers = submittedQuestion.SelectedAnswers[:len(selectedAnswers)]
	}

	return nil
}

func SubmissionNewTestVerify(submittedTest *SubmittedTest) error {
	test, _ := Step2Test(&submittedTest.Step)

	for i := 0; i < len(submittedTest.SubmittedQuestions); i++ {
		submittedQuestion := &submittedTest.SubmittedQuestions[i]
		question := &test.Questions[i]

		if len(submittedQuestion.SelectedAnswers) == 0 {
			return http.BadRequest("question %d: select at least one answer", i+1)
		}
		if (len(question.CorrectAnswers) == 1) && (len(submittedQuestion.SelectedAnswers) > 1) {
			return http.ClientError(nil)
		}
	}

	return nil
}

func SubmissionNewProgrammingFillFromRequest(vs url.Values, submittedTask *SubmittedProgramming) error {
	id, err := GetValidIndex(vs.Get("LanguageID"), len(ProgrammingLanguages))
	if err != nil {
		return http.ClientError(err)
	}

	submittedTask.LanguageID = int32(id)
	submittedTask.Solution = vs.Get("Solution")
	return nil
}

func SubmissionNewProgrammingVerify(submittedTask *SubmittedProgramming) error {
	if !ProgrammingLanguages[submittedTask.LanguageID].Available {
		return http.BadRequest("selected language is not available")
	}

	if !strings.LengthInRange(submittedTask.Solution, MinSolutionLen, MaxSolutionLen) {
		return http.BadRequest("solution length must be between %d and %d characters long", MinSolutionLen, MaxSolutionLen)
	}

	return nil
}

func SubmissionNewMainPageHandler(w *http.Response, r *http.Request, submission *Submission) error {
	var lesson Lesson

	if err := GetLessonByID(DB2, submission.LessonID, &lesson); err != nil {
		return http.ServerError(err)
	}

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>`)
	w.AppendString(`Evaluation for `)
	w.WriteHTMLString(lesson.Name)
	w.AppendString(`</title></head>`)
	w.AppendString(`<body>`)

	w.AppendString(`<h1>`)
	w.AppendString(`Evaluation for `)
	w.WriteHTMLString(lesson.Name)
	w.AppendString(`</h1>`)

	DisplayErrorMessage(w, r.Form.Get("Error"))

	w.AppendString(`<form style="min-width: 300px; max-width: max-content;" method="POST" action="/submission/new">`)

	w.AppendString(`<input type="hidden" name="CurrentPage" value="Main">`)

	w.AppendString(`<input type="hidden" name="ID" value="`)
	w.WriteHTMLString(r.Form.Get("ID"))
	w.AppendString(`">`)

	w.AppendString(`<input type="hidden" name="SubmissionIndex" value="`)
	w.WriteHTMLString(r.Form.Get("SubmissionIndex"))
	w.AppendString(`">`)

	for i := 0; i < len(submission.SubmittedSteps); i++ {
		submittedStep := &submission.SubmittedSteps[i]

		w.AppendString(`<fieldset>`)

		w.AppendString(`<legend>Step #`)
		w.WriteInt(i + 1)
		if submittedStep.Flags == SubmittedStepDraft {
			w.AppendString(` (draft)`)
		}
		w.AppendString(`</legend>`)

		w.AppendString(`<p>Name: `)
		w.WriteHTMLString(submittedStep.Step.Name)
		w.AppendString(`</p>`)

		w.AppendString(`<p>Type: `)
		w.AppendString(StepStringType(&submittedStep.Step))
		w.AppendString(`</p>`)

		if submittedStep.Flags == SubmittedStepSkipped {
			DisplayIndexedCommand(w, i, "Pass")
		} else {
			DisplayIndexedCommand(w, i, "Edit")
		}

		w.AppendString(`</fieldset>`)
		w.AppendString(`<br>`)
	}

	w.AppendString(`<input type="submit" name="NextPage" value="Finish">`)

	w.AppendString(`</form>`)

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil

}

func SubmissionNewTestPageHandler(w *http.Response, r *http.Request, submittedTest *SubmittedTest) error {
	test, _ := Step2Test(&submittedTest.Step)

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>`)
	w.AppendString(`Test: `)
	w.WriteHTMLString(test.Name)
	w.AppendString(`</title></head>`)
	w.AppendString(`<body>`)

	w.AppendString(`<h1>`)
	w.AppendString(`Test: `)
	w.WriteHTMLString(test.Name)
	w.AppendString(`</h1>`)

	DisplayErrorMessage(w, r.Form.Get("Error"))

	w.AppendString(`<form style="min-width: 300px; max-width: max-content;" method="POST" action="/submission/new">`)

	w.AppendString(`<input type="hidden" name="CurrentPage" value="Test">`)

	w.AppendString(`<input type="hidden" name="ID" value="`)
	w.WriteHTMLString(r.Form.Get("ID"))
	w.AppendString(`">`)

	w.AppendString(`<input type="hidden" name="SubmissionIndex" value="`)
	w.WriteHTMLString(r.Form.Get("SubmissionIndex"))
	w.AppendString(`">`)

	w.AppendString(`<input type="hidden" name="StepIndex" value="`)
	w.WriteHTMLString(r.Form.Get("StepIndex"))
	w.AppendString(`">`)

	if len(submittedTest.SubmittedQuestions) == 0 {
		submittedTest.SubmittedQuestions = make([]SubmittedQuestion, len(test.Questions))
	}
	for i := 0; i < len(submittedTest.SubmittedQuestions); i++ {
		submittedQuestion := &submittedTest.SubmittedQuestions[i]
		question := &test.Questions[i]

		w.AppendString(`<fieldset>`)
		w.AppendString(`<legend>Question #`)
		w.WriteInt(i + 1)
		w.AppendString(`</legend>`)
		w.AppendString(`<p>`)
		w.WriteHTMLString(question.Name)
		w.AppendString(`</p>`)

		w.AppendString(`<ol>`)
		for j := 0; j < len(question.Answers); j++ {
			answer := question.Answers[j]

			if j > 0 {
				w.AppendString(`<br>`)
			}

			w.AppendString(`<li>`)

			w.AppendString(`<input type="`)
			if len(question.CorrectAnswers) > 1 {
				w.AppendString(`checkbox`)
			} else {
				w.AppendString(`radio`)
			}
			w.AppendString(`" name="SelectedAnswer`)
			w.WriteInt(i)
			w.AppendString(`" value="`)
			w.WriteInt(j)
			w.AppendString(`"`)

			for k := 0; k < len(submittedQuestion.SelectedAnswers); k++ {
				if j == submittedQuestion.SelectedAnswers[k] {
					w.AppendString(` checked`)
					break
				}
			}

			w.AppendString(`> `)

			w.AppendString(`<span>`)
			w.WriteHTMLString(answer)
			w.AppendString(`</span>`)

			w.AppendString(`</li>`)
		}
		w.AppendString(`</ol>`)

		w.AppendString(`</fieldset>`)
		w.AppendString(`<br>`)
	}

	w.AppendString(`<input type="submit" name="NextPage" value="Save"> `)
	w.AppendString(`<input type="submit" name="NextPage" value="Discard">`)

	w.AppendString(`</form>`)

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func SubmissionNewDisplayProgrammingChecks(w *http.Response, task *StepProgramming, checkType CheckType) {
	w.AppendString(`<ol>`)
	for i := 0; i < len(task.Checks[checkType]); i++ {
		check := &task.Checks[checkType][i]

		w.AppendString(`<li>`)

		w.AppendString(`<label>Input: `)

		w.AppendString(`<textarea rows="1" readonly>`)
		w.WriteHTMLString(check.Input)
		w.AppendString(`</textarea>`)

		w.AppendString(`</label> `)

		w.AppendString(`<label>output: `)

		w.AppendString(`<textarea rows="1" readonly>`)
		w.WriteHTMLString(check.Output)
		w.AppendString(`</textarea>`)

		w.AppendString(`</label>`)

		w.AppendString(`</li>`)
	}
	w.AppendString(`</ol>`)
}

func SubmissionNewProgrammingPageHandler(w *http.Response, r *http.Request, submittedTask *SubmittedProgramming) error {
	task, _ := Step2Programming(&submittedTask.Step)

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>`)
	w.AppendString(`Programming task: `)
	w.WriteHTMLString(task.Name)
	w.AppendString(`</title></head>`)
	w.AppendString(`<body>`)

	w.AppendString(`<h1>`)
	w.AppendString(`Programming task: `)
	w.WriteHTMLString(task.Name)
	w.AppendString(`</h1>`)

	DisplayErrorMessage(w, r.Form.Get("Error"))

	w.AppendString(`<h2>Description</h2>`)
	w.AppendString(`<p>`)
	w.WriteHTMLString(task.Description)
	w.AppendString(`</p>`)

	w.AppendString(`<h2>Examples</h2>`)
	SubmissionNewDisplayProgrammingChecks(w, task, CheckTypeExample)

	w.AppendString(`<form method="POST" action="/submission/new">`)

	w.AppendString(`<input type="hidden" name="CurrentPage" value="Programming">`)

	w.AppendString(`<input type="hidden" name="ID" value="`)
	w.WriteHTMLString(r.Form.Get("ID"))
	w.AppendString(`">`)

	w.AppendString(`<input type="hidden" name="SubmissionIndex" value="`)
	w.WriteHTMLString(r.Form.Get("SubmissionIndex"))
	w.AppendString(`">`)

	w.AppendString(`<input type="hidden" name="StepIndex" value="`)
	w.WriteHTMLString(r.Form.Get("StepIndex"))
	w.AppendString(`">`)

	w.AppendString(`<h2>Solution</h2>`)

	w.AppendString(`<label>Programming language: `)
	SubmissionDisplayLanguageSelect(w, submittedTask, true)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<textarea cols="80" rows="24" name="Solution">`)
	w.WriteHTMLString(submittedTask.Solution)
	w.AppendString(`</textarea>`)

	w.AppendString(`<br><br>`)
	w.AppendString(`<input type="submit" name="NextPage" value="Save"> `)
	w.AppendString(`<input type="submit" name="NextPage" value="Discard">`)

	w.AppendString(`</form>`)

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)
	return nil
}

func SubmissionNewStepPageHandler(w *http.Response, r *http.Request, submittedStep *SubmittedStep) error {
	switch submittedStep.Type {
	default:
		panic("invalid step type")
	case SubmittedTypeTest:
		submittedTest, _ := Submitted2Test(submittedStep)
		return SubmissionNewTestPageHandler(w, r, submittedTest)
	case SubmittedTypeProgramming:
		submittedProgramming, _ := Submitted2Programming(submittedStep)
		return SubmissionNewProgrammingPageHandler(w, r, submittedProgramming)
	}
}

func SubmissionNewHandleCommand(w *http.Response, r *http.Request, submission *Submission, currentPage, k, command string) error {
	pindex, spindex, _, _, err := GetIndicies(k[len("Command"):])
	if err != nil {
		return http.ClientError(err)
	}

	switch currentPage {
	default:
		return http.ClientError(nil)
	case "Main":
		switch command {
		default:
			return http.ClientError(nil)
		case "Pass", "Edit":
			if (pindex < 0) || (pindex >= len(submission.SubmittedSteps)) {
				return http.ClientError(nil)
			}
			submittedStep := &submission.SubmittedSteps[pindex]
			submittedStep.Flags = SubmittedStepDraft
			submittedStep.Type = SubmittedType(submittedStep.Step.Type)

			r.Form.Set("StepIndex", spindex)
			return SubmissionNewStepPageHandler(w, r, submittedStep)
		}
	}
}

func SubmissionNewPageHandler(w *http.Response, r *http.Request) error {
	var submission Submission
	var subject Subject
	var lesson Lesson

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}

	if err := r.ParseForm(); err != nil {
		return http.ClientError(err)
	}

	currentPage := r.Form.Get("CurrentPage")
	nextPage := r.Form.Get("NextPage")

	lessonID, err := r.Form.GetInt("ID")
	if err != nil {
		return http.ClientError(err)
	}
	if err := GetLessonByID(DB2, int32(lessonID), &lesson); err != nil {
		if err == DBNotFound {
			return http.NotFound("lesson with this ID does not exist")
		}
		return http.ServerError(err)
	}
	if lesson.ContainerType != ContainerTypeSubject {
		return http.ClientError(nil)
	}

	if err := GetSubjectByID(DB2, int32(lesson.ContainerID), &subject); err != nil {
		return http.ServerError(err)
	}
	who, err := WhoIsUserInSubject(session.ID, &subject)
	if err != nil {
		return http.ServerError(err)
	}
	if who != SubjectUserStudent {
		return http.ForbiddenError
	}

	submissionIndex := r.Form.Get("SubmissionIndex")
	if submissionIndex == "" {
		submission.Flags = SubmissionDraft
		submission.UserID = session.ID
		submission.LessonID = lesson.ID
		submission.StartedAt = time.Now().Unix()
		submission.SubmittedSteps = make([]SubmittedStep, len(lesson.Steps))
		for i := 0; i < len(submission.SubmittedSteps); i++ {
			submittedStep := &submission.SubmittedSteps[i]
			StepDeepCopy(&submittedStep.Step, &lesson.Steps[i])
		}
		if err := CreateSubmission(DB2, &submission); err != nil {
			return http.ServerError(err)
		}

		lesson.Submissions = append(lesson.Submissions, submission.ID)
		if err := SaveLesson(DB2, &lesson); err != nil {
			return http.ServerError(err)
		}
		r.Form.SetInt("SubmissionIndex", len(lesson.Submissions)-1)
	} else {
		si, err := GetValidIndex(submissionIndex, len(lesson.Submissions))
		if err != nil {
			return http.ClientError(err)
		}
		if err := GetSubmissionByID(DB2, lesson.Submissions[si], &submission); err != nil {
			return http.ServerError(err)
		}
	}
	defer SaveSubmission(DB2, &submission)

	for i := 0; i < len(r.Form); i++ {
		k := r.Form[i].Key
		if len(r.Form[i].Values) == 0 {
			continue
		}
		v := r.Form[i].Values[0]

		/* 'command' is button, which modifies content of a current page. */
		if strings.StartsWith(k, "Command") {
			/* NOTE(anton2920): after command is executed, function must return. */
			return SubmissionNewHandleCommand(w, r, &submission, currentPage, k, v)
		}
	}

	stepIndex := r.Form.Get("StepIndex")
	if stepIndex != "" {
		si, err := GetValidIndex(r.Form.Get("StepIndex"), len(lesson.Steps))
		if err != nil {
			return http.ClientError(err)
		}
		submittedStep := &submission.SubmittedSteps[si]

		if nextPage != "Discard" {
			switch currentPage {
			case "Test":
				submittedTest, err := Submitted2Test(submittedStep)
				if err != nil {
					return http.ClientError(err)
				}

				if err := SubmissionNewTestFillFromRequest(r.Form, submittedTest); err != nil {
					return WritePageEx(w, r, SubmissionNewTestPageHandler, submittedTest, err)
				}
				if err := SubmissionNewTestVerify(submittedTest); err != nil {
					return WritePageEx(w, r, SubmissionNewTestPageHandler, submittedTest, err)
				}
			case "Programming":
				submittedTask, err := Submitted2Programming(submittedStep)
				if err != nil {
					return http.ClientError(err)
				}

				if err := SubmissionNewProgrammingFillFromRequest(r.Form, submittedTask); err != nil {
					return WritePageEx(w, r, SubmissionNewProgrammingPageHandler, submittedTask, err)
				}
				if err := SubmissionNewProgrammingVerify(submittedTask); err != nil {
					return WritePageEx(w, r, SubmissionNewProgrammingPageHandler, submittedTask, err)
				}

				if err := SubmissionVerifyProgramming(submittedTask, CheckTypeExample); err != nil {
					return WritePageEx(w, r, SubmissionNewProgrammingPageHandler, submittedTask, http.BadRequest(err.Error()))
				}

				scores := submittedTask.Scores[CheckTypeExample]
				messages := submittedTask.Messages[CheckTypeExample]
				for i := 0; i < len(scores); i++ {
					if scores[i] == 0 {
						return WritePageEx(w, r, SubmissionNewProgrammingPageHandler, submittedTask, http.BadRequest("example %d: %s", i+1, messages[i]))
					}
				}
			}

			submittedStep.Flags = SubmittedStepPassed
		} else {
			submittedStep.Flags = SubmittedStepSkipped
			clear(submittedStep.Data[:])
		}
	}

	switch nextPage {
	default:
		return SubmissionNewMainPageHandler(w, r, &submission)
	case "Finish":
		if err := SubmissionNewVerify(&submission); err != nil {
			return WritePageEx(w, r, SubmissionNewMainPageHandler, &submission, err)
		}
		submission.Flags = SubmissionActive
		submission.FinishedAt = time.Now().Unix()

		if err := SaveSubmission(DB2, &submission); err != nil {
			return http.ServerError(err)
		}
		SubmissionVerifyChannel <- submission.ID

		w.RedirectID("/lesson/", lessonID, http.StatusSeeOther)
		return nil
	}
}
