package main

import (
	"errors"
	"fmt"
	"time"
	"unsafe"

	"github.com/anton2920/gofa/database"
	"github.com/anton2920/gofa/syscall"
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
	users := [...]User{
		AdminID: {ID: AdminID, FirstName: "Admin", LastName: "Admin", Email: "admin@masters.com", Password: "admin", CreatedOn: time.Now().Unix(), Courses: []database.ID{0}},
		{FirstName: "Larisa", LastName: "Sidorova", Email: "teacher@masters.com", Password: "teacher", CreatedOn: time.Now().Unix(), Courses: []database.ID{1}},
		{FirstName: "Anatolii", LastName: "Ivanov", Email: "student@masters.com", Password: "student", CreatedOn: time.Now().Unix()},
		{FirstName: "Robert", LastName: "Martin", Email: "student2@masters.com", Password: "student2", CreatedOn: time.Now().Unix()},
	}

	if err := database.Drop(UsersDB); err != nil {
		return fmt.Errorf("failed to drop users data: %w", err)
	}
	for id := database.ID(0); id < database.ID(len(users)); id++ {
		if err := CreateUser(&users[id]); err != nil {
			return err
		}
	}

	groups := [...]Group{
		{Name: "18-SWE", Students: []database.ID{2, 3}, CreatedOn: time.Now().Unix()},
	}
	if err := database.Drop(GroupsDB); err != nil {
		return fmt.Errorf("failed to drop groups data: %w", err)
	}
	for id := database.ID(0); id < database.ID(len(groups)); id++ {
		if err := CreateGroup(&groups[id]); err != nil {
			return err
		}
	}

	lessons := [...]Lesson{
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
			Submissions:   []database.ID{0},
		},
	}
	*((*StepTest)(unsafe.Pointer(&lessons[0].Steps[0]))) = StepTest{
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
	*((*StepProgramming)(unsafe.Pointer(&lessons[0].Steps[1]))) = StepProgramming{
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
	if err := database.Drop(LessonsDB); err != nil {
		return fmt.Errorf("failed to drop lessons data: %w", err)
	}
	for id := database.ID(0); id < database.ID(len(lessons)); id++ {
		if err := CreateLesson(&lessons[id]); err != nil {
			return err
		}
	}

	courses := [...]Course{
		{Name: "Programming basics", Lessons: []database.ID{0}},
		{Name: "Test course", Lessons: []database.ID{1}},
	}
	if err := database.Drop(CoursesDB); err != nil {
		return fmt.Errorf("failed to drop courses data: %w", err)
	}
	for id := database.ID(0); id < database.ID(len(courses)); id++ {
		if err := CreateCourse(&courses[id]); err != nil {
			return err
		}
	}

	subjects := [...]Subject{
		{Name: "Programming", TeacherID: 0, GroupID: 0, CreatedOn: time.Now().Unix()},
		{Name: "Physics", TeacherID: 1, GroupID: 0, CreatedOn: time.Now().Unix(), Lessons: []database.ID{2}},
	}
	if err := database.Drop(SubjectsDB); err != nil {
		return fmt.Errorf("failed to drop subjects data: %w", err)
	}
	for id := database.ID(0); id < database.ID(len(subjects)); id++ {
		if err := CreateSubject(&subjects[id]); err != nil {
			return err
		}
	}

	submissions := [...]Submission{
		{LessonID: 2, UserID: 2, SubmittedSteps: make([]SubmittedStep, 2), Status: SubmissionCheckDone},
	}
	if err := database.Drop(SubmissionsDB); err != nil {
		return fmt.Errorf("failed to drop submissions data: %w", err)
	}
	for id := database.ID(0); id < database.ID(len(submissions)); id++ {
		if err := CreateSubmission(&submissions[id]); err != nil {
			return err
		}
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
