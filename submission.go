package main

import (
	"fmt"
	"time"
	"unsafe"

	"github.com/anton2920/gofa/database"
	"github.com/anton2920/gofa/errors"
	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/net/url"
	"github.com/anton2920/gofa/prof"
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
	_
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
	defer prof.End(prof.Begin(""))

	if submittedStep.Type != SubmittedTypeTest {
		return nil, errors.New("invalid submitted type for test")
	}
	return (*SubmittedTest)(unsafe.Pointer(submittedStep)), nil
}

func Submitted2Programming(submittedStep *SubmittedStep) (*SubmittedProgramming, error) {
	defer prof.End(prof.Begin(""))

	if submittedStep.Type != SubmittedTypeProgramming {
		return nil, errors.New("invalid submitted type for programming")
	}
	return (*SubmittedProgramming)(unsafe.Pointer(submittedStep)), nil
}

func CreateSubmission(submission *Submission) error {
	defer prof.End(prof.Begin(""))

	var err error

	submission.ID, err = database.IncrementNextID(SubmissionsDB)
	if err != nil {
		return fmt.Errorf("failed to increment submission ID: %w", err)
	}

	return SaveSubmission(submission)
}

func DBSubmitted2Submitted(submittedStep *SubmittedStep, data *byte) {
	defer prof.End(prof.Begin(""))

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
	defer prof.End(prof.Begin(""))

	data := &submission.Data[0]

	submission.SubmittedSteps = database.Offset2Slice(submission.SubmittedSteps, data)
	for i := 0; i < len(submission.SubmittedSteps); i++ {
		DBSubmitted2Submitted(&submission.SubmittedSteps[i], data)
	}
}

func GetSubmissionByID(id database.ID, submission *Submission) error {
	defer prof.End(prof.Begin(""))

	if err := database.Read(SubmissionsDB, id, submission); err != nil {
		return err
	}

	DBSubmission2Submission(submission)
	return nil
}

func GetSubmissions(pos *int64, submissions []Submission) (int, error) {
	defer prof.End(prof.Begin(""))

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
	defer prof.End(prof.Begin(""))

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
	defer prof.End(prof.Begin(""))

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
	defer prof.End(prof.Begin(""))

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
	defer prof.End(prof.Begin(""))

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
	defer prof.End(prof.Begin(""))

	w.WriteString(`<p>`)
	w.WriteString(Ls(l, "Score"))
	w.WriteString(`: `)
	w.WriteInt(GetSubmittedStepScore(submittedStep))
	w.WriteString(`/`)
	w.WriteInt(GetStepMaximumScore(&submittedStep.Step))
	w.WriteString(`</p>`)
}

func DisplaySubmissionTotalScore(w *http.Response, submission *Submission) {
	defer prof.End(prof.Begin(""))

	var score, maximum int
	for i := 0; i < len(submission.SubmittedSteps); i++ {
		score += GetSubmittedStepScore(&submission.SubmittedSteps[i])
		maximum += GetStepMaximumScore(&submission.SubmittedSteps[i].Step)
	}

	w.WriteInt(score)
	w.WriteString(`/`)
	w.WriteInt(maximum)
}

func DisplaySubmissionLanguageSelect(w *http.Response, submittedTask *SubmittedProgramming, enabled bool) {
	defer prof.End(prof.Begin(""))

	w.WriteString(` <select name="LanguageID"`)
	if !enabled {
		w.WriteString(` disabled`)
	}
	w.WriteString(`>`)
	for i := database.ID(0); i < database.ID(len(ProgrammingLanguages)); i++ {
		lang := &ProgrammingLanguages[i]

		w.WriteString(`<option value="`)
		w.WriteID(i)
		w.WriteString(`"`)
		if i == submittedTask.LanguageID {
			w.WriteString(` selected`)
		}
		w.WriteString(`>`)
		w.WriteString(lang.Name)
		w.WriteString(`</option>`)
	}
	w.WriteString(`</select>`)
}

func DisplaySubmissionTitle(w *http.Response, l Language, subject *Subject, lesson *Lesson, user *User) {
	defer prof.End(prof.Begin(""))

	w.WriteString(Ls(l, "Submission"))
	w.WriteString(` `)
	w.WriteString(Ls(GL, "for"))
	w.WriteString(` «`)
	w.WriteHTMLString(subject.Name)
	w.WriteString(`: `)
	w.WriteHTMLString(lesson.Name)
	w.WriteString(`» `)
	w.WriteString(Ls(GL, "by"))
	w.WriteString(` `)
	w.WriteHTMLString(user.LastName)
	w.WriteString(` `)
	w.WriteHTMLString(user.FirstName)
}

