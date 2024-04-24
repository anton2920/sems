package main

import (
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	APIPrefix = "/api"

	PageSize = 4096
)

var (
	DebugMode string
	Debug     bool
)

var WorkingDirectory string

func HandlePageRequest(w *HTTPResponse, r *HTTPRequest, path string) error {
	switch {
	default:
		switch path {
		case "/":
			return IndexPageHandler(w, r)
		}
	case StringStartsWith(path, "/course"):
		switch path[len("/course"):] {
		default:
			return CoursePageHandler(w, r)
		case "/create", "/edit":
			return CourseCreateEditPageHandler(w, r)
		}
	case StringStartsWith(path, "/group"):
		switch path[len("/group"):] {
		default:
			return GroupPageHandler(w, r)
		case "/create":
			return GroupCreatePageHandler(w, r)
		case "/edit":
			return GroupEditPageHandler(w, r)
		}
	case StringStartsWith(path, "/subject"):
		path = path[len("/subject"):]

		switch {
		default:
			switch path {
			default:
				return SubjectPageHandler(w, r)
			case "/create":
				return SubjectCreatePageHandler(w, r)
			case "/edit":
				return SubjectEditPageHandler(w, r)
			}
		case StringStartsWith(path, "/lesson"):
			switch path[len("/lesson"):] {
			default:
				return SubjectLessonPageHandler(w, r)
			case "/edit":
				return SubjectLessonEditPageHandler(w, r)
			}
		}
	case StringStartsWith(path, "/submission"):
		switch path[len("/submission"):] {
		default:
			return SubmissionPageHandler(w, r)
		case "/new":
			return SubmissionNewPageHandler(w, r)
		}
	case StringStartsWith(path, "/user"):
		switch path[len("/user"):] {
		default:
			return UserPageHandler(w, r)
		case "/create":
			return UserCreatePageHandler(w, r)
		case "/edit":
			return UserEditPageHandler(w, r)
		case "/signin":
			return UserSigninPageHandler(w, r)
		}
	}

	return NotFound("requested page does not exist")
}

func HandleAPIRequest(w *HTTPResponse, r *HTTPRequest, path string) error {
	switch {
	case StringStartsWith(path, "/course"):
		switch path[len("/course"):] {
		case "/delete":
			return CourseDeleteHandler(w, r)
		}
	case StringStartsWith(path, "/group"):
		switch path[len("/group"):] {
		case "/create":
			return GroupCreateHandler(w, r)
		case "/edit":
			return GroupEditHandler(w, r)
		}
	case StringStartsWith(path, "/subject"):
		switch path[len("/subject"):] {
		case "/create":
			return SubjectCreateHandler(w, r)
		case "/edit":
			return SubjectEditHandler(w, r)
		}
	case StringStartsWith(path, "/user"):
		switch path[len("/user"):] {
		case "/create":
			return UserCreateHandler(w, r)
		case "/edit":
			return UserEditHandler(w, r)
		case "/signin":
			return UserSigninHandler(w, r)
		case "/signout":
			return UserSignoutHandler(w, r)
		}
	}

	return NotFound("requested API endpoint does not exist")
}

func RouterFunc(w *HTTPResponse, r *HTTPRequest) (err error) {
	defer func() {
		if p := recover(); p != nil {
			err = NewPanicError(p)
		}
	}()

	path := r.URL.Path
	switch {
	default:
		return HandlePageRequest(w, r, path)
	case StringStartsWith(path, APIPrefix):
		return HandleAPIRequest(w, r, path[len(APIPrefix):])

	case path == "/plaintext":
		w.AppendString("Hello, world!\n")
		return nil
	case path == "/error":
		return ServerError(NewError("test error"))
	case path == "/panic":
		panic("test panic")
	}
}

func Router(w *HTTPResponse, r *HTTPRequest) {
	level := LevelDebug
	start := time.Now()

	err := RouterFunc(w, r)
	if err != nil {
		var panicError PanicError
		var httpError HTTPError
		var message string

		if errors.As(err, &httpError) {
			w.StatusCode = httpError.StatusCode
			message = httpError.DisplayMessage
			if (w.StatusCode >= HTTPStatusBadRequest) && (w.StatusCode < HTTPStatusInternalServerError) {
				level = LevelWarn
			} else {
				level = LevelError
			}
		} else if errors.As(err, &panicError) {
			w.StatusCode = ServerError(nil).StatusCode
			message = ServerError(nil).DisplayMessage
			level = LevelPanic
		} else {
			panic("unsupported error")
		}

		if Debug {
			message = err.Error()
		}
		ErrorPageHandler(w, r, message)
	}

	Logf(level, "%7s %s -> %d (%v), %v", r.Method, r.URL.Path, w.StatusCode, err, time.Since(start))
}

func main() {
	var err error

	if DebugMode == "on" {
		Debug = true
		SetLogLevel(LevelDebug)
	}

	WorkingDirectory, err = os.Getwd()
	if err != nil {
		Fatalf("Failed to get current working directory: %v", err)
	}

	if err := RestoreSessionsFromFile(SessionsFile); err != nil {
		Warnf("Failed to restore sessions from file: %v", err)
	}
	if err := RestoreDBFromFile(DBFile); err != nil {
		Warnf("Failed to restore DB from file: %v", err)
		CreateInitialDB()
	}

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		signal := <-sigchan
		Infof("Received %s, exitting...", signal)

		if err := StoreDBToFile(DBFile); err != nil {
			Warnf("Failed to store DB to file: %v", err)
		}
		if err := StoreSessionsToFile(SessionsFile); err != nil {
			Warnf("Failed to store sessions to file: %v", err)
		}

		os.Exit(0)
	}()

	go SubmissionVerifyWorker()

	Infof("Listening on 0.0.0.0:7072...")

	if err := ListenAndServe(7072, Router); err != nil {
		Fatalf("Failed to listen on port: %v", err)
	}
}
