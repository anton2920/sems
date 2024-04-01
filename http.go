package main

import (
	"runtime"
	"unsafe"
)

type HTTPRequest struct {
	Method  string
	URL     URL
	Version string
}

type HTTPState int

const (
	HTTP_STATE_UNKNOWN HTTPState = iota

	HTTP_STATE_METHOD
	HTTP_STATE_URI
	HTTP_STATE_VERSION
	HTTP_STATE_HEADER

	HTTP_STATE_DONE
)

type HTTPRequestParser struct {
	State   HTTPState
	Request HTTPRequest
}

type HTTPResponse struct {
	/* Holds pointers to response. Data must live long enough. */
	Iovs *[]Iovec

	/* ContentLength is calculated with each WriteBodyNoCopy call. */
	ContentLength int

	/* ContentLengthBuf points to stack-allocated buffer enough to hold 'Content-Length' header. */
	ContentLengthBuf []byte

	/* DateBuf points to array with current date in RFC822 format, which updates every second by kevent timer. */
	DateBuf []byte
}

type HTTPContext struct {
	RequestBuffer CircularBuffer
	ResponseIovs  []Iovec
	ResponsePos   int

	Parser HTTPRequestParser

	/* Check must be the same as pointer's bit, if context is in use. */
	Check uintptr
}

type HTTPRouter func(w *HTTPResponse, r *HTTPRequest)

const (
	HTTPStatusOK                    = 200
	HTTPStatusBadRequest            = 400
	HTTPStatusNotFound              = 404
	HTTPStatusMethodNotAllowed      = 405
	HTTPStatusRequestTimeout        = 408
	HTTPStatusRequestEntityTooLarge = 413
)

const (
	HTTPResponseBadRequest            = "HTTP/1.1 400 Bad HTTPRequest\r\nContent-Type: text/html\r\nContent-Length: 175\r\nConnection: close\r\n\r\n<!DOCTYPE html><head><title>400 Bad HTTPRequest</title></head><body><h1>400 Bad HTTPRequest</h1><p>Your browser sent a request that this server could not understand.</p></body></html>"
	HTTPResponseNotFound              = "HTTP/1.1 404 Not Found\r\nContent-Type: text/html\r\nContent-Length: 152\r\nConnection: close\r\n\r\n<!DOCTYPE html><head><title>404 Not Found</title></head><body><h1>404 Not Found</h1><p>The requested URL was not found on this server.</p></body></html>"
	HTTPResponseMethodNotAllowed      = "HTTP/1.1 405 Method Not Allowed\r\nContent-Type: text/html\r\nContent-Length: ...\r\nConnection: close\r\n\r\n"
	HTTPResponseRequestTimeout        = "HTTP/1.1 408 HTTPRequest Timeout\r\nContent-Type: text/html\r\nContent-Length: ...\r\nConnection: close\r\n\r\n"
	HTTPResponseRequestEntityTooLarge = "HTTP/1.1 413 HTTPRequest Entity Too Large\r\nContent-Type: text/html\r\nConent-Length: ...\r\nConnection: close\r\n\r\n"
)

func (w *HTTPResponse) WriteResponseNoCopyFunc(contentType string, f func(*HTTPResponse)) {
	*w.Iovs = append(*w.Iovs, IovecForString("HTTP/1.1 200 OK\r\nHost: rant\r\nDate: "), IovecForByteSlice(w.DateBuf), IovecForString("\r\nContent-Type: "), IovecForString(contentType), IovecForString("\r\nContent-Length: "), IovecForByteSlice(w.ContentLengthBuf), IovecForString("\r\n\r\n"))
	contentLengthIdx := len(*w.Iovs) - 2

	f(w)

	n := SlicePutInt(w.ContentLengthBuf, int(w.ContentLength))
	(*w.Iovs)[contentLengthIdx].Len = uint64(n)
}

func (w *HTTPResponse) WriteBodyNoCopy(body []byte) {
	*w.Iovs = append(*w.Iovs, IovecForByteSlice(body))
	w.ContentLength += len(body)
}

func (w *HTTPResponse) WriteResponseNoCopy(contentType string, body []byte) {
	n := SlicePutInt(w.ContentLengthBuf, len(body))

	*w.Iovs = append(*w.Iovs, IovecForString("HTTP/1.1 200 OK\r\nHost: rant\r\nDate: "), IovecForByteSlice(w.DateBuf), IovecForString("\r\nContent-Type: "), IovecForString(contentType), IovecForString("\r\nContent-Length: "), IovecForByteSlice(w.ContentLengthBuf[:n]), IovecForString("\r\n\r\n"), IovecForByteSlice(body))
}

func (w *HTTPResponse) WriteBuiltinError(code int) {
	switch code {
	case HTTPStatusBadRequest:
		*w.Iovs = append(*w.Iovs, IovecForString(HTTPResponseBadRequest))
	case HTTPStatusNotFound:
		*w.Iovs = append(*w.Iovs, IovecForString(HTTPResponseNotFound))
	}
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
	c.Parser.State = HTTP_STATE_METHOD

	if c.RequestBuffer, err = NewCircularBuffer(PageSize); err != nil {
		println("ERROR: failed to create request buffer:", err.Error())
		return nil
	}
	c.ResponseIovs = make([]Iovec, 0, 256)

	return unsafe.Pointer(c)
}

