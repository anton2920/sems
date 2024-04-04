package main

import "time"

/* TODO(anton2920): this is all temporary! */
type UserEntry struct {
	FirstName string
	LastName  string
	Email     string
	Password  string
	CreatedOn time.Time
}

var UsersDB = map[int]UserEntry{
	1: {"Anton", "Pavlovskii", "anton2920@gmail.com", "pass&word", time.Now()},
}
