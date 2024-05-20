package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	sys "syscall"
	"time"
	"unsafe"

	"github.com/anton2920/gofa/jail"
	"github.com/anton2920/gofa/log"
	"github.com/anton2920/gofa/syscall"
)

var SubmissionVerifyChannel = make(chan *Submission)

func SubmissionVerifyTest(submittedTest *SubmittedTest) error {
	test := submittedTest.Test

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
	var n int

	n += jail.PutEnv(buffer[n:], j)

	buffer[n] = '/'
	n++

	n += copy(buffer[n:], lang.SourceFile)

	return n
}

func PutProgrammingExecutable(buffer []byte, j jail.Jail, lang *ProgrammingLanguage) int {
	var n int

	n += jail.PutEnv(buffer[n:], j)

	buffer[n] = '/'
	n++

	n += copy(buffer[n:], lang.Executable)

	return n
}

func SubmissionVerifyProgrammingCreateSource(j jail.Jail, lang *ProgrammingLanguage, solution string) error {
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

/* TODO(anton2920): rewrite without using standard library. */
func SubmissionVerifyProgrammingCompile(j jail.Jail, lang *ProgrammingLanguage) error {
	var buffer bytes.Buffer

	const timeout = 5
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, lang.Compiler, append(lang.CompilerArgs, lang.SourceFile)...)
	cmd.Dir = "/tmp"
	cmd.SysProcAttr = &sys.SysProcAttr{Setsid: true, Jail: int(j.ID)}
	cmd.Cancel = func() error {
		return syscall.Kill(int32(-cmd.Process.Pid), syscall.SIGKILL)
	}
	cmd.Stdout = &buffer
	cmd.Stderr = &buffer

	if err := cmd.Run(); err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("failed to compile program: exceeded compilation timeout of %d seconds", timeout)
		}
		return fmt.Errorf("failed to compile program: %s %w", buffer.String(), err)
	}

	return nil
}

func SubmissionVerifyProgrammingRun(j jail.Jail, lang *ProgrammingLanguage, input string, output *bytes.Buffer) error {
	var exe string
	var args []string
	if lang.Executable != "" {
		exe = lang.Executable
	} else {
		exe = lang.Runner
		args = append(lang.RunnerArgs, lang.SourceFile)
	}

	const timeout = 2
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, exe, args...)
	cmd.Dir = "/tmp"
	cmd.SysProcAttr = &sys.SysProcAttr{Setsid: true, Jail: int(j.ID)}
	cmd.Cancel = func() error {
		return syscall.Kill(int32(-cmd.Process.Pid), syscall.SIGKILL)
	}
	cmd.Stdout = output
	cmd.Stderr = output

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	if _, err := io.WriteString(stdin, input); err != nil {
		return fmt.Errorf("failed to write input string: %w", err)
	}
	stdin.Close()

	if err := cmd.Run(); err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("failed to run program: exceeded timeout of %d seconds", timeout)
		}
		return fmt.Errorf("failed to run program: %s %w", output.String(), err)
	}

	return nil
}

func SubmissionVerifyProgrammingCheck(j jail.Jail, submittedTask *SubmittedProgramming, checkType CheckType) {
	var output bytes.Buffer

	task := submittedTask.Task
	lang := submittedTask.Language

	scores := make([]int, len(task.Checks[checkType]))
	messages := make([]string, len(task.Checks[checkType]))
	for i := 0; i < len(task.Checks[checkType]); i++ {
		output.Reset()

		check := &task.Checks[checkType][i]
		input := strings.Replace(strings.TrimSpace(check.Input), "\r\n", "\n", -1)
		if err := SubmissionVerifyProgrammingRun(j, lang, input, &output); err != nil {
			messages[i] = err.Error()
			if checkType == CheckTypeExample {
				break
			}
			continue
		}

		expectedOutput := strings.Replace(strings.TrimSpace(check.Output), "\r\n", "\n", -1)
		actualOutput := strings.Replace(strings.TrimSpace(output.String()), "\r\n", "\n", -1)
		if actualOutput != expectedOutput {
			messages[i] = fmt.Sprintf("expected %q, got %q", expectedOutput, actualOutput)
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

func SubmissionVerifyProgramming(submittedTask *SubmittedProgramming, checkType CheckType) error {
	lang := submittedTask.Language

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
		if err := SubmissionVerifyProgrammingCompile(j, lang); err != nil {
			return err
		}
	}

	if err := jail.Protect(j); err != nil {
		return err
	}

	SubmissionVerifyProgrammingCheck(j, submittedTask, checkType)
	return nil
}

func SubmissionVerifyStep(step interface{}) {
	switch step := step.(type) {
	case *SubmittedTest:
		SubmissionVerifyTest(step)
	case *SubmittedProgramming:
		step.Error = SubmissionVerifyProgramming(step, CheckTypeTest)
	}
}

func SubmissionVerify(submission *Submission) {
	for i := 0; i < len(submission.SubmittedSteps); i++ {
		SubmissionVerifyStep(submission.SubmittedSteps[i])
	}
}

func SubmissionVerifyWorker() {
	for submission := range SubmissionVerifyChannel {
		SubmissionVerify(submission)
	}
}
