package main

import (
	"fmt"
	"strconv"
	"unsafe"
)

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

	SOL_SOCKET = 0xFFFF

	SO_REUSEADDR    = 0x00000004
	SO_REUSEPORT    = 0x00000200
	SO_REUSEPORT_LB = 0x00010000
	SO_RCVTIMEO     = 0x00001006

	SHUT_RD = 0
	SHUT_WR = 1

	/* From <netinet/in.h>. */
	INADDR_ANY = 0
)

func SwapBytesInWord(x uint16) uint16 {
	return ((x << 8) & 0xFF00) | (x >> 8)
}

func ParseAddressString(address string) (uint32, uint16, error) {
	var addr uint32

	colon := FindChar(address, ':')
	if colon == -1 {
		return 0, 0, NewError("no port specified")
	}

	part, err := strconv.Atoi(address[colon+1:])
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse port value: %w", err)
	}
	port := SwapBytesInWord(uint16(part))

	address = address[:colon]
	dot := FindChar(address, '.')
	if dot == -1 {
		return INADDR_ANY, port, nil
	}
	part, err = strconv.Atoi(address[:dot])
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse first address octet: %w", err)
	}
	addr |= uint32(part)

	address = address[dot+1:]
	dot = FindChar(address, '.')
	if dot == -1 {
		return 0, 0, fmt.Errorf("expected second address octet, found nothing")
	}
	part, err = strconv.Atoi(address[:dot])
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse second address octet: %w", err)
	}
	addr |= uint32(part) << 8

	address = address[dot+1:]
	dot = FindChar(address, '.')
	if dot == -1 {
		return 0, 0, fmt.Errorf("expected third address octet, found nothing")
	}
	part, err = strconv.Atoi(address[:dot])
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse third address octet: %w", err)
	}
	addr |= uint32(part) << 16

	address = address[dot+1:]
	part, err = strconv.Atoi(address)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse fourth address octet: %w", err)
	}
	addr |= uint32(part) << 24

	return addr, port, nil
}

/* TCPListen creates TCP/IPv4 socket and starts listening on a specified address. */
func TCPListen(address string, backlog int) (int32, error) {
	l, err := Socket(PF_INET, SOCK_STREAM, 0)
	if err != nil {
		return -1, fmt.Errorf("failed to create new socket: %w", err)
	}

	var enable int32 = 1
	if err := Setsockopt(l, SOL_SOCKET, SO_REUSEPORT_LB, unsafe.Pointer(&enable), uint32(unsafe.Sizeof(enable))); err != nil {
		return -1, fmt.Errorf("failed to apply options to socket: %w", err)
	}

	addr, port, err := ParseAddressString(address)
	if err != nil {
		return -1, fmt.Errorf("failed to parse address string: %w", err)
	}
	sin := SockAddrIn{Family: AF_INET, Addr: addr, Port: port}
	if err := Bind(l, &sin, uint32(unsafe.Sizeof(sin))); err != nil {
		return -1, fmt.Errorf("failed to bind socket to address: %w", err)
	}

	if err := Listen(l, int32(backlog)); err != nil {
		return -1, fmt.Errorf("failed to listen for incoming connections: %w", err)
	}

	return l, nil
}
