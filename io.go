package main

const (
	/* From <fcntl.h>. */
	O_RDONLY   = 0x0000
	O_RDWR     = 0x0002
	O_NONBLOCK = 0x0004

	F_SETFL = 4

	SEEK_SET = 0
	SEEK_END = 2

	PATH_MAX = 1024
)

func ReadFull(fd int32, buf []byte) (int64, error) {
	var read int64
	for read < int64(len(buf)) {
		n, err := Read(fd, buf[read:])
		if err != nil {
			code := err.(E).Code
			if code != EINTR {
				return n, err
			}
			continue
		} else if n == 0 {
			break
		}
		read += n
	}

	return read, nil
}

func ReadEntireFile(fd int32) ([]byte, error) {
	flen, err := Lseek(fd, 0, SEEK_END)
	if err != nil {
		return nil, err
	}

	data := make([]byte, flen)
	if _, err := Lseek(fd, 0, SEEK_SET); err != nil {
		return nil, err
	}

	if _, err := ReadFull(fd, data); err != nil {
		return nil, err
	}

	return data, nil
}
