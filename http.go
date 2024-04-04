package main

import (
	"errors"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

type HTTPState int

const (
	HTTPStateUnknown HTTPState = iota

	HTTPStateMethod
	HTTPStateURI
	HTTPStateVersion
	HTTPStateHeader
	HTTPStateBody

	HTTPStateDone
)

type HTTPRequest struct {
	Arena *Arena

	Method  string
	URL     URL
	Version string

	Headers []string
	Body    []byte

	Form URLValues
}

type HTTPRequestParser struct {
	State   HTTPState
	Request HTTPRequest

	ContentLength int
}

type HTTPStatus int

const (
	HTTPStatusOK                    HTTPStatus = 200
	HTTPStatusSeeOther                         = 303
	HTTPStatusBadRequest                       = 400
	HTTPStatusUnauthorized                     = 401
	HTTPStatusNotFound                         = 404
	HTTPStatusMethodNotAllowed                 = 405
	HTTPStatusRequestTimeout                   = 408
	HTTPStatusConflict                         = 409
	HTTPStatusRequestEntityTooLarge            = 413
	HTTPStatusInternalServerError              = 500
)

var Status2String = [...]string{
	HTTPStatusOK:                    "200",
	HTTPStatusSeeOther:              "303",
	HTTPStatusBadRequest:            "400",
	HTTPStatusUnauthorized:          "401",
	HTTPStatusNotFound:              "404",
	HTTPStatusMethodNotAllowed:      "405",
	HTTPStatusRequestTimeout:        "408",
	HTTPStatusConflict:              "409",
	HTTPStatusRequestEntityTooLarge: "413",
	HTTPStatusInternalServerError:   "500",
}

var Status2Reason = [...]string{
	HTTPStatusOK:                    "OK",
	HTTPStatusSeeOther:              "See Other",
	HTTPStatusBadRequest:            "Bad Request",
	HTTPStatusUnauthorized:          "Unauthorized",
	HTTPStatusNotFound:              "Not Found",
	HTTPStatusMethodNotAllowed:      "Method Not Allowed",
	HTTPStatusRequestTimeout:        "Request Timeout",
	HTTPStatusConflict:              "Conflict",
	HTTPStatusRequestEntityTooLarge: "Request Entity Too Large",
	HTTPStatusInternalServerError:   "Internal Server Error",
}

type HTTPResponse struct {
	Arena *Arena

	StatusCode HTTPStatus
	Headers    []Iovec
	Bodies     []Iovec
}

type HTTPContext struct {
	RequestBuffer CircularBuffer
	ResponseArena Arena
	ResponseIovs  []Iovec
	ResponsePos   int

	Parser HTTPRequestParser

	/* Check must be the same as pointer's bit, if context is in use. */
	Check uintptr
}

type HTTPRouter func(w *HTTPResponse, r *HTTPRequest)

func (r *HTTPRequest) Cookie(name string) string {
	for i := 0; i < len(r.Headers); i++ {
		header := r.Headers[i]
		if StringStartsWith(header, "Cookie: ") {
			cookie := header[len("Cookie: "):]
			if StringStartsWith(cookie, name) {
				cookie = cookie[len(name):]
				if cookie[0] != '=' {
					return ""
				}
				return cookie[1:]
			}

		}
	}

	return ""
}

func (r *HTTPRequest) ParseForm() error {
	var err error

	if len(r.Form) != 0 {
		return nil
	}

	query := unsafe.String(unsafe.SliceData(r.Body), len(r.Body))

forQuery:
	for query != "" {
		var key string
		key, query, _ = strings.Cut(query, "&")
		if strings.Contains(key, ";") {
			err = errors.New("invalid semicolon separator in query")
			continue
		}
		if key == "" {
			continue
		}
		key, value, _ := strings.Cut(key, "=")

		keyBuffer := r.Arena.NewSlice(len(key))
		n, ok := URLDecode(keyBuffer, key)
		if !ok {
			if err == nil {
				err = errors.New("invalid key")
			}
			continue
		}
		key = unsafe.String(unsafe.SliceData(keyBuffer), n)

		valueBuffer := r.Arena.NewSlice(len(value))
		n, ok = URLDecode(valueBuffer, value)
		if !ok {
			if err == nil {
				err = errors.New("invalid value")
			}
			continue
		}
		value = unsafe.String(unsafe.SliceData(valueBuffer), n)

		for i := 0; i < len(r.Form); i++ {
			if key == r.Form[i].Key {
				r.Form[i].Values = append(r.Form[i].Values, value)
				continue forQuery
			}
		}
		if len(r.Form) < cap(r.Form) {
			l := len(r.Form)
			r.Form = r.Form[:l+1]
			r.Form[l].Key = key
			r.Form[l].Values = r.Form[l].Values[:1]
			r.Form[l].Values[0] = value
		} else {
			r.Form = append(r.Form, URLValue{Key: key, Values: []string{value}})
		}
	}

	return err
}

func (w *HTTPResponse) Append(b []byte) {
	w.Bodies = append(w.Bodies, IovecForByteSlice(b))
}

func (w *HTTPResponse) AppendString(s string) {
	w.Bodies = append(w.Bodies, IovecForString(s))
}

func (w *HTTPResponse) DelCookie(name string) {
	/* TODO(anton2920): replace with minimum required size. */
	cookie := w.Arena.NewSlice(128)

	var n int
	n += copy(cookie[n:], name)
	n += copy(cookie[n:], "=; Path=/; Max-Age=0; HttpOnly; Secure; SameSite=Strict")

	w.SetHeader("Set-Cookie", unsafe.String(unsafe.SliceData(cookie), n))
}

func (w *HTTPResponse) SetCookie(name, value string, expiry time.Time) {
	/* TODO(anton2920): replace with minimum required size. */
	cookie := w.Arena.NewSlice(128)

	var n int
	n += copy(cookie[n:], name)
	n += copy(cookie[n:], "=")
	n += copy(cookie[n:], value)
	n += copy(cookie[n:], "; Path=/; Expires=")
	n += SlicePutTmRFC822(cookie[n:], TimeToTm(int(expiry.Unix())))
	n += copy(cookie[n:], "; HttpOnly; Secure; SameSite=Strict")

	w.SetHeader("Set-Cookie", unsafe.String(unsafe.SliceData(cookie), n))
}

func (w *HTTPResponse) SetHeader(header string, value string) {
	buffer := w.Arena.NewSlice(len(header) + len(": ") + len(value) + len("\r\n"))

	var n int
	n += copy(buffer[n:], header)
	n += copy(buffer[n:], ": ")
	n += copy(buffer[n:], value)
	n += copy(buffer[n:], "\r\n")

	w.Headers = append(w.Headers, IovecForByteSlice(buffer[:n]))
}

func (w *HTTPResponse) Redirect(path string, code HTTPStatus) {
	w.SetHeader("Location", path)
	w.Bodies = w.Bodies[:0]
	w.StatusCode = code
}

/* TODO(anton2920): allow large data to be written. */
func (w *HTTPResponse) Write(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}

	buffer := w.Arena.NewSlice(len(b))
	copy(buffer, b)
	w.Append(buffer)

	return len(b), nil
}