func GetContextAndCheck(ptr unsafe.Pointer) (*HTTPContext, uintptr) {
	uptr := uintptr(ptr)

	check := uptr & 0x1
	ctx := (*HTTPContext)(unsafe.Pointer(uptr - check))

	return ctx, check
}

func HTTPHandleRequests(wIovs *[]Iovec, rBuf *CircularBuffer, rp *HTTPRequestParser, contentLengthBuf, dateBuf []byte, router HTTPRouter) {
	var w HTTPResponse
	r := &rp.Request

	for {
		for rp.State != HTTP_STATE_DONE {
			switch rp.State {
			default:
				println(rp.State)
				panic("unknown HTTP parser state")
			case HTTP_STATE_UNKNOWN:
				unconsumed := rBuf.UnconsumedString()
				if len(unconsumed) < 2 {
					return
				}
				if unconsumed[:2] == "\r\n" {
					rBuf.Consume(len("\r\n"))
					rp.State = HTTP_STATE_DONE
				} else {
					rp.State = HTTP_STATE_HEADER
				}

			case HTTP_STATE_METHOD:
				unconsumed := rBuf.UnconsumedString()
				if len(unconsumed) < 3 {
					return
				}
				switch unconsumed[:3] {
				case "GET":
					r.Method = "GET"
				default:
					*wIovs = append(*wIovs, IovecForString(HTTPResponseMethodNotAllowed))
					return
				}
				rBuf.Consume(len(r.Method) + 1)
				rp.State = HTTP_STATE_URI
			case HTTP_STATE_URI:
				unconsumed := rBuf.UnconsumedString()
				lineEnd := FindChar(unconsumed, '\r')
				if lineEnd == -1 {
					return
				}

				uriEnd := FindChar(unconsumed[:lineEnd], ' ')
				if uriEnd == -1 {
					*wIovs = append(*wIovs, IovecForString(HTTPResponseBadRequest))
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
					*wIovs = append(*wIovs, IovecForString(HTTPResponseBadRequest))
					return
				}
				r.Version = httpVersion[len(httpVersionPrefix):]
				rBuf.Consume(len(r.URL.Path) + len(r.URL.Query) + 1 + len(httpVersionPrefix) + len(r.Version) + len("\r\n"))
				rp.State = HTTP_STATE_UNKNOWN
			case HTTP_STATE_HEADER:
				unconsumed := rBuf.UnconsumedString()
				lineEnd := FindChar(unconsumed, '\r')
				if lineEnd == -1 {
					return
				}
				header := unconsumed[:lineEnd]
				rBuf.Consume(len(header) + len("\r\n"))
				rp.State = HTTP_STATE_UNKNOWN
			}
		}

		w.Iovs = wIovs
		w.DateBuf = dateBuf
		w.ContentLength = 0
		w.ContentLengthBuf = contentLengthBuf
		router((*HTTPResponse)(Noescape(unsafe.Pointer(&w))), r)

		// println("Executed:", r.Method, r.URL.Path, r.URL.Query)

		rp.State = HTTP_STATE_METHOD
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
		FatalError("Failed to open kernel queue: ", err)
	}
	chlist := [...]Kevent_t{
		{Ident: uintptr(l), Filter: EVFILT_READ, Flags: EV_ADD | EV_CLEAR},
		{Ident: 1, Filter: EVFILT_TIMER, Flags: EV_ADD, Fflags: NOTE_SECONDS, Data: 1},
	}
	if _, err := Kevent(kq, unsafe.Slice(&chlist[0], len(chlist)), nil, nil); err != nil {
		FatalError("Failed to add event for listener socket: ", err)
	}

	if err := ClockGettime(CLOCK_REALTIME, &tp); err != nil {
		FatalError("Failed to get current walltime: ", err)
	}
	tp.Nsec = 0 /* NOTE(anton2920): we don't care about nanoseconds. */
	dateBuf := make([]byte, 31)

	contentLengthBuf := make([]byte, 10)

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
					Fatal("Failed to acquire new HTTP context")
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
				rBuf := &ctx.RequestBuffer
				parser := &ctx.Parser

				if rBuf.RemainingSpace() == 0 {
					Shutdown(c, SHUT_RD)
					Writev(c, []Iovec{IovecForString(HTTPResponseRequestEntityTooLarge)})
					goto closeConnection
				}

				n, err := Read(c, rBuf.RemainingSlice())
				if err != nil {
					println("ERROR: failed to read data from socket:", err.Error())
					goto closeConnection
				}
				rBuf.Produce(int(n))

				HTTPHandleRequests(&ctx.ResponseIovs, rBuf, parser, contentLengthBuf, dateBuf, router)

				wIovs := ctx.ResponseIovs
				n, err = Writev(c, wIovs[ctx.ResponsePos:])
				if err != nil {
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

	nworkers := runtime.GOMAXPROCS(0) / 2
	for i := 0; i < nworkers-1; i++ {
		go HTTPWorker(l, router)
	}
	HTTPWorker(l, router)

	return nil
}
