package main

import (
	"fmt"
	"net/mail"
	"time"
	"unicode"
	"unicode/utf8"
	"unsafe"

	"github.com/anton2920/gofa/database"
	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/strings"
	"github.com/anton2920/gofa/syscall"
)

type User struct {
	ID    database.ID
	Flags int32

	FirstName string
	LastName  string
	Email     string
	Password  string
	Courses   []database.ID
	CreatedOn int64

	Data [1024]byte
}

const (
	UserActive int32 = iota
	UserDeleted
)

const (
	MinUserNameLen = 1
	MaxUserNameLen = 45

	MinPasswordLen = 5
	MaxPasswordLen = 45
)

func UserNameValid(l Language, name string) error {
	if !strings.LengthInRange(name, MinUserNameLen, MaxUserNameLen) {
		return http.BadRequest(Ls(l, "length of the name must be between %d and %d characters"), MinUserNameLen, MaxUserNameLen)
	}

	/* Fist character must be a letter. */
	r, nbytes := utf8.DecodeRuneInString(name)
	if !unicode.IsLetter(r) {
		return http.BadRequest(Ls(l, "first character of the name must be a letter"))
	}

	/* Latter characters may include: letters, spaces, dots, hyphens and apostrophes. */
	for _, r := range name[nbytes:] {
		if (!unicode.IsLetter(r)) && (r != ' ') && (r != '.') && (r != '-') && (r != '\'') {
			return http.BadRequest(Ls(l, "second and latter characters of the name must be letters, spaces, dots, hyphens or apostrophes"))
		}
	}

	return nil
}

func GetUserByEmail(email string, user *User) error {
	users := make([]User, 32)
	var pos int64

	for {
		n, err := GetUsers(&pos, users)
		if err != nil {
			return err
		}
		if n == 0 {
			break
		}
		for i := 0; i < n; i++ {
			if (users[i].Flags != UserDeleted) && (users[i].Email == email) {
				*user = users[i]
				return nil
			}
		}
	}

	return database.NotFound
}

func CreateUser(user *User) error {
	var err error

	user.ID, err = database.IncrementNextID(UsersDB)
	if err != nil {
		return fmt.Errorf("failed to increment user ID: %w", err)
	}

	return SaveUser(user)
}

func DBUser2User(user *User) {
	data := &user.Data[0]

	user.FirstName = database.Offset2String(user.FirstName, data)
	user.LastName = database.Offset2String(user.LastName, data)
	user.Email = database.Offset2String(user.Email, data)
	user.Password = database.Offset2String(user.Password, data)
	user.Courses = database.Offset2Slice(user.Courses, data)
}

func GetUserByID(id database.ID, user *User) error {
	if err := database.Read(UsersDB, id, user); err != nil {
		return err
	}

	DBUser2User(user)
	return nil
}

func GetUsers(pos *int64, users []User) (int, error) {
	n, err := database.ReadMany(UsersDB, pos, users)
	if err != nil {
		return 0, err
	}

	for i := 0; i < n; i++ {
		DBUser2User(&users[i])
	}
	return n, nil
}

func DeleteUserByID(id database.ID) error {
	flags := UserDeleted
	var user User

	offset := int64(int(id)*int(unsafe.Sizeof(user))) + database.DataOffset + int64(unsafe.Offsetof(user.Flags))
	_, err := syscall.Pwrite(UsersDB.FD, unsafe.Slice((*byte)(unsafe.Pointer(&flags)), unsafe.Sizeof(flags)), offset)
	if err != nil {
		return fmt.Errorf("failed to delete user from DB: %w", err)
	}

	return nil
}

func SaveUser(user *User) error {
	var userDB User
	var n int

	userDB.ID = user.ID
	userDB.Flags = user.Flags

	/* TODO(anton2920): save up to a sizeof(user.Data). */
	data := unsafe.Slice(&userDB.Data[0], len(userDB.Data))
	n += database.String2DBString(&userDB.FirstName, user.FirstName, data, n)
	n += database.String2DBString(&userDB.LastName, user.LastName, data, n)
	n += database.String2DBString(&userDB.Email, user.Email, data, n)
	n += database.String2DBString(&userDB.Password, user.Password, data, n)
	n += database.Slice2DBSlice(&userDB.Courses, user.Courses, data, n)

	userDB.CreatedOn = user.CreatedOn

	return database.Write(UsersDB, userDB.ID, &userDB)
}

func UserOwnsCourse(user *User, courseID database.ID) bool {
	/* TODO(anton2920): move this out to caller? */
	if user.ID == AdminID {
		return true
	}
	for i := 0; i < len(user.Courses); i++ {
		cID := user.Courses[i]
		if cID == courseID {
			return true
		}
	}
	return false
}