// WriteHTML writes to w the escaped HTML equivalent of the plain text data b.
func (w *HTTPResponse) WriteHTML(b []byte) {
	last := 0
	for i, c := range b {
		var html string
		switch c {
		case '\000':
			html = HTMLNull
		case '"':
			html = HTMLQuot
		case '\'':
			html = HTMLApos
		case '&':
			html = HTMLAmp
		case '<':
			html = HTMLLt
		case '>':
			html = HTMLGt
		default:
			continue
		}
		w.Write(b[last:i])
		w.AppendString(html)
		last = i + 1
	}
	w.Write(b[last:])
}

func (w *HTTPResponse) WriteString(s string) (int, error) {
	return w.Write(unsafe.Slice(unsafe.StringData(s), len(s)))
}

func (w *HTTPResponse) WriteHTMLString(s string) {
	w.WriteHTML(unsafe.Slice(unsafe.StringData(s), len(s)))
}

/* NOTE(anton2920): Noescape hides a pointer from escape analysis. Noescape is the identity function but escape analysis doesn't think the output depends on the input. Noescape is inlined and currently compiles down to zero instructions. */
//go:nosplit
func Noescape(p unsafe.Pointer) unsafe.Pointer {
	x := uintptr(p)
	return unsafe.Pointer(x ^ 0)
}

