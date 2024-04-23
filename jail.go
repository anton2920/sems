package main

import (
	"errors"
	"fmt"
	"sync/atomic"
	"unsafe"
)

type Jail struct {
	ID    int32
	Index uint32
}

const (
	JAIL_CREATE = 1

	JailTemplatePath = "/usr/local/jails/templates/workster"
)

const (
	JailNamePrefix = "sems-"
	MaxJailNameLen = len(JailNamePrefix) + 20

	MaxJailRctlPrefixLen = len("jail:") + MaxJailNameLen + len(":")
	MaxJailRctlRuleLen   = MaxJailRctlPrefixLen + 20
)

var (
	JailsRootDir = "./jails"

	JailLastIndex uint32
)

func SlicePutJailName(buffer []byte, index int) int {
	var n int

	n += copy(buffer[n:], JailNamePrefix)
	n += SlicePutInt(buffer[n:], index)

	return n
}

func SlicePutJailPath(buffer []byte, index int) int {
	var n int

	n += copy(buffer[n:], WorkingDirectory)

	buffer[n] = '/'
	n++

	n += copy(buffer[n:], JailsRootDir)

	buffer[n] = '/'
	n++

	n += copy(buffer[n:], "containers/")
	n += SlicePutInt(buffer[n:], index)

	return n
}

func SlicePutJailTmp(buffer []byte, index int) int {
	var n int

	n += SlicePutJailPath(buffer[n:], index)
	n += copy(buffer[n:], "/tmp")

	return n
}

func SlicePutJailEnv(buffer []byte, index int) int {
	var n int

	n += copy(buffer[n:], WorkingDirectory)

	buffer[n] = '/'
	n++

	n += copy(buffer[n:], JailsRootDir)

	buffer[n] = '/'
	n++

	n += copy(buffer[n:], "envs/")
	n += SlicePutInt(buffer[n:], index)

	return n

}

func SlicePutJailRctlPrefix(buffer []byte, name []byte) int {
	var n int

	n += copy(buffer[n:], "jail:")
	n += copy(buffer[n:], name)

	buffer[n] = ':'
	n++

	return n
}

func SlicePutJailRctlRule(buffer []byte, prefix []byte, rule string) int {
	var n int

	n += copy(buffer[n:], prefix)
	n += copy(buffer[n:], rule)
	buffer[n] = '\x00'

	return n
}

