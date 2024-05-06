package main

import "unsafe"

type SockAddrIn struct {
	Len    uint8
	Family uint8
	Port   uint16
	Addr   uint32
	_      [8]byte
}

/* NOTE(anton2920): actually SockAddr is the following structure:
 * struct sockaddr {
 *	unsigned char	sa_len;		// total length
 *	sa_family_t	sa_family;	// address family
 *	char		sa_data[14];	// actually longer; address value
 * };
 * But because I don't really care, and sizes are the same, I made them synonyms.
 */
type SockAddr = SockAddrIn

const (
	/* From <sys/socket.h>. */
	AF_INET = 2
	PF_INET = AF_INET

	SOCK_STREAM = 1

	SOL_SOCKET   = 0xFFFF
	SO_REUSEADDR = 0x00000004
	SO_RCVTIMEO  = 0x1006

	SHUT_RD = 0
	SHUT_WR = 1

	/* From <netinet/in.h>. */
	INADDR_ANY = 0
)

func SwapBytesInWord(x uint16) uint16 {
	return ((x << 8) & 0xFF00) | (x >> 8)
}

func TCPListen(port uint16) (int32, error) {
	l, err := Socket(PF_INET, SOCK_STREAM, 0)
	if err != nil {
		return -1, err
	}

	var enable int32 = 1
	if err := Setsockopt(l, SOL_SOCKET, SO_REUSEADDR, unsafe.Pointer(&enable), uint32(unsafe.Sizeof(enable))); err != nil {
		return -1, err
	}

	addr := SockAddrIn{Family: AF_INET, Addr: INADDR_ANY, Port: SwapBytesInWord(port)}
	if err := Bind(l, &addr, uint32(unsafe.Sizeof(addr))); err != nil {
		return -1, err
	}

	const backlog = 128
	if err := Listen(l, backlog); err != nil {
		return -1, err
	}

	return l, nil
}
