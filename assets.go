package main

import (
	"fmt"

	"github.com/anton2920/gofa/syscall"
)

const AssetsDir = "./assets"

var (
	BootstrapContents []byte
)

func LoadAssets() error {
	var stat syscall.Stat_t

	fd, err := syscall.Open(AssetsDir+"/bootstrap.min.css", syscall.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("failed to open bootstrap CSS file: %w", err)
	}

	if err := syscall.Fstat(fd, &stat); err != nil {
		return fmt.Errorf("failed to query stat of bootstrap file: %w", err)
	}

	BootstrapContents = make([]byte, stat.Size)
	n, err := syscall.Read(fd, BootstrapContents)
	if (err != nil) || (n != len(BootstrapContents)) {
		return fmt.Errorf("failed to read bootstrap contents: %w", err)
	}

	syscall.Close(fd)

	return nil
}
