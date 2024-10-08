package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync/atomic"
	sys "syscall"
	"time"
	"unsafe"

	"github.com/anton2920/gofa/database"
	"github.com/anton2920/gofa/jail"
	"github.com/anton2920/gofa/log"
	"github.com/anton2920/gofa/syscall"
	"github.com/anton2920/gofa/trace"
)

type SubmissionCheckStatus int

const (
	SubmissionCheckPending SubmissionCheckStatus = iota
	SubmissionCheckInProgress
	SubmissionCheckDone
)

var SubmissionVerifyChannel = make(chan database.ID, 128)

func SubmissionVerifyTest(submittedTest *SubmittedTest) error {
	defer trace.End(trace.Begin(""))

	test, _ := Step2Test(&submittedTest.Step)

	scores := make([]int, len(test.Questions))
	for i := 0; i < len(test.Questions); i++ {
		question := test.Questions[i]
		submittedQuestion := submittedTest.SubmittedQuestions[i]

		correctAnswers := question.CorrectAnswers
		selectedAnswers := submittedQuestion.SelectedAnswers

		if len(correctAnswers) != len(selectedAnswers) {
			continue
		}

		var nfound int
		for j := 0; j < len(selectedAnswers); j++ {
			selectedAnswer := selectedAnswers[j]

			for k := 0; k < len(correctAnswers); k++ {
				correctAnswer := correctAnswers[k]
				if correctAnswer == selectedAnswer {
					nfound++
					break
				}
			}
		}

		if nfound == len(selectedAnswers) {
			scores[i] = 1
		}
	}
	submittedTest.Scores = scores

	return nil
}

func PutProgrammingSource(buffer []byte, j jail.Jail, lang *ProgrammingLanguage) int {
	defer trace.End(trace.Begin(""))

	var n int

	n += jail.PutEnv(buffer[n:], j)

	buffer[n] = '/'
	n++

	n += copy(buffer[n:], lang.SourceFile)

	return n
}

func PutProgrammingExecutable(buffer []byte, j jail.Jail, lang *ProgrammingLanguage) int {
	defer trace.End(trace.Begin(""))

	var n int

	n += jail.PutEnv(buffer[n:], j)

	buffer[n] = '/'
	n++

	n += copy(buffer[n:], lang.Executable)

	return n
}

func SubmissionVerifyProgrammingCreateSource(j jail.Jail, lang *ProgrammingLanguage, solution string) error {
	defer trace.End(trace.Begin(""))

	buffer := make([]byte, syscall.PATH_MAX)
	n := PutProgrammingSource(buffer, j, lang)
	source := unsafe.String(unsafe.SliceData(buffer), n)

	fd, err := syscall.Open(source, syscall.O_WRONLY|syscall.O_CREAT, 0644)
	if err != nil {
		return fmt.Errorf("failed to create source file: %w", err)
	}

	if _, err := syscall.Write(fd, unsafe.Slice(unsafe.StringData(solution), len(solution))); err != nil {
		if err := syscall.Close(fd); err != nil {
			log.Warnf("Failed to close source file: %v", err)
		}
		return fmt.Errorf("failed to write data to a source file: %w", err)
	}

	if err := syscall.Close(fd); err != nil {
		log.Warnf("Failed to close a source file: %v", err)
	}

	return nil
}

func SubmissionVerifyProgrammingCleanup(j jail.Jail, lang *ProgrammingLanguage) error {
	defer trace.End(trace.Begin(""))

	var err error

	buffer := make([]byte, syscall.PATH_MAX)
	n := PutProgrammingSource(buffer, j, lang)
	source := unsafe.String(unsafe.SliceData(buffer), n)

	if err1 := syscall.Unlink(source); err1 != nil {
		if err1.(syscall.Error).Errno != syscall.ENOENT {
			err = errors.Join(err, fmt.Errorf("failed to remove source file: %w", err1))
		}
	}

	if lang.Executable != "" {
		n = PutProgrammingExecutable(buffer, j, lang)
		executable := unsafe.String(unsafe.SliceData(buffer), n)

		if err1 := syscall.Unlink(executable); err1 != nil {
			if err1.(syscall.Error).Errno != syscall.ENOENT {
				err = errors.Join(err, fmt.Errorf("failed to remove executable: %w", err1))
			}
		}
	}

	return nil
}

