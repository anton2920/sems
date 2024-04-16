/* TODO(anton2920): this is all temporary! */
package main

import (
	"encoding/gob"
	"os"
	"strconv"
	"time"
)

type (
	Group struct {
		StringID  string
		Name      string
		Users     []*User
		CreatedOn time.Time
	}

	Subject struct {
		StringID  string
		Name      string
		Teacher   *User
		Group     *Group
		CreatedOn time.Time
	}

	User struct {
		StringID  string
		FirstName string
		LastName  string
		Email     string
		Password  string
		CreatedOn time.Time

		Courses []*Course
	}
)

const AdminID = 0

const DBFile = "db.gob"

var DB struct {
	Users    []User
	Groups   []Group
	Subjects []Subject
}

func init() {
	DB.Users = []User{
		AdminID: {FirstName: "Admin", LastName: "Admin", Email: "admin@masters.com", Password: "admin", CreatedOn: time.Now()},
		{FirstName: "Larisa", LastName: "Sidorova", Email: "teacher@masters.com", Password: "teacher", CreatedOn: time.Now()},
		{FirstName: "Anatolii", LastName: "Ivanov", Email: "student@masters.com", Password: "student", CreatedOn: time.Now()},
		{FirstName: "Robert", LastName: "Martin", Email: "student2@masters.com", Password: "student2", CreatedOn: time.Now()},
	}
	for id := 0; id < len(DB.Users); id++ {
		user := &DB.Users[id]
		user.StringID = strconv.Itoa(id)
	}

	DB.Groups = []Group{
		{Name: "18-SWE", Users: []*User{&DB.Users[2], &DB.Users[3]}, CreatedOn: time.Now()},
	}
	for id := 0; id < len(DB.Groups); id++ {
		group := &DB.Groups[id]
		group.StringID = strconv.Itoa(id)
	}
}

func StoreDBToFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return WrapErrorWithTrace(err)
	}
	defer f.Close()

	if err := gob.NewEncoder(f).Encode(DB); err != nil {
		return WrapErrorWithTrace(err)
	}

	return nil
}

func RestoreDBFromFile(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return WrapErrorWithTrace(err)
	}
	defer f.Close()

	if err := gob.NewDecoder(f).Decode(&DB); err != nil {
		return WrapErrorWithTrace(err)
	}

	return nil
}