func NewHTTPContext() unsafe.Pointer {
	var err error

	c := new(HTTPContext)
	c.Parser.State = HTTPStateMethod

	if c.RequestBuffer, err = NewCircularBuffer(PageSize); err != nil {
		println("ERROR: failed to create request buffer:", err.Error())
		return nil
	}
	c.ResponseIovs = make([]Iovec, 0, 512)

	return unsafe.Pointer(c)
}

func GetContextAndCheck(ptr unsafe.Pointer) (*HTTPContext, uintptr) {
	uptr := uintptr(ptr)

	check := uptr & 0x1
	ctx := (*HTTPContext)(unsafe.Pointer(uptr - check))

	return ctx, check
}

func HTTPWriteError(wIovs *[]Iovec, arena *Arena, statusCode HTTPStatus) {
	body := Status2Reason[statusCode]

	lengthBuf := arena.NewSlice(20)
	n := SlicePutInt(lengthBuf, len(body))

	*wIovs = append(*wIovs, IovecForString("HTTP/1.1"), IovecForString(" "), IovecForString(Status2String[statusCode]), IovecForString(" "), IovecForString(Status2Reason[statusCode]), IovecForString("\r\n"))
	*wIovs = append(*wIovs, IovecForString("Content-Length: "), IovecForByteSlice(lengthBuf[:n]), IovecForString("\r\n"))
	*wIovs = append(*wIovs, IovecForString("Connection: close\r\n"))
	*wIovs = append(*wIovs, IovecForString("\r\n"))
	*wIovs = append(*wIovs, IovecForString(body))
}

func HTTPHandleRequests(wIovs *[]Iovec, rBuf *CircularBuffer, arena *Arena, rp *HTTPRequestParser, dateBuf []byte, router HTTPRouter) {
	var w HTTPResponse
	var err error

	r := &rp.Request

	w.Arena = arena
	w.Headers = make([]Iovec, 32)
	w.Bodies = make([]Iovec, 128)

	for {
		r.Headers = r.Headers[:0]
		r.Form = r.Form[:0]
		rp.ContentLength = 0

		for rp.State != HTTPStateDone {
			switch rp.State {
			default:
				println(rp.State)
				panic("unknown HTTP parser state")
			case HTTPStateUnknown:
				unconsumed := rBuf.UnconsumedString()
				if len(unconsumed) < 2 {
					return
				}
				if unconsumed[:2] == "\r\n" {
					rBuf.Consume(len("\r\n"))

					if rp.ContentLength != 0 {
						rp.State = HTTPStateBody
					} else {
						rp.State = HTTPStateDone
					}
				} else {
					rp.State = HTTPStateHeader
				}

			case HTTPStateMethod:
				unconsumed := rBuf.UnconsumedString()
				if len(unconsumed) < 4 {
					return
				}
				switch unconsumed[:4] {
				case "GET ":
					r.Method = "GET"
				case "POST":
					r.Method = "POST"
				default:
					HTTPWriteError(wIovs, arena, HTTPStatusMethodNotAllowed)
					return
				}
				rBuf.Consume(len(r.Method) + 1)
				rp.State = HTTPStateURI
			case HTTPStateURI:
				unconsumed := rBuf.UnconsumedString()
				lineEnd := FindChar(unconsumed, '\r')
				if lineEnd == -1 {
					return
				}

				uriEnd := FindChar(unconsumed[:lineEnd], ' ')
				if uriEnd == -1 {
					HTTPWriteError(wIovs, arena, HTTPStatusBadRequest)
					return
				}

				queryStart := FindChar(unconsumed[:lineEnd], '?')
				if queryStart != -1 {
					r.URL.Path = unconsumed[:queryStart]
					r.URL.Query = unconsumed[queryStart+1 : uriEnd]
				} else {
					r.URL.Path = unconsumed[:uriEnd]
					r.URL.Query = ""
				}

				const httpVersionPrefix = "HTTP/"
				httpVersion := unconsumed[uriEnd+1 : lineEnd]
				if httpVersion[:len(httpVersionPrefix)] != httpVersionPrefix {
					HTTPWriteError(wIovs, arena, HTTPStatusBadRequest)
					return
				}
				r.Version = httpVersion[len(httpVersionPrefix):]
				rBuf.Consume(len(r.URL.Path) + len(r.URL.Query) + 1 + len(httpVersionPrefix) + len(r.Version) + len("\r\n"))
				rp.State = HTTPStateUnknown
			case HTTPStateHeader:
				unconsumed := rBuf.UnconsumedString()
				lineEnd := FindChar(unconsumed, '\r')
				if lineEnd == -1 {
					return
				}
				header := unconsumed[:lineEnd]
				r.Headers = append(r.Headers, header)
				rBuf.Consume(len(header) + len("\r\n"))

				if StringStartsWith(header, "Content-Length: ") {
					header = header[len("Content-Length: "):]
					rp.ContentLength, err = strconv.Atoi(header)
					if err != nil {
						HTTPWriteError(wIovs, arena, HTTPStatusBadRequest)
						return
					}
				}

				rp.State = HTTPStateUnknown
			case HTTPStateBody:
				unconsumed := rBuf.UnconsumedSlice()
				if len(unconsumed) < rp.ContentLength {
					return
				}

				r.Body = unconsumed[:rp.ContentLength]
				rBuf.Consume(len(r.Body))
				rp.State = HTTPStateDone
			}
		}

		w.Headers = w.Headers[:0]
		w.Bodies = w.Bodies[:0]
		router((*HTTPResponse)(Noescape(unsafe.Pointer(&w))), r)
		if w.StatusCode == 0 {
			w.StatusCode = HTTPStatusOK
		}

		var length int
		for i := 0; i < len(w.Bodies); i++ {
			length += int(w.Bodies[i].Len)
		}

		lengthBuf := arena.NewSlice(20)
		n := SlicePutInt(lengthBuf, length)

		*wIovs = append(*wIovs, IovecForString("HTTP/1.1"), IovecForString(" "), IovecForString(Status2String[w.StatusCode]), IovecForString(" "), IovecForString(Status2Reason[w.StatusCode]), IovecForString("\r\n"))
		*wIovs = append(*wIovs, w.Headers...)
		*wIovs = append(*wIovs, IovecForString("Date: "), IovecForByteSlice(dateBuf), IovecForString("\r\n"))
		*wIovs = append(*wIovs, IovecForString("Content-Length: "), IovecForByteSlice(lengthBuf[:n]), IovecForString("\r\n"))
		*wIovs = append(*wIovs, IovecForString("\r\n"))
		*wIovs = append(*wIovs, w.Bodies...)

		rp.State = HTTPStateMethod
	}
}