func SubmissionVerifyProgramWatchdog(cmd *exec.Cmd, seconds time.Duration, done <-chan struct{}, timeout *int32) {
	select {
	case <-time.After(seconds * time.Second):
		pid := atomic.LoadInt32((*int32)(unsafe.Pointer(&cmd.Process.Pid)))
		syscall.Kill(-pid, syscall.SIGKILL)
		atomic.StoreInt32(timeout, 1)
	case <-done:
		atomic.StoreInt32(timeout, 0)
	}
}

/* TODO(anton2920): rewrite without using standard library. */
func SubmissionVerifyProgrammingCompile(l Language, j jail.Jail, lang *ProgrammingLanguage) error {
	defer trace.End(trace.Begin(""))

	var buffer bytes.Buffer

	cmd := exec.Command(lang.Compiler, append(lang.CompilerArgs, lang.SourceFile)...)
	cmd.Dir = "/tmp"
	cmd.SysProcAttr = &sys.SysProcAttr{Setsid: true, Jail: int(j.ID)}
	cmd.Stdout = &buffer
	cmd.Stderr = &buffer

	done := make(chan struct{})

	const timeout = 5
	var timeoutExceeded int32
	go SubmissionVerifyProgramWatchdog(cmd, timeout, done, &timeoutExceeded)

	err := cmd.Run()
	close(done)

	if atomic.LoadInt32(&timeoutExceeded) == 1 {
		return fmt.Errorf(Ls(l, "failed to compile program: exceeded compilation timeout of %d seconds"), timeout)
	}
	if err != nil {
		return fmt.Errorf(Ls(l, "failed to compile program: %s %w"), buffer.String(), err)
	}
	return nil
}

func SubmissionVerifyProgrammingRun(l Language, j jail.Jail, lang *ProgrammingLanguage, input string, output *bytes.Buffer) error {
	defer trace.End(trace.Begin(""))

	var exe string
	var args []string
	if lang.Executable != "" {
		exe = lang.Executable
	} else {
		exe = lang.Runner
		args = append(lang.RunnerArgs, lang.SourceFile)
	}

	cmd := exec.Command(exe, args...)
	cmd.Dir = "/tmp"
	cmd.SysProcAttr = &sys.SysProcAttr{Setsid: true, Jail: int(j.ID)}
	cmd.Stdout = output
	cmd.Stderr = output

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf(Ls(l, "failed to create stdin pipe: %w"), err)
	}
	if _, err := io.WriteString(stdin, input); err != nil {
		return fmt.Errorf(Ls(l, "failed to write input string: %w"), err)
	}
	stdin.Close()

	done := make(chan struct{})

	const timeout = 2
	var timeoutExceeded int32
	go SubmissionVerifyProgramWatchdog(cmd, timeout, done, &timeoutExceeded)

	err = cmd.Run()
	close(done)

	if atomic.LoadInt32(&timeoutExceeded) == 1 {
		return fmt.Errorf(Ls(l, "failed to run program: exceeded timeout of %d seconds"), timeout)
	}
	if err != nil {
		return fmt.Errorf(Ls(l, "failed to run program: %s %w"), output.String(), err)
	}
	return nil
}

