package main

import (
	"fmt"
	"strconv"
	"unsafe"

	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/net/url"
	"github.com/anton2920/gofa/strings"
	"github.com/anton2920/gofa/syscall"
)

type Course struct {
	ID    int32
	Flags int32

	Name    string
	Lessons []int32

	Data [1024]byte
}

const (
	CourseActive  int32 = 0
	CourseDeleted       = 1
	CourseDraft         = 2
)

const (
	MinNameLen = 1
	MaxNameLen = 45
)

func CreateCourse(db *Database, course *Course) error {
	var err error

	course.ID, err = IncrementNextID(db.CoursesFile)
	if err != nil {
		return fmt.Errorf("failed to increment course ID: %w", err)
	}

	return SaveCourse(db, course)
}

func DBCourse2Course(course *Course) {
	data := &course.Data[0]

	course.Name = Offset2String(course.Name, data)
	course.Lessons = Offset2Slice(course.Lessons, data)
}

func GetCourseByID(db *Database, id int32, course *Course) error {
	size := int(unsafe.Sizeof(*course))
	offset := int64(int(id)*size) + DataOffset

	n, err := syscall.Pread(db.CoursesFile, unsafe.Slice((*byte)(unsafe.Pointer(course)), size), offset)
	if err != nil {
		return fmt.Errorf("failed to read course from DB: %w", err)
	}
	if n < size {
		return DBNotFound
	}

	DBCourse2Course(course)
	return nil
}

func GetCourses(db *Database, pos *int64, courses []Course) (int, error) {
	if *pos < DataOffset {
		*pos = DataOffset
	}
	size := int(unsafe.Sizeof(courses[0]))

	n, err := syscall.Pread(db.CoursesFile, unsafe.Slice((*byte)(unsafe.Pointer(unsafe.SliceData(courses))), len(courses)*size), *pos)
	if err != nil {
		return 0, fmt.Errorf("failed to read course from DB: %w", err)
	}
	*pos += int64(n)

	n /= size
	for i := 0; i < n; i++ {
		DBCourse2Course(&courses[i])
	}

	return n, nil
}

func DeleteCourseByID(db *Database, id int32) error {
	flags := CourseDeleted
	var course Course

	offset := int64(int(id)*int(unsafe.Sizeof(course))) + DataOffset + int64(unsafe.Offsetof(course.Flags))
	_, err := syscall.Pwrite(db.CoursesFile, unsafe.Slice((*byte)(unsafe.Pointer(&flags)), unsafe.Sizeof(flags)), offset)
	if err != nil {
		return fmt.Errorf("failed to delete course from DB: %w", err)
	}

	return nil
}

func SaveCourse(db *Database, course *Course) error {
	var courseDB Course
	var n int

	size := int(unsafe.Sizeof(*course))
	offset := int64(int(course.ID)*size) + DataOffset
	data := unsafe.Slice(&courseDB.Data[0], len(courseDB.Data))

	courseDB.ID = course.ID
	courseDB.Flags = course.Flags

	/* TODO(anton2920): save up to a sizeof(course.Data). */
	n += String2DBString(&courseDB.Name, course.Name, data, n)
	n += Slice2DBSlice(&courseDB.Lessons, course.Lessons, data, n)

	_, err := syscall.Pwrite(db.CoursesFile, unsafe.Slice((*byte)(unsafe.Pointer(&courseDB)), size), offset)
	if err != nil {
		return fmt.Errorf("failed to write course to DB: %w", err)
	}

	return nil
}

func DisplayCourseLink(w *http.Response, course *Course) {
	w.AppendString(`<a href="/course/`)
	w.WriteInt(int(course.ID))
	w.AppendString(`">`)
	w.WriteHTMLString(course.Name)
	if course.Flags == CourseDraft {
		w.AppendString(` (draft)`)
	}
	w.AppendString(`</a>`)
}

