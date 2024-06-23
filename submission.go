package main

import (
	"fmt"
	"time"
	"unsafe"

	"github.com/anton2920/gofa/database"
	"github.com/anton2920/gofa/errors"
	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/net/url"
	"github.com/anton2920/gofa/slices"
	"github.com/anton2920/gofa/strings"
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
		LanguageID database.ID
		Solution   string

		Scores   [2][]int
		Messages [2][]string
	}
	SubmittedStep/* union */ struct {
		SubmittedCommon

		_ [max(unsafe.Sizeof(stdt), unsafe.Sizeof(stdp)) - unsafe.Sizeof(stdc)]byte
	}

	Submission struct {
		ID       database.ID
		Flags    int32
		UserID   database.ID
		LessonID database.ID

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

func CreateSubmission(submission *Submission) error {
	var err error

	submission.ID, err = database.IncrementNextID(SubmissionsDB)
	if err != nil {
		return fmt.Errorf("failed to increment submission ID: %w", err)
	}

	return SaveSubmission(submission)
}

func DBSubmitted2Submitted(submittedStep *SubmittedStep, data *byte) {
	submittedStep.Error = database.Offset2String(submittedStep.Error, data)
	DBStep2Step(&submittedStep.Step, data)

	switch submittedStep.Type {
	default:
		panic("invalid step type")
	case SubmittedTypeTest:
		submittedTest, _ := Submitted2Test(submittedStep)

		submittedTest.SubmittedQuestions = database.Offset2Slice(submittedTest.SubmittedQuestions, data)
		for i := 0; i < len(submittedTest.SubmittedQuestions); i++ {
			submittedQuestion := &submittedTest.SubmittedQuestions[i]
			submittedQuestion.SelectedAnswers = database.Offset2Slice(submittedQuestion.SelectedAnswers, data)
		}
		submittedTest.Scores = database.Offset2Slice(submittedTest.Scores, data)
	case SubmittedTypeProgramming:
		submittedTask, _ := Submitted2Programming(submittedStep)

		submittedTask.Solution = database.Offset2String(submittedTask.Solution, data)

		for i := 0; i < 2; i++ {
			submittedTask.Scores[i] = database.Offset2Slice(submittedTask.Scores[i], data)

			submittedTask.Messages[i] = database.Offset2Slice(submittedTask.Messages[i], data)
			for j := 0; j < len(submittedTask.Messages[i]); j++ {
				submittedTask.Messages[i][j] = database.Offset2String(submittedTask.Messages[i][j], data)
			}
		}
	}

}

func DBSubmission2Submission(submission *Submission) {
	data := &submission.Data[0]

	submission.SubmittedSteps = database.Offset2Slice(submission.SubmittedSteps, data)
	for i := 0; i < len(submission.SubmittedSteps); i++ {
		DBSubmitted2Submitted(&submission.SubmittedSteps[i], data)
	}
}

func GetSubmissionByID(id database.ID, submission *Submission) error {
	if err := database.Read(SubmissionsDB, id, submission); err != nil {
		return err
	}

	DBSubmission2Submission(submission)
	return nil
}

func GetSubmissions(pos *int64, submissions []Submission) (int, error) {
	n, err := database.ReadMany(SubmissionsDB, pos, submissions)
	if err != nil {
		return 0, err
	}

	for i := 0; i < n; i++ {
		DBSubmission2Submission(&submissions[i])
	}
	return n, nil
}

func Submitted2DBSubmitted(ds *SubmittedStep, ss *SubmittedStep, data []byte, n int) int {
	ds.Flags = ss.Flags
	ds.Status = ss.Status

	n += Step2DBStep(&ds.Step, &ss.Step, data, n)
	n += database.String2DBString(&ds.Error, ss.Error, data, n)

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
			n += database.Slice2DBSlice(&dq.SelectedAnswers, sq.SelectedAnswers, data, n)
		}
		n += database.Slice2DBSlice(&dt.SubmittedQuestions, dt.SubmittedQuestions, data, n)

		n += database.Slice2DBSlice(&dt.Scores, st.Scores, data, n)
	case SubmittedTypeProgramming:
		st, _ := Submitted2Programming(ss)

		ds.Type = SubmittedTypeProgramming
		dt, _ := Submitted2Programming(ds)

		dt.LanguageID = st.LanguageID

		n += database.String2DBString(&dt.Solution, st.Solution, data, n)

		for i := 0; i < 2; i++ {
			n += database.Slice2DBSlice(&dt.Scores[i], st.Scores[i], data, n)

			dt.Messages[i] = make([]string, len(st.Messages[i]))
			for j := 0; j < len(st.Messages[i]); j++ {
				n += database.String2DBString(&dt.Messages[i][j], st.Messages[i][j], data, n)
			}
			n += database.Slice2DBSlice(&dt.Messages[i], dt.Messages[i], data, n)
		}
	}

	return n
}

