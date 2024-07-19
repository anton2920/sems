package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"

	"github.com/anton2920/gofa/errors"
	"github.com/anton2920/gofa/event"
	"github.com/anton2920/gofa/intel"
	"github.com/anton2920/gofa/log"
	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/net/http/http1"
	"github.com/anton2920/gofa/net/tcp"
	"github.com/anton2920/gofa/strings"
	"github.com/anton2920/gofa/syscall"
	"github.com/anton2920/gofa/time"
)

const (
	APIPrefix = "/api"
	FSPrefix  = "/fs"

	PageSize = 4096
)

var (
	BuildMode string
	Debug     bool
)

var WorkingDirectory string

var DateBuffer = make([]byte, time.RFC822Len)

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
		case "s":
			return CoursesPageHandler(w, r)
		case "/create", "/edit":
			return CourseCreateEditPageHandler(w, r)
		}
	case strings.StartsWith(path, "/group"):
		switch path[len("/group"):] {
		default:
			return GroupPageHandler(w, r)
		case "s":
			return GroupsPageHandler(w, r)
		case "/create":
			return GroupCreatePageHandler(w, r, nil)
		case "/edit":
			return GroupEditPageHandler(w, r, nil)
		}
	case strings.StartsWith(path, "/lesson"):
		switch path[len("/lesson"):] {
		default:
			return LessonPageHandler(w, r)
		}
	case strings.StartsWith(path, "/step"):
		switch path[len("/step"):] {
		case "s":
			return StepsPageHandler(w, r)
		}
	case strings.StartsWith(path, "/subject"):
		switch path[len("/subject"):] {
		default:
			return SubjectPageHandler(w, r)
		case "s":
			return SubjectsPageHandler(w, r)
		case "/create":
			return SubjectCreatePageHandler(w, r, nil)
		case "/edit":
			return SubjectEditPageHandler(w, r, nil)
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
		case "s":
			return UsersPageHandler(w, r)
		case "/create":
			return UserCreatePageHandler(w, r, nil)
		case "/edit":
			return UserEditPageHandler(w, r, nil)
		case "/signin":
			return UserSigninPageHandler(w, r, nil)
		}
	}

	return http.NotFound(Ls(GL, "requested page does not exist"))
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

	return http.NotFound(Ls(GL, "requested API endpoint does not exist"))
}

/* TODO(anton2920): maybe switch to sendfile(2)? */
func HandleFSRequest(w *http.Response, r *http.Request, path string) error {
	switch path {
	case "/bootstrap.min.css":
		w.Headers.Set("Content-Type", "text/css")
		w.Headers.Set("Cache-Control", "max-age=604800")
		w.Append(BootstrapCSS)
		return nil
	case "/bootstrap.min.js":
		w.Headers.Set("Content-Type", "application/js")
		w.Headers.Set("Cache-Control", "max-age=604800")
		w.Append(BootstrapJS)
		return nil
	}

	return http.NotFound(Ls(GL, "requested file does not exist"))
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
	case strings.StartsWith(path, FSPrefix):
		return HandleFSRequest(w, r, path[len(FSPrefix):])

	case path == "/error":
		return http.ServerError(errors.New(Ls(GL, "test error")))
	case path == "/panic":
		panic(Ls(GL, "test panic"))
	}
}

func Router(ctx *http.Context, ws []http.Response, rs []http.Request) {
	for i := 0; i < len(rs); i++ {
		w := &ws[i]
		r := &rs[i]

		if r.URL.Path == "/plaintext" {
			w.AppendString("Hello, world!\n")
			continue
		}
		w.Headers.Set("Content-Type", `text/html; charset="UTF-8"`)

		level := log.LevelDebug

		start := intel.RDTSC()
		err := RouterFunc(w, r)
		if err != nil {
			ErrorPageHandler(w, r, GL, err)
			if (w.StatusCode >= http.StatusBadRequest) && (w.StatusCode < http.StatusInternalServerError) {
				level = log.LevelWarn
			} else {
				level = log.LevelError
			}
			http.CloseAfterWrite(ctx)
		}

		if r.Headers.Get("Connection") == "close" {
			w.Headers.Set("Connection", "close")
			http.CloseAfterWrite(ctx)
		}

		addr := ctx.ClientAddress
		if r.Headers.Has("X-Forwarded-For") {
			addr = r.Headers.Get("X-Forwarded-For")
		}
		end := intel.RDTSC()
		elapsed := end - start

		log.Logf(level, "[%21s] %7s %s -> %v (%v), %4dµs", addr, r.Method, r.URL.Path, w.StatusCode, err, elapsed.ToUsec())
	}
}