func SubmissionVerifyProgrammingCheck(l Language, j jail.Jail, submittedTask *SubmittedProgramming, checkType CheckType) {
	defer trace.End(trace.Begin(""))

	var output bytes.Buffer

	task, _ := Step2Programming(&submittedTask.Step)
	lang := &ProgrammingLanguages[submittedTask.LanguageID]

	scores := make([]int, len(task.Checks[checkType]))
	messages := make([]string, len(task.Checks[checkType]))
	for i := 0; i < len(task.Checks[checkType]); i++ {
		output.Reset()

		check := &task.Checks[checkType][i]
		input := strings.Replace(strings.TrimSpace(check.Input), "\r\n", "\n", -1)
		if err := SubmissionVerifyProgrammingRun(l, j, lang, input, &output); err != nil {
			messages[i] = err.Error()
			if checkType == CheckTypeExample {
				break
			}
			continue
		}

		expectedOutput := strings.Replace(strings.TrimSpace(check.Output), "\r\n", "\n", -1)
		actualOutput := strings.Replace(strings.TrimSpace(output.String()), "\r\n", "\n", -1)
		if actualOutput != expectedOutput {
			messages[i] = fmt.Sprintf(Ls(l, "expected %q, got %q"), expectedOutput, actualOutput)
			if checkType == CheckTypeExample {
				break
			}
			continue
		}

		scores[i] = 1
	}

	submittedTask.Scores[checkType] = scores
	submittedTask.Messages[checkType] = messages
}

func SubmissionVerifyProgramming(l Language, submittedTask *SubmittedProgramming, checkType CheckType) error {
	defer trace.End(trace.Begin(""))

	lang := &ProgrammingLanguages[submittedTask.LanguageID]

	j, err := jail.New("/usr/local/jails/templates/workster", WorkingDirectory)
	if err != nil {
		return err
	}
	defer func(j jail.Jail) {
		if err := jail.Remove(j); err != nil {
			log.Warnf("Failed to remove jail: %v", err)
		}
	}(j)

	if err := SubmissionVerifyProgrammingCreateSource(j, lang, submittedTask.Solution); err != nil {
		return err
	}
	defer func(j jail.Jail, lang *ProgrammingLanguage) {
		if err := SubmissionVerifyProgrammingCleanup(j, lang); err != nil {
			log.Warnf("Failed to cleanup jail environment: %v", err)
		}
	}(j, lang)

	if lang.Compiler != "" {
		if err := SubmissionVerifyProgrammingCompile(l, j, lang); err != nil {
			return err
		}
	}

	if err := jail.Protect(j); err != nil {
		return err
	}

	SubmissionVerifyProgrammingCheck(l, j, submittedTask, checkType)
	return nil
}

func SubmissionVerifyStep(submittedStep *SubmittedStep) {
	defer trace.End(trace.Begin(""))

	if submittedStep.Flags == SubmittedStepPassed {
		switch submittedStep.Type {
		case SubmittedTypeTest:
			submittedTest, _ := Submitted2Test(submittedStep)
			if submittedTest.Status == SubmissionCheckPending {
				submittedTest.Status = SubmissionCheckInProgress
				SubmissionVerifyTest(submittedTest)
				submittedTest.Status = SubmissionCheckDone
			}
		case SubmittedTypeProgramming:
			submittedTask, _ := Submitted2Programming(submittedStep)
			if submittedTask.Status == SubmissionCheckPending {
				submittedTask.Status = SubmissionCheckInProgress
				if err := SubmissionVerifyProgramming(GL, submittedTask, CheckTypeTest); err != nil {
					submittedTask.Error = err.Error()
				}
				submittedTask.Status = SubmissionCheckDone
			}
		}
	}
}

func SubmissionVerify(submission *Submission) {
	defer trace.End(trace.Begin(""))

	for i := 0; i < len(submission.SubmittedSteps); i++ {
		SubmissionVerifyStep(&submission.SubmittedSteps[i])
	}
}

func SubmissionVerifyWorker() {
	defer trace.End(trace.Begin(""))

	var submission Submission

	for submissionID := range SubmissionVerifyChannel {
		start := time.Now()

		if err := GetSubmissionByID(submissionID, &submission); err != nil {
			/* TODO(anton2920): report error. */
		}
		if submission.Status == SubmissionCheckPending {
			submission.Status = SubmissionCheckInProgress
			SubmissionVerify(&submission)
			submission.Status = SubmissionCheckDone
		}
		if err := SaveSubmission(&submission); err != nil {
			/* TODO(anton2920): report error. */
		}

		log.Debugf("Verified submission with ID = %d, took %v", submission.ID, time.Since(start))
	}
}
