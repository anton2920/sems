package main

import "unsafe"

type E struct {
	Message string
	Code    int
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

func (e E) Error() string {
	var buf [512]byte

	n := copy(buf[:], e.Message)
	buf[n] = ' '
	n++

	if e.Code != 0 {
		n += SlicePutInt(buf[n:], e.Code)
	}

	return string(unsafe.Slice(&buf[0], n))
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