func CoursePageHandler(w *http.Response, r *http.Request) error {
	var course Course
	var user User

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}
	if err := GetUserByID(DB2, session.ID, &user); err != nil {
		return http.ServerError(err)
	}

	id, err := GetIDFromURL(r.URL, "/course/")
	if err != nil {
		return http.ClientError(err)
	}
	if !UserOwnsCourse(&user, int32(id)) {
		return http.ForbiddenError
	}
	if err := GetCourseByID(DB2, int32(id), &course); err != nil {
		if err == DBNotFound {
			return http.NotFound("course with this ID does not exist")
		}
		return http.ServerError(err)
	}

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>`)
	w.WriteHTMLString(course.Name)
	if course.Flags == CourseDraft {
		w.AppendString(` (draft)`)
	}
	w.AppendString(`</title></head>`)
	w.AppendString(`<body>`)

	w.AppendString(`<h1>`)
	w.WriteHTMLString(course.Name)
	if course.Flags == CourseDraft {
		w.AppendString(` (draft)`)
	}
	w.AppendString(`</h1>`)

	w.AppendString(`<h2>Lessons</h2>`)
	for i := 0; i < len(course.Lessons); i++ {
		var lesson Lesson
		if err := GetLessonByID(DB2, course.Lessons[i], &lesson); err != nil {
			return http.ServerError(err)
		}

		w.AppendString(`<fieldset>`)

		w.AppendString(`<legend>Lesson #`)
		w.WriteInt(i + 1)
		if lesson.Flags == LessonDraft {
			w.AppendString(` (draft)`)
		}
		w.AppendString(`</legend>`)

		w.AppendString(`<p>Name: `)
		w.WriteHTMLString(lesson.Name)
		w.AppendString(`</p>`)

		w.AppendString(`<p>Theory: `)
		DisplayShortenedString(w, lesson.Theory, LessonTheoryMaxDisplayLen)
		w.AppendString(`</p>`)

		DisplayLessonLink(w, &lesson)

		w.AppendString(`</fieldset>`)
		w.AppendString(`<br>`)
	}

	w.AppendString(`<div>`)

	w.AppendString(`<form style="display:inline" method="POST" action="/course/edit">`)
	w.AppendString(`<input type="hidden" name="ID" value="`)
	w.WriteHTMLString(r.URL.Path[len("/course/"):])
	w.AppendString(`">`)
	w.AppendString(`<input type="submit" value="Edit">`)
	w.AppendString(`</form> `)

	w.AppendString(`<form style="display:inline" method="POST" action="/api/course/delete">`)
	w.AppendString(`<input type="hidden" name="ID" value="`)
	w.WriteHTMLString(r.URL.Path[len("/course/"):])
	w.AppendString(`">`)
	w.AppendString(`<input type="submit" value="Delete">`)
	w.AppendString(`</form>`)

	w.AppendString(`</div>`)

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func CourseFillFromRequest(vs url.Values, course *Course) {
	course.Name = vs.Get("Name")
}

func CourseVerify(course *Course) error {
	var lesson Lesson

	if !strings.LengthInRange(course.Name, MinNameLen, MaxNameLen) {
		return http.BadRequest("course name length must be between %d and %d characters long", MinNameLen, MaxNameLen)
	}

	if len(course.Lessons) == 0 {
		return http.BadRequest("create at least one lesson")
	}
	for li := 0; li < len(course.Lessons); li++ {
		if err := GetLessonByID(DB2, course.Lessons[li], &lesson); err != nil {
			return http.ServerError(err)
		}
		if lesson.Flags == LessonDraft {
			return http.BadRequest("lesson %d is a draft", li+1)
		}
	}

	return nil
}

