package main

import (
	"errors"
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
	case StringStartsWith(path, "/user"):
		switch path[len("/user"):] {
		case "/signin":
			return SigninPageHandler(w, r)
		}
	}

	return NotFoundError
}

func HandleAPIRequest(w *HTTPResponse, r *HTTPRequest, path string) error {
	switch {
	case StringStartsWith(path, "/user"):
		switch path[len("/user"):] {
		case "/signin":
			return SigninHandler(w, r)
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
	statusCode := HTTPStatusOK
	level := LevelDebug
	start := time.Now()

	err := RouterFunc(w, r)
	if err != nil {
		var httpError HTTPError
		displayError := err
		if (errors.As(err, &httpError)) && (httpError.StatusCode != HTTPStatusInternalServerError) {
			statusCode = httpError.StatusCode
			level = LevelWarn
		} else {
			statusCode = HTTPStatusInternalServerError
			if DebugMode != "on" {
				displayError = TryAgainLaterError
			}
			if _, ok := err.(PanicError); ok {
				level = LevelPanic
			} else {
				level = LevelError
			}
		}

		ErrorPageHandler(w, r, statusCode, displayError)
	}

	Logf(level, "%7s %s -> %d (%v), %v", r.Method, r.URL.Path, statusCode, err, time.Since(start))
}

func main() {
	if DebugMode == "on" {
		SetLogLevel(LevelDebug)
	}

	if err := ListenAndServe(7072, Router); err != nil {
		Fatalf("Failed to listen on port: %v", err)
	}
}
