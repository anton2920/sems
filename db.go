package main

import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/anton2920/gofa/database"
	"github.com/anton2920/gofa/log"
	"github.com/anton2920/gofa/syscall"
	"github.com/anton2920/gofa/time"
)

var (
	UsersDB       *database.DB
	GroupsDB      *database.DB
	CoursesDB     *database.DB
	LessonsDB     *database.DB
	SubjectsDB    *database.DB
	SubmissionsDB *database.DB
)

const AdminID database.ID = 0

func CreateInitialDBs() error {
	user := User{ID: AdminID, FirstName: "Admin", LastName: "Admin", Email: "admin@masters.com", Password: "admin", CreatedOn: int64(time.Unix())}

	if err := CreateUser(&user); err != nil {
		return fmt.Errorf("failed to create administrator: %w", err)
	}

	return nil
}

func PutPath(buf []byte, dir string, name string) int {
	var n int

	n += copy(buf[n:], dir)

	buf[n] = '/'
	n++

	n += copy(buf[n:], name)

	return n
}

func OpenDB(dir string, name string) (*database.DB, error) {
	buf := make([]byte, syscall.PATH_MAX)
	n := PutPath(buf, dir, name)
	return database.Open(unsafe.String(unsafe.SliceData(buf), n))
}

func OpenDBs(dir string) error {
	var shouldCreate bool

	err := syscall.Mkdir(dir, 0755)
	if err != nil {
		if err.(syscall.Error).Errno != syscall.EEXIST {
			return fmt.Errorf("failed to create DB directory: %w", err)
		}
	}
	if err == nil {
		shouldCreate = true
	}

	UsersDB, err = OpenDB(dir, "Users.db")
	if err != nil {
		return fmt.Errorf("failed to open users DB file: %w", err)
	}

	GroupsDB, err = OpenDB(dir, "Groups.db")
	if err != nil {
		return fmt.Errorf("failed to open groups DB file: %w", err)
	}

	CoursesDB, err = OpenDB(dir, "Courses.db")
	if err != nil {
		return fmt.Errorf("failed to open courses DB file: %w", err)
	}

	LessonsDB, err = OpenDB(dir, "Lessons.db")
	if err != nil {
		return fmt.Errorf("failed to open lessons DB file: %w", err)
	}

	SubjectsDB, err = OpenDB(dir, "Subjects.db")
	if err != nil {
		return fmt.Errorf("failed to open subjects DB file: %w", err)
	}

	SubmissionsDB, err = OpenDB(dir, "Submissions.db")
	if err != nil {
		return fmt.Errorf("failed to open subjects DB file: %w", err)
	}

	if shouldCreate {
		log.Infof("Creating new DBs...")
		CreateInitialDBs()
	}

	return nil
}

func CloseDBs() error {
	var err error

	if err1 := database.Close(UsersDB); err1 != nil {
		err = errors.Join(err, err1)
	}

	if err1 := database.Close(GroupsDB); err1 != nil {
		err = errors.Join(err, err1)
	}

	if err1 := database.Close(CoursesDB); err1 != nil {
		err = errors.Join(err, err1)
	}

	if err1 := database.Close(LessonsDB); err1 != nil {
		err = errors.Join(err, err1)
	}

	if err1 := database.Close(SubjectsDB); err1 != nil {
		err = errors.Join(err, err1)
	}

	if err1 := database.Close(SubmissionsDB); err1 != nil {
		err = errors.Join(err, err1)
	}

	return err
}