func DisplaySubmissionLink(w *http.Response, l Language, submission *Submission) {
	defer prof.End(prof.Begin(""))

	var user User

	if err := GetUserByID(submission.UserID, &user); err != nil {
		/* TODO(anton2920): report error. */
	}

	w.WriteString(`<a href="/submission/`)
	w.WriteID(submission.ID)
	w.WriteString(`">`)
	w.WriteHTMLString(user.LastName)
	w.WriteString(` `)
	w.WriteHTMLString(user.FirstName)
	w.WriteString(` (`)
	switch submission.Status {
	case SubmissionCheckPending:
		w.WriteString(`<i>`)
		w.WriteString(Ls(l, "pending"))
		w.WriteString(` `)
		w.WriteString(Ls(l, "verification"))
		w.WriteString(`</i>`)
	case SubmissionCheckInProgress:
		w.WriteString(`<i>`)
		w.WriteString(Ls(l, "verification"))
		w.WriteString(` `)
		w.WriteString(Ls(l, "in progress"))
		w.WriteString(`</i>`)
	case SubmissionCheckDone:
		DisplaySubmissionTotalScore(w, submission)
	}
	w.WriteString(`)`)
	w.WriteString(`</a>`)
}

func SubmissionPageHandler(w *http.Response, r *http.Request) error {
	defer prof.End(prof.Begin(""))

	const width = WidthLarge

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
		w.WriteString(`<title>`)
		DisplaySubmissionTitle(w, GL, &subject, &lesson, &user)
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
			DisplayCrumbsLinkID(w, "/subject", subject.ID, subject.Name)
			DisplayCrumbsLinkID(w, "/lesson", lesson.ID, lesson.Name)
			DisplayCrumbsItem(w, GL, "Submission")
		}
		DisplayCrumbsEnd(w)

		DisplayPageStart(w, width)
		{
			w.WriteString(`<h2>`)
			DisplaySubmissionTitle(w, GL, &subject, &lesson, &user)
			w.WriteString(`</h2>`)
			w.WriteString(`<br>`)

			w.WriteString(`<p>`)
			w.WriteString(Ls(GL, "Started at"))
			w.WriteString(`: `)
			DisplayFormattedTime(w, submission.StartedAt)
			w.WriteString(`</p>`)

			w.WriteString(`<p>`)
			w.WriteString(Ls(GL, "Finished at"))
			w.WriteString(`: `)
			DisplayFormattedTime(w, submission.FinishedAt)
			w.WriteString(`</p>`)

			w.WriteString(`<form method="POST" action="/submission/results">`)

			DisplayHiddenID(w, "ID", id)

			for i := 0; i < len(submission.SubmittedSteps); i++ {
				submittedStep := &submission.SubmittedSteps[i]

				DisplayFrameStart(w)

				w.WriteString(`<p><b>`)
				w.WriteString(Ls(GL, "Step"))
				w.WriteString(` #`)
				w.WriteInt(i + 1)
				w.WriteString(`</b></p>`)

				w.WriteString(`<p>`)
				w.WriteString(Ls(GL, "Name"))
				w.WriteString(`: `)
				w.WriteHTMLString(submittedStep.Step.Name)
				w.WriteString(`</p>`)

				w.WriteString(`<p>`)
				w.WriteString(Ls(GL, "Type"))
				w.WriteString(`: `)
				w.WriteString(StepStringType(GL, &submittedStep.Step))
				w.WriteString(`</p>`)

				if submittedStep.Flags == SubmittedStepSkipped {
					DisplaySubmittedStepScore(w, GL, submittedStep)

					w.WriteString(`<p><i>`)
					w.WriteString(Ls(GL, "This step has been skipped"))
					w.WriteString(`.</i></p>`)
				} else {
					switch submittedStep.Status {
					case SubmissionCheckPending:
						w.WriteString(`<p><i>`)
						w.WriteString(Ls(GL, "Pending"))
						w.WriteString(` `)
						w.WriteString(Ls(GL, "verification"))
						w.WriteString(`...`)
						w.WriteString(`</i></p>`)
					case SubmissionCheckInProgress:
						w.WriteString(`<p><i>`)
						w.WriteString(Ls(GL, "Verification"))
						w.WriteString(` `)
						w.WriteString(Ls(GL, "in progress"))
						w.WriteString(`...`)
						w.WriteString(`</i></p>`)
					case SubmissionCheckDone:
						DisplaySubmittedStepScore(w, GL, submittedStep)
						DisplayErrorMessage(w, GL, submittedStep.Error)

						DisplayIndexedCommand(w, GL, i, "Open")
						if teacher {
							DisplayIndexedCommand(w, GL, i, "Re-check")
						}
					}
				}

				DisplayFrameEnd(w)
			}

			switch submission.Status {
			case SubmissionCheckPending:
				w.WriteString(`<p><i>`)
				w.WriteString(Ls(GL, "Pending"))
				w.WriteString(` `)
				w.WriteString(Ls(GL, "verification"))
				w.WriteString(`...`)
				w.WriteString(`</i></p>`)
			case SubmissionCheckInProgress:
				w.WriteString(`<p><i>`)
				w.WriteString(Ls(GL, "Verification"))
				w.WriteString(` `)
				w.WriteString(Ls(GL, "in progress"))
				w.WriteString(`...`)
				w.WriteString(`</i></p>`)
			case SubmissionCheckDone:
				w.WriteString(`<p>`)
				w.WriteString(Ls(GL, "Total score"))
				w.WriteString(`: `)
				DisplaySubmissionTotalScore(w, &submission)
				w.WriteString(`</p>`)
				if teacher {
					DisplayCommand(w, GL, "Re-check")
				}
			}

			w.WriteString(`</form>`)
		}
		DisplayPageEnd(w)
		DisplayMainEnd(w)
	}
	DisplayBodyEnd(w)

	DisplayHTMLEnd(w)
	return nil

}

