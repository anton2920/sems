package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

type HTTPState int

const (
	HTTPStateRequestLine HTTPState = iota
	HTTPStateHeader
	HTTPStateBody

	HTTPStateUnknown
	HTTPStateDone
)

type HTTPRequest struct {
	Arena Arena

	RemoteAddr string

	Method string
	URL    URL
	Proto  string

	Headers []string
	Body    []byte

	Form URLValues
}

type HTTPRequestParser struct {
	State HTTPState
	Pos   int

	ContentLength int
}

type HTTPStatus int

const (
	HTTPStatusOK                    HTTPStatus = 200
	HTTPStatusSeeOther                         = 303
	HTTPStatusBadRequest                       = 400
	HTTPStatusUnauthorized                     = 401
	HTTPStatusForbidden                        = 403
	HTTPStatusNotFound                         = 404
	HTTPStatusMethodNotAllowed                 = 405
	HTTPStatusRequestTimeout                   = 408
	HTTPStatusConflict                         = 409
	HTTPStatusRequestEntityTooLarge            = 413
	HTTPStatusInternalServerError              = 500
)

var HTTPStatus2String = [...]string{
	0:                               "200",
	HTTPStatusOK:                    "200",
	HTTPStatusSeeOther:              "303",
	HTTPStatusBadRequest:            "400",
	HTTPStatusUnauthorized:          "401",
	HTTPStatusForbidden:             "403",
	HTTPStatusNotFound:              "404",
	HTTPStatusMethodNotAllowed:      "405",
	HTTPStatusRequestTimeout:        "408",
	HTTPStatusConflict:              "409",
	HTTPStatusRequestEntityTooLarge: "413",
	HTTPStatusInternalServerError:   "500",
}

var HTTPStatus2Reason = [...]string{
	0:                               "OK",
	HTTPStatusOK:                    "OK",
	HTTPStatusSeeOther:              "See Other",
	HTTPStatusBadRequest:            "Bad Request",
	HTTPStatusUnauthorized:          "Unauthorized",
	HTTPStatusForbidden:             "Forbidden",
	HTTPStatusNotFound:              "Not Found",
	HTTPStatusMethodNotAllowed:      "Method Not Allowed",
	HTTPStatusRequestTimeout:        "Request Timeout",
	HTTPStatusConflict:              "Conflict",
	HTTPStatusRequestEntityTooLarge: "Request Entity Too Large",
	HTTPStatusInternalServerError:   "Internal Server Error",
}

type HTTPHeaders struct {
	Values []Iovec

	OmitDate          bool
	OmitServer        bool
	OmitContentType   bool
	OmitContentLength bool
}

type HTTPResponse struct {
	Arena Arena

	StatusCode HTTPStatus
	Headers    HTTPHeaders
	Bodies     []Iovec
}

type HTTPError struct {
	StatusCode     HTTPStatus
	DisplayMessage string
	LogError       error
}

type HTTPContext struct {
	/* Check must be the same as the last pointer's bit, if context is in use. */
	Check int32

	Connection    int32
	ClientAddress string

	RequestParser HTTPRequestParser
	RequestBuffer CircularBuffer

	ResponseIovs []Iovec
	ResponsePos  int

	/* DateRFC822 could be set by client to reduce unnecessary syscalls and date formatting. */
	DateRFC822 []byte
}

var (
	UnauthorizedError = HTTPError{StatusCode: HTTPStatusUnauthorized, DisplayMessage: "whoops... You have to sign in to see this page", LogError: NewError("whoops... You have to sign in to see this page")}
	ForbiddenError    = HTTPError{StatusCode: HTTPStatusForbidden, DisplayMessage: "whoops... Your permissions are insufficient", LogError: NewError("whoops... Your permissions are insufficient")}
)

func (s HTTPStatus) String() string {
	return HTTPStatus2String[s]
}

