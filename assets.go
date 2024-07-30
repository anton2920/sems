package main

import (
	"fmt"

	"github.com/anton2920/gofa/prof"
	"github.com/anton2920/gofa/syscall"
)

const AssetsDir = "./assets"

var (
	BootstrapCSS []byte
	BootstrapJS  []byte
)

func LoadAssetFile(path string) ([]byte, error) {
	defer prof.End(prof.Begin(""))

	var stat syscall.Stat_t

	fd, err := syscall.Open(path, syscall.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open asset file: %w", err)
	}
	defer syscall.Close(fd)

	if err := syscall.Fstat(fd, &stat); err != nil {
		return nil, fmt.Errorf("failed to query stat of asset file: %w", err)
	}

	buffer := make([]byte, stat.Size)
	n, err := syscall.Read(fd, buffer)
	if (err != nil) || (n != len(buffer)) {
		return nil, fmt.Errorf("failed to read asset file contents: %w", err)
	}

	return buffer, nil
}

func LoadAssets() error {
	defer prof.End(prof.Begin(""))

	var err error

	BootstrapCSS, err = LoadAssetFile(AssetsDir + "/bootstrap.min.css")
	if err != nil {
		return fmt.Errorf("failed to read bootstrap CSS file: %w", err)
	}

	BootstrapJS, err = LoadAssetFile(AssetsDir + "/bootstrap.min.js")
	if err != nil {
		return fmt.Errorf("failed to read bootstrap JS file: %w", err)
	}

	return nil
}