func SubmissionResultsTestPageHandler(w *http.Response, r *http.Request, session *Session, subject *Subject, lesson *Lesson, submission *Submission, submittedTest *SubmittedTest) error {
	defer prof.End(prof.Begin(""))

	const width = WidthLarge

	test, _ := Step2Test(&submittedTest.Step)
	teacher := r.Form.Get("Teacher") != ""

	DisplayHTMLStart(w)

	DisplayHeadStart(w)
	{
		w.WriteString(`<title>`)
		w.WriteString(Ls(GL, "Submitted test"))
		w.WriteString(`: «`)
		w.WriteHTMLString(test.Name)
		w.WriteString(`»</title>`)
	}
	DisplayHeadEnd(w)

	DisplayBodyStart(w)
	{
		DisplayHeader(w, GL)
		DisplaySidebarWithLessons(w, GL, session, subject.Lessons)

		DisplayMainStart(w)

		DisplayCrumbsStart(w, width)
		{
			DisplayCrumbsLinkID(w, "/subject", subject.ID, subject.Name)
			DisplayCrumbsLinkID(w, "/lesson", lesson.ID, lesson.Name)
			DisplayCrumbsLinkID(w, "/submission", submission.ID, Ls(GL, "Submission"))
			DisplayCrumbsItem(w, GL, "Submitted test")
		}
		DisplayCrumbsEnd(w)

		DisplayPageStart(w, width)
		{
			w.WriteString(`<h2>`)
			w.WriteString(Ls(GL, "Submitted test"))
			w.WriteString(`: «`)
			w.WriteHTMLString(test.Name)
			w.WriteString(`»</h2>`)
			w.WriteString(`<br>`)

			if teacher {
				w.WriteString(`<p><i>`)
				w.WriteString(Ls(GL, "Note: answers marked with [x] are correct"))
				w.WriteString(`.</i></p>`)
			}

			for i := 0; i < len(test.Questions); i++ {
				question := &test.Questions[i]

				DisplayFrameStart(w)

				w.WriteString(`<p><b>`)
				w.WriteHTMLString(question.Name)
				w.WriteString(`</b></p>`)

				w.WriteString(`<ol>`)
				for j := 0; j < len(question.Answers); j++ {
					answer := question.Answers[j]

					if j > 0 {
						w.WriteString(`<br>`)
					}

					w.WriteString(`<li>`)

					w.WriteString(`<input type="`)
					if len(question.CorrectAnswers) > 1 {
						w.WriteString(`checkbox`)
					} else {
						w.WriteString(`radio`)
					}
					w.WriteString(`" name="SelectedAnswer`)
					w.WriteInt(i)
					w.WriteString(`" value="`)
					w.WriteInt(j)
					w.WriteString(`"`)

					for k := 0; k < len(submittedTest.SubmittedQuestions[i].SelectedAnswers); k++ {
						selectedAnswer := submittedTest.SubmittedQuestions[i].SelectedAnswers[k]
						if j == selectedAnswer {
							w.WriteString(` checked`)
							break
						}
					}

					w.WriteString(` disabled> `)

					w.WriteString(`<span>`)
					w.WriteHTMLString(answer)
					w.WriteString(`</span>`)

					if teacher {
						for k := 0; k < len(question.CorrectAnswers); k++ {
							correctAnswer := question.CorrectAnswers[k]
							if j == correctAnswer {
								w.WriteString(` <span>[x]</span>`)
								break
							}
						}
					}

					w.WriteString(`</li>`)
				}
				w.WriteString(`</ol>`)

				w.WriteString(`<span>`)
				w.WriteString(Ls(GL, "Score"))
				w.WriteString(`: `)
				w.WriteInt(submittedTest.Scores[i])
				w.WriteString(`/1</span>`)

				DisplayFrameEnd(w)
			}
		}
		DisplayPageEnd(w)
		DisplayMainEnd(w)
	}
	DisplayBodyEnd(w)

	DisplayHTMLEnd(w)
	return nil
}

