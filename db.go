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

const (
	Version int32 = 0x0

	VersionOffset = 0
	NextIDOffset  = VersionOffset + 4
	DataOffset    = NextIDOffset + 4

	DataStartOffset = 1
)

const AdminID = 0

const DBFile = "db.gob"

var DB struct {
	Lessons     []Lesson
	Submissions []Submission
}

var DBNotFound = errors.New("not found")

func Offset2String(s string, base *byte) string {
	return unsafe.String((*byte)(unsafe.Add(unsafe.Pointer(base), uintptr(unsafe.Pointer(unsafe.StringData(s))))), len(s))
}

func Offset2Slice[T any](s []T, base *byte) []T {
	return unsafe.Slice((*T)(unsafe.Add(unsafe.Pointer(base), uintptr(unsafe.Pointer(unsafe.SliceData(s))))), len(s))
}

func String2Offset(s string, offset int) string {
	return unsafe.String((*byte)(unsafe.Pointer(uintptr(offset))), len(s))
}

func Slice2Offset[T any](s []T, offset int) []T {
	return unsafe.Slice((*T)(unsafe.Pointer(uintptr(offset))), len(s))
}

func CreateInitialDB() error {
	users := [...]User{
		AdminID: {ID: AdminID, FirstName: "Admin", LastName: "Admin", Email: "admin@masters.com", Password: "admin", CreatedOn: time.Now().Unix(), Courses: []int32{0}},
		{FirstName: "Larisa", LastName: "Sidorova", Email: "teacher@masters.com", Password: "teacher", CreatedOn: time.Now().Unix(), Courses: []int32{1}},
		{FirstName: "Anatolii", LastName: "Ivanov", Email: "student@masters.com", Password: "student", CreatedOn: time.Now().Unix()},
		{FirstName: "Robert", LastName: "Martin", Email: "student2@masters.com", Password: "student2", CreatedOn: time.Now().Unix()},
	}

	if err := DropData(DB2.UsersFile); err != nil {
		return fmt.Errorf("failed to drop users data: %w", err)
	}
	for id := int32(0); id < int32(len(users)); id++ {
		if err := CreateUser(DB2, &users[id]); err != nil {
			return err
		}
	}

	groups := [...]Group{
		{Name: "18-SWE", Students: []int32{2, 3}, CreatedOn: time.Now().Unix()},
	}
	if err := DropData(DB2.GroupsFile); err != nil {
		return fmt.Errorf("failed to drop groups data: %w", err)
	}
	for id := int32(0); id < int32(len(groups)); id++ {
		if err := CreateGroup(DB2, &groups[id]); err != nil {
			return err
		}
	}

	DB.Lessons = []Lesson{
		Lesson{
			ContainerID:   0,
			ContainerType: ContainerTypeCourse,
			Name:          "Introduction",
			Theory:        "This is an introduction.",
			Steps:         make([]Step, 2),
		},
		Lesson{
			ContainerID:   1,
			ContainerType: ContainerTypeCourse,
			Name:          "Test lesson",
			Theory:        "This is a test lesson.",
		},
		Lesson{
			ContainerID:   1,
			ContainerType: ContainerTypeSubject,
			Name:          "New lesson",
			Theory:        "New theory",
			Steps:         make([]Step, 2),
			Submissions:   []int32{0},
		},
	}
	for id := int32(0); id < int32(len(DB.Lessons)); id++ {
		DB.Lessons[id].ID = id
	}
	*((*StepTest)(unsafe.Pointer(&DB.Lessons[0].Steps[0]))) = StepTest{
		StepCommon: StepCommon{Name: "Back-end development basics", Type: StepTypeTest},
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
	}
	*((*StepProgramming)(unsafe.Pointer(&DB.Lessons[0].Steps[1]))) = StepProgramming{
		StepCommon:  StepCommon{Name: "Hello, world", Type: StepTypeProgramming},
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
	}

	courses := [...]Course{
		{Name: "Programming basics", Lessons: []int32{0}},
		{Name: "Test course", Lessons: []int32{1}},
	}
	if err := DropData(DB2.CoursesFile); err != nil {
		return fmt.Errorf("failed to drop courses data: %w", err)
	}
	for id := int32(0); id < int32(len(courses)); id++ {
		if err := CreateCourse(DB2, &courses[id]); err != nil {
			return err
		}
	}

	subjects := [...]Subject{
		{Name: "Programming", TeacherID: 0, GroupID: 0, CreatedOn: time.Now().Unix()},
		{Name: "Physics", TeacherID: 1, GroupID: 0, CreatedOn: time.Now().Unix(), Lessons: []int32{2}},
	}
	if err := DropData(DB2.SubjectsFile); err != nil {
		return fmt.Errorf("failed to drop subjects data: %w", err)
	}
	for id := int32(0); id < int32(len(subjects)); id++ {
		if err := CreateSubject(DB2, &subjects[id]); err != nil {
			return err
		}
	}

	DB.Submissions = make([]Submission, 0, 1024)
	DB.Submissions = append(DB.Submissions, Submission{Steps: make([]Step, 2), LessonID: 2, UserID: 2, SubmittedSteps: make([]interface{}, 2), Status: SubmissionCheckDone})
	for i := int32(0); i < int32(len(DB.Submissions)); i++ {
		DB.Submissions[i].ID = i
	}

	return nil
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

	fd, err := syscall.Open(path, syscall.O_RDWR|syscall.O_CREAT, 0644)
	if err != nil {
		return -1, err
	}

	var version int32

	n, err = syscall.Pread(fd, unsafe.Slice((*byte)(unsafe.Pointer(&version)), unsafe.Sizeof(version)), VersionOffset)
	if err != nil {
		syscall.Close(fd)
		return -1, err
	}
	if n < int(unsafe.Sizeof(version)) {
		version = Version

		_, err := syscall.Pwrite(fd, unsafe.Slice((*byte)(unsafe.Pointer(&version)), unsafe.Sizeof(version)), VersionOffset)
		if err != nil {
			syscall.Close(fd)
			return -1, err
		}
	} else if version != Version {
		syscall.Close(fd)
		return -1, fmt.Errorf("incompatible DB file version %d, expected %d", version, Version)
	}

	return fd, nil
}