func DisplayUserGroups(w *http.Response, l Language, userID database.ID) {
	groups := make([]Group, 32)
	var displayed bool
	var pos int64

	for {
		n, err := GetGroups(&pos, groups)
		if err != nil {
			/* TODO(anton2920): report error. */
		}
		if n == 0 {
			break
		}
		for i := 0; i < n; i++ {
			group := &groups[i]
			if (group.Flags == GroupDeleted) || (!UserInGroup(userID, group)) {
				continue
			}

			if !displayed {
				w.AppendString(`<h2>`)
				w.AppendString(Ls(l, "Groups"))
				w.AppendString(`</h2>`)
				w.AppendString(`<ul>`)
				displayed = true
			}

			w.AppendString(`<li>`)
			DisplayGroupLink(w, l, group)
			w.AppendString(`</li>`)
		}
	}
	if displayed {
		w.AppendString(`</ul>`)
	}
}

func DisplayUserCourses(w *http.Response, l Language, user *User) {
	var course Course

	w.AppendString(`<h2>`)
	w.AppendString(Ls(l, "Courses"))
	w.AppendString(`</h2>`)
	w.AppendString(`<ul>`)
	for i := 0; i < len(user.Courses); i++ {
		if err := GetCourseByID(user.Courses[i], &course); err != nil {
			/* TODO(anton2920): report error. */
		}
		if course.Flags == CourseDeleted {
			continue
		}

		w.AppendString(`<li>`)
		DisplayCourseLink(w, l, &course)
		w.AppendString(`</li>`)
	}
	w.AppendString(`</ul>`)
	w.AppendString(`<form method="POST" action="/course/create">`)
	DisplaySubmit(w, l, "", "Create course", true)
	w.AppendString(`</form>`)
}

func DisplayUserSubjects(w *http.Response, l Language, userID database.ID) {
	subjects := make([]Subject, 32)
	var displayed bool
	var pos int64

	for {
		n, err := GetSubjects(&pos, subjects)
		if err != nil {
			/* TODO(anton2920): report error. */
		}
		if n == 0 {
			break
		}
		for i := 0; i < n; i++ {
			subject := &subjects[i]
			if subject.Flags == SubjectDeleted {
				continue
			}

			who, err := WhoIsUserInSubject(userID, subject)
			if err != nil {
				/* TODO(anton2920): report error. */
			}
			if who == SubjectUserNone {
				continue
			}

			if !displayed {
				w.AppendString(`<h2>`)
				w.AppendString(Ls(GL, "Subjects"))
				w.AppendString(`</h2>`)
				w.AppendString(`<ul>`)
				displayed = true
			}

			w.AppendString(`<li>`)
			DisplaySubjectLink(w, l, subject)
			w.AppendString(`</li>`)
		}
	}
	if displayed {
		w.AppendString(`</ul>`)
	}
}

func DisplayUserTitle(w *http.Response, l Language, user *User) {
	w.WriteHTMLString(user.LastName)
	w.AppendString(` `)
	w.WriteHTMLString(user.FirstName)
	w.AppendString(` (ID: `)
	w.WriteID(user.ID)
	w.AppendString(`)`)
	DisplayDeleted(w, l, user.Flags == UserDeleted)
}

func DisplayUserLink(w *http.Response, l Language, user *User) {
	w.AppendString(`<a href="/user/`)
	w.WriteID(user.ID)
	w.AppendString(`">`)
	DisplayUserTitle(w, l, user)
	w.AppendString(`</a>`)
}