func (rp *HTTPRequestParser) Parse(request string, r *HTTPRequest) (int, error) {
	var err error

	rp.State = HTTPStateRequestLine
	rp.ContentLength = 0
	rp.Pos = 0

	for rp.State != HTTPStateDone {
		switch rp.State {
		default:
			Panicf("Unknown HTTP parser state %d", rp.State)
		case HTTPStateUnknown:
			if len(request[rp.Pos:]) < 2 {
				return 0, nil
			}
			if request[rp.Pos:rp.Pos+2] == "\r\n" {
				rp.Pos += len("\r\n")

				if rp.ContentLength != 0 {
					rp.State = HTTPStateBody
				} else {
					rp.State = HTTPStateDone
				}
			} else {
				rp.State = HTTPStateHeader
			}
		case HTTPStateRequestLine:
			lineEnd := FindChar(request[rp.Pos:], '\r')
			if lineEnd == -1 {
				return 0, nil
			}

			sp := FindChar(request[rp.Pos:rp.Pos+lineEnd], ' ')
			if sp == -1 {
				return 0, fmt.Errorf("expected method, found %q", request[rp.Pos:])
			}
			r.Method = request[rp.Pos : rp.Pos+sp]
			rp.Pos += len(r.Method) + 1
			lineEnd -= len(r.Method) + 1

			uriEnd := FindChar(request[rp.Pos:rp.Pos+lineEnd], ' ')
			if uriEnd == -1 {
				return 0, fmt.Errorf("expected space after URI, found %q", request[rp.Pos:lineEnd])
			}

			queryStart := FindChar(request[rp.Pos:rp.Pos+uriEnd], '?')
			if queryStart != -1 {
				r.URL.Path = request[rp.Pos : rp.Pos+queryStart]
				r.URL.Query = request[rp.Pos+queryStart+1 : rp.Pos+uriEnd]
			} else {
				r.URL.Path = request[rp.Pos : rp.Pos+uriEnd]
				r.URL.Query = ""
			}
			rp.Pos += len(r.URL.Path) + len(r.URL.Query) + 1
			lineEnd -= len(r.URL.Path) + len(r.URL.Query) + 1

			if request[rp.Pos:rp.Pos+len("HTTP/")] != "HTTP/" {
				return 0, fmt.Errorf("expected HTTP version prefix, found %q", request[rp.Pos:rp.Pos+lineEnd])
			}
			r.Proto = request[rp.Pos : rp.Pos+lineEnd]

			rp.Pos += len(r.Proto) + len("\r\n")
			rp.State = HTTPStateUnknown
		case HTTPStateHeader:
			lineEnd := FindChar(request[rp.Pos:], '\r')
			if lineEnd == -1 {
				return 0, nil
			}
			header := request[rp.Pos : rp.Pos+lineEnd]
			r.Headers = append(r.Headers, header)

			if StringStartsWith(header, "Content-Length: ") {
				header = header[len("Content-Length: "):]
				rp.ContentLength, err = strconv.Atoi(header)
				if err != nil {
					return 0, fmt.Errorf("failed to parse Content-Length value: %w", err)
				}
			}

			rp.Pos += len(header) + len("\r\n")
			rp.State = HTTPStateUnknown
		case HTTPStateBody:
			if len(request[rp.Pos:]) < rp.ContentLength {
				return 0, nil
			}

			r.Body = unsafe.Slice(unsafe.StringData(request[rp.Pos:]), rp.ContentLength)
			rp.Pos += len(r.Body)
			rp.State = HTTPStateDone
		}
	}

	return rp.Pos, nil
}

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
	for query != "" {
		var key string
		key, query, _ = strings.Cut(query, "&")
		if strings.Contains(key, ";") {
			err = NewError("invalid semicolon separator in query")
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
				err = NewError("invalid key")
			}
			continue
		}
		key = unsafe.String(unsafe.SliceData(keyBuffer), n)

		valueBuffer := r.Arena.NewSlice(len(value))
		n, ok = URLDecode(valueBuffer, value)
		if !ok {
			if err == nil {
				err = NewError("invalid value")
			}
			continue
		}
		value = unsafe.String(unsafe.SliceData(valueBuffer), n)

		r.Form.Add(key, value)
	}

	return err
}

func (r *HTTPRequest) Reset() {
	r.Headers = r.Headers[:0]
	r.Body = r.Body[:0]
	r.Form = r.Form[:0]
	r.Arena.Reset()
}

func (w *HTTPResponse) Append(b []byte) {
	w.Bodies = append(w.Bodies, IovecForByteSlice(b))
}

func (w *HTTPResponse) AppendString(s string) {
	w.Bodies = append(w.Bodies, IovecForString(s))
}

func (w *HTTPResponse) DelCookie(name string) {
	const finisher = "=; Path=/; Max-Age=0; HttpOnly; Secure; SameSite=Strict"

	cookie := w.Arena.NewSlice(len(name) + len(finisher))

	var n int
	n += copy(cookie[n:], name)
	n += copy(cookie[n:], finisher)

	w.SetHeaderUnsafe("Set-Cookie", unsafe.String(unsafe.SliceData(cookie), n))
}

