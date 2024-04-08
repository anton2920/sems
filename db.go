/* TODO(anton2920): this is all temporary! */
package main

import "time"

type ID = int

type User struct {
	FirstName string
	LastName  string
	Email     string
	Password  string
	Role      UserRole
	CreatedOn time.Time
}

type Group struct {
	Students []ID
}

var DB struct {
	Users  map[ID]*User
	Groups map[ID]*Group
}

func init() {
	DB.Users = map[ID]*User{
		1: &User{"Anton", "Pavlovskii", "anton2920@gmail.com", "pass&word", UserRoleAdmin, time.Now()},
		2: &User{"Larisa", "Sidorova", "teacher@masters.com", "teacher", UserRoleTeacher, time.Now()},
		3: &User{"Anatolii", "Ivanov", "student@masters.com", "student", UserRoleStudent, time.Now()},
	}
}