func SubmissionResultsProgrammingDisplayChecks(w *http.Response, l Language, submittedTask *SubmittedProgramming, checkType CheckType) {
	defer prof.End(prof.Begin(""))

	task, _ := Step2Programming(&submittedTask.Step)
	scores := submittedTask.Scores[checkType]
	messages := submittedTask.Messages[checkType]

	w.WriteString(`<ol>`)
	for i := 0; i < len(task.Checks[checkType]); i++ {
		check := &task.Checks[checkType][i]
		score := scores[i]
		message := messages[i]

		w.WriteString(`<li class="mt-3">`)

		w.WriteString(`<label>`)
		w.WriteString(Ls(l, "Input"))
		w.WriteString(`: `)

		w.WriteString(`<textarea class="btn btn-outline-dark" rows="1" readonly>`)
		w.WriteHTMLString(check.Input)
		w.WriteString(`</textarea>`)

		w.WriteString(`</label> `)

		w.WriteString(`<label>`)
		w.WriteString(Ls(l, "output"))
		w.WriteString(`: `)

		w.WriteString(`<textarea class="btn btn-outline-dark" rows="1" readonly>`)
		w.WriteHTMLString(check.Output)
		w.WriteString(`</textarea>`)

		w.WriteString(`</label> `)

		w.WriteString(`<span>`)
		w.WriteString(Ls(GL, "score"))
		w.WriteString(`: `)
		w.WriteInt(score)
		w.WriteString(`/1</span>`)

		if message != "" {
			w.WriteString(` <span>`)
			w.WriteHTMLString(message)
			w.WriteString(`</span>`)
		}

		w.WriteString(`</li>`)
	}
	w.WriteString(`</ol>`)
}

func SubmissionResultsProgrammingPageHandler(w *http.Response, r *http.Request, session *Session, subject *Subject, lesson *Lesson, submission *Submission, submittedTask *SubmittedProgramming) error {
	defer prof.End(prof.Begin(""))

	const width = WidthLarge

	task, _ := Step2Programming(&submittedTask.Step)
	teacher := r.Form.Get("Teacher") != ""

	DisplayHTMLStart(w)

	DisplayHeadStart(w)
	{
		w.WriteString(`<title>`)
		w.WriteString(Ls(GL, "Submitted programming task"))
		w.WriteString(`: «`)
		w.WriteHTMLString(task.Name)
		w.WriteString(`»</title>`)
	}
	DisplayHeadEnd(w)

	DisplayBodyStart(w)
	{
		DisplayHeader(w, GL)
		DisplaySidebarWithLessons(w, GL, session, subject.Lessons)

		DisplayMainStart(w)

		DisplayCrumbsStart(w, width)
		{
			DisplayCrumbsLinkID(w, "/subject", subject.ID, subject.Name)
			DisplayCrumbsLinkID(w, "/lesson", lesson.ID, lesson.Name)
			DisplayCrumbsLinkID(w, "/submission", submission.ID, Ls(GL, "Submission"))
			DisplayCrumbsItem(w, GL, "Submitted programming task")
		}
		DisplayCrumbsEnd(w)

		DisplayPageStart(w, width)
		{
			w.WriteString(`<h2>`)
			w.WriteString(Ls(GL, "Submitted programming task"))
			w.WriteString(`: «`)
			w.WriteHTMLString(task.Name)
			w.WriteString(`»</h2>`)
			w.WriteString(`<br>`)

			w.WriteString(`<h3>`)
			w.WriteString(Ls(GL, "Description"))
			w.WriteString(`</h3>`)
			w.WriteString(`<p>`)
			w.WriteHTMLString(task.Description)
			w.WriteString(`</p>`)
			w.WriteString(`<br>`)

			w.WriteString(`<h3>`)
			w.WriteString(Ls(GL, "Examples"))
			w.WriteString(`</h3>`)
			SubmissionNewDisplayProgrammingChecks(w, GL, task, CheckTypeExample)
			w.WriteString(`<br>`)

			w.WriteString(`<h3>`)
			w.WriteString(Ls(GL, "Solution"))
			w.WriteString(`</h3>`)

			DisplayLabel(w, GL, "Programming language")
			DisplaySubmissionLanguageSelect(w, submittedTask, false)
			w.WriteString(`<br>`)

			w.WriteString(`<textarea class="form-control" rows="10" readonly>`)
			w.WriteHTMLString(submittedTask.Solution)
			w.WriteString(`</textarea>`)

			if teacher {
				w.WriteString(`<br><br>`)
				w.WriteString(`<h3>`)
				w.WriteString(Ls(GL, "Tests"))
				w.WriteString(`</h3>`)
				SubmissionResultsProgrammingDisplayChecks(w, GL, submittedTask, CheckTypeTest)
			}
		}
		DisplayPageEnd(w)
		DisplayMainEnd(w)
	}
	DisplayBodyEnd(w)

	DisplayHTMLEnd(w)
	return nil
}

func SubmissionResultsStepPageHandler(w *http.Response, r *http.Request, session *Session, subject *Subject, lesson *Lesson, submission *Submission, submittedStep *SubmittedStep) error {
	defer prof.End(prof.Begin(""))

	switch submittedStep.Type {
	default:
		panic("invalid step type")
	case SubmittedTypeTest:
		submittedTest, _ := Submitted2Test(submittedStep)
		return SubmissionResultsTestPageHandler(w, r, session, subject, lesson, submission, submittedTest)
	case SubmittedTypeProgramming:
		submittedTask, _ := Submitted2Programming(submittedStep)
		return SubmissionResultsProgrammingPageHandler(w, r, session, subject, lesson, submission, submittedTask)
	}
}

