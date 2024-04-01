package main

import (
	"fmt"
	"runtime/debug"
)

type E struct {
	Message string
	Code    int
}

/* TODO(anton2920): think about placing this into 'http.go' file. */
type HTTPError struct {
	StatusCode HTTPStatus
	Message    string
}

type PanicError struct {
	Value interface{}
	Trace []byte
}

const (
	/* From <errno.h>. */
	ENOENT      = 2      /* No such file or directory */
	EINTR       = 4      /* Interrupted system call */
	EPIPE       = 32     /* Broken pipe */
	EAGAIN      = 35     /* Resource temporarily unavailable */
	EWOULDBLOCK = EAGAIN /* Operation would block */
	EINPROGRESS = 36     /* Operation now in progress */
	EOPNOTSUPP  = 45     /* Operation not supported */
	ECONNRESET  = 54     /* Connection reset by peer */
	ENOSYS      = 78     /* Function not implemented */
)

var (
	NotFoundError      = HTTPError{StatusCode: HTTPStatusNotFound, Message: "whoops... Requested page not found"}
	TryAgainLaterError = HTTPError{StatusCode: HTTPStatusInternalServerError, Message: "whoops... Something went wrong. Please try again later"}
)

func (e E) Error() string {
	buffer := make([]byte, 512)
	n := copy(buffer, e.Message)
	buffer[n] = ' '
	n++

	if e.Code != 0 {
		n += SlicePutInt(buffer[n:], e.Code)
	}

	return string(buffer[:n])
}

func Error(msg string) error {
	return error(E{Message: msg})
}

func ErrorWithCode(msg string, code int) error {
	return error(E{Message: msg, Code: code})
}

func SyscallError(msg string, errno uintptr) error {
	if errno == 0 {
		return nil
	}
	return error(E{Message: msg, Code: int(errno)})
}

func NewHTTPError(statusCode HTTPStatus, message string) HTTPError {
	return HTTPError{StatusCode: statusCode, Message: message}
}

func (e HTTPError) Error() string {
	return e.Message
}

func NewPanicError(value interface{}) PanicError {
	return PanicError{Value: value, Trace: debug.Stack()}
}

func (e PanicError) Error() string {
	buffer := make([]byte, 0, 1024)
	buffer = fmt.Appendf(buffer, "%v\n", e.Value)
	buffer = append(buffer, e.Trace...)
	return string(buffer)
}

func ErrorPageHandler(w *HTTPResponse, r *HTTPRequest, statusCode HTTPStatus, err error) {
	const pageFormat = `
<!DOCTYPE html>
<head>
	<title>Error</title>
</head>
<body>
	<h1>Master's degree</h1>
	<h2>Error</h2>

	<p>Error: %v.</p>
</body>
</html>
`

	w.StatusCode = statusCode
	fmt.Fprintf(w, pageFormat, err)
}
