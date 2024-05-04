/* TODO(anton2920): this is all temporary! */
package main

import (
	"encoding/gob"
	"os"
	"time"
)

type (
	Group struct {
		ID        int
		Name      string
		Students  []*User
		CreatedOn time.Time
	}

	Subject struct {
		ID        int
		Name      string
		Teacher   *User
		Group     *Group
		CreatedOn time.Time

		Lessons []*Lesson
	}

	User struct {
		ID        int
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
		{FirstName: "Larisa", LastName: "Sidorova", Email: "teacher@masters.com", Password: "teacher", CreatedOn: time.Now()},
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
