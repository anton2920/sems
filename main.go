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
		case "/lesson":
			return CourseLessonPageHandler(w, r)
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
			level = LevelError
		} else {
			Panicf("Unsupported error type %T", err)
		}

		if Debug {
			message = err.Error()
		}
		ErrorPageHandler(w, message)
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

	l, err := TCPListen(7072)
	if err != nil {
		Fatalf("Failed to listen on port: %v", err)
	}
	Infof("Listening on 0.0.0.0:7072...")

	q, err := NewEventQueue()
	if err != nil {
		Fatalf("Failed to create event queue: %v", err)
	}

	q.AddSocket(l, EventRequestRead, EventTriggerEdge, nil)

	signal.Ignore(syscall.Signal(SIGINT), syscall.Signal(SIGTERM))
	q.AddSignal(SIGINT)
	q.AddSignal(SIGTERM)

	ctxPool := NewPool(NewHTTPContext, (*HTTPContext).Reset)

	var quit bool
	for !quit {
		event, err := q.GetEvent()
		if (event == nil) || (err != nil) {
			Errorf("Failed to get event: %v", err)
			continue
		}

		switch event := event.(type) {
		default:
			Panicf("Unhandled event %T", event)
		case ReadEvent:
			switch event.Handle {
			case l: /* ready to accept new connection. */
				var addr SockAddrIn
				var addrLen uint32

				c, err := Accept(l, &addr, &addrLen)
				if err != nil {
					Errorf("Failed to accept new connection: %v", err)
					continue
				}

				ctx, err := ctxPool.Get()
				if err != nil {
					Errorf("Failed to create new HTTP context: %v", err)
					Close(c)
					continue
				}

				q.AddSocket(c, EventRequestRead|EventRequestWrite, EventTriggerEdge, ctx.CheckedPointer())
			default: /* ready to serve new HTTP request. */
				ctx, check := HTTPContextFromCheckedPointer(event.UserData)
				if ctx.Check != check {
					continue
				}

				if event.EndOfFile {
					ctxPool.Put(ctx)
					Close(event.Handle)
					continue
				}

				if err := HTTPRead(event.Handle, ctx); err != nil {
					Errorf("Failed to read data from socket: %v", err)
					ctxPool.Put(ctx)
					Close(event.Handle)
					continue
				}

				HTTPProcessRequests(ctx, Router, true)

				if err := HTTPWrite(event.Handle, ctx); err != nil {
					Errorf("Failed to write HTTP response: %v", err)
					ctxPool.Put(ctx)
					Close(event.Handle)
				}
			}
		case WriteEvent:
			ctx, check := HTTPContextFromCheckedPointer(event.UserData)
			if ctx.Check != check {
				continue
			}

			if event.EndOfFile {
				ctxPool.Put(ctx)
				Close(event.Handle)
				continue
			}

			if err := HTTPWrite(event.Handle, ctx); err != nil {
				Errorf("Failed to write HTTP response: %v", err)
				ctxPool.Put(ctx)
				Close(event.Handle)
			}
		case SignalEvent:
			Infof("Received signal %d, exitting...", event.Signal)
			quit = true
		}
	}

	q.Close()
	Close(l)

	if err := StoreDBToFile(DBFile); err != nil {
		Warnf("Failed to store DB to file: %v", err)
	}
	if err := StoreSessionsToFile(SessionsFile); err != nil {
		Warnf("Failed to store sessions to file: %v", err)
	}
}
