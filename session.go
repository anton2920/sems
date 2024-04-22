package main

import (
	"encoding/base64"
	"encoding/gob"
	"os"
	"sync"
	"time"
)

type Session struct {
	GobMutex
	ID     int
	Expiry time.Time
}

const OneWeek = time.Hour * 24 * 7

const SessionsFile = "sessions.gob"

var (
	Sessions     = make(map[string]*Session)
	SessionsLock sync.RWMutex
)

func GetSessionFromToken(token string) (*Session, error) {
	SessionsLock.RLock()
	session, ok := Sessions[token]
	SessionsLock.RUnlock()
	if !ok {
		return nil, NewError("session for this token does not exist")
	}

	session.Lock()
	defer session.Unlock()

	now := time.Now()
	if session.Expiry.Before(now) {
		SessionsLock.Lock()
		delete(Sessions, token)
		SessionsLock.Unlock()
		return nil, NewError("session for this token has expired")
	}

	session.Expiry = now.Add(OneWeek)
	return session, nil
}

func GetSessionFromRequest(r *HTTPRequest) (*Session, error) {
	return GetSessionFromToken(r.Cookie("Token"))
}

func GenerateSessionToken() (string, error) {
	const n = 64
	buffer := make([]byte, n)

	/* NOTE(anton2920): see encoding/base64/base64.go:294. */
	token := make([]byte, (n+2)/3*4)

	if _, err := Getrandom(buffer, 0); err != nil {
		return "", err
	}

	base64.StdEncoding.Encode(token, buffer)

	return string(token), nil
}

func StoreSessionsToFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := gob.NewEncoder(f)
	SessionsLock.Lock()
	defer SessionsLock.Unlock()

	if err := enc.Encode(Sessions); err != nil {
		return err
	}

	return nil
}

func RestoreSessionsFromFile(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	dec := gob.NewDecoder(f)
	if err := dec.Decode(&Sessions); err != nil {
		return err
	}

	return nil
}