func SubmissionResultsPageHandler(w *http.Response, r *http.Request) error {
	defer prof.End(prof.Begin(""))

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

	for i := 0; i < len(r.Form.Keys); i++ {
		k := r.Form.Keys[i]

		/* 'command' is button, which modifies content of a current page. */
		if strings.StartsWith(k, "Command") {
			if len(r.Form.Values[i]) == 0 {
				continue
			}
			v := r.Form.Values[i][0]

			pindex, spindex, _, _, err := GetIndicies(k[len("Command"):])
			if err != nil {
				return http.ClientError(err)
			}

			switch v {
			default:
				return http.ClientError(nil)
			case Ls(GL, "Open"):
				if (pindex < 0) || (pindex >= len(submission.SubmittedSteps)) {
					return http.ClientError(nil)
				}
				submittedStep := &submission.SubmittedSteps[pindex]

				return SubmissionResultsStepPageHandler(w, r, session, &subject, &lesson, &submission, submittedStep)
			case Ls(GL, "Re-check"):
				if spindex != "" {
					if (pindex < 0) || (pindex >= len(submission.SubmittedSteps)) {
						return http.ClientError(nil)
					}
					submittedStep := &submission.SubmittedSteps[pindex]

					submission.Status = SubmissionCheckPending
					submittedStep.Status = SubmissionCheckPending
					submittedStep.Error = ""
				} else {
					submission.Status = SubmissionCheckPending
					for i := 0; i < len(submission.SubmittedSteps); i++ {
						submittedStep := &submission.SubmittedSteps[i]
						submittedStep.Status = SubmissionCheckPending
						submittedStep.Error = ""
					}
				}
				if err := SaveSubmission(&submission); err != nil {
					return http.ServerError(err)
				}

				SubmissionVerifyChannel <- submission.ID

				w.RedirectID("/submission/", submission.ID, http.StatusSeeOther)
				return nil
			}
		}
	}

	return http.ClientError(nil)
}

