package main

import "unsafe"

const (
	/* See <sys/syscall.h>. */
	SYS_accept        = 30
	SYS_bind          = 104
	SYS_clock_gettime = 232
	SYS_close         = 6
	SYS_exit          = 1
	SYS_fcntl         = 92
	SYS_fstat         = 551
	SYS_ftruncate     = 480
	SYS_getrandom     = 563
	SYS_kevent        = 560
	SYS_kqueue        = 362
	SYS_listen        = 106
	SYS_lseek         = 478
	SYS_mmap          = 477
	SYS_nanosleep     = 240
	SYS_open          = 5
	SYS_read          = 3
	SYS_setsockopt    = 105
	SYS_shm_open2     = 571
	SYS_shutdown      = 134
	SYS_socket        = 97
	SYS_write         = 4
	SYS_writev        = 121
)

//go:linkname SyscallEnter runtime.entersyscall
func SyscallEnter()

//go:linkname SyscallExit runtime.exitsyscall
func SyscallExit()

func RawSyscall(trap, a1, a2, a3 uintptr) (r1, r2, errno uintptr)

func Syscall(trap, a1, a2, a3 uintptr) (r1, r2, errno uintptr) {
	SyscallEnter()
	r1, r2, errno = RawSyscall(trap, a1, a2, a3)
	SyscallExit()
	return
}

func RawSyscall6(trap, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2, errno uintptr)

func Syscall6(trap, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2, errno uintptr) {
	SyscallEnter()
	r1, r2, errno = RawSyscall6(trap, a1, a2, a3, a4, a5, a6)
	SyscallExit()
	return
}

func Accept(s int32, addr *SockAddr, addrlen *uint32) (int32, error) {
	r1, _, errno := Syscall(SYS_accept, uintptr(s), uintptr(unsafe.Pointer(addr)), uintptr(unsafe.Pointer(addrlen)))
	return int32(r1), NewSyscallError("accept failed with code", errno)
}

func Bind(s int32, addr *SockAddr, addrlen uint32) error {
	_, _, errno := RawSyscall(SYS_bind, uintptr(s), uintptr(unsafe.Pointer(addr)), uintptr(addrlen))
	return NewSyscallError("bind failed with code", errno)
}

func ClockGettime(clockID int32, tp *Timespec) error {
	_, _, errno := RawSyscall(SYS_clock_gettime, uintptr(clockID), uintptr(unsafe.Pointer(tp)), 0)
	return NewSyscallError("clock_gettime failed with code", errno)
}

func Close(fd int32) error {
	_, _, errno := Syscall(SYS_close, uintptr(fd), 0, 0)
	return NewSyscallError("close failed with code", errno)
}

func Exit(status int32) {
	RawSyscall(SYS_exit, uintptr(status), 0, 0)
}

func Fcntl(fd, cmd, arg int32) error {
	_, _, errno := Syscall(SYS_fcntl, uintptr(fd), uintptr(cmd), uintptr(arg))
	return NewSyscallError("fcntl failed with code", errno)
}

func Fstat(fd int32, sb *Stat) error {
	_, _, errno := Syscall(SYS_fstat, uintptr(fd), uintptr(unsafe.Pointer(sb)), 0)
	return NewSyscallError("fstat failed with code", errno)
}

func Ftruncate(fd int32, length int64) error {
	_, _, errno := Syscall(SYS_ftruncate, uintptr(fd), uintptr(length), 0)
	return NewSyscallError("ftruncate failed with code", errno)
}

func Getrandom(buf []byte, flags uint32) (int64, error) {
	r1, _, errno := Syscall(SYS_getrandom, uintptr(unsafe.Pointer(unsafe.SliceData(buf))), uintptr(len(buf)), uintptr(flags))
	return int64(r1), NewSyscallError("getrandom failed with code", errno)
}

func Kevent(kq int32, changelist []Kevent_t, eventlist []Kevent_t, timeout *Timespec) (int, error) {
	r1, _, errno := Syscall6(SYS_kevent, uintptr(kq), uintptr(unsafe.Pointer(unsafe.SliceData(changelist))), uintptr(len(changelist)), uintptr(unsafe.Pointer(unsafe.SliceData(eventlist))), uintptr(len(eventlist)), uintptr(unsafe.Pointer(timeout)))
	return int(r1), NewSyscallError("kevent failed with code", errno)
}