func (w *HTTPResponse) SetCookie(name, value string, expiry time.Time) {
	const secure = "; HttpOnly; Secure; SameSite=Strict"
	const expires = "; Expires="
	const path = "; Path=/"
	const eq = "="

	cookie := w.Arena.NewSlice(len(name) + len(eq) + len(value) + len(path) + len(expires) + RFC822Len + len(secure))

	var n int
	n += copy(cookie[n:], name)
	n += copy(cookie[n:], eq)
	n += copy(cookie[n:], value)
	n += copy(cookie[n:], path)
	n += copy(cookie[n:], expires)
	n += SlicePutTmRFC822(cookie[n:], TimeToTm(int(expiry.Unix())))
	n += copy(cookie[n:], secure)

	w.SetHeaderUnsafe("Set-Cookie", unsafe.String(unsafe.SliceData(cookie), n))
}

/* SetCookieUnsafe is useful for debugging purposes. It's also more compatible with older browsers. */
func (w *HTTPResponse) SetCookieUnsafe(name, value string, expiry time.Time) {
	const expires = "; Expires="
	const path = "; Path=/"
	const eq = "="

	cookie := w.Arena.NewSlice(len(name) + len(eq) + len(value) + len(path) + len(expires) + RFC822Len)

	var n int
	n += copy(cookie[n:], name)
	n += copy(cookie[n:], eq)
	n += copy(cookie[n:], value)
	n += copy(cookie[n:], path)
	n += copy(cookie[n:], expires)
	n += SlicePutTmRFC822(cookie[n:], TimeToTm(int(expiry.Unix())))

	w.SetHeaderUnsafe("Set-Cookie", unsafe.String(unsafe.SliceData(cookie), n))
}

/* SetHeaderUnsafe sets new 'value' for 'header' relying on that memory lives long enough. */
func (w *HTTPResponse) SetHeaderUnsafe(header string, value string) {
	switch header {
	case "Date":
		w.Headers.OmitDate = true
	case "Server":
		w.Headers.OmitServer = true
	case "ContentType":
		w.Headers.OmitContentType = true
	case "ContentLength":
		w.Headers.OmitContentLength = true
	}

	for i := 0; i < len(w.Headers.Values); i += 4 {
		key := w.Headers.Values[i]
		if header == unsafe.String((*byte)(key.Base), key.Len) {
			w.Headers.Values[i+2] = IovecForString(value)
			return
		}
	}

	w.Headers.Values = append(w.Headers.Values, IovecForString(header), IovecForString(": "), IovecForString(value), IovecForString("\r\n"))
}

func (w *HTTPResponse) Redirect(path string, code HTTPStatus) {
	pathBuf := w.Arena.NewSlice(len(path))
	copy(pathBuf, path)

	w.SetHeaderUnsafe("Location", unsafe.String(unsafe.SliceData(pathBuf), len(pathBuf)))
	w.Bodies = w.Bodies[:0]
	w.StatusCode = code
}

func (w *HTTPResponse) RedirectID(prefix string, id int, code HTTPStatus) {
	buffer := w.Arena.NewSlice(len(prefix) + 20)
	n := copy(buffer, prefix)
	n += SlicePutInt(buffer[n:], id)

	w.SetHeaderUnsafe("Location", unsafe.String(unsafe.SliceData(buffer), n))
	w.Bodies = w.Bodies[:0]
	w.StatusCode = code
}

func (w *HTTPResponse) Reset() {
	w.StatusCode = HTTPStatusOK
	w.Headers.Values = w.Headers.Values[:0]
	w.Headers.OmitDate = false
	w.Headers.OmitServer = false
	w.Headers.OmitContentType = false
	w.Headers.OmitContentLength = false
	w.Bodies = w.Bodies[:0]
	w.Arena.Reset()
}

func (w *HTTPResponse) Write(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}

	buffer := w.Arena.NewSlice(len(b))
	copy(buffer, b)
	w.Append(buffer)

	return len(b), nil
}

/* WriteHTML writes to w the escaped HTML equivalent of the plain text data b. */
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

func (w *HTTPResponse) WriteInt(i int) (int, error) {
	buffer := w.Arena.NewSlice(20)
	n := SlicePutInt(buffer, i)
	w.Append(buffer[:n])
	return n, nil
}

func (w *HTTPResponse) WriteString(s string) (int, error) {
	return w.Write(unsafe.Slice(unsafe.StringData(s), len(s)))
}