func SubmittedStepClear(submittedStep *SubmittedStep) {
	defer prof.End(prof.Begin(""))

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
	defer prof.End(prof.Begin(""))

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
	defer prof.End(prof.Begin(""))

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
	defer prof.End(prof.Begin(""))

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

func SubmissionNewTestPageHandler(w *http.Response, r *http.Request, session *Session, subject *Subject, lesson *Lesson, submittedTest *SubmittedTest, err error) error {
	defer prof.End(prof.Begin(""))

	const width = WidthMedium

	test, _ := Step2Test(&submittedTest.Step)

	DisplayHTMLStart(w)

	DisplayHeadStart(w)
	{
		w.WriteString(`<title>`)
		w.WriteString(Ls(GL, "Test"))
		w.WriteString(`: «`)
		w.WriteHTMLString(test.Name)
		w.WriteString(`»</title>`)
	}
	DisplayHeadEnd(w)

	DisplayBodyStart(w)
	{
		DisplayHeader(w, GL)
		DisplaySidebar(w, GL, session)

		DisplayMainStart(w)

		DisplayFormStart(w, r, "/submission/new")
		DisplayHiddenString(w, "CurrentPage", "Test")
		DisplayHiddenString(w, "SubmissionIndex", r.Form.Get("SubmissionIndex"))
		DisplayHiddenString(w, "StepIndex", r.Form.Get("StepIndex"))

		DisplayCrumbsStart(w, width)
		{
			DisplayCrumbsLinkID(w, "/subject", subject.ID, subject.Name)
			DisplayCrumbsLinkID(w, "/lesson", lesson.ID, lesson.Name)
			DisplayCrumbsSubmit(w, GL, "Back", "Evaluation pass")
			DisplayCrumbsItemRaw(w, test.Name)
		}
		DisplayCrumbsEnd(w)

		DisplayPageStart(w, width)
		{
			w.WriteString(`<h3 class="text-center">`)
			w.WriteString(Ls(GL, "Test"))
			w.WriteString(`: «`)
			w.WriteHTMLString(test.Name)
			w.WriteString(`»</h3>`)
			w.WriteString(`<br>`)

			DisplayError(w, GL, err)

			if len(submittedTest.SubmittedQuestions) == 0 {
				submittedTest.SubmittedQuestions = make([]SubmittedQuestion, len(test.Questions))
			}
			for i := 0; i < len(submittedTest.SubmittedQuestions); i++ {
				submittedQuestion := &submittedTest.SubmittedQuestions[i]
				question := &test.Questions[i]

				DisplayFrameStart(w)

				w.WriteString(`<p><b>`)
				w.WriteHTMLString(question.Name)
				w.WriteString(`</b></p>`)

				w.WriteString(`<ol>`)
				for j := 0; j < len(question.Answers); j++ {
					answer := question.Answers[j]

					if j > 0 {
						w.WriteString(`<br>`)
					}

					w.WriteString(`<li>`)

					w.WriteString(`<input type="`)
					if len(question.CorrectAnswers) > 1 {
						w.WriteString(`checkbox`)
					} else {
						w.WriteString(`radio`)
					}
					w.WriteString(`" name="SelectedAnswer`)
					w.WriteInt(i)
					w.WriteString(`" value="`)
					w.WriteInt(j)
					w.WriteString(`"`)

					for k := 0; k < len(submittedQuestion.SelectedAnswers); k++ {
						if j == submittedQuestion.SelectedAnswers[k] {
							w.WriteString(` checked`)
							break
						}
					}

					w.WriteString(`> `)

					w.WriteString(`<span>`)
					w.WriteHTMLString(answer)
					w.WriteString(`</span>`)

					w.WriteString(`</li>`)
				}
				w.WriteString(`</ol>`)

				DisplayFrameEnd(w)
			}
			DisplaySubmit(w, GL, "NextPage", "Save", true)
			DisplaySubmit(w, GL, "NextPage", "Discard", true)
		}
		DisplayPageEnd(w)
		DisplayFormEnd(w)
		DisplayMainEnd(w)
	}
	DisplayBodyEnd(w)

	DisplayHTMLEnd(w)
	return nil
}

func SubmissionNewProgrammingFillFromRequest(vs url.Values, submittedTask *SubmittedProgramming) error {
	defer prof.End(prof.Begin(""))

	id, err := GetValidID(vs.Get("LanguageID"), database.ID(len(ProgrammingLanguages)))
	if err != nil {
		return http.ClientError(err)
	}
	submittedTask.LanguageID = id

	submittedTask.Solution = vs.Get("Solution")
	return nil
}

func SubmissionNewProgrammingVerify(submittedTask *SubmittedProgramming, l Language) error {
	defer prof.End(prof.Begin(""))

	if !ProgrammingLanguages[submittedTask.LanguageID].Available {
		return http.BadRequest(Ls(l, "selected language is not available"))
	}

	if !strings.LengthInRange(submittedTask.Solution, MinSolutionLen, MaxSolutionLen) {
		return http.BadRequest(Ls(l, "solution length must be between %d and %d characters long"), MinSolutionLen, MaxSolutionLen)
	}

	return nil
}

func SubmissionNewDisplayProgrammingChecks(w *http.Response, l Language, task *StepProgramming, checkType CheckType) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`<ol>`)
	for i := 0; i < len(task.Checks[checkType]); i++ {
		check := &task.Checks[checkType][i]

		w.WriteString(`<li class="mt-3">`)

		w.WriteString(`<label>`)
		w.WriteString(Ls(l, "Input"))
		w.WriteString(`: `)

		w.WriteString(`<textarea class="btn btn-outline-dark" rows="1" readonly>`)
		w.WriteHTMLString(check.Input)
		w.WriteString(`</textarea>`)

		w.WriteString(`</label> `)

		w.WriteString(`<label>`)
		w.WriteString(Ls(l, "output"))
		w.WriteString(`: `)

		w.WriteString(`<textarea class="btn btn-outline-dark" rows="1" readonly>`)
		w.WriteHTMLString(check.Output)
		w.WriteString(`</textarea>`)

		w.WriteString(`</label>`)

		w.WriteString(`</li>`)
	}
	w.WriteString(`</ol>`)
}

func SubmissionNewProgrammingPageHandler(w *http.Response, r *http.Request, session *Session, subject *Subject, lesson *Lesson, submittedTask *SubmittedProgramming, err error) error {
	defer prof.End(prof.Begin(""))

	const width = WidthLarge

	task, _ := Step2Programming(&submittedTask.Step)

	DisplayHTMLStart(w)

	DisplayHeadStart(w)
	{
		w.WriteString(`<title>`)
		w.WriteString(Ls(GL, "Programming task"))
		w.WriteString(`: «`)
		w.WriteHTMLString(task.Name)
		w.WriteString(`»</title>`)
	}
	DisplayHeadEnd(w)

	DisplayBodyStart(w)
	{
		DisplayHeader(w, GL)
		DisplaySidebar(w, GL, session)

		DisplayMainStart(w)

		DisplayFormStart(w, r, "/submission/new")
		DisplayHiddenString(w, "CurrentPage", "Programming")
		DisplayHiddenString(w, "SubmissionIndex", r.Form.Get("SubmissionIndex"))
		DisplayHiddenString(w, "StepIndex", r.Form.Get("StepIndex"))

		DisplayCrumbsStart(w, width)
		{
			DisplayCrumbsLinkID(w, "/subject", subject.ID, subject.Name)
			DisplayCrumbsLinkID(w, "/lesson", lesson.ID, lesson.Name)
			DisplayCrumbsSubmit(w, GL, "Back", "Evaluation pass")
			DisplayCrumbsItemRaw(w, task.Name)
		}
		DisplayCrumbsEnd(w)

		DisplayPageStart(w, width)
		{
			w.WriteString(`<h3 class="text-center">`)
			w.WriteString(Ls(GL, "Programming task"))
			w.WriteString(`: «`)
			w.WriteHTMLString(task.Name)
			w.WriteString(`»</h3>`)
			w.WriteString(`<br>`)

			DisplayError(w, GL, err)

			w.WriteString(`<h4>`)
			w.WriteString(Ls(GL, "Description"))
			w.WriteString(`</h4>`)
			w.WriteString(`<p>`)
			w.WriteHTMLString(task.Description)
			w.WriteString(`</p>`)

			w.WriteString(`<h4>`)
			w.WriteString(Ls(GL, "Examples"))
			w.WriteString(`</h4>`)
			SubmissionNewDisplayProgrammingChecks(w, GL, task, CheckTypeExample)

			w.WriteString(`<h4>`)
			w.WriteString(Ls(GL, "Solution"))
			w.WriteString(`</h4>`)

			DisplayLabel(w, GL, "Programming language")
			DisplaySubmissionLanguageSelect(w, submittedTask, true)
			w.WriteString(`</label>`)
			w.WriteString(`<br>`)

			w.WriteString(`<textarea class="form-control" rows="10" name="Solution">`)
			w.WriteHTMLString(submittedTask.Solution)
			w.WriteString(`</textarea>`)

			w.WriteString(`<br><br>`)

			DisplaySubmit(w, GL, "NextPage", "Save", true)
			DisplaySubmit(w, GL, "NextPage", "Discard", true)
		}
		DisplayPageEnd(w)
		DisplayFormEnd(w)
		DisplayMainEnd(w)
	}
	DisplayBodyEnd(w)

	DisplayHTMLEnd(w)
	return nil
}