func NewJail() (Jail, error) {
	var jail Jail
	var err error

	if err := Access(JailTemplatePath+"/tmp", 0); err != nil {
		return Jail{}, err
	}

	jail.Index = atomic.AddUint32(&JailLastIndex, 1)

	name := make([]byte, MaxJailNameLen)
	n := SlicePutJailName(name, int(jail.Index))
	name = name[:n+1]

	path := make([]byte, PATH_MAX)
	n = SlicePutJailPath(path, int(jail.Index))
	path = path[:n+1]

	tmp := make([]byte, PATH_MAX)
	n = SlicePutJailTmp(tmp, int(jail.Index))
	tmp = tmp[:n+1]

	env := make([]byte, PATH_MAX)
	n = SlicePutJailEnv(env, int(jail.Index))
	env = env[:n+1]

	if err := Mkdir(unsafe.String(unsafe.SliceData(path), len(path)), 0755); err != nil {
		if err.(ErrorWithCode).Code != EEXIST {
			return Jail{}, fmt.Errorf("failed to create path: %w", err)
		}
	}

	if err := Mkdir(unsafe.String(unsafe.SliceData(env), len(env)), 0755); err != nil {
		if err.(ErrorWithCode).Code != EEXIST {
			return Jail{}, fmt.Errorf("failed to create environment directory: %w", err)
		}
	}

	if err := Nmount([]Iovec{
		IovecForString("target\x00"), IovecForString(JailTemplatePath + "\x00"),
		IovecForString("fspath\x00"), IovecForByteSlice(path),
		IovecForString("fstype\x00"), IovecForString("nullfs\x00"),
		IovecForString("ro\x00"), Iovec{},
	}, 0); err != nil {
		return Jail{}, fmt.Errorf("failed to mount container directory: %w", err)
	}

	if err := Nmount([]Iovec{
		IovecForString("target\x00"), IovecForByteSlice(env),
		IovecForString("fspath\x00"), IovecForByteSlice(tmp),
		IovecForString("fstype\x00"), IovecForString("nullfs\x00"),
		IovecForString("rw\x00"), Iovec{},
	}, 0); err != nil {
		Unmount(unsafe.String(unsafe.SliceData(path), len(path)), 0)
		return Jail{}, fmt.Errorf("failed to mount environment directory: %w", err)
	}

	jail.ID, err = JailSet([]Iovec{
		IovecForString("host.hostname\x00"), IovecForString("sems-jail\x00"),
		IovecForString("name\x00"), IovecForByteSlice(name),
		IovecForString("path\x00"), IovecForByteSlice(path),
		IovecForString("persist\x00"), Iovec{},
	}, JAIL_CREATE)
	if err != nil {
		Unmount(unsafe.String(unsafe.SliceData(tmp), len(tmp)), 0)
		Unmount(unsafe.String(unsafe.SliceData(path), len(path)), 0)
		return Jail{}, err
	}

	prefix := make([]byte, MaxJailRctlPrefixLen)
	n = SlicePutJailRctlPrefix(prefix, name[:len(name)-1])
	prefix = prefix[:n+1]

	rule := make([]byte, MaxJailRctlRuleLen)
	rules := [...]string{
		"maxproc:deny=16",
		"vmemoryuse:deny=2684354560",
		"memoryuse:deny=536870912",
		"swapuse:deny=536870912",
	}
	for i := 0; i < len(rules); i++ {
		n := SlicePutJailRctlRule(rule, prefix[:len(prefix)-1], rules[i])

		if err := RctlAddRule(rule[:n+1]); err != nil {
			RctlRemoveRule(prefix)
			JailRemove(jail.ID)
			Unmount(unsafe.String(unsafe.SliceData(tmp), len(tmp)), 0)
			Unmount(unsafe.String(unsafe.SliceData(path), len(path)), 0)
			return Jail{}, fmt.Errorf("failed to add rule %d for jail: %w", i, err)
		}
	}

	return jail, nil
}

func RemoveJail(jail Jail) error {
	var err error

	name := make([]byte, MaxJailNameLen)
	n := SlicePutJailName(name, int(jail.Index))
	name = name[:n+1]

	path := make([]byte, PATH_MAX)
	n = SlicePutJailPath(path, int(jail.Index))
	path = path[:n+1]

	tmp := make([]byte, PATH_MAX)
	n = SlicePutJailTmp(tmp, int(jail.Index))
	tmp = tmp[:n+1]

	env := make([]byte, PATH_MAX)
	n = SlicePutJailEnv(env, int(jail.Index))
	env = env[:n+1]

	prefix := make([]byte, MaxJailRctlPrefixLen)
	n = SlicePutJailRctlPrefix(prefix, name[:len(name)-1])
	prefix = prefix[:n+1]

	if err1 := RctlRemoveRule(prefix); err1 != nil {
		err = fmt.Errorf("failed to remove jail rules: %w", err1)
	}

	if err1 := JailRemove(jail.ID); err1 != nil {
		err = errors.Join(err, fmt.Errorf("failed to remove jail: %w", err1))
	}

	if err1 := Unmount(unsafe.String(unsafe.SliceData(tmp), len(tmp)), 0); err1 != nil {
		err = errors.Join(err, fmt.Errorf("failed to unmount environment: %w", err1))
	}

	if err1 := Unmount(unsafe.String(unsafe.SliceData(path), len(path)), 0); err1 != nil {
		err = errors.Join(err, fmt.Errorf("failed to unmount container: %w", err1))
	}

	if err1 := Rmdir(unsafe.String(unsafe.SliceData(env), len(env))); err1 != nil {
		err = errors.Join(err, fmt.Errorf("failed to remove environment directory: %w", err1))
	}

	if err1 := Rmdir(unsafe.String(unsafe.SliceData(path), len(path))); err1 != nil {
		err = errors.Join(err, fmt.Errorf("failed to remove container directory: %w", err1))
	}

	return nil
}
