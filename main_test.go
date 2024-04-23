package main

import (
	"os"
	"testing"
)

func testWaitForJails() {
	/* TODO(anton2920): implement. */
}

func TestMain(m *testing.M) {
	var err error

	WorkingDirectory, err = os.Getwd()
	if err != nil {
		Fatalf("Failed to get current working directory: %v", err)
	}

	code := m.Run()

	testWaitForJails()
	os.Exit(code)
}