func SubmissionNewStepVerify(l Language, submittedStep *SubmittedStep) error {
	defer prof.End(prof.Begin(""))

	switch submittedStep.Type {
	default:
		panic("invalid step type")
	case SubmittedTypeTest:
		submittedTest, _ := Submitted2Test(submittedStep)
		return SubmissionNewTestVerify(l, submittedTest)
	case SubmittedTypeProgramming:
		submittedTask, _ := Submitted2Programming(submittedStep)
		if err := SubmissionNewProgrammingVerify(submittedTask, GL); err != nil {
			return err
		}

		if err := SubmissionVerifyProgramming(GL, submittedTask, CheckTypeExample); err != nil {
			return http.BadRequest(err.Error())
		}

		scores := submittedTask.Scores[CheckTypeExample]
		messages := submittedTask.Messages[CheckTypeExample]
		for i := 0; i < len(scores); i++ {
			if scores[i] == 0 {
				return http.BadRequest(Ls(GL, "example %d: %s"), i+1, messages[i])
			}
		}
	}
	return nil
}

func SubmissionNewStepPageHandler(w *http.Response, r *http.Request, session *Session, subject *Subject, lesson *Lesson, submittedStep *SubmittedStep, err error) error {
	defer prof.End(prof.Begin(""))

	switch submittedStep.Type {
	default:
		panic("invalid step type")
	case SubmittedTypeTest:
		submittedTest, _ := Submitted2Test(submittedStep)
		return SubmissionNewTestPageHandler(w, r, session, subject, lesson, submittedTest, err)
	case SubmittedTypeProgramming:
		submittedTask, _ := Submitted2Programming(submittedStep)
		return SubmissionNewProgrammingPageHandler(w, r, session, subject, lesson, submittedTask, err)
	}
}

func SubmissionNewMainPageHandler(w *http.Response, r *http.Request, session *Session, subject *Subject, lesson *Lesson, submission *Submission, err error) error {
	defer prof.End(prof.Begin(""))

	const width = WidthSmall

	DisplayHTMLStart(w)

	DisplayHeadStart(w)
	{
		w.WriteString(`<title>`)
		w.WriteString(Ls(GL, "Evaluation"))
		w.WriteString(` `)
		w.WriteString(Ls(GL, "for"))
		w.WriteString(` «`)
		w.WriteHTMLString(lesson.Name)
		w.WriteString(`»</title>`)
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
			DisplayCrumbsLinkID(w, "/lesson", lesson.ID, lesson.Name)
			DisplayCrumbsItem(w, GL, "Evaluation pass")
		}
		DisplayCrumbsEnd(w)

		DisplayPageStart(w, width)
		{
			w.WriteString(`<h3 class="text-center">`)
			w.WriteString(Ls(GL, "Evaluation"))
			w.WriteString(` `)
			w.WriteString(Ls(GL, "for"))
			w.WriteString(` «`)
			w.WriteHTMLString(lesson.Name)
			w.WriteString(`»</h3>`)
			w.WriteString(`<br>`)

			DisplayError(w, GL, err)

			DisplayFormStart(w, r, "/submission/new")
			DisplayHiddenString(w, "CurrentPage", "Main")
			DisplayHiddenString(w, "SubmissionIndex", r.Form.Get("SubmissionIndex"))

			for i := 0; i < len(submission.SubmittedSteps); i++ {
				submittedStep := &submission.SubmittedSteps[i]

				DisplayFrameStart(w)

				w.WriteString(`<p><b>`)
				w.WriteString(Ls(GL, "Step"))
				w.WriteString(` #`)
				w.WriteInt(i + 1)
				DisplayDraft(w, GL, submittedStep.Flags == SubmittedStepDraft)
				w.WriteString(`</b></p>`)

				w.WriteString(`<p>`)
				w.WriteString(Ls(GL, "Name"))
				w.WriteString(`: `)
				w.WriteHTMLString(submittedStep.Step.Name)
				w.WriteString(`</p>`)

				w.WriteString(`<p>`)
				w.WriteString(Ls(GL, "Type"))
				w.WriteString(`: `)
				w.WriteString(StepStringType(GL, &submittedStep.Step))
				w.WriteString(`</p>`)

				if submittedStep.Flags == SubmittedStepSkipped {
					DisplayIndexedCommand(w, GL, i, "Pass")
				} else {
					DisplayIndexedCommand(w, GL, i, "Edit")
				}

				DisplayFrameEnd(w)
			}

			DisplaySubmit(w, GL, "NextPage", "Finish", true)

			DisplayFormEnd(w)
		}
		DisplayPageEnd(w)
		DisplayMainEnd(w)
	}
	DisplayBodyEnd(w)

	DisplayHTMLEnd(w)
	return nil
}