func CourseCreateEditCoursePageHandler(w *http.Response, r *http.Request, course *Course) error {
	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>Course</title></head>`)
	w.AppendString(`<body>`)

	w.AppendString(`<h1>Course</h1>`)

	DisplayErrorMessage(w, r.Form.Get("Error"))

	w.AppendString(`<form method="POST" action="`)
	w.WriteString(r.URL.Path)
	w.AppendString(`">`)

	w.AppendString(`<input type="hidden" name="CurrentPage" value="Course">`)

	w.AppendString(`<input type="hidden" name="ID" value="`)
	w.WriteHTMLString(r.Form.Get("ID"))
	w.AppendString(`">`)

	w.AppendString(`<label>Name: `)
	DisplayConstraintInput(w, "text", MinNameLen, MaxNameLen, "Name", course.Name, true)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	DisplayLessonsEditableList(w, course.Lessons)

	w.AppendString(`<input type="submit" name="NextPage" value="Add lesson" formnovalidate>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<input type="submit" name="NextPage" value="Save">`)

	w.AppendString(`</form>`)

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func CourseCreateEditHandleCommand(w *http.Response, r *http.Request, course *Course, currentPage, k, command string) error {
	var lesson Lesson

	pindex, spindex, _, _, err := GetIndicies(k[len("Command"):])
	if err != nil {
		return http.ClientError(err)
	}

	switch currentPage {
	default:
		return LessonAddHandleCommand(w, r, course.Lessons, currentPage, k, command)
	case "Course":
		switch command {
		case "Delete":
			course.Lessons = RemoveAtIndex(course.Lessons, pindex)
		case "Edit":
			if (pindex < 0) || (pindex >= len(course.Lessons)) {
				return http.ClientError(nil)
			}
			if err := GetLessonByID(DB2, course.Lessons[pindex], &lesson); err != nil {
				return http.ServerError(err)
			}
			lesson.Flags = LessonDraft
			if err := SaveLesson(DB2, &lesson); err != nil {
				return http.ServerError(err)
			}

			r.Form.Set("LessonIndex", spindex)
			return LessonAddPageHandler(w, r, &lesson)
		case "↑", "^|":
			MoveUp(course.Lessons, pindex)
		case "↓", "|v":
			MoveDown(course.Lessons, pindex)
		}

		return CourseCreateEditCoursePageHandler(w, r, course)
	}
}

func CourseCreateEditPageHandler(w *http.Response, r *http.Request) error {
	var course Course
	var lesson Lesson
	var user User

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}
	if err := GetUserByID(DB2, session.ID, &user); err != nil {
		return http.ServerError(err)
	}

	if err := r.ParseForm(); err != nil {
		return http.ClientError(err)
	}

	currentPage := r.Form.Get("CurrentPage")
	nextPage := r.Form.Get("NextPage")

	id := r.Form.Get("ID")
	if id == "" {
		if err := CreateCourse(DB2, &course); err != nil {
			return http.ServerError(err)
		}

		user.Courses = append(user.Courses, course.ID)
		if err := SaveUser(DB2, &user); err != nil {
			return http.ServerError(err)
		}

		r.Form.SetInt("ID", int(course.ID))
	} else {
		courseID, err := strconv.Atoi(id)
		if err != nil {
			return http.ClientError(err)
		}
		if !UserOwnsCourse(&user, int32(courseID)) {
			return http.ForbiddenError
		}
		if err := GetCourseByID(DB2, int32(courseID), &course); err != nil {
			if err == DBNotFound {
				return http.NotFound("course with this ID does not exist")
			}
			return http.ServerError(err)
		}
	}
	defer SaveCourse(DB2, &course)

	course.Flags = CourseDraft

	for i := 0; i < len(r.Form); i++ {
		k := r.Form[i].Key
		if len(r.Form[i].Values) == 0 {
			continue
		}
		v := r.Form[i].Values[0]

		/* 'command' is button, which modifies content of a current page. */
		if strings.StartsWith(k, "Command") {
			/* NOTE(anton2920): after command is executed, function must return. */
			return CourseCreateEditHandleCommand(w, r, &course, currentPage, k, v)
		}
	}

	/* 'currentPage' is the page to save/check before leaving it. */
	switch currentPage {
	case "Course":
		CourseFillFromRequest(r.Form, &course)
	case "Lesson":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), len(course.Lessons))
		if err != nil {
			return http.ClientError(err)
		}
		if err := GetLessonByID(DB2, course.Lessons[li], &lesson); err != nil {
			return http.ServerError(err)
		}
		defer SaveLesson(DB2, &lesson)

		LessonFillFromRequest(r.Form, &lesson)
	case "Test":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), len(course.Lessons))
		if err != nil {
			return http.ClientError(err)
		}
		if err := GetLessonByID(DB2, course.Lessons[li], &lesson); err != nil {
			return http.ServerError(err)
		}
		defer SaveLesson(DB2, &lesson)

		si, err := GetValidIndex(r.Form.Get("StepIndex"), len(lesson.Steps))
		if err != nil {
			return http.ClientError(err)
		}
		test, err := Step2Test(&lesson.Steps[si])
		if err != nil {
			return http.ClientError(err)
		}

		if err := LessonTestFillFromRequest(r.Form, test); err != nil {
			return WritePageEx(w, r, LessonAddTestPageHandler, test, err)
		}
		if err := LessonTestVerify(test); err != nil {
			return WritePageEx(w, r, LessonAddTestPageHandler, test, err)
		}
	case "Programming":
		li, err := GetValidIndex(r.Form.Get("LessonIndex"), len(course.Lessons))
		if err != nil {
			return http.ClientError(err)
		}
		if err := GetLessonByID(DB2, course.Lessons[li], &lesson); err != nil {
			return http.ServerError(err)
		}
		defer SaveLesson(DB2, &lesson)

		si, err := GetValidIndex(r.Form.Get("StepIndex"), len(lesson.Steps))
		if err != nil {
			return http.ClientError(err)
		}
		task, err := Step2Programming(&lesson.Steps[si])
		if err != nil {
			return http.ClientError(err)
		}

		if err := LessonProgrammingFillFromRequest(r.Form, task); err != nil {
			return WritePageEx(w, r, LessonAddProgrammingPageHandler, task, err)
		}
		if err := LessonProgrammingVerify(task); err != nil {
			return WritePageEx(w, r, LessonAddProgrammingPageHandler, task, err)
		}
	}

	switch nextPage {
	default:
		return CourseCreateEditCoursePageHandler(w, r, &course)
	case "Next":
		if err := LessonVerify(&lesson); err != nil {
			return WritePageEx(w, r, LessonAddPageHandler, &lesson, err)
		}
		lesson.Flags = LessonActive
		if err := SaveLesson(DB2, &lesson); err != nil {
			return http.ServerError(err)
		}

		return CourseCreateEditCoursePageHandler(w, r, &course)
	case "Add lesson":
		lesson.Flags = LessonDraft
		lesson.ContainerID = course.ID
		lesson.ContainerType = ContainerTypeCourse

		if err := CreateLesson(DB2, &lesson); err != nil {
			return http.ServerError(err)
		}

		course.Lessons = append(course.Lessons, lesson.ID)
		r.Form.SetInt("LessonIndex", len(course.Lessons)-1)

		return LessonAddPageHandler(w, r, &lesson)
	case "Continue":
		si, err := GetValidIndex(r.Form.Get("StepIndex"), len(lesson.Steps))
		if err != nil {
			return http.ClientError(err)
		}
		step := &lesson.Steps[si]
		step.Draft = false

		return LessonAddPageHandler(w, r, &lesson)
	case "Add test":
		lesson.Flags = LessonDraft

		lesson.Steps = append(lesson.Steps, Step{StepCommon: StepCommon{Type: StepTypeTest, Draft: true}})
		test, _ := Step2Test(&lesson.Steps[len(lesson.Steps)-1])

		r.Form.SetInt("StepIndex", len(lesson.Steps)-1)
		return LessonAddTestPageHandler(w, r, test)
	case "Add programming task":
		lesson.Flags = LessonDraft

		lesson.Steps = append(lesson.Steps, Step{StepCommon: StepCommon{Type: StepTypeProgramming, Draft: true}})
		task, _ := Step2Programming(&lesson.Steps[len(lesson.Steps)-1])

		r.Form.SetInt("StepIndex", len(lesson.Steps)-1)
		return LessonAddProgrammingPageHandler(w, r, task)
	case "Save":
		if err := CourseVerify(&course); err != nil {
			return WritePageEx(w, r, CourseCreateEditCoursePageHandler, &course, err)
		}
		course.Flags = CourseActive

		w.RedirectID("/course/", int(course.ID), http.StatusSeeOther)
		return nil
	}
}
