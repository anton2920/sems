/* TODO(anton2920): this is all temporary! */
package main

import "time"

type ID = int

type User struct {
	StringID  string
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
		1: &User{"1", "Anton", "Pavlovskii", "anton2920@gmail.com", "pass&word", UserRoleAdmin, time.Now()},
		2: &User{"2", "Larisa", "Sidorova", "teacher@masters.com", "teacher", UserRoleTeacher, time.Now()},
		3: &User{"3", "Anatolii", "Ivanov", "student@masters.com", "student", UserRoleStudent, time.Now()},
		4: &User{"4", "Robert", "Martin", "prestudent@masters.com", "prestudent", UserRolePrestudent, time.Now()},
	}
}