func SubmissionNewHandleCommand(w *http.Response, r *http.Request, l Language, session *Session, subject *Subject, lesson *Lesson, submission *Submission, currentPage, k, command string) error {
	defer prof.End(prof.Begin(""))

	pindex, spindex, _, _, err := GetIndicies(k[len("Command"):])
	if err != nil {
		return http.ClientError(err)
	}

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
		return SubmissionNewStepPageHandler(w, r, session, subject, lesson, submittedStep, nil)
	}

}

func SubmissionNewPageHandler(w *http.Response, r *http.Request) error {
	defer prof.End(prof.Begin(""))

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
	if lesson.ContainerType != LessonContainerSubject {
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

	for i := 0; i < len(r.Form.Keys); i++ {
		k := r.Form.Keys[i]

		/* 'command' is button, which modifies content of a current page. */
		if strings.StartsWith(k, "Command") {
			if len(r.Form.Values[i]) == 0 {
				continue
			}
			v := r.Form.Values[i][0]

			/* NOTE(anton2920): after command is executed, function must return. */
			return SubmissionNewHandleCommand(w, r, GL, session, &subject, &lesson, &submission, currentPage, k, v)
		}
	}

	/* 'currentPage' is the page to save before leaving it. */
	switch currentPage {
	case "Test":
		si, err := GetValidIndex(r.Form.Get("StepIndex"), len(lesson.Steps))
		if err != nil {
			return http.ClientError(err)
		}
		submittedStep := &submission.SubmittedSteps[si]

		submittedTest, err := Submitted2Test(submittedStep)
		if err != nil {
			return http.ClientError(err)
		}
		if err := SubmissionNewTestFillFromRequest(r.Form, submittedTest); err != nil {
			return SubmissionNewTestPageHandler(w, r, session, &subject, &lesson, submittedTest, err)
		}
	case "Programming":
		si, err := GetValidIndex(r.Form.Get("StepIndex"), len(lesson.Steps))
		if err != nil {
			return http.ClientError(err)
		}
		submittedStep := &submission.SubmittedSteps[si]

		submittedTask, err := Submitted2Programming(submittedStep)
		if err != nil {
			return http.ClientError(err)
		}
		if err := SubmissionNewProgrammingFillFromRequest(r.Form, submittedTask); err != nil {
			return SubmissionNewProgrammingPageHandler(w, r, session, &subject, &lesson, submittedTask, err)
		}
	}

	switch nextPage {
	default:
		return SubmissionNewMainPageHandler(w, r, session, &subject, &lesson, &submission, nil)
	case Ls(GL, "Save"):
		si, err := GetValidIndex(r.Form.Get("StepIndex"), len(lesson.Steps))
		if err != nil {
			return http.ClientError(err)
		}
		submittedStep := &submission.SubmittedSteps[si]

		if err := SubmissionNewStepVerify(GL, submittedStep); err != nil {
			return SubmissionNewStepPageHandler(w, r, session, &subject, &lesson, submittedStep, err)
		}
		submittedStep.Flags = SubmittedStepPassed

		return SubmissionNewMainPageHandler(w, r, session, &subject, &lesson, &submission, nil)
	case Ls(GL, "Discard"):
		si, err := GetValidIndex(r.Form.Get("StepIndex"), len(lesson.Steps))
		if err != nil {
			return http.ClientError(err)
		}
		submittedStep := &submission.SubmittedSteps[si]
		submittedStep.Flags = SubmittedStepSkipped
		SubmittedStepClear(submittedStep)

		return SubmissionNewMainPageHandler(w, r, session, &subject, &lesson, &submission, nil)
	case Ls(GL, "Finish"):
		if err := SubmissionNewVerify(GL, &submission); err != nil {
			return SubmissionNewMainPageHandler(w, r, session, &subject, &lesson, &submission, err)
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
