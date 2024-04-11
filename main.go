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

var DebugMode = "off"

func HandlePageRequest(w *HTTPResponse, r *HTTPRequest, path string) error {
	switch {
	default:
		switch path {
		case "/":
			return IndexPageHandler(w, r)
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
		switch path[len("/subject"):] {
		case "/create":
			return SubjectCreatePageHandler(w, r)
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

	return NotFoundError
}

func HandleAPIRequest(w *HTTPResponse, r *HTTPRequest, path string) error {
	switch {
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

	return NotFoundError
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

	case path == "/error":
		return TryAgainLaterError
	case path == "/panic":
		panic("test panic")
	}
}

func Router(w *HTTPResponse, r *HTTPRequest) {
	level := LevelDebug
	start := time.Now()

	err := RouterFunc(w, r)
	if err != nil {
		var httpError HTTPError
		displayError := err
		if (errors.As(err, &httpError)) && (httpError.StatusCode != HTTPStatusInternalServerError) {
			w.StatusCode = httpError.StatusCode
			level = LevelWarn
		} else {
			w.StatusCode = HTTPStatusInternalServerError
			if DebugMode != "on" {
				displayError = TryAgainLaterError
			}
			if _, ok := err.(PanicError); ok {
				level = LevelPanic
			} else {
				level = LevelError
			}
		}

		ErrorPageHandler(w, r, displayError)
	}

	Logf(level, "%7s %s -> %d (%v), %v", r.Method, r.URL.Path, w.StatusCode, err, time.Since(start))
}

func main() {
	if DebugMode == "on" {
		SetLogLevel(LevelDebug)
	}

	if err := RestoreSessionsFromFile(SessionsFile); err != nil {
		Warnf("Failed to restore sessions from file: %v", err)
	}
	if err := RestoreDBFromFile(DBFile); err != nil {
		Warnf("Failed to restore DB from file: %v", err)
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

	Infof("Listening on 0.0.0.0:7072...")

	if err := ListenAndServe(7072, Router); err != nil {
		Fatalf("Failed to listen on port: %v", err)
	}
}
