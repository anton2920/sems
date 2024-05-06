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
	HTTPStateMethod HTTPState = iota
	HTTPStateURI
	HTTPStateVersion
	HTTPStateHeader
	HTTPStateBody

	HTTPStateUnknown
	HTTPStateDone
)

type HTTPRequest struct {
	Arena Arena

	Address string

	Method  string
	URL     URL
	Version string

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

var Status2String = [...]string{
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

var Status2Reason = [...]string{
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

type HTTPResponse struct {
	Arena Arena

	StatusCode HTTPStatus
	Headers    []Iovec
	Bodies     []Iovec
}

type HTTPError struct {
	StatusCode     HTTPStatus
	DisplayMessage string
	LogError       error
}

type HTTPContext struct {
	RequestParser HTTPRequestParser
	RequestBuffer CircularBuffer
	Request       HTTPRequest

	ResponseArena Arena
	ResponseIovs  []Iovec
	ResponsePos   int
	Response      HTTPResponse

	ClientAddressBuffer [21]byte
	ClientAddress       string

	DateBuf [31]byte

	/* Check must be the same as pointer's bit, if context is in use. */
	Check uintptr
}

type HTTPRouter func(w *HTTPResponse, r *HTTPRequest)

var (
	UnauthorizedError = HTTPError{StatusCode: HTTPStatusUnauthorized, DisplayMessage: "whoops... You have to sign in to see this page", LogError: NewError("whoops... You have to sign in to see this page")}
	ForbiddenError    = HTTPError{StatusCode: HTTPStatusForbidden, DisplayMessage: "whoops... Your permissions are insufficient", LogError: NewError("whoops... Your permissions are insufficient")}
)

func (rp *HTTPRequestParser) Parse(request string, r *HTTPRequest) (int, error) {
	var err error

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

		case HTTPStateMethod:
			if len(request[rp.Pos:]) < 4 {
				return 0, nil
			}
			switch request[rp.Pos : rp.Pos+4] {
			case "GET ":
				r.Method = "GET"
			case "POST":
				r.Method = "POST"
			default:
				return 0, NewError("Method not allowed")
			}
			rp.Pos += len(r.Method) + 1
			rp.State = HTTPStateURI
		case HTTPStateURI:
			lineEnd := FindChar(request[rp.Pos:], '\r')
			if lineEnd == -1 {
				return 0, nil
			}

			uriEnd := FindChar(request[rp.Pos:rp.Pos+lineEnd], ' ')
			if uriEnd == -1 {
				return 0, NewError("Bad Request")
			}

			queryStart := FindChar(request[rp.Pos:rp.Pos+lineEnd], '?')
			if queryStart != -1 {
				r.URL.Path = request[rp.Pos : rp.Pos+queryStart]
				r.URL.Query = request[rp.Pos+queryStart+1 : rp.Pos+uriEnd]
			} else {
				r.URL.Path = request[rp.Pos : rp.Pos+uriEnd]
				r.URL.Query = ""
			}

			const httpVersionPrefix = "HTTP/"
			httpVersion := request[rp.Pos+uriEnd+1 : rp.Pos+lineEnd]
			if httpVersion[:len(httpVersionPrefix)] != httpVersionPrefix {
				return 0, NewError("Bad Request")
			}
			r.Version = httpVersion[len(httpVersionPrefix):]
			rp.Pos += len(r.URL.Path) + len(r.URL.Query) + 1 + len(httpVersionPrefix) + len(r.Version) + len("\r\n")
			rp.State = HTTPStateUnknown
		case HTTPStateHeader:
			lineEnd := FindChar(request[rp.Pos:], '\r')
			if lineEnd == -1 {
				return 0, nil
			}
			header := request[rp.Pos : rp.Pos+lineEnd]
			r.Headers = append(r.Headers, header)
			rp.Pos += len(header) + len("\r\n")

			if StringStartsWith(header, "Content-Length: ") {
				header = header[len("Content-Length: "):]
				rp.ContentLength, err = strconv.Atoi(header)
				if err != nil {
					return 0, NewError("Bad Request")
				}
			}

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

	n := rp.Pos
	rp.State = HTTPStateMethod
	rp.ContentLength = 0
	rp.Pos = 0

	return n, nil
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

	w.SetHeader("Set-Cookie", unsafe.String(unsafe.SliceData(cookie), n))
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

	w.SetHeader("Set-Cookie", unsafe.String(unsafe.SliceData(cookie), n))
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

func (w *HTTPResponse) RedirectID(prefix string, id int, code HTTPStatus) {
	buffer := w.Arena.NewSlice(len(prefix) + 20)
	n := copy(buffer, prefix)
	n += SlicePutInt(buffer[n:], id)

	w.SetHeader("Location", unsafe.String(unsafe.SliceData(buffer), n))
	w.Bodies = w.Bodies[:0]
	w.StatusCode = code
}

func (w *HTTPResponse) Reset() {
	w.StatusCode = HTTPStatusOK
	w.Headers = w.Headers[:0]
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

func HTTPAppendResponse(wIovs *[]Iovec, w *HTTPResponse, dateBuf []byte) {
	var length int
	for i := 0; i < len(w.Bodies); i++ {
		length += int(w.Bodies[i].Len)
	}

	lengthBuf := w.Arena.NewSlice(20)
	n := SlicePutInt(lengthBuf, length)

	*wIovs = append(*wIovs, IovecForString("HTTP/1.1"), IovecForString(" "), IovecForString(Status2String[w.StatusCode]), IovecForString(" "), IovecForString(Status2Reason[w.StatusCode]), IovecForString("\r\n"))
	*wIovs = append(*wIovs, w.Headers...)
	*wIovs = append(*wIovs, IovecForString("Date: "), IovecForByteSlice(dateBuf), IovecForString("\r\n"))
	*wIovs = append(*wIovs, IovecForString("Content-Length: "), IovecForByteSlice(lengthBuf[:n]), IovecForString("\r\n"))
	if ContentTypeHTML(w.Bodies) {
		*wIovs = append(*wIovs, IovecForString("Content-Type: text/html; charset=\"UTF-8\"\r\n"))
	}
	*wIovs = append(*wIovs, IovecForString("\r\n"))
	*wIovs = append(*wIovs, w.Bodies...)
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

func NewHTTPContext() (*HTTPContext, error) {
	rb, err := NewCircularBuffer(PageSize)
	if err != nil {
		return nil, err
	}

	ctx := new(HTTPContext)
	ctx.RequestBuffer = rb
	ctx.ResponseIovs = make([]Iovec, 0, 512)

	ctx.Response.Headers = make([]Iovec, 0, 32)
	ctx.Response.Bodies = make([]Iovec, 0, 128)

	return ctx, nil
}

func HTTPContextFromCheckedPointer(ptr unsafe.Pointer) (*HTTPContext, uintptr) {
	uptr := uintptr(ptr)

	check := uptr & 0x1
	ctx := (*HTTPContext)(unsafe.Pointer(uptr - check))

	return ctx, check
}

func (ctx *HTTPContext) CheckedPointer() unsafe.Pointer {
	return unsafe.Pointer(uintptr(unsafe.Pointer(ctx)) | ctx.Check)
}

func (ctx *HTTPContext) Reset() {
	ctx.Check = 1 - ctx.Check
	ctx.RequestBuffer.Reset()
	ctx.ResponsePos = 0
	ctx.ResponseIovs = ctx.ResponseIovs[:0]
}

func (ctx *HTTPContext) SetClientAddress(addr SockAddrIn) {
	var n int

	n += SlicePutInt(ctx.ClientAddressBuffer[n:], int((addr.Addr&0x000000FF)>>0))
	ctx.ClientAddressBuffer[n] = ':'
	n++

	n += SlicePutInt(ctx.ClientAddressBuffer[n:], int((addr.Addr&0x0000FF00)>>8))
	ctx.ClientAddressBuffer[n] = '.'
	n++

	n += SlicePutInt(ctx.ClientAddressBuffer[n:], int((addr.Addr&0x00FF0000)>>16))
	ctx.ClientAddressBuffer[n] = '.'
	n++

	n += SlicePutInt(ctx.ClientAddressBuffer[n:], int((addr.Addr&0xFF000000)>>24))
	ctx.ClientAddressBuffer[n] = '.'
	n++

	n += SlicePutInt(ctx.ClientAddressBuffer[n:], int(SwapBytesInWord(addr.Port)))

	ctx.ClientAddress = unsafe.String(&ctx.ClientAddressBuffer[0], n)
}

func HTTPAccept(l int32, ctxPool Pool[HTTPContext]) (int32, *HTTPContext, error) {
	var addr SockAddrIn
	var addrLen uint32 = uint32(unsafe.Sizeof(addr))

	c, err := Accept(l, &addr, &addrLen)
	if err != nil {
		return -1, nil, fmt.Errorf("failed to accept incoming connection: %w", err)
	}

	ctx, err := ctxPool.Get()
	if err != nil {
		Close(c)
		return -1, nil, fmt.Errorf("failed to create new HTTP context: %w", err)
	}
	ctx.SetClientAddress(addr)

	return c, ctx, nil
}

func HTTPRead(c int32, ctx *HTTPContext) error {
	rBuf := &ctx.RequestBuffer
	n, err := Read(c, rBuf.RemainingSlice())
	if err != nil {
		return err
	}
	rBuf.Produce(int(n))
	return nil
}

func HTTPProcessRequests(ctx *HTTPContext, router HTTPRouter) {
	dateBuf := unsafe.Slice(&ctx.DateBuf[0], len(ctx.DateBuf))

	rBuf := &ctx.RequestBuffer
	parser := &ctx.RequestParser
	wIovs := &ctx.ResponseIovs

	w := &ctx.Response
	r := &ctx.Request
	r.Address = ctx.ClientAddress

	const pipelining = true
	for pipelining {
		n, err := parser.Parse(rBuf.UnconsumedString(), r)
		if err != nil {
			Errorf("Failed to parse HTTP request: %v", err)
		}
		if n == 0 {
			break
		}
		rBuf.Consume(n)

		w.Reset()
		Router(w, r)
		r.Reset()

		HTTPAppendResponse(wIovs, w, dateBuf)
	}
}

func HTTPWrite(c int32, ctx *HTTPContext) error {
	wIovs := ctx.ResponseIovs
	if len(wIovs[ctx.ResponsePos:]) > 0 {
		n, err := Writev(c, wIovs[ctx.ResponsePos:])
		if err != nil {
			return err
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
	return nil
}
