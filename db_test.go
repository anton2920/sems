package main

import (
	"fmt"
	"unsafe"

	"github.com/anton2920/gofa/database"
	"github.com/anton2920/gofa/time"
)

func testCreateInitialDBs() error {
	users := [...]User{
		AdminID: {ID: AdminID, FirstName: "Admin", LastName: "Admin", Email: "admin@masters.com", Password: "admin", CreatedOn: int64(time.Unix()), Courses: []database.ID{0}},
		{FirstName: "Larisa", LastName: "Sidorova", Email: "teacher@masters.com", Password: "teacher", CreatedOn: int64(time.Unix()), Courses: []database.ID{1}},
		{FirstName: "Anatolii", LastName: "Ivanov", Email: "student@masters.com", Password: "student", CreatedOn: int64(time.Unix())},
		{FirstName: "Robert", LastName: "Martin", Email: "student2@masters.com", Password: "student2", CreatedOn: int64(time.Unix())},
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
		{Name: "18-SWE", Students: []database.ID{2, 3}, CreatedOn: int64(time.Unix())},
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
			ContainerType: LessonContainerCourse,
			Name:          "Introduction",
			Theory:        "This is an introduction.",
			Steps:         make([]Step, 2),
		},
		Lesson{
			ContainerID:   1,
			ContainerType: LessonContainerCourse,
			Name:          "Test lesson",
			Theory:        "This is a test lesson.",
		},
		Lesson{
			ContainerID:   1,
			ContainerType: LessonContainerSubject,
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
		{LessonContainer{Name: "Programming basics", Lessons: []database.ID{0}}, [1024]byte{}},
		{LessonContainer{Name: "Test course", Lessons: []database.ID{1}}, [1024]byte{}},
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
		{LessonContainer{Name: "Programming"}, 0, 0, int64(time.Unix()), [1024]byte{}},
		{LessonContainer{Name: "Physics", Lessons: []database.ID{2}}, 1, 0, int64(time.Unix()), [1024]byte{}},
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
