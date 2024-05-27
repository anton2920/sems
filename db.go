/* TODO(anton2920): this is all temporary! */
package main

import (
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"time"
	"unsafe"

	"github.com/anton2920/gofa/syscall"
)

type Database struct {
	UsersFile    int32
	GroupsFile   int32
	CoursesFile  int32
	SubjectsFile int32
}

const AdminID = 0

const DBFile = "db.gob"

var DB struct {
	Users    []User
	Groups   []Group
	Subjects []Subject
}

//go:noescape
//go:nosplit
func DBString(ptr *byte, len int) string

//go:noescape
//go:nosplit
func DBSlice(ptr *byte, len int) []byte

func Offset2String(s string, base *byte) string {
	return unsafe.String((*byte)(unsafe.Add(unsafe.Pointer(base), uintptr(unsafe.Pointer(unsafe.StringData(s))))), len(s))
}

func Offset2Slice[T any](s []T, base *byte) []T {
	return unsafe.Slice((*T)(unsafe.Add(unsafe.Pointer(base), uintptr(unsafe.Pointer(unsafe.SliceData(s))))), len(s))
}

func String2Offset(s string, offset int) string {
	return DBString((*byte)(unsafe.Pointer(uintptr(offset-len(s)))), len(s))
}

func Slice2Offset[T any](s []T, offset int) []T {
	var t T
	return unsafe.Slice((*T)(unsafe.Pointer((unsafe.SliceData(DBSlice((*byte)(unsafe.Pointer(uintptr(offset-len(s)*int(unsafe.Sizeof(t))))), len(s)))))), len(s))
}

func CreateInitialDB() {
	DB.Users = []User{
		AdminID: {ID: AdminID, FirstName: "Admin", LastName: "Admin", Email: "admin@masters.com", Password: "admin", CreatedOn: time.Now(), Courses: []*Course{
			&Course{
				Name: "Programming basics",
				Lessons: []*Lesson{
					&Lesson{
						Name:   "Introduction",
						Theory: "This is an introduction.",
						Steps: []interface{}{
							&StepTest{
								Name: "Back-end development basics",
								Questions: []Question{
									Question{
										Name: "What is an API?",
										Answers: []string{
											"One",
											"Two",
											"Three",
											"Four",
										},
										CorrectAnswers: []int{2},
									},
									Question{
										Name: "To be or not to be?",
										Answers: []string{
											"To be",
											"Not to be",
										},
										CorrectAnswers: []int{0, 1},
									},
									Question{
										Name: "Third question",
										Answers: []string{
											"What?",
											"Where?",
											"When?",
											"Correct",
										},
										CorrectAnswers: []int{3},
									},
								},
							},
							&StepProgramming{
								Name:        "Hello, world",
								Description: "Print 'hello, world' in your favorite language",
								Checks: [2][]Check{
									CheckTypeExample: []Check{
										Check{Input: "aaa", Output: "bbb"},
										Check{Input: "ccc", Output: "ddd"},
									},
									CheckTypeTest: []Check{
										Check{Input: "fff", Output: "eee"},
									},
								},
							},
						},
					},
				},
			},
		}},
		{FirstName: "Larisa", LastName: "Sidorova", Email: "teacher@masters.com", Password: "teacher", CreatedOn: time.Now(), Courses: []*Course{
			&Course{
				Name: "Test course",
				Lessons: []*Lesson{
					&Lesson{
						Name:   "Test lesson",
						Theory: "This is a test lesson.",
					},
				},
			},
		}},
		{FirstName: "Anatolii", LastName: "Ivanov", Email: "student@masters.com", Password: "student", CreatedOn: time.Now()},
		{FirstName: "Robert", LastName: "Martin", Email: "student2@masters.com", Password: "student2", CreatedOn: time.Now()},
	}
	for id := AdminID + 1; id < len(DB.Users); id++ {
		DB.Users[id].ID = id
	}

	DB.Groups = []Group{
		{Name: "18-SWE", Students: []*User{&DB.Users[2], &DB.Users[3]}, CreatedOn: time.Now()},
	}
	for id := 0; id < len(DB.Groups); id++ {
		DB.Groups[id].ID = id
	}

	DB.Subjects = []Subject{
		{Name: "Programming", Teacher: &DB.Users[AdminID], Group: &DB.Groups[0], CreatedOn: time.Now()},
	}
	for id := 0; id < len(DB.Subjects); id++ {
		DB.Subjects[id].ID = id
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

func SlicePutDBPath(buf []byte, dir string, name string) int {
	var n int

	n += copy(buf[n:], dir)

	buf[n] = '/'
	n++

	n += copy(buf[n:], name)

	return n
}

func OpenDBFile(dir string, name string) (int32, error) {
	buf := make([]byte, syscall.PATH_MAX)
	n := SlicePutDBPath(buf, dir, name)
	path := unsafe.String(unsafe.SliceData(buf), n)
	return syscall.Open(path, syscall.O_RDWR|syscall.O_CREAT, 0644)
}

func OpenDB(dir string) (Database, error) {
	var db Database
	var err error

	db.UsersFile, err = OpenDBFile(dir, "Users.db")
	if err != nil {
		return db, fmt.Errorf("failed to open users DB file: %w", err)
	}

	db.GroupsFile, err = OpenDBFile(dir, "Groups.db")
	if err != nil {
		return db, fmt.Errorf("failed to open groups DB file: %w", err)
	}

	db.CoursesFile, err = OpenDBFile(dir, "Courses.db")
	if err != nil {
		return db, fmt.Errorf("failed to open courses DB file: %w", err)
	}

	db.SubjectsFile, err = OpenDBFile(dir, "Subjects.db")
	if err != nil {
		return db, fmt.Errorf("failed to open subjects DB file: %w", err)
	}

	return db, nil
}

func CloseDB(db *Database) error {
	var err error

	if err1 := syscall.Close(db.UsersFile); err1 != nil {
		err = errors.Join(err, err1)
	}

	if err1 := syscall.Close(db.GroupsFile); err1 != nil {
		err = errors.Join(err, err1)
	}

	if err1 := syscall.Close(db.CoursesFile); err1 != nil {
		err = errors.Join(err, err1)
	}

	if err1 := syscall.Close(db.SubjectsFile); err1 != nil {
		err = errors.Join(err, err1)
	}

	return err
}