func (w *HTTPResponse) WriteHTMLString(s string) {
	w.WriteHTML(unsafe.Slice(unsafe.StringData(s), len(s)))
}

func BadRequest(format string, args ...interface{}) HTTPError {
	message := fmt.Sprintf(format, args...)
	return HTTPError{StatusCode: HTTPStatusBadRequest, DisplayMessage: message, LogError: WrapErrorWithTrace(NewError(message), 2)}
}

func NotFound(format string, args ...interface{}) HTTPError {
	message := fmt.Sprintf(format, args...)
	return HTTPError{StatusCode: HTTPStatusNotFound, DisplayMessage: message, LogError: WrapErrorWithTrace(NewError(message), 2)}
}

func Conflict(format string, args ...interface{}) HTTPError {
	message := fmt.Sprintf(format, args...)
	return HTTPError{StatusCode: HTTPStatusConflict, DisplayMessage: message, LogError: WrapErrorWithTrace(NewError(message), 2)}
}

func ClientError(err error) HTTPError {
	return HTTPError{StatusCode: HTTPStatusBadRequest, DisplayMessage: "whoops... Something went wrong. Please reload this page or try again later", LogError: WrapErrorWithTrace(err, 2)}
}

func ServerError(err error) HTTPError {
	return HTTPError{StatusCode: HTTPStatusInternalServerError, DisplayMessage: "whoops... Something went wrong. Please try again later", LogError: WrapErrorWithTrace(err, 2)}
}

func (e HTTPError) Error() string {
	if e.LogError == nil {
		return "<nil>"
	}
	return e.LogError.Error()
}

func SlicePutAddress(buffer []byte, addr uint32, port uint16) int {
	var n int

	n += SlicePutInt(buffer[n:], int((addr&0x000000FF)>>0))
	buffer[n] = ':'
	n++

	n += SlicePutInt(buffer[n:], int((addr&0x0000FF00)>>8))
	buffer[n] = '.'
	n++

	n += SlicePutInt(buffer[n:], int((addr&0x00FF0000)>>16))
	buffer[n] = '.'
	n++

	n += SlicePutInt(buffer[n:], int((addr&0xFF000000)>>24))
	buffer[n] = '.'
	n++

	n += SlicePutInt(buffer[n:], int(SwapBytesInWord(port)))

	return n
}

func NewHTTPContext(c int32, addr SockAddrIn) (*HTTPContext, error) {
	rb, err := NewCircularBuffer(PageSize)
	if err != nil {
		return nil, err
	}

	ctx := new(HTTPContext)
	ctx.Connection = c
	ctx.RequestBuffer = rb
	ctx.ResponseIovs = make([]Iovec, 0, 1024)

	buffer := make([]byte, 21)
	n := SlicePutAddress(buffer, addr.Addr, addr.Port)
	ctx.ClientAddress = string(buffer[:n])

	return ctx, nil
}

func HTTPContextFromEvent(event Event) (*HTTPContext, bool) {
	if event.UserData == nil {
		return nil, false
	}
	uptr := uintptr(event.UserData)

	check := uptr & 0x1
	ctx := (*HTTPContext)(unsafe.Pointer(uptr - check))

	return ctx, ctx.Check == int32(check)
}

func (ctx *HTTPContext) Reset() {
	ctx.Check = 1 - ctx.Check
	ctx.RequestBuffer.Reset()
	ctx.ResponsePos = 0
	ctx.ResponseIovs = ctx.ResponseIovs[:0]
}

func FreeHTTPContext(ctx *HTTPContext) {
	ctx.Reset()
	FreeCircularBuffer(&ctx.RequestBuffer)
}

func HTTPAccept(l int32) (*HTTPContext, error) {
	var addr SockAddrIn
	var addrLen uint32 = uint32(unsafe.Sizeof(addr))

	c, err := Accept(l, &addr, &addrLen)
	if err != nil {
		return nil, fmt.Errorf("failed to accept incoming connection: %w", err)
	}

	ctx, err := NewHTTPContext(c, addr)
	if err != nil {
		Close(c)
		return nil, fmt.Errorf("failed to create new HTTP context: %w", err)
	}

	return ctx, nil
}

