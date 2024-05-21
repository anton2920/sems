package main

import (
	"fmt"
	"os"
	"testing"
	"time"
	"unsafe"

	"github.com/anton2920/gofa/jail"
	"github.com/anton2920/gofa/log"
	"github.com/anton2920/gofa/net/http"
)

var (
	testInvalidToken = "invalid-token"
	testTokens       [4]string
)

func testGet(t *testing.T, endpoint string, expectedStatus http.Status) {
	t.Helper()

	var w http.Response
	var r http.Request

	r.URL.Path = endpoint
	w.StatusCode = http.StatusOK

	Router(unsafe.Slice(&w, 1), unsafe.Slice(&r, 1))

	if w.StatusCode != expectedStatus {
		t.Errorf("GET %s -> %d, expected %d", endpoint, w.StatusCode, expectedStatus)
	}
}

func testGetAuth(t *testing.T, endpoint string, token string, expectedStatus http.Status) {
	t.Helper()

	var w http.Response
	var r http.Request

	r.URL.Path = endpoint
	r.Headers = []string{fmt.Sprintf("Cookie: Token=%s", token)}
	w.StatusCode = http.StatusOK

	Router(unsafe.Slice(&w, 1), unsafe.Slice(&r, 1))

	if w.StatusCode != expectedStatus {
		t.Errorf("GET %s -> %d, expected %d", endpoint, w.StatusCode, expectedStatus)
	}
}

func testWaitForJails() {
	/* TODO(anton2920): implement. */
}

func TestMain(m *testing.M) {
	var err error

	log.SetLevel(log.LevelError)

	WorkingDirectory, err = os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}

	CreateInitialDB()

	jail.JailsRootDir = "./jails_test"

	now := time.Now()
	for i := 0; i < len(DB.Users); i++ {
		testTokens[i], err = GenerateSessionToken()
		if err != nil {
			log.Fatalf("Failed to generate session token: %v", err)
		}
		Sessions[testTokens[i]] = &Session{ID: i, Expiry: now.Add(OneWeek)}
	}

	code := m.Run()

	testWaitForJails()
	os.Exit(code)
}