func OpenDB(dir string) (*Database, error) {
	var err error

	if err := syscall.Mkdir(dir, 0755); err != nil {
		if err.(syscall.Error).Errno != syscall.EEXIST {
			return nil, fmt.Errorf("failed to create DB directory: %w", err)
		}
	}

	db := new(Database)

	db.UsersFile, err = OpenDBFile(dir, "Users.db")
	if err != nil {
		return nil, fmt.Errorf("failed to open users DB file: %w", err)
	}

	db.GroupsFile, err = OpenDBFile(dir, "Groups.db")
	if err != nil {
		return nil, fmt.Errorf("failed to open groups DB file: %w", err)
	}

	db.CoursesFile, err = OpenDBFile(dir, "Courses.db")
	if err != nil {
		return nil, fmt.Errorf("failed to open courses DB file: %w", err)
	}

	db.SubjectsFile, err = OpenDBFile(dir, "Subjects.db")
	if err != nil {
		return nil, fmt.Errorf("failed to open subjects DB file: %w", err)
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

func GetNextID(fd int32) (int32, error) {
	var id int32

	_, err := syscall.Pread(fd, unsafe.Slice((*byte)(unsafe.Pointer(&id)), unsafe.Sizeof(id)), NextIDOffset)
	if err != nil {
		return -1, fmt.Errorf("failed to read next ID: %w", err)
	}

	return id, nil
}

func SetNextID(fd int32, id int32) error {
	_, err := syscall.Pwrite(fd, unsafe.Slice((*byte)(unsafe.Pointer(&id)), unsafe.Sizeof(id)), NextIDOffset)
	if err != nil {
		return fmt.Errorf("failed to write next ID in users DB: %w", err)
	}
	return nil
}

/* TODO(anton2920): make that atomic. */
func IncrementNextID(fd int32) (int32, error) {
	id, err := GetNextID(fd)
	if err != nil {
		return -1, err
	}
	if err := SetNextID(fd, id+1); err != nil {
		return -1, err
	}
	return id, nil
}

func DropData(fd int32) error {
	if err := syscall.Ftruncate(fd, DataOffset); err != nil {
		return fmt.Errorf("failed to truncate courses file: %w", err)
	}
	if err := SetNextID(fd, 0); err != nil {
		return fmt.Errorf("failed to set next course ID: %w", err)
	}
	return nil
}
