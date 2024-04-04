package main

import (
	"errors"
	"sync"
	"time"
)

type Session struct {
	sync.Mutex
	ID     int
	Expiry time.Time
}

const OneWeek = time.Hour * 24 * 7

var (
	Sessions     = make(map[string]*Session)
	SessionsLock sync.RWMutex
)

func GetSessionFromToken(token string) (*Session, error) {
	SessionsLock.RLock()
	session, ok := Sessions[token]
	SessionsLock.RUnlock()
	if !ok {
		return nil, errors.New("session for this token does not exist")
	}

	session.Lock()
	defer session.Unlock()

	now := time.Now()
	if session.Expiry.Before(now) {
		SessionsLock.Lock()
		delete(Sessions, token)
		SessionsLock.Unlock()
		return nil, errors.New("session for this token has expired")
	}

	session.Expiry = now.Add(OneWeek)
	return session, nil
}

func GetSessionFromRequest(r *HTTPRequest) (*Session, error) {
	return GetSessionFromToken(r.Cookie("Token"))
}

func GenerateSessionToken() (string, error) {
	return "some-secret-string", nil
}
