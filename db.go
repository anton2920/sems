/* TODO(anton2920): this is all temporary! */
package main

import (
	"encoding/gob"
	"os"
	"time"
)

type ID = int

type User struct {
	StringID  string
	FirstName string
	LastName  string
	Email     string
	Password  string
	RoleID    UserRole
	CreatedOn time.Time
}

type Group struct {
	StringID  string
	Name      string
	Teacher   *User
	Students  []*User
	CreatedOn time.Time
}

const DBFile = "db.gob"

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
		5: &User{"5", "Sidor", "Sidorov", "student2@masters.com", "student2", UserRoleStudent, time.Now()},
	}
	DB.Groups = map[ID]*Group{
		1: &Group{"1", "18-SWE", DB.Users[2], []*User{DB.Users[3], DB.Users[5]}, time.Now()},
	}
}

func StoreDBToFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := gob.NewEncoder(f).Encode(DB); err != nil {
		return err
	}

	return nil
}

func RestoreDBFromFile(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := gob.NewDecoder(f).Decode(&DB); err != nil {
		return err
	}

	return nil
}