/* HTTPReadRequests reads data from socket and parses HTTP requests. Returns the number of requests parsed. */
func HTTPReadRequests(ctx *HTTPContext, rs []HTTPRequest) (int, error) {
	rBuf := &ctx.RequestBuffer
	if rBuf.RemainingSpace() == 0 {
		return 0, NewError("no space left in the buffer")
	}
	n, err := Read(ctx.Connection, rBuf.RemainingSlice())
	if err != nil {
		return 0, err
	}
	rBuf.Produce(int(n))

	parser := &ctx.RequestParser

	var i int
	for i = 0; i < len(rs); i++ {
		r := &rs[i]
		r.RemoteAddr = ctx.ClientAddress
		r.Reset()

		n, err := parser.Parse(rBuf.UnconsumedString(), r)
		if err != nil {
			return i, err
		}
		if n == 0 {
			break
		}
		rBuf.Consume(n)
	}

	return i, nil
}

/* HTTPWriteResponses generates HTTP responses and writes them on wire. Returns the number of processed responses. */
func HTTPWriteResponses(ctx *HTTPContext, ws []HTTPResponse) (int, error) {
	dateBuf := ctx.DateRFC822
	if dateBuf == nil {
		dateBuf := make([]byte, 31)

		var tp Timespec
		ClockGettime(CLOCK_REALTIME, &tp)
		SlicePutTmRFC822(dateBuf, TimeToTm(int(tp.Sec)))
	}

	for i := 0; i < len(ws); i++ {
		w := &ws[i]

		ctx.ResponseIovs = append(ctx.ResponseIovs, IovecForString("HTTP/1.1"), IovecForString(" "), IovecForString(HTTPStatus2String[w.StatusCode]), IovecForString(" "), IovecForString(HTTPStatus2Reason[w.StatusCode]), IovecForString("\r\n"))

		if !w.Headers.OmitDate {
			ctx.ResponseIovs = append(ctx.ResponseIovs, IovecForString("Date: "), IovecForByteSlice(dateBuf), IovecForString("\r\n"))
		}

		if !w.Headers.OmitServer {
			ctx.ResponseIovs = append(ctx.ResponseIovs, IovecForString("Server: gofa/http\r\n"))
		}

		if !w.Headers.OmitContentType {
			if ContentTypeHTML(w.Bodies) {
				ctx.ResponseIovs = append(ctx.ResponseIovs, IovecForString("Content-Type: text/html; charset=\"UTF-8\"\r\n"))
			} else {
				ctx.ResponseIovs = append(ctx.ResponseIovs, IovecForString("Content-Type: text/plain; charset=\"UTF-8\"\r\n"))
			}
		}

		if !w.Headers.OmitContentLength {
			var length int
			for i := 0; i < len(w.Bodies); i++ {
				length += int(w.Bodies[i].Len)
			}

			lengthBuf := w.Arena.NewSlice(20)
			n := SlicePutInt(lengthBuf, length)

			ctx.ResponseIovs = append(ctx.ResponseIovs, IovecForString("Content-Length: "), IovecForByteSlice(lengthBuf[:n]), IovecForString("\r\n"))
		}

		ctx.ResponseIovs = append(ctx.ResponseIovs, w.Headers.Values...)
		ctx.ResponseIovs = append(ctx.ResponseIovs, IovecForString("\r\n"))
		ctx.ResponseIovs = append(ctx.ResponseIovs, w.Bodies...)

		w.Reset()
	}

	if len(ctx.ResponseIovs[ctx.ResponsePos:]) > 0 {
		/* TODO(anton2920): think about using CircularBuffer with 'sendfile(2)' to speed up things. */
		n, err := Writev(ctx.Connection, ctx.ResponseIovs[ctx.ResponsePos:])
		if err != nil {
			return 0, err
		}

		/* TODO(anton2920): if written less than max, do something with memory that is going to be reused outside. */
		for (ctx.ResponsePos < len(ctx.ResponseIovs)) && (n >= int64(ctx.ResponseIovs[ctx.ResponsePos].Len)) {
			n -= int64(ctx.ResponseIovs[ctx.ResponsePos].Len)
			ctx.ResponsePos++
		}
		if ctx.ResponsePos == len(ctx.ResponseIovs) {
			ctx.ResponsePos = 0
			ctx.ResponseIovs = ctx.ResponseIovs[:0]
		} else {
			ctx.ResponseIovs[ctx.ResponsePos].Base = unsafe.Add(ctx.ResponseIovs[ctx.ResponsePos].Base, n)
			ctx.ResponseIovs[ctx.ResponsePos].Len -= uint64(n)
		}
	}
	return len(ws), nil
}

func HTTPClose(ctx *HTTPContext) error {
	c := ctx.Connection
	FreeHTTPContext(ctx)
	return Close(c)
}