func Kqueue() (int32, error) {
	r1, _, errno := RawSyscall(SYS_kqueue, 0, 0, 0)
	return int32(r1), NewSyscallError("kqueue failed with code", errno)
}

func Listen(s int32, backlog int32) error {
	_, _, errno := RawSyscall(SYS_listen, uintptr(s), uintptr(backlog), 0)
	return NewSyscallError("listen failed with code", errno)
}

func Lseek(fd int32, offset int64, whence int32) (int64, error) {
	r1, _, errno := RawSyscall(SYS_lseek, uintptr(fd), uintptr(offset), uintptr(whence))
	return int64(r1), NewSyscallError("lseek failed with code", errno)
}

func Mmap(addr unsafe.Pointer, len uint64, prot, flags, fd int32, offset int64) (unsafe.Pointer, error) {
	r1, _, errno := RawSyscall6(SYS_mmap, uintptr(addr), uintptr(len), uintptr(prot), uintptr(flags), uintptr(fd), uintptr(offset))
	return unsafe.Pointer(r1), NewSyscallError("mmap failed with code", errno)
}

func Nanosleep(rqtp, rmtp *Timespec) error {
	_, _, errno := Syscall(SYS_nanosleep, uintptr(unsafe.Pointer(rqtp)), uintptr(unsafe.Pointer(rmtp)), 0)
	return NewSyscallError("nanosleep failed with code", errno)
}

func Open(path string, flags int32, mode uint16) (int32, error) {
	r1, _, errno := Syscall(SYS_open, uintptr(unsafe.Pointer(unsafe.StringData(path+"\x00"))), uintptr(flags), uintptr(mode))
	return int32(r1), NewSyscallError("open failed with code", errno)
}

func Read(fd int32, buf []byte) (int64, error) {
	r1, _, errno := Syscall(SYS_read, uintptr(fd), uintptr(unsafe.Pointer(unsafe.SliceData(buf))), uintptr(len(buf)))
	return int64(r1), NewSyscallError("read failed with code", errno)
}

func Setsockopt(s, level, optname int32, optval unsafe.Pointer, optlen uint32) error {
	_, _, errno := RawSyscall6(SYS_setsockopt, uintptr(s), uintptr(level), uintptr(optname), uintptr(optval), uintptr(optlen), 0)
	return NewSyscallError("setsockopt failed with code", errno)
}

func ShmOpen2(path string, flags int32, mode uint16, shmflags int32, name string) (int32, error) {
	r1, _, errno := RawSyscall6(SYS_shm_open2, uintptr(unsafe.Pointer(unsafe.StringData(path))), uintptr(flags), uintptr(mode), uintptr(shmflags), uintptr(unsafe.Pointer(unsafe.StringData(name))), 0)
	return int32(r1), NewSyscallError("shm_open2 failed with code", errno)
}

func Shutdown(s int32, how int32) error {
	_, _, errno := RawSyscall(SYS_shutdown, uintptr(s), uintptr(how), 0)
	return NewSyscallError("shutdown failed with code", errno)
}

func Socket(domain, typ, protocol int32) (int32, error) {
	r1, _, errno := RawSyscall(SYS_socket, uintptr(domain), uintptr(typ), uintptr(protocol))
	return int32(r1), NewSyscallError("socket failed with code", errno)
}

func Write(fd int32, buf []byte) (int64, error) {
	r1, _, errno := Syscall(SYS_write, uintptr(fd), uintptr(unsafe.Pointer(unsafe.SliceData(buf))), uintptr(len(buf)))
	return int64(r1), NewSyscallError("write failed with code", errno)
}

func Writev(fd int32, iov []Iovec) (int64, error) {
	r1, _, errno := Syscall(SYS_writev, uintptr(fd), uintptr(unsafe.Pointer(unsafe.SliceData(iov))), uintptr(len(iov)))
	return int64(r1), NewSyscallError("writev failed with code", errno)
}
