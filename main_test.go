package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/anton2920/gofa/database"
	"github.com/anton2920/gofa/jail"
	"github.com/anton2920/gofa/log"
	"github.com/anton2920/gofa/net/http"
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

	var ctx http.Context
	var w http.Response
	var r http.Request

	r.URL.Path = endpoint

	w.StatusCode = http.StatusOK

	Router(&ctx, unsafe.Slice(&w, 1), unsafe.Slice(&r, 1))

	if w.StatusCode != expectedStatus {
		t.Errorf("GET %s -> %d, expected %d", endpoint, w.StatusCode, expectedStatus)
	}
}

func testGetAuth(t *testing.T, endpoint string, token string, expectedStatus http.Status) {
	t.Helper()

	var ctx http.Context
	var w http.Response
	var r http.Request

	r.Headers.Set("Cookie", fmt.Sprintf("Token=%s", token))
	r.URL.Path = endpoint

	w.StatusCode = http.StatusOK

	Router(&ctx, unsafe.Slice(&w, 1), unsafe.Slice(&r, 1))

	if w.StatusCode != expectedStatus {
		t.Errorf("GET %s -> %d, expected %d", endpoint, w.StatusCode, expectedStatus)
	}
}

func testPost(t *testing.T, endpoint string, form url.Values, expectedStatus http.Status) {
	t.Helper()

	var ctx http.Context
	var w http.Response
	var r http.Request

	r.Method = http.MethodPost
	r.URL.Path = endpoint
	r.Headers.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Body = []byte(form.Encode())

	w.StatusCode = http.StatusOK

	Router(&ctx, unsafe.Slice(&w, 1), unsafe.Slice(&r, 1))

	if w.StatusCode != expectedStatus {
		t.Errorf("POST %s -> %d (with form %v), expected %d", endpoint, w.StatusCode, form, expectedStatus)
	}
}

func testPostAuth(t *testing.T, endpoint string, token string, form url.Values, expectedStatus http.Status) {
	t.Helper()

	var ctx http.Context
	var w http.Response
	var r http.Request

	r.Method = http.MethodPost
	r.URL.Path = endpoint
	r.Headers.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Headers.Set("Cookie", fmt.Sprintf("Token=%s", token))
	r.Body = []byte(form.Encode())

	w.StatusCode = http.StatusOK

	Router(&ctx, unsafe.Slice(&w, 1), unsafe.Slice(&r, 1))

	if w.StatusCode != expectedStatus {
		t.Errorf("POST %s -> %d (with form %v), expected %d", endpoint, w.StatusCode, form, expectedStatus)
	}
}

func testPostInvalidFormAuth(t *testing.T, endpoint string, token string) {
	t.Helper()

	var ctx http.Context
	var w http.Response
	var r http.Request

	r.Method = http.MethodPost
	r.URL.Path = endpoint
	r.Headers.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Headers.Set("Cookie", fmt.Sprintf("Token=%s", token))
	r.Body = testInvalidForm

	w.StatusCode = http.StatusOK

	Router(&ctx, unsafe.Slice(&w, 1), unsafe.Slice(&r, 1))

	if w.StatusCode != http.StatusBadRequest {
		t.Errorf("POST %s -> %d (with invalid payload), expected 400", endpoint, w.StatusCode)
	}
}

func testWaitForJails() {
	for len(SubmissionVerifyChannel) > 0 {
		time.Sleep(time.Millisecond * 10)
	}

	for {
		ents, err := os.ReadDir(jail.JailsRootDir + "/containers")
		if (err != nil) || (len(ents) == 0) {
			return
		}
		time.Sleep(time.Millisecond * 100)
	}
}

func TestMain(m *testing.M) {
	var err error

	log.SetLevel(log.LevelError)

	WorkingDirectory, err = os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}

	if err := OpenDBs("db_test"); err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}
	defer CloseDBs()

	testCreateInitialDBs()

	jail.JailsRootDir = "./jails_test"
	os.MkdirAll(jail.JailsRootDir+"/containers", 0755)
	os.MkdirAll(jail.JailsRootDir+"/envs", 0755)

	go SubmissionVerifyWorker()

	now := time.Now()
	for i := database.ID(0); i < database.ID(len(testTokens)); i++ {
		testTokens[i], err = GenerateSessionToken()
		if err != nil {
			log.Fatalf("Failed to generate session token: %v", err)
		}
		Sessions[testTokens[i]] = &Session{ID: i, Expiry: now.Add(OneWeek)}
	}

	code := m.Run()

	testWaitForJails()
	os.RemoveAll("jails_test")
	os.RemoveAll("db_test")
	os.Exit(code)
}