func SaveSubmission(submission *Submission) error {
	var submissionDB Submission
	var n int

	submissionDB.ID = submission.ID
	submissionDB.Flags = submission.Flags
	submissionDB.Status = submission.Status
	submissionDB.UserID = submission.UserID
	submissionDB.LessonID = submission.LessonID
	submissionDB.StartedAt = submission.StartedAt
	submissionDB.FinishedAt = submission.FinishedAt

	/* TODO(anton2920): save up to a sizeof(lesson.Data). */
	data := unsafe.Slice(&submissionDB.Data[0], len(submissionDB.Data))
	submissionDB.SubmittedSteps = make([]SubmittedStep, len(submission.SubmittedSteps))
	for i := 0; i < len(submission.SubmittedSteps); i++ {
		n += Submitted2DBSubmitted(&submissionDB.SubmittedSteps[i], &submission.SubmittedSteps[i], data, n)
	}
	n += database.Slice2DBSlice(&submissionDB.SubmittedSteps, submissionDB.SubmittedSteps, data, n)

	return database.Write(SubmissionsDB, submissionDB.ID, &submissionDB)
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

func DisplaySubmittedStepScore(w *http.Response, l Language, submittedStep *SubmittedStep) {
	w.AppendString(`<p>`)
	w.AppendString(Ls(l, "Score"))
	w.AppendString(`: `)
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

func DisplaySubmissionLanguageSelect(w *http.Response, submittedTask *SubmittedProgramming, enabled bool) {
	w.AppendString(` <select name="LanguageID"`)
	if !enabled {
		w.AppendString(` disabled`)
	}
	w.AppendString(`>`)
	for i := database.ID(0); i < database.ID(len(ProgrammingLanguages)); i++ {
		lang := &ProgrammingLanguages[i]

		w.AppendString(`<option value="`)
		w.WriteID(i)
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

func DisplaySubmissionTitle(w *http.Response, l Language, subject *Subject, lesson *Lesson, user *User) {
	w.AppendString(Ls(l, "Submission"))
	w.AppendString(` `)
	w.AppendString(Ls(GL, "for"))
	w.AppendString(` «`)
	w.WriteHTMLString(subject.Name)
	w.AppendString(`: `)
	w.WriteHTMLString(lesson.Name)
	w.AppendString(`» `)
	w.AppendString(Ls(GL, "by"))
	w.AppendString(` `)
	w.WriteHTMLString(user.LastName)
	w.AppendString(` `)
	w.WriteHTMLString(user.FirstName)
}

func DisplaySubmissionLink(w *http.Response, l Language, submission *Submission) {
	var user User

	if err := GetUserByID(submission.UserID, &user); err != nil {
		/* TODO(anton2920): report error. */
	}

	w.AppendString(`<a href="/submission/`)
	w.WriteID(submission.ID)
	w.AppendString(`">`)
	w.WriteHTMLString(user.LastName)
	w.AppendString(` `)
	w.WriteHTMLString(user.FirstName)
	w.AppendString(` (`)
	switch submission.Status {
	case SubmissionCheckPending:
		w.AppendString(`<i>`)
		w.AppendString(Ls(l, "pending"))
		w.AppendString(` `)
		w.AppendString(Ls(l, "verification"))
		w.AppendString(`</i>`)
	case SubmissionCheckInProgress:
		w.AppendString(`<i>`)
		w.AppendString(Ls(l, "verification"))
		w.AppendString(` `)
		w.AppendString(Ls(l, "in progress"))
		w.AppendString(`</i>`)
	case SubmissionCheckDone:
		DisplaySubmissionTotalScore(w, submission)
	}
	w.AppendString(`)`)
	w.AppendString(`</a>`)
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

	id, err := GetIDFromURL(GL, r.URL, "/submission/")
	if err != nil {
		return http.ClientError(err)
	}
	if err := GetSubmissionByID(id, &submission); err != nil {
		if err == database.NotFound {
			return http.NotFound("lesson with this ID does not exist")
		}
		return http.ServerError(err)
	}

	if err := GetLessonByID(submission.LessonID, &lesson); err != nil {
		return http.ServerError(err)
	}

	if err := GetSubjectByID(lesson.ContainerID, &subject); err != nil {
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

	if err := GetUserByID(submission.UserID, &user); err != nil {
		return http.ServerError(err)
	}

	DisplayHTMLStart(w)

	DisplayHeadStart(w)
	{
		w.AppendString(`<title>`)
		DisplaySubmissionTitle(w, GL, &subject, &lesson, &user)
		w.AppendString(`</title>`)
	}
	DisplayHeadEnd(w)

	DisplayBodyStart(w)
	{
		DisplayHeader(w, GL)
		DisplaySidebar(w, GL, session)

		DisplayPageStart(w)
		{
			w.AppendString(`<h2>`)
			DisplaySubmissionTitle(w, GL, &subject, &lesson, &user)
			w.AppendString(`</h2>`)
			w.AppendString(`<br>`)

			w.AppendString(`<form method="POST" action="/submission/results">`)

			DisplayHiddenID(w, "ID", submission.ID)

			for i := 0; i < len(submission.SubmittedSteps); i++ {
				submittedStep := &submission.SubmittedSteps[i]

				w.AppendString(`<div class="border rounded p-4">`)

				w.AppendString(`<p><b>`)
				w.AppendString(Ls(GL, "Step"))
				w.AppendString(` #`)
				w.WriteInt(i + 1)
				w.AppendString(`</b></p>`)

				w.AppendString(`<p>`)
				w.AppendString(Ls(GL, "Name"))
				w.AppendString(`: `)
				w.WriteHTMLString(submittedStep.Step.Name)
				w.AppendString(`</p>`)

				w.AppendString(`<p>`)
				w.AppendString(Ls(GL, "Type"))
				w.AppendString(`: `)
				w.AppendString(StepStringType(GL, &submittedStep.Step))
				w.AppendString(`</p>`)

				if submittedStep.Flags == SubmittedStepSkipped {
					DisplaySubmittedStepScore(w, GL, submittedStep)

					w.AppendString(`<p><i>`)
					w.AppendString(Ls(GL, "This step has been skipped"))
					w.AppendString(`.</i></p>`)
				} else {
					switch submittedStep.Status {
					case SubmissionCheckPending:
						w.AppendString(`<p><i>`)
						w.AppendString(Ls(GL, "Pending"))
						w.AppendString(` `)
						w.AppendString(Ls(GL, "verification"))
						w.AppendString(`...`)
						w.AppendString(`</i></p>`)
					case SubmissionCheckInProgress:
						w.AppendString(`<p><i>`)
						w.AppendString(Ls(GL, "Verification"))
						w.AppendString(` `)
						w.AppendString(Ls(GL, "in progress"))
						w.AppendString(`...`)
						w.AppendString(`</i></p>`)
					case SubmissionCheckDone:
						DisplaySubmittedStepScore(w, GL, submittedStep)
						DisplayErrorMessage(w, GL, submittedStep.Error)

						DisplayIndexedCommand(w, GL, i, "Open")
						if teacher {
							DisplayIndexedCommand(w, GL, i, "Re-check")
						}
					}
				}

				w.AppendString(`</div>`)
				w.AppendString(`<br>`)
			}

			switch submission.Status {
			case SubmissionCheckPending:
				w.AppendString(`<p><i>`)
				w.AppendString(Ls(GL, "Pending"))
				w.AppendString(` `)
				w.AppendString(Ls(GL, "verification"))
				w.AppendString(`...`)
				w.AppendString(`</i></p>`)
			case SubmissionCheckInProgress:
				w.AppendString(`<p><i>`)
				w.AppendString(Ls(GL, "Verification"))
				w.AppendString(` `)
				w.AppendString(Ls(GL, "in progress"))
				w.AppendString(`...`)
				w.AppendString(`</i></p>`)
			case SubmissionCheckDone:
				w.AppendString(`<p>`)
				w.AppendString(Ls(GL, "Total score"))
				w.AppendString(`: `)
				DisplaySubmissionTotalScore(w, &submission)
				w.AppendString(`</p>`)
				if teacher {
					DisplayCommand(w, GL, "Re-check")
				}
			}

			w.AppendString(`</form>`)
		}
		DisplayPageEnd(w)
	}
	DisplayBodyEnd(w)

	DisplayHTMLEnd(w)
	return nil

}

func SubmissionResultsTestPageHandler(w *http.Response, r *http.Request, session *Session, submittedTest *SubmittedTest) error {
	test, _ := Step2Test(&submittedTest.Step)
	teacher := r.Form.Get("Teacher") != ""

	DisplayHTMLStart(w)

	DisplayHeadStart(w)
	{
		w.AppendString(`<title>`)
		w.AppendString(Ls(GL, "Submitted test"))
		w.AppendString(`: «`)
		w.WriteHTMLString(test.Name)
		w.AppendString(`»</title>`)
	}
	DisplayHeadEnd(w)

	DisplayBodyStart(w)
	{
		DisplayHeader(w, GL)
		DisplaySidebar(w, GL, session)

		DisplayPageStart(w)
		{
			w.AppendString(`<h2>`)
			w.AppendString(Ls(GL, "Submitted test"))
			w.AppendString(`: «`)
			w.WriteHTMLString(test.Name)
			w.AppendString(`»</h2>`)
			w.AppendString(`<br>`)

			if teacher {
				w.AppendString(`<p><i>`)
				w.AppendString(Ls(GL, "Note: answers marked with [x] are correct"))
				w.AppendString(`.</i></p>`)
			}

			for i := 0; i < len(test.Questions); i++ {
				question := &test.Questions[i]

				w.AppendString(`<div class="border round p-4">`)

				w.AppendString(`<p><b>`)
				w.WriteHTMLString(question.Name)
				w.AppendString(`</b></p>`)

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

				w.AppendString(`<span>`)
				w.AppendString(Ls(GL, "Score"))
				w.AppendString(`: `)
				w.WriteInt(submittedTest.Scores[i])
				w.AppendString(`/1</span>`)

				w.AppendString(`</div>`)
				w.AppendString(`<br>`)
			}
		}
		DisplayPageEnd(w)
	}
	DisplayBodyEnd(w)

	DisplayHTMLEnd(w)
	return nil
}

func SubmissionResultsProgrammingDisplayChecks(w *http.Response, l Language, submittedTask *SubmittedProgramming, checkType CheckType) {
	task, _ := Step2Programming(&submittedTask.Step)
	scores := submittedTask.Scores[checkType]
	messages := submittedTask.Messages[checkType]

	w.AppendString(`<ol>`)
	for i := 0; i < len(task.Checks[checkType]); i++ {
		check := &task.Checks[checkType][i]
		score := scores[i]
		message := messages[i]

		w.AppendString(`<li class="mt-3">`)

		w.AppendString(`<label>`)
		w.AppendString(Ls(l, "Input"))
		w.AppendString(`: `)

		w.AppendString(`<textarea class="btn btn-outline-dark" rows="1" readonly>`)
		w.WriteHTMLString(check.Input)
		w.AppendString(`</textarea>`)

		w.AppendString(`</label> `)

		w.AppendString(`<label>`)
		w.AppendString(Ls(l, "output"))
		w.AppendString(`: `)

		w.AppendString(`<textarea class="btn btn-outline-dark" rows="1" readonly>`)
		w.WriteHTMLString(check.Output)
		w.AppendString(`</textarea>`)

		w.AppendString(`</label> `)

		w.AppendString(`<span>`)
		w.AppendString(Ls(GL, "score"))
		w.AppendString(`: `)
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

func SubmissionResultsProgrammingPageHandler(w *http.Response, r *http.Request, session *Session, submittedTask *SubmittedProgramming) error {
	task, _ := Step2Programming(&submittedTask.Step)
	teacher := r.Form.Get("Teacher") != ""

	DisplayHTMLStart(w)

	DisplayHeadStart(w)
	{
		w.AppendString(`<title>`)
		w.AppendString(Ls(GL, "Submitted programming task"))
		w.AppendString(`: «`)
		w.WriteHTMLString(task.Name)
		w.AppendString(`»</title>`)
	}
	DisplayHeadEnd(w)

	DisplayBodyStart(w)
	{
		DisplayHeader(w, GL)
		DisplaySidebar(w, GL, session)

		DisplayPageStart(w)
		{
			w.AppendString(`<h2>`)
			w.AppendString(Ls(GL, "Submitted programming task"))
			w.AppendString(`: «`)
			w.WriteHTMLString(task.Name)
			w.AppendString(`»</h2>`)
			w.AppendString(`<br>`)

			w.AppendString(`<h3>`)
			w.AppendString(Ls(GL, "Description"))
			w.AppendString(`</h3>`)
			w.AppendString(`<p>`)
			w.WriteHTMLString(task.Description)
			w.AppendString(`</p>`)
			w.AppendString(`<br>`)

			w.AppendString(`<h3>`)
			w.AppendString(Ls(GL, "Examples"))
			w.AppendString(`</h3>`)
			SubmissionNewDisplayProgrammingChecks(w, GL, task, CheckTypeExample)
			w.AppendString(`<br>`)

			w.AppendString(`<h3>`)
			w.AppendString(Ls(GL, "Solution"))
			w.AppendString(`</h3>`)

			DisplayInputLabel(w, GL, "Programming language")
			DisplaySubmissionLanguageSelect(w, submittedTask, false)
			w.AppendString(`<br>`)

			w.AppendString(`<textarea class="form-control" rows="10" readonly>`)
			w.WriteHTMLString(submittedTask.Solution)
			w.AppendString(`</textarea>`)

			if teacher {
				w.AppendString(`<br><br>`)
				w.AppendString(`<h3>`)
				w.AppendString(Ls(GL, "Tests"))
				w.AppendString(`</h3>`)
				SubmissionResultsProgrammingDisplayChecks(w, GL, submittedTask, CheckTypeTest)
			}
		}
		DisplayPageEnd(w)
	}
	DisplayBodyEnd(w)

	DisplayHTMLEnd(w)
	return nil
}

func SubmissionResultsStepPageHandler(w *http.Response, r *http.Request, session *Session, submittedStep *SubmittedStep) error {
	switch submittedStep.Type {
	default:
		panic("invalid step type")
	case SubmittedTypeTest:
		submittedTest, _ := Submitted2Test(submittedStep)
		return SubmissionResultsTestPageHandler(w, r, session, submittedTest)
	case SubmittedTypeProgramming:
		submittedTask, _ := Submitted2Programming(submittedStep)
		return SubmissionResultsProgrammingPageHandler(w, r, session, submittedTask)
	}
}

func SubmissionResultsHandleCommand(w *http.Response, r *http.Request, l Language, session *Session, submission *Submission, k, command string) error {
	pindex, spindex, _, _, err := GetIndicies(k[len("Command"):])
	if err != nil {
		return http.ClientError(err)
	}

	switch command {
	default:
		return http.ClientError(nil)
	case Ls(l, "Open"):
		if (pindex < 0) || (pindex >= len(submission.SubmittedSteps)) {
			return http.ClientError(nil)
		}
		submittedStep := &submission.SubmittedSteps[pindex]

		return SubmissionResultsStepPageHandler(w, r, session, submittedStep)
	case Ls(l, "Re-check"):
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
		if err := SaveSubmission(submission); err != nil {
			return http.ServerError(err)
		}

		SubmissionVerifyChannel <- submission.ID

		w.RedirectID("/submission/", submission.ID, http.StatusSeeOther)
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

	submissionID, err := r.Form.GetID("ID")
	if err != nil {
		return http.ClientError(err)
	}
	if err := GetSubmissionByID(submissionID, &submission); err != nil {
		if err == database.NotFound {
			return http.NotFound(Ls(GL, "submission with this ID does not exist"))
		}
		return http.ServerError(err)
	}

	if err := GetLessonByID(submission.LessonID, &lesson); err != nil {
		return http.ServerError(err)
	}

	if err := GetSubjectByID(lesson.ContainerID, &subject); err != nil {
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
			return SubmissionResultsHandleCommand(w, r, GL, session, &submission, k, v)
		}
	}

	return http.ClientError(nil)
}

func SubmittedStepClear(submittedStep *SubmittedStep) {
	switch submittedStep.Type {
	default:
		panic("invalid step type")
	case SubmittedTypeTest:
		submittedTest, _ := Submitted2Test(submittedStep)
		submittedTest.SubmittedQuestions = submittedTest.SubmittedQuestions[:0]
	case SubmittedTypeProgramming:
		submittedTask, _ := Submitted2Programming(submittedStep)
		submittedTask.LanguageID = 0
		submittedTask.Solution = ""
	}
}

func SubmissionNewVerify(l Language, submission *Submission) error {
	empty := true
	for i := 0; i < len(submission.SubmittedSteps); i++ {
		submittedStep := &submission.SubmittedSteps[i]
		if submittedStep.Flags != SubmittedStepSkipped {
			empty = false

			if submittedStep.Flags == SubmittedStepDraft {
				return http.BadRequest(Ls(l, "step %d is still a draft"), i+1)
			}
		}
	}
	if empty {
		return http.BadRequest(Ls(l, "you have to pass at least one step"))
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

func SubmissionNewTestVerify(l Language, submittedTest *SubmittedTest) error {
	test, _ := Step2Test(&submittedTest.Step)

	for i := 0; i < len(submittedTest.SubmittedQuestions); i++ {
		submittedQuestion := &submittedTest.SubmittedQuestions[i]
		question := &test.Questions[i]

		if len(submittedQuestion.SelectedAnswers) == 0 {
			return http.BadRequest(Ls(l, "question %d: select at least one answer"), i+1)
		}
		if (len(question.CorrectAnswers) == 1) && (len(submittedQuestion.SelectedAnswers) > 1) {
			return http.ClientError(nil)
		}
	}

	return nil
}

func SubmissionNewTestPageHandler(w *http.Response, r *http.Request, session *Session, submittedTest *SubmittedTest) error {
	test, _ := Step2Test(&submittedTest.Step)

	DisplayHTMLStart(w)

	DisplayHeadStart(w)
	{
		w.AppendString(`<title>`)
		w.AppendString(Ls(GL, "Test"))
		w.AppendString(`: «`)
		w.WriteHTMLString(test.Name)
		w.AppendString(`»</title>`)
	}
	DisplayHeadEnd(w)

	DisplayBodyStart(w)
	{
		DisplayHeader(w, GL)
		DisplaySidebar(w, GL, session)

		DisplayFormStart(w, r, GL, "", r.URL.Path, 6)
		{
			w.AppendString(`<h3 class="text-center">`)
			w.AppendString(Ls(GL, "Test"))
			w.AppendString(`: «`)
			w.WriteHTMLString(test.Name)
			w.AppendString(`»</h3>`)
			w.AppendString(`<br>`)

			DisplayErrorMessage(w, GL, r.Form.Get("Error"))

			DisplayHiddenString(w, "CurrentPage", "Test")
			DisplayHiddenString(w, "SubmissionIndex", r.Form.Get("SubmissionIndex"))
			DisplayHiddenString(w, "StepIndex", r.Form.Get("StepIndex"))

			if len(submittedTest.SubmittedQuestions) == 0 {
				submittedTest.SubmittedQuestions = make([]SubmittedQuestion, len(test.Questions))
			}
			for i := 0; i < len(submittedTest.SubmittedQuestions); i++ {
				submittedQuestion := &submittedTest.SubmittedQuestions[i]
				question := &test.Questions[i]

				w.AppendString(`<div class="border round p-4">`)

				w.AppendString(`<p><b>`)
				w.WriteHTMLString(question.Name)
				w.AppendString(`</b></p>`)

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

				w.AppendString(`</div>`)
				w.AppendString(`<br>`)
			}
			DisplaySubmit(w, GL, "NextPage", "Save", true)
			DisplaySubmit(w, GL, "NextPage", "Discard", true)
		}
		DisplayFormEnd(w)
	}
	DisplayBodyEnd(w)

	DisplayHTMLEnd(w)
	return nil
}

func SubmissionNewDisplayProgrammingChecks(w *http.Response, l Language, task *StepProgramming, checkType CheckType) {
	w.AppendString(`<ol>`)
	for i := 0; i < len(task.Checks[checkType]); i++ {
		check := &task.Checks[checkType][i]

		w.AppendString(`<li class="mt-3">`)

		w.AppendString(`<label>`)
		w.AppendString(Ls(l, "Input"))
		w.AppendString(`: `)

		w.AppendString(`<textarea class="btn btn-outline-dark" rows="1" readonly>`)
		w.WriteHTMLString(check.Input)
		w.AppendString(`</textarea>`)

		w.AppendString(`</label> `)

		w.AppendString(`<label>`)
		w.AppendString(Ls(l, "output"))
		w.AppendString(`: `)

		w.AppendString(`<textarea class="btn btn-outline-dark" rows="1" readonly>`)
		w.WriteHTMLString(check.Output)
		w.AppendString(`</textarea>`)

		w.AppendString(`</label>`)

		w.AppendString(`</li>`)
	}
	w.AppendString(`</ol>`)
}

func SubmissionNewProgrammingPageHandler(w *http.Response, r *http.Request, session *Session, submittedTask *SubmittedProgramming) error {
	task, _ := Step2Programming(&submittedTask.Step)

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>`)
	w.AppendString(Ls(GL, "Programming task"))
	w.AppendString(`: `)
	w.WriteHTMLString(task.Name)
	w.AppendString(`</title></head>`)

	w.AppendString(`<body>`)

	w.AppendString(`<h1>`)
	w.AppendString(Ls(GL, "Programming task"))
	w.AppendString(`: `)
	w.WriteHTMLString(task.Name)
	w.AppendString(`</h1>`)

	DisplayErrorMessage(w, GL, r.Form.Get("Error"))

	w.AppendString(`<h2>`)
	w.AppendString(Ls(GL, "Description"))
	w.AppendString(`</h2>`)
	w.AppendString(`<p>`)
	w.WriteHTMLString(task.Description)
	w.AppendString(`</p>`)

	w.AppendString(`<h2>`)
	w.AppendString(Ls(GL, "Examples"))
	w.AppendString(`</h2>`)
	SubmissionNewDisplayProgrammingChecks(w, GL, task, CheckTypeExample)

	w.AppendString(`<form method="POST" action="/submission/new">`)

	DisplayHiddenString(w, "ID", r.Form.Get("ID"))
	DisplayHiddenString(w, "SubmissionIndex", r.Form.Get("SubmissionIndex"))
	DisplayHiddenString(w, "StepIndex", r.Form.Get("StepIndex"))

	DisplayHiddenString(w, "CurrentPage", "Programming")

	w.AppendString(`<h2>`)
	w.AppendString(Ls(GL, "Solution"))
	w.AppendString(`</h2>`)

	w.AppendString(`<label>`)
	w.AppendString(Ls(GL, "Programming language"))
	w.AppendString(`: `)
	DisplaySubmissionLanguageSelect(w, submittedTask, true)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<textarea cols="80" rows="24" name="Solution">`)
	w.WriteHTMLString(submittedTask.Solution)
	w.AppendString(`</textarea>`)

	w.AppendString(`<br><br>`)

	DisplaySubmit(w, GL, "NextPage", "Save", true)
	DisplaySubmit(w, GL, "NextPage", "Discard", true)
	w.AppendString(`</form>`)

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)
	return nil
}

func SubmissionNewStepPageHandler(w *http.Response, r *http.Request, session *Session, submittedStep *SubmittedStep) error {
	switch submittedStep.Type {
	default:
		panic("invalid step type")
	case SubmittedTypeTest:
		submittedTest, _ := Submitted2Test(submittedStep)
		return SubmissionNewTestPageHandler(w, r, session, submittedTest)
	case SubmittedTypeProgramming:
		submittedProgramming, _ := Submitted2Programming(submittedStep)
		return SubmissionNewProgrammingPageHandler(w, r, session, submittedProgramming)
	}
}

func SubmissionNewProgrammingFillFromRequest(vs url.Values, submittedTask *SubmittedProgramming) error {
	id, err := GetValidID(vs.Get("LanguageID"), database.ID(len(ProgrammingLanguages)))
	if err != nil {
		return http.ClientError(err)
	}
	submittedTask.LanguageID = id

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

func SubmissionNewMainPageHandler(w *http.Response, r *http.Request, session *Session, submission *Submission) error {
	var lesson Lesson

	if err := GetLessonByID(submission.LessonID, &lesson); err != nil {
		return http.ServerError(err)
	}

	DisplayHTMLStart(w)

	DisplayHeadStart(w)
	{
		w.AppendString(`<title>`)
		w.AppendString(Ls(GL, "Evaluation"))
		w.AppendString(` `)
		w.AppendString(Ls(GL, "for"))
		w.AppendString(` «`)
		w.WriteHTMLString(lesson.Name)
		w.AppendString(`»</title>`)
	}
	DisplayHeadEnd(w)

	DisplayBodyStart(w)
	{
		DisplayHeader(w, GL)
		DisplaySidebar(w, GL, session)

		DisplayFormStart(w, r, GL, "", r.URL.Path, 4)
		{
			w.AppendString(`<h3 class="text-center">`)
			w.AppendString(Ls(GL, "Evaluation"))
			w.AppendString(` `)
			w.AppendString(Ls(GL, "for"))
			w.AppendString(` «`)
			w.WriteHTMLString(lesson.Name)
			w.AppendString(`»</h3>`)
			w.AppendString(`<br>`)

			DisplayErrorMessage(w, GL, r.Form.Get("Error"))

			DisplayHiddenString(w, "CurrentPage", "Main")
			DisplayHiddenString(w, "SubmissionIndex", r.Form.Get("SubmissionIndex"))

			for i := 0; i < len(submission.SubmittedSteps); i++ {
				submittedStep := &submission.SubmittedSteps[i]

				w.AppendString(`<div class="border round p-4">`)

				w.AppendString(`<p><b>`)
				w.AppendString(Ls(GL, "Step"))
				w.AppendString(` #`)
				w.WriteInt(i + 1)
				DisplayDraft(w, GL, submittedStep.Flags == SubmittedStepDraft)
				w.AppendString(`</b></p>`)

				w.AppendString(`<p>`)
				w.AppendString(Ls(GL, "Name"))
				w.AppendString(`: `)
				w.WriteHTMLString(submittedStep.Step.Name)
				w.AppendString(`</p>`)

				w.AppendString(`<p>`)
				w.AppendString(Ls(GL, "Type"))
				w.AppendString(`: `)
				w.AppendString(StepStringType(GL, &submittedStep.Step))
				w.AppendString(`</p>`)

				if submittedStep.Flags == SubmittedStepSkipped {
					DisplayIndexedCommand(w, GL, i, "Pass")
				} else {
					DisplayIndexedCommand(w, GL, i, "Edit")
				}

				w.AppendString(`</div>`)
				w.AppendString(`<br>`)
			}

			DisplaySubmit(w, GL, "NextPage", "Finish", true)
		}
		DisplayFormEnd(w)
	}
	DisplayBodyEnd(w)

	DisplayHTMLEnd(w)
	return nil
}

func SubmissionNewHandleCommand(w *http.Response, r *http.Request, l Language, session *Session, submission *Submission, currentPage, k, command string) error {
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
		case Ls(l, "Pass"), Ls(l, "Edit"):
			if (pindex < 0) || (pindex >= len(submission.SubmittedSteps)) {
				return http.ClientError(nil)
			}
			submittedStep := &submission.SubmittedSteps[pindex]
			submittedStep.Flags = SubmittedStepDraft
			submittedStep.Type = SubmittedType(submittedStep.Step.Type)

			r.Form.Set("StepIndex", spindex)
			return SubmissionNewStepPageHandler(w, r, session, submittedStep)
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

	lessonID, err := r.Form.GetID("ID")
	if err != nil {
		return http.ClientError(err)
	}
	if err := GetLessonByID(lessonID, &lesson); err != nil {
		if err == database.NotFound {
			return http.NotFound("lesson with this ID does not exist")
		}
		return http.ServerError(err)
	}
	if lesson.ContainerType != ContainerTypeSubject {
		return http.ClientError(nil)
	}

	if err := GetSubjectByID(lesson.ContainerID, &subject); err != nil {
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
		if err := CreateSubmission(&submission); err != nil {
			return http.ServerError(err)
		}

		lesson.Submissions = append(lesson.Submissions, submission.ID)
		if err := SaveLesson(&lesson); err != nil {
			return http.ServerError(err)
		}
		r.Form.SetInt("SubmissionIndex", len(lesson.Submissions)-1)
	} else {
		si, err := GetValidIndex(submissionIndex, len(lesson.Submissions))
		if err != nil {
			return http.ClientError(err)
		}
		if err := GetSubmissionByID(lesson.Submissions[si], &submission); err != nil {
			return http.ServerError(err)
		}
	}
	defer SaveSubmission(&submission)

	for i := 0; i < len(r.Form); i++ {
		k := r.Form[i].Key
		if len(r.Form[i].Values) == 0 {
			continue
		}
		v := r.Form[i].Values[0]

		/* 'command' is button, which modifies content of a current page. */
		if strings.StartsWith(k, "Command") {
			/* NOTE(anton2920): after command is executed, function must return. */
			return SubmissionNewHandleCommand(w, r, GL, session, &submission, currentPage, k, v)
		}
	}

	stepIndex := r.Form.Get("StepIndex")
	if stepIndex != "" {
		si, err := GetValidIndex(r.Form.Get("StepIndex"), len(lesson.Steps))
		if err != nil {
			return http.ClientError(err)
		}
		submittedStep := &submission.SubmittedSteps[si]

		if nextPage != Ls(GL, "Discard") {
			switch currentPage {
			case "Test":
				submittedTest, err := Submitted2Test(submittedStep)
				if err != nil {
					return http.ClientError(err)
				}

				if err := SubmissionNewTestFillFromRequest(r.Form, submittedTest); err != nil {
					return WritePageEx(w, r, session, SubmissionNewTestPageHandler, submittedTest, err)
				}
				if err := SubmissionNewTestVerify(GL, submittedTest); err != nil {
					return WritePageEx(w, r, session, SubmissionNewTestPageHandler, submittedTest, err)
				}
			case "Programming":
				submittedTask, err := Submitted2Programming(submittedStep)
				if err != nil {
					return http.ClientError(err)
				}

				if err := SubmissionNewProgrammingFillFromRequest(r.Form, submittedTask); err != nil {
					return WritePageEx(w, r, session, SubmissionNewProgrammingPageHandler, submittedTask, err)
				}
				if err := SubmissionNewProgrammingVerify(submittedTask); err != nil {
					return WritePageEx(w, r, session, SubmissionNewProgrammingPageHandler, submittedTask, err)
				}

				if err := SubmissionVerifyProgramming(GL, submittedTask, CheckTypeExample); err != nil {
					return WritePageEx(w, r, session, SubmissionNewProgrammingPageHandler, submittedTask, http.BadRequest(err.Error()))
				}

				scores := submittedTask.Scores[CheckTypeExample]
				messages := submittedTask.Messages[CheckTypeExample]
				for i := 0; i < len(scores); i++ {
					if scores[i] == 0 {
						return WritePageEx(w, r, session, SubmissionNewProgrammingPageHandler, submittedTask, http.BadRequest(Ls(GL, "example %d: %s"), i+1, messages[i]))
					}
				}
			}

			submittedStep.Flags = SubmittedStepPassed
		} else {
			submittedStep.Flags = SubmittedStepSkipped
			SubmittedStepClear(submittedStep)
		}
	}

	switch nextPage {
	default:
		return SubmissionNewMainPageHandler(w, r, session, &submission)
	case Ls(GL, "Finish"):
		if err := SubmissionNewVerify(GL, &submission); err != nil {
			return WritePageEx(w, r, session, SubmissionNewMainPageHandler, &submission, err)
		}
		submission.Flags = SubmissionActive
		submission.FinishedAt = time.Now().Unix()

		if err := SaveSubmission(&submission); err != nil {
			return http.ServerError(err)
		}
		SubmissionVerifyChannel <- submission.ID

		w.RedirectID("/lesson/", lessonID, http.StatusSeeOther)
		return nil
	}
}
