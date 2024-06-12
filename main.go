package main

import (
	"fmt"
	"os"
	"runtime/pprof"
	"time"

	"github.com/anton2920/gofa/errors"
	"github.com/anton2920/gofa/event"
	"github.com/anton2920/gofa/log"
	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/net/tcp"
	"github.com/anton2920/gofa/strings"
	"github.com/anton2920/gofa/syscall"
)

const (
	APIPrefix = "/api"

	PageSize = 4096
)

var (
	BuildMode string
	Debug     bool
)

var WorkingDirectory string

func HandlePageRequest(w *http.Response, r *http.Request, path string) error {
	switch {
	default:
		switch path {
		case "/":
			return IndexPageHandler(w, r)
		}
	case strings.StartsWith(path, "/course"):
		switch path[len("/course"):] {
		default:
			return CoursePageHandler(w, r)
		case "/create", "/edit":
			return CourseCreateEditPageHandler(w, r)
		}
	case strings.StartsWith(path, "/group"):
		switch path[len("/group"):] {
		default:
			return GroupPageHandler(w, r)
		case "/create":
			return GroupCreatePageHandler(w, r)
		case "/edit":
			return GroupEditPageHandler(w, r)
		}
	case strings.StartsWith(path, "/lesson"):
		switch path[len("/lesson"):] {
		default:
			return LessonPageHandler(w, r)
		}
	case strings.StartsWith(path, "/subject"):
		switch path[len("/subject"):] {
		default:
			return SubjectPageHandler(w, r)
		case "/create":
			return SubjectCreatePageHandler(w, r)
		case "/edit":
			return SubjectEditPageHandler(w, r)
		case "/lessons":
			return SubjectLessonsPageHandler(w, r)
		}
	case strings.StartsWith(path, "/submission"):
		switch path[len("/submission"):] {
		default:
			return SubmissionPageHandler(w, r)
		case "/new":
			return SubmissionNewPageHandler(w, r)
		case "/results":
			return SubmissionResultsPageHandler(w, r)
		}
	case strings.StartsWith(path, "/user"):
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

	return http.NotFound("requested page does not exist")
}

func HandleAPIRequest(w *http.Response, r *http.Request, path string) error {
	switch {
	case strings.StartsWith(path, "/course"):
		switch path[len("/course"):] {
		case "/delete":
			return CourseDeleteHandler(w, r)
		}
	case strings.StartsWith(path, "/group"):
		switch path[len("/group"):] {
		case "/create":
			return GroupCreateHandler(w, r)
		case "/delete":
			return GroupDeleteHandler(w, r)
		case "/edit":
			return GroupEditHandler(w, r)
		}
	case strings.StartsWith(path, "/subject"):
		switch path[len("/subject"):] {
		case "/create":
			return SubjectCreateHandler(w, r)
		case "/delete":
			return SubjectDeleteHandler(w, r)
		case "/edit":
			return SubjectEditHandler(w, r)
		}
	case strings.StartsWith(path, "/user"):
		switch path[len("/user"):] {
		case "/create":
			return UserCreateHandler(w, r)
		case "/delete":
			return UserDeleteHandler(w, r)
		case "/edit":
			return UserEditHandler(w, r)
		case "/signin":
			return UserSigninHandler(w, r)
		case "/signout":
			return UserSignoutHandler(w, r)
		}
	}

	return http.NotFound("requested API endpoint does not exist")
}

func RouterFunc(w *http.Response, r *http.Request) (err error) {
	defer func() {
		if p := recover(); p != nil {
			err = errors.NewPanic(p)
		}
	}()

	path := r.URL.Path
	switch {
	default:
		return HandlePageRequest(w, r, path)
	case strings.StartsWith(path, APIPrefix):
		return HandleAPIRequest(w, r, path[len(APIPrefix):])

	case path == "/error":
		return http.ServerError(errors.New("test error"))
	case path == "/panic":
		panic("test panic")
	}
}

func Router(ws []http.Response, rs []http.Request) {
	for i := 0; i < min(len(ws), len(rs)); i++ {
		w := &ws[i]
		r := &rs[i]

		if r.URL.Path == "/plaintext" {
			w.AppendString("Hello, world!\n")
			continue
		}

		level := log.LevelDebug
		start := time.Now()

		err := RouterFunc(w, r)
		if err != nil {
			var message string

			if httpError, ok := err.(http.Error); ok {
				w.StatusCode = httpError.StatusCode
				message = httpError.DisplayMessage
				if (w.StatusCode >= http.StatusBadRequest) && (w.StatusCode < http.StatusInternalServerError) {
					level = log.LevelWarn
				} else {
					level = log.LevelError
				}
			} else if _, ok := err.(errors.Panic); ok {
				w.StatusCode = http.ServerError(nil).StatusCode
				message = http.ServerError(nil).DisplayMessage
				level = log.LevelError
			} else {
				log.Panicf("Unsupported error type %T", err)
			}

			if Debug {
				message = err.Error()
			}
			ErrorPageHandler(w, message)
		}

		addr := r.RemoteAddr
		for i := 0; i < len(r.Headers); i++ {
			header := r.Headers[i]
			if strings.StartsWith(header, "X-Forwarded-For: ") {
				addr = header[len("X-Forwarded-For: "):]
				break
			}
		}
		log.Logf(level, "[%21s] %7s %s -> %v (%v), %v", addr, r.Method, r.URL.Path, w.StatusCode, err, time.Since(start))
	}
}

func main() {
	var err error

	switch BuildMode {
	default:
		log.Fatalf("Build mode %q is not recognized", BuildMode)
	case "Release":
	case "Debug":
		Debug = true
		log.SetLevel(log.LevelDebug)
	case "Profiling":
		f, err := os.Create(fmt.Sprintf("masters-%d-cpu.pprof", os.Getpid()))
		if err != nil {
			log.Fatalf("Failed to create a profiling file: %v", err)
		}
		defer f.Close()

		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	log.Infof("Starting SEMS in %s mode...", BuildMode)

	WorkingDirectory, err = os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}

	if err = OpenDBs("db"); err != nil {
		log.Fatalf("Failed to open DBs: %v", err)
	}
	defer CloseDBs()

	CreateInitialDBs()

	if err := RestoreSessionsFromFile(SessionsFile); err != nil {
		log.Warnf("Failed to restore sessions from file: %v", err)
	}

	go SubmissionVerifyWorker()

	const address = "0.0.0.0:7072"
	l, err := tcp.Listen(address, 128)
	if err != nil {
		log.Fatalf("Failed to listen on port: %v", err)
	}
	defer syscall.Close(l)

	log.Infof("Listening on %s...", address)

	q, err := event.NewQueue()
	if err != nil {
		log.Fatalf("Failed to create event queue: %v", err)
	}
	defer q.Close()

	_ = q.AddSocket(l, event.RequestRead, event.TriggerEdge, nil)

	_ = syscall.IgnoreSignals(syscall.SIGINT, syscall.SIGTERM)
	_ = q.AddSignals(syscall.SIGINT, syscall.SIGTERM)

	ws := make([]http.Response, 32)
	rs := make([]http.Request, 32)

	var quit bool
	for !quit {
		for q.HasEvents() {
			e, err := q.GetEvent()
			if err != nil {
				log.Errorf("Failed to get event: %v", err)
				continue
			}

			switch e.Type {
			default:
				log.Panicf("Unhandled event: %#v", e)
			case event.Read:
				switch e.Identifier {
				case l: /* ready to accept new connection. */
					ctx, err := http.Accept(l, PageSize)
					if err != nil {
						log.Errorf("Failed to accept new HTTP connection: %v", err)
						continue
					}

					/* TODO(anton2920): remove this. */
					ctx.DateRFC822 = []byte("Thu, 09 May 2024 16:30:39 +0300")

					_ = http.AddClientToQueue(q, ctx, event.RequestRead|event.RequestWrite, event.TriggerEdge)
				default: /* ready to serve new HTTP request. */
					ctx, ok := http.ContextFromEvent(e)
					if !ok {
						continue
					}

					if e.EndOfFile {
						http.Close(ctx)
						continue
					}

					n, err := http.ReadRequests(ctx, rs)
					if err != nil {
						log.Errorf("Failed to read HTTP requests: %v", err)
						http.Close(ctx)
						continue
					}

					Router(ws[:n], rs[:n])

					if _, err := http.WriteResponses(ctx, ws[:n]); err != nil {
						log.Errorf("Failed to write HTTP responses: %v", err)
						http.Close(ctx)
						continue
					}
				}
			case event.Write:
				ctx, ok := http.ContextFromEvent(e)
				if !ok {
					continue
				}

				if e.EndOfFile {
					http.Close(ctx)
					continue
				}

				if _, err := http.WriteResponses(ctx, nil); err != nil {
					http.Close(ctx)
					continue
				}
			case event.Signal:
				log.Infof("Received signal %d, exitting...", e.Identifier)
				quit = true
				break
			}
		}

		const FPS = 60
		q.Pause(FPS)
	}

	if err := StoreSessionsToFile(SessionsFile); err != nil {
		log.Warnf("Failed to store sessions to file: %v", err)
	}
}
