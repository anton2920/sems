package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/anton2920/gofa/jail"
	"github.com/anton2920/gofa/log"
	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/net/url"
)

var (
	testInvalidToken = "invalid-token"
	testTokens       [4]string

	testInvalidForm = []byte("a=1;b=2")
)

func testString(len int) string {
	return strings.Repeat("a", len)
}

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

	r.Headers = []string{fmt.Sprintf("Cookie: Token=%s", token)}
	r.URL.Path = endpoint

	w.StatusCode = http.StatusOK

	Router(unsafe.Slice(&w, 1), unsafe.Slice(&r, 1))

	if w.StatusCode != expectedStatus {
		t.Errorf("GET %s -> %d, expected %d", endpoint, w.StatusCode, expectedStatus)
	}
}

func testPost(t *testing.T, endpoint string, form url.Values, expectedStatus http.Status) {
	t.Helper()

	var w http.Response
	var r http.Request

	r.URL.Path = endpoint
	r.Form = form

	w.StatusCode = http.StatusOK

	Router(unsafe.Slice(&w, 1), unsafe.Slice(&r, 1))

	if w.StatusCode != expectedStatus {
		t.Errorf("POST %s -> %d (with form %v), expected %d", endpoint, w.StatusCode, form, expectedStatus)
	}
}

func testPostAuth(t *testing.T, endpoint string, token string, form url.Values, expectedStatus http.Status) {
	t.Helper()

	var w http.Response
	var r http.Request

	r.Headers = []string{fmt.Sprintf("Cookie: Token=%s", token)}
	r.URL.Path = endpoint
	r.Form = form

	w.StatusCode = http.StatusOK

	Router(unsafe.Slice(&w, 1), unsafe.Slice(&r, 1))

	if w.StatusCode != expectedStatus {
		t.Errorf("POST %s -> %d (with form %v), expected %d", endpoint, w.StatusCode, form, expectedStatus)
	}
}

func testPostInvalidFormAuth(t *testing.T, endpoint string, token string) {
	t.Helper()

	var w http.Response
	var r http.Request

	r.Headers = []string{fmt.Sprintf("Cookie: Token=%s", token)}
	r.Body = testInvalidForm
	r.URL.Path = endpoint

	w.StatusCode = http.StatusOK

	Router(unsafe.Slice(&w, 1), unsafe.Slice(&r, 1))

	if w.StatusCode != http.StatusBadRequest {
		t.Errorf("POST %s -> %d (with invalid payload), expected 400", endpoint, w.StatusCode)
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

	DB2, err = OpenDB("db_test")
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}
	defer CloseDB(DB2)

	CreateInitialDB()

	jail.JailsRootDir = "./jails_test"

	now := time.Now()
	for i := int32(0); i < int32(len(testTokens)); i++ {
		testTokens[i], err = GenerateSessionToken()
		if err != nil {
			log.Fatalf("Failed to generate session token: %v", err)
		}
		Sessions[testTokens[i]] = &Session{ID: i, Expiry: now.Add(OneWeek)}
	}

	code := m.Run()

	testWaitForJails()
	os.RemoveAll("db_test")
	os.Exit(code)
}