func HTTPWorker(l int32, router HTTPRouter) {
	var pinner runtime.Pinner
	var events [256]Kevent_t
	var ctx *HTTPContext
	var check uintptr
	var tp Timespec

	kq, err := Kqueue()
	if err != nil {
		Fatalf("Failed to open kernel queue: %v", err)
	}
	chlist := [...]Kevent_t{
		{Ident: uintptr(l), Filter: EVFILT_READ, Flags: EV_ADD | EV_CLEAR},
		{Ident: 1, Filter: EVFILT_TIMER, Flags: EV_ADD, Fflags: NOTE_SECONDS, Data: 1},
	}
	if _, err := Kevent(kq, unsafe.Slice(&chlist[0], len(chlist)), nil, nil); err != nil {
		Fatalf("Failed to add event for listener socket: %v", err)
	}

	if err := ClockGettime(CLOCK_REALTIME, &tp); err != nil {
		Fatalf("Failed to get current walltime: %v", err)
	}
	tp.Nsec = 0 /* NOTE(anton2920): we don't care about nanoseconds. */
	dateBuf := make([]byte, 31)

	ctxPool := NewPool(NewHTTPContext)

	for {
		nevents, err := Kevent(kq, nil, unsafe.Slice(&events[0], len(events)), nil)
		if err != nil {
			code := err.(E).Code
			if code == EINTR {
				continue
			}
			println("ERROR: failed to get requested kernel events: ", code)
		}
		for i := 0; i < nevents; i++ {
			e := events[i]
			c := int32(e.Ident)

			// println("EVENT", e.Ident, e.Filter, e.Fflags&0xF, e.Data)

			switch c {
			case l:
				c, err := Accept(l, nil, nil)
				if err != nil {
					code := err.(E).Code
					if code == EAGAIN {
						continue
					}
					println("ERROR: failed to accept new connection:", err.Error())
					continue
				}

				ctx = (*HTTPContext)(ctxPool.Get())
				if ctx == nil {
					Fatalf("Failed to acquire new HTTP context")
				}
				pinner.Pin(ctx)

				udata := unsafe.Pointer(uintptr(unsafe.Pointer(ctx)) | ctx.Check)
				events := [...]Kevent_t{
					{Ident: uintptr(c), Filter: EVFILT_READ, Flags: EV_ADD | EV_CLEAR, Udata: udata},
					{Ident: uintptr(c), Filter: EVFILT_WRITE, Flags: EV_ADD | EV_CLEAR, Udata: udata},
				}
				if _, err := Kevent(kq, unsafe.Slice(&events[0], len(events)), nil, nil); err != nil {
					println("ERROR: failed to add new events to kqueue:", err.Error())
					goto closeConnection
				}
				continue
			case 1:
				tp.Sec += int64(e.Data)
				SlicePutTmRFC822(dateBuf, TimeToTm(int(tp.Sec)))
				continue
			}

			ctx, check = GetContextAndCheck(e.Udata)
			if check != ctx.Check {
				continue
			}

			if (e.Flags & EV_EOF) != 0 {
				goto closeConnection
			}

			switch e.Filter {
			case EVFILT_READ:
				arena := &ctx.ResponseArena
				rBuf := &ctx.RequestBuffer
				parser := &ctx.Parser

				if rBuf.RemainingSpace() == 0 {
					Shutdown(c, SHUT_RD)
					HTTPWriteError(&ctx.ResponseIovs, arena, HTTPStatusRequestEntityTooLarge)
				} else {
					n, err := Read(c, rBuf.RemainingSlice())
					if err != nil {
						println("ERROR: failed to read data from socket:", err.Error())
						goto closeConnection
					}
					rBuf.Produce(int(n))

					HTTPHandleRequests(&ctx.ResponseIovs, rBuf, arena, parser, dateBuf, router)
				}

				fallthrough
			case EVFILT_WRITE:
				wIovs := ctx.ResponseIovs
				if len(wIovs[ctx.ResponsePos:]) > 0 {
					n, err := Writev(c, wIovs[ctx.ResponsePos:])
					if err != nil {
						code := err.(E).Code
						if code == EAGAIN {
							continue
						}
						println("ERROR: failed to write data to socket:", err.Error())
						goto closeConnection
					}

					for (ctx.ResponsePos < len(wIovs)) && (n >= int64(wIovs[ctx.ResponsePos].Len)) {
						n -= int64(wIovs[ctx.ResponsePos].Len)
						ctx.ResponsePos++
					}
					if ctx.ResponsePos == len(wIovs) {
						ctx.ResponsePos = 0
						ctx.ResponseIovs = ctx.ResponseIovs[:0]
					} else {
						wIovs[ctx.ResponsePos].Base = unsafe.Add(wIovs[ctx.ResponsePos].Base, n)
						wIovs[ctx.ResponsePos].Len -= uint64(n)
					}
				}
			}
			continue

		closeConnection:
			ctx.Check = 1 - ctx.Check
			ctx.RequestBuffer.Reset()
			ctx.ResponsePos = 0
			ctx.ResponseIovs = ctx.ResponseIovs[:0]
			ctxPool.Put(unsafe.Pointer(ctx))

			Shutdown(c, SHUT_WR)
			Close(c)
			continue
		}
	}
}

func ListenAndServe(port uint16, router HTTPRouter) error {
	l, err := Socket(PF_INET, SOCK_STREAM, 0)
	if err != nil {
		return err
	}

	var enable int32 = 1
	if err := Setsockopt(l, SOL_SOCKET, SO_REUSEADDR, unsafe.Pointer(&enable), uint32(unsafe.Sizeof(enable))); err != nil {
		return err
	}

	if err := Fcntl(l, F_SETFL, O_NONBLOCK); err != nil {
		return err
	}

	addr := SockAddrIn{Family: AF_INET, Addr: INADDR_ANY, Port: SwapBytesInWord(port)}
	if err := Bind(l, &addr, uint32(unsafe.Sizeof(addr))); err != nil {
		return err
	}

	const backlog = 128
	if err := Listen(l, backlog); err != nil {
		return err
	}

	// nworkers := runtime.GOMAXPROCS(0) / 2
	const nworkers = 1
	for i := 0; i < nworkers-1; i++ {
		go HTTPWorker(l, router)
	}
	HTTPWorker(l, router)

	return nil
}