func UserPageHandler(w *http.Response, r *http.Request) error {
	var user User

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}

	id, err := GetIDFromURL(GL, r.URL, "/user/")
	if err != nil {
		return http.ClientError(err)
	}
	if err := GetUserByID(id, &user); err != nil {
		if err == database.NotFound {
			return http.NotFound(Ls(GL, "user with this ID does not exist"))
		}
		return http.ServerError(err)
	}
	if (session.ID != AdminID) && (session.ID != user.ID) {
		return http.ForbiddenError
	}

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>`)
	DisplayUserTitle(w, GL, &user)
	w.AppendString(`</title></head>`)

	w.AppendString(`<body>`)

	w.AppendString(`<h1>`)
	DisplayUserTitle(w, GL, &user)
	w.AppendString(`</h1>`)

	w.AppendString(`<h2>`)
	w.AppendString(Ls(GL, "Info"))
	w.AppendString(`</h2>`)

	w.AppendString(`<p>`)
	w.AppendString(Ls(GL, "Email"))
	w.AppendString(`: `)
	w.WriteHTMLString(user.Email)
	w.AppendString(`</p>`)

	w.AppendString(`<p>`)
	w.AppendString(Ls(GL, "Created on"))
	w.AppendString(`: `)
	DisplayFormattedTime(w, user.CreatedOn)
	w.AppendString(`</p>`)

	w.AppendString(`<div>`)

	w.AppendString(`<form style="display:inline" method="POST" action="/user/edit">`)
	DisplayHiddenID(w, "ID", user.ID)
	DisplayHiddenString(w, "FirstName", user.FirstName)
	DisplayHiddenString(w, "LastName", user.LastName)
	DisplayHiddenString(w, "Email", user.Email)
	DisplaySubmit(w, GL, "", "Edit", true)
	w.AppendString(`</form>`)

	if (session.ID == AdminID) && (id != AdminID) {
		w.AppendString(` <form style="display:inline" method="POST" action="/api/user/delete">`)
		DisplayHiddenID(w, "ID", user.ID)
		DisplaySubmit(w, GL, "", "Delete", true)
		w.AppendString(`</form>`)
	}
	w.AppendString(`</div>`)

	DisplayUserGroups(w, GL, user.ID)
	DisplayUserCourses(w, GL, &user)
	DisplayUserSubjects(w, GL, user.ID)

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func UserCreatePageHandler(w *http.Response, r *http.Request) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}
	if session.ID != AdminID {
		return http.ForbiddenError
	}

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>`)
	w.AppendString(Ls(GL, "Create user"))
	w.AppendString(`</title></head>`)

	w.AppendString(`<body>`)

	w.AppendString(`<h1>`)
	w.AppendString(Ls(GL, "User"))
	w.AppendString(`</h1>`)
	w.AppendString(`<h2>`)
	w.AppendString(Ls(GL, "Create"))
	w.AppendString(`</h2>`)

	DisplayErrorMessage(w, GL, r.Form.Get("Error"))

	w.AppendString(`<form method="POST" action="/api/user/create">`)

	w.AppendString(`<label>`)
	w.AppendString(Ls(GL, "First name"))
	w.AppendString(`:<br>`)
	DisplayConstraintInput(w, "text", MinNameLen, MaxNameLen, "FirstName", r.Form.Get("FirstName"), true)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>`)
	w.AppendString(Ls(GL, "Last name"))
	w.AppendString(`:<br>`)
	DisplayConstraintInput(w, "text", MinNameLen, MaxNameLen, "LastName", r.Form.Get("LastName"), true)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>`)
	w.AppendString(Ls(GL, "Email"))
	w.AppendString(`:<br>`)
	w.AppendString(`<input type="email" name="Email" value="`)
	w.WriteHTMLString(r.Form.Get("Email"))
	w.AppendString(`" required>`)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>`)
	w.AppendString(Ls(GL, "Password"))
	w.AppendString(`:<br>`)
	DisplayConstraintInput(w, "password", MinPasswordLen, MaxPasswordLen, "Password", "", true)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>`)
	w.AppendString(Ls(GL, "Repeat password"))
	w.AppendString(`:<br>`)
	DisplayConstraintInput(w, "password", MinPasswordLen, MaxPasswordLen, "RepeatPassword", "", true)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	DisplaySubmit(w, GL, "", "Create", true)
	w.AppendString(`</form>`)

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func UserEditPageHandler(w *http.Response, r *http.Request) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}

	if err := r.ParseForm(); err != nil {
		return http.ClientError(err)
	}

	userID, err := r.Form.GetID("ID")
	if err != nil {
		return http.ClientError(err)
	}
	if (session.ID != userID) && (session.ID != AdminID) {
		return http.ForbiddenError
	}

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>`)
	w.AppendString(Ls(GL, "Edit user"))
	w.AppendString(`</title></head>`)

	w.AppendString(`<body>`)

	w.AppendString(`<h1>`)
	w.AppendString(Ls(GL, "User"))
	w.AppendString(`</h1>`)
	w.AppendString(`<h2>`)
	w.AppendString(Ls(GL, "Edit"))
	w.AppendString(`</h2>`)

	DisplayErrorMessage(w, GL, r.Form.Get("Error"))

	w.AppendString(`<form method="POST" action="/api/user/edit">`)

	DisplayHiddenID(w, "ID", userID)

	w.AppendString(`<label>`)
	w.AppendString(Ls(GL, "First name"))
	w.AppendString(`:<br>`)
	DisplayConstraintInput(w, "text", MinNameLen, MaxNameLen, "FirstName", r.Form.Get("FirstName"), true)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>`)
	w.AppendString(Ls(GL, "Last name"))
	w.AppendString(`:<br>`)
	DisplayConstraintInput(w, "text", MinNameLen, MaxNameLen, "LastName", r.Form.Get("LastName"), true)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>`)
	w.AppendString(Ls(GL, "Email"))
	w.AppendString(`:<br>`)
	w.AppendString(`<input type="email" name="Email" value="`)
	w.WriteHTMLString(r.Form.Get("Email"))
	w.AppendString(`" required>`)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>`)
	w.AppendString(Ls(GL, "Password"))
	w.AppendString(`:<br>`)
	DisplayConstraintInput(w, "password", MinPasswordLen, MaxPasswordLen, "Password", "", true)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>`)
	w.AppendString(Ls(GL, "Repeat password"))
	w.AppendString(`:<br>`)
	DisplayConstraintInput(w, "password", MinPasswordLen, MaxPasswordLen, "RepeatPassword", "", true)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	DisplaySubmit(w, GL, "", "Save", true)
	w.AppendString(`</form>`)

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func UserSigninPageHandler(w *http.Response, r *http.Request) error {
	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>`)
	w.AppendString(Ls(GL, "Sign in"))
	w.AppendString(`</title></head>`)
	w.AppendString(`<body>`)

	w.AppendString(`<h1>`)
	w.AppendString(Ls(GL, "Master's degree"))
	w.AppendString(`</h1>`)
	w.AppendString(`<h2>`)
	w.AppendString(Ls(GL, "Sign in"))
	w.AppendString(`</h2>`)

	DisplayErrorMessage(w, GL, r.Form.Get("Error"))

	w.AppendString(`<form method="POST" action="/api/user/signin">`)

	w.AppendString(`<label>`)
	w.AppendString(Ls(GL, "Email"))
	w.AppendString(`:<br>`)
	w.AppendString(`<input type="email" name="Email" value="`)
	w.WriteHTMLString(r.Form.Get("Email"))
	w.AppendString(`" required>`)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>`)
	w.AppendString(Ls(GL, "Password"))
	w.AppendString(`:<br>`)
	w.AppendString(`<input type="password" name="Password" required>`)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	DisplaySubmit(w, GL, "", "Sign in", true)
	w.AppendString(`</form>`)

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func UserCreateHandler(w *http.Response, r *http.Request) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}
	if session.ID != AdminID {
		return http.ForbiddenError
	}

	if err := r.ParseForm(); err != nil {
		return http.ClientError(err)
	}

	firstName := r.Form.Get("FirstName")
	if err := UserNameValid(GL, firstName); err != nil {
		return WritePage(w, r, UserCreatePageHandler, err)
	}

	lastName := r.Form.Get("LastName")
	if err := UserNameValid(GL, lastName); err != nil {
		return WritePage(w, r, UserCreatePageHandler, err)
	}

	address, err := mail.ParseAddress(r.Form.Get("Email"))
	if err != nil {
		return WritePage(w, r, UserCreatePageHandler, http.BadRequest(Ls(GL, "provided email is not valid")))
	}
	email := address.Address

	password := r.Form.Get("Password")
	repeatPassword := r.Form.Get("RepeatPassword")
	if !strings.LengthInRange(password, MinPasswordLen, MaxPasswordLen) {
		return WritePage(w, r, UserCreatePageHandler, http.BadRequest(Ls(GL, "password length must be between %d and %d characters long"), MinPasswordLen, MaxPasswordLen))
	}
	if password != repeatPassword {
		return WritePage(w, r, UserCreatePageHandler, http.BadRequest(Ls(GL, "passwords do not match each other")))
	}

	var user User
	if err := GetUserByEmail(email, &user); err == nil {
		return WritePage(w, r, UserCreatePageHandler, http.Conflict(Ls(GL, "user with this email already exists")))
	}

	user.FirstName = firstName
	user.LastName = lastName
	user.Email = email
	user.Password = password
	user.CreatedOn = time.Now().Unix()

	if err := CreateUser(&user); err != nil {
		return http.ServerError(err)
	}

	w.Redirect("/", http.StatusSeeOther)
	return nil

}

func UserDeleteHandler(w *http.Response, r *http.Request) error {
	var user User

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}

	if err := r.ParseForm(); err != nil {
		return http.ClientError(err)
	}

	userID, err := r.Form.GetID("ID")
	if err != nil {
		return http.ClientError(err)
	}
	if err := GetUserByID(userID, &user); err != nil {
		if err == database.NotFound {
			return http.NotFound(Ls(GL, "user with this ID does not exist"))
		}
		return http.ServerError(err)
	}
	if session.ID != AdminID {
		return http.ForbiddenError
	}
	if userID == AdminID {
		return http.Conflict(Ls(GL, "cannot delete Admin user"))
	}

	/* TODO(anton2920): maybe in race with 'UserSigninHandler'. */
	RemoveAllUserSessions(userID)
	if err := DeleteUserByID(userID); err != nil {
		return http.ServerError(err)
	}
	for i := 0; i < len(user.Courses); i++ {
		_ = DeleteCourseByID(user.Courses[i])
	}

	w.Redirect("/", http.StatusSeeOther)
	return nil
}

func UserEditHandler(w *http.Response, r *http.Request) error {
	var user User

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}

	if err := r.ParseForm(); err != nil {
		return http.ClientError(err)
	}

	userID, err := r.Form.GetID("ID")
	if err != nil {
		return http.ClientError(err)
	}
	if err := GetUserByID(userID, &user); err != nil {
		if err == database.NotFound {
			return http.NotFound(Ls(GL, "user with this ID does not exist"))
		}
		return http.ServerError(err)
	}
	if (session.ID != userID) && (session.ID != AdminID) {
		return http.ForbiddenError
	}

	firstName := r.Form.Get("FirstName")
	if err := UserNameValid(GL, firstName); err != nil {
		return WritePage(w, r, UserEditPageHandler, err)
	}

	lastName := r.Form.Get("LastName")
	if err := UserNameValid(GL, lastName); err != nil {
		return WritePage(w, r, UserEditPageHandler, err)
	}

	address, err := mail.ParseAddress(r.Form.Get("Email"))
	if err != nil {
		return WritePage(w, r, UserEditPageHandler, http.BadRequest(Ls(GL, "provided email is not valid")))
	}
	email := address.Address

	password := r.Form.Get("Password")
	repeatPassword := r.Form.Get("RepeatPassword")
	if !strings.LengthInRange(password, MinPasswordLen, MaxPasswordLen) {
		return WritePage(w, r, UserEditPageHandler, http.BadRequest(Ls(GL, "password length must be between %d and %d characters long"), MinPasswordLen, MaxPasswordLen))
	}
	if password != repeatPassword {
		return WritePage(w, r, UserEditPageHandler, http.BadRequest(Ls(GL, "passwords do not match each other")))
	}

	var user2 User
	if err := GetUserByEmail(email, &user2); (err == nil) && (user2.ID != userID) {
		return WritePage(w, r, UserEditPageHandler, http.Conflict(Ls(GL, "user with this email already exists")))
	}

	user.FirstName = firstName
	user.LastName = lastName
	user.Email = email
	user.Password = password

	if err := SaveUser(&user); err != nil {
		return http.ServerError(err)
	}

	w.RedirectID("/user/", userID, http.StatusSeeOther)
	return nil
}

func UserSigninHandler(w *http.Response, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return http.ClientError(err)
	}

	address, err := mail.ParseAddress(r.Form.Get("Email"))
	if err != nil {
		return WritePage(w, r, UserSigninPageHandler, http.BadRequest(Ls(GL, "provided email is not valid")))
	}
	email := address.Address

	var user User
	if err := GetUserByEmail(email, &user); err != nil {
		if err == database.NotFound {
			return WritePage(w, r, UserSigninPageHandler, http.NotFound(Ls(GL, "user with this email does not exist")))
		}
		return http.ServerError(err)
	}

	password := r.Form.Get("Password")
	if user.Password != password {
		return WritePage(w, r, UserSigninPageHandler, http.Conflict(Ls(GL, "provided password is incorrect")))
	}

	token, err := GenerateSessionToken()
	if err != nil {
		return http.ServerError(err)
	}
	expiry := time.Now().Add(OneWeek)

	SessionsLock.Lock()
	Sessions[token] = &Session{
		ID:     user.ID,
		Expiry: expiry,
	}
	SessionsLock.Unlock()

	if Debug {
		w.SetCookieUnsafe("Token", token, int(expiry.Unix()))
	} else {
		w.SetCookie("Token", token, int(expiry.Unix()))
	}
	w.Redirect("/", http.StatusSeeOther)
	return nil
}

func UserSignoutHandler(w *http.Response, r *http.Request) error {
	token := r.Cookie("Token")
	if token == "" {
		return http.UnauthorizedError
	}

	if _, err := GetSessionFromToken(token); err != nil {
		return http.UnauthorizedError
	}

	SessionsLock.Lock()
	delete(Sessions, token)
	SessionsLock.Unlock()

	w.DelCookie("Token")
	w.Redirect("/", http.StatusSeeOther)
	return nil
}