func ServerWorker(q *event.Queue) {
	events := make([]event.Event, 64)

	const batchSize = 32
	ws := make([]http.Response, batchSize)
	rs := make([]http.Request, batchSize)

	for {
		n, err := q.GetEvents(events)
		if err != nil {
			log.Errorf("Failed to get events from client queue: %v", err)
			continue
		}

		for i := 0; i < n; i++ {
			e := &events[i]
			if errno := e.Error(); errno != 0 {
				log.Errorf("Event for %v returned code %d (%s)", e.Identifier, errno, errno)
				continue
			}

			ctx, ok := http.GetContextFromPointer(e.UserData)
			if !ok {
				continue
			}
			if e.EndOfFile() {
				http.Close(ctx)
				continue
			}

			switch e.Type {
			case event.Read:
				var read int
				for read < e.Data {
					n, err := http.Read(ctx)
					if err != nil {
						if err == http.NoSpaceLeft {
							http1.FillError(ctx, err, DateBuffer)
							http.CloseAfterWrite(ctx)
							break
						}
						log.Errorf("Failed to read data from client: %v", err)
						http.Close(ctx)
						break
					}
					read += n

					for n > 0 {
						n, err = http1.ParseRequestsUnsafe(ctx, rs)
						if err != nil {
							http1.FillError(ctx, err, DateBuffer)
							http.CloseAfterWrite(ctx)
							break
						}
						Router(ctx, ws[:n], rs[:n])
						http1.FillResponses(ctx, ws[:n], DateBuffer)
					}
				}
				fallthrough
			case event.Write:
				_, err = http.Write(ctx)
				if err != nil {
					log.Errorf("Failed to write data to client: %v", err)
					http.Close(ctx)
					continue
				}
			}
		}
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
	case "Tracing":
		f, err := os.Create(fmt.Sprintf("masters-%d.trace", os.Getpid()))
		if err != nil {
			log.Fatalf("Failed to create a tracing file: %v", err)
		}
		defer f.Close()

		trace.Start(f)
		defer trace.Stop()
	}
	log.Infof("Starting SEMS in %s mode...", BuildMode)

	WorkingDirectory, err = os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}

	if err := LoadAssets(); err != nil {
		log.Fatalf("Failed to load assets: %v", err)
	}

	if err = OpenDBs("db"); err != nil {
		log.Fatalf("Failed to open DBs: %v", err)
	}
	defer CloseDBs()

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
		log.Fatalf("Failed to create listener event queue: %v", err)
	}
	defer q.Close()

	_ = q.AddSocket(l, event.RequestRead, event.TriggerEdge, nil)
	_ = q.AddTimer(1, 1, event.Seconds, nil)

	_ = syscall.IgnoreSignals(syscall.SIGINT, syscall.SIGTERM)
	_ = q.AddSignals(syscall.SIGINT, syscall.SIGTERM)

	nworkers := runtime.GOMAXPROCS(0) / 2
	qs := make([]*event.Queue, nworkers)
	for i := 0; i < nworkers; i++ {
		qs[i], err = event.NewQueue()
		if err != nil {
			log.Fatalf("Failed to create new client queue: %v", err)
		}
		go ServerWorker(qs[i])
	}

	events := make([]event.Event, 64)
	now := time.Unix()
	var counter int

	var quit bool
	for !quit {
		n, err := q.GetEvents(events)
		if err != nil {
			log.Errorf("Failed to get events: %v", err)
			continue
		}

		for i := 0; i < n; i++ {
			e := &events[i]

			switch e.Type {
			default:
				log.Panicf("Unhandled event: %#v", e)
			case event.Read:
				ctx, err := http.Accept(l, PageSize)
				if err != nil {
					log.Errorf("Failed to accept new HTTP connection: %v", err)
					continue
				}
				_ = http.AddClientToQueue(qs[counter%len(qs)], ctx, event.RequestRead, event.TriggerEdge)
				counter++
			case event.Timer:
				now += e.Data
				time.PutTmRFC822(DateBuffer, time.ToTm(now))
			case event.Signal:
				log.Infof("Received signal %d, exitting...", e.Identifier)
				quit = true
				break
			}
		}
	}

	if err := StoreSessionsToFile(SessionsFile); err != nil {
		log.Warnf("Failed to store sessions to file: %v", err)
	}
}
