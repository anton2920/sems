package main

import (
	"fmt"
	"net/mail"
	"time"
	"unicode"
	"unicode/utf8"
	"unsafe"

	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/strings"
	"github.com/anton2920/gofa/syscall"
)

type User struct {
	ID        int
	FirstName string
	LastName  string
	Email     string
	Password  string
	CreatedOn time.Time

	Courses []*Course
}

type User2 struct {
	ID        int
	FirstName string
	LastName  string
	Email     string
	Password  string
	Courses   []int
	CreatedOn int64

	Data [1024]byte
}

const (
	MinUserNameLen = 1
	MaxUserNameLen = 45

	MinPasswordLen = 5
	MaxPasswordLen = 45
)

func UserNameValid(name string) error {
	if !strings.LengthInRange(name, MinUserNameLen, MaxUserNameLen) {
		return http.BadRequest("length of the name must be between %d and %d characters", MinUserNameLen, MaxUserNameLen)
	}

	/* Fist character must be a letter. */
	r, nbytes := utf8.DecodeRuneInString(name)
	if !unicode.IsLetter(r) {
		return http.BadRequest("first character of the name must be a letter")
	}

	/* Latter characters may include: letters, spaces, dots, hyphens and apostrophes. */
	for _, r := range name[nbytes:] {
		if (!unicode.IsLetter(r)) && (r != ' ') && (r != '.') && (r != '-') && (r != '\'') {
			return http.BadRequest("second and latter characters of the name must be letters, spaces, dots, hyphens or apostrophes")
		}
	}

	return nil
}

func GetUserByEmail(email string) *User {
	for i := 0; i < len(DB.Users); i++ {
		user := &DB.Users[i]

		if user.Email == email {
			return user
		}
	}

	return nil
}

func GetUserByID(db *Database, id int, user *User2) error {
	size := int(unsafe.Sizeof(*user))
	offset := int64(id * size)

	n, err := syscall.Pread(db.UsersFile, unsafe.Slice((*byte)(unsafe.Pointer(user)), size), offset)
	if (err != nil) || (n < size) {
		return fmt.Errorf("failed to read user from DB: %w", err)
	}

	user.FirstName = Offset2String(user.FirstName, &user.Data[0])
	user.LastName = Offset2String(user.LastName, &user.Data[0])
	user.Email = Offset2String(user.Email, &user.Data[0])
	user.Password = Offset2String(user.Password, &user.Data[0])
	user.Courses = Offset2Slice(user.Courses, &user.Data[0])

	return nil
}

func SaveUser(db *Database, user User2) error {
	size := int(unsafe.Sizeof(user))
	offset := int64(user.ID * size)

	var n int

	/* TODO(anton2920): saving up to a sizeof(user.Data). */
	n += copy(user.Data[n:], user.FirstName)
	user.FirstName = String2Offset(user.FirstName, n)

	n += copy(user.Data[n:], user.LastName)
	user.LastName = String2Offset(user.LastName, n)

	n += copy(user.Data[n:], user.Email)
	user.Email = String2Offset(user.Email, n)

	n += copy(user.Data[n:], user.Password)
	user.Password = String2Offset(user.Password, n)

	n += copy(user.Data[n:], unsafe.Slice((*byte)(unsafe.Pointer(&user.Courses[0])), len(user.Courses)*int(unsafe.Sizeof(n))))
	user.Courses = Slice2Offset(user.Courses, n)

	n, err := syscall.Pwrite(db.UsersFile, unsafe.Slice((*byte)(unsafe.Pointer(&user)), size), offset)
	if (err != nil) || (n < size) {
		return fmt.Errorf("failed to write user to DB: %w", err)
	}

	return nil
}

func DisplayUserLink(w *http.Response, user *User) {
	w.AppendString(`<a href="/user/`)
	w.WriteInt(user.ID)
	w.AppendString(`">`)
	w.WriteHTMLString(user.LastName)
	w.AppendString(` `)
	w.WriteHTMLString(user.FirstName)
	w.AppendString(` (ID: `)
	w.WriteInt(user.ID)
	w.AppendString(`)`)
	w.AppendString(`</a>`)
}

func UserPageHandler(w *http.Response, r *http.Request) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}

	id, err := GetIDFromURL(r.URL, "/user/")
	if err != nil {
		return http.ClientError(err)
	}
	if (id < 0) || (id >= len(DB.Users)) {
		return http.NotFound("user with this ID does not exist")
	}
	user := &DB.Users[id]

	if (session.ID != AdminID) && (session.ID != user.ID) {
		return http.ForbiddenError
	}

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>`)
	w.WriteHTMLString(user.LastName)
	w.AppendString(` `)
	w.WriteHTMLString(user.FirstName)
	w.AppendString(`</title></head>`)

	w.AppendString(`<body>`)

	w.AppendString(`<h1>`)
	w.WriteHTMLString(user.LastName)
	w.AppendString(` `)
	w.WriteHTMLString(user.FirstName)
	w.AppendString(`</h1>`)

	w.AppendString(`<h2>Info</h2>`)

	w.AppendString(`<p>ID: `)
	w.WriteInt(user.ID)
	w.AppendString(`</p>`)

	w.AppendString(`<p>Email: `)
	w.WriteHTMLString(user.Email)
	w.AppendString(`</p>`)

	w.AppendString(`<p>Created on: `)
	DisplayFormattedTime(w, user.CreatedOn)
	w.AppendString(`</p>`)

	if (session.ID == id) || (session.ID == AdminID) {
		w.AppendString(`<form method="POST" action="/user/edit">`)

		w.AppendString(`<input type="hidden" name="ID" value="`)
		w.WriteString(r.URL.Path[len("/user/"):])
		w.AppendString(`">`)

		w.AppendString(`<input type="hidden" name="FirstName" value="`)
		w.WriteHTMLString(user.FirstName)
		w.AppendString(`">`)

		w.AppendString(`<input type="hidden" name="LastName" value="`)
		w.WriteHTMLString(user.LastName)
		w.AppendString(`">`)

		w.AppendString(`<input type="hidden" name="Email" value="`)
		w.WriteHTMLString(user.Email)
		w.AppendString(`">`)

		w.AppendString(`<input type="submit" value="Edit">`)

		w.AppendString(`</form>`)
	}

	var displayGroups bool
	for i := 0; i < len(DB.Groups); i++ {
		group := &DB.Groups[i]

		if UserInGroup(user.ID, group) {
			displayGroups = true
			break
		}
	}
	if displayGroups {
		w.AppendString(`<h2>Groups</h2>`)
		w.AppendString(`<ul>`)
		for i := 0; i < len(DB.Groups); i++ {
			group := &DB.Groups[i]

			if UserInGroup(user.ID, group) {
				w.AppendString(`<li>`)
				DisplayGroupLink(w, group)
				w.AppendString(`</li>`)
			}
		}
		w.AppendString(`</ul>`)
	}

	if session.ID == user.ID {
		w.AppendString(`<h2>Courses</h2>`)
		w.AppendString(`<ul>`)
		for i := 0; i < len(user.Courses); i++ {
			course := user.Courses[i]

			w.AppendString(`<li>`)
			DisplayCourseLink(w, i, course)
			w.AppendString(`</li>`)
		}
		w.AppendString(`</ul>`)
		w.AppendString(`<form method="POST" action="/course/create">`)
		w.AppendString(`<input type="submit" value="Create course">`)
		w.AppendString(`</form>`)
	}

	var displaySubjects bool
	for i := 0; i < len(DB.Subjects); i++ {
		subject := &DB.Subjects[i]

		if WhoIsUserInSubject(user.ID, subject) != SubjectUserNone {
			displaySubjects = true
			break
		}
	}
	if displaySubjects {
		w.AppendString(`<h2>Subjects</h2>`)
		w.AppendString(`<ul>`)
		for i := 0; i < len(DB.Subjects); i++ {
			subject := &DB.Subjects[i]

			if WhoIsUserInSubject(user.ID, subject) != SubjectUserNone {
				w.AppendString(`<li>`)
				DisplaySubjectLink(w, subject)
				w.AppendString(`</li>`)
			}
		}
		w.AppendString(`</ul>`)
	}

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
	w.AppendString(`<head><title>Create user</title></head>`)
	w.AppendString(`<body>`)

	w.AppendString(`<h1>User</h1>`)
	w.AppendString(`<h2>Create</h2>`)

	DisplayErrorMessage(w, r.Form.Get("Error"))

	w.AppendString(`<form method="POST" action="/api/user/create">`)

	w.AppendString(`<label>First name:<br>`)
	DisplayConstraintInput(w, "text", MinNameLen, MaxNameLen, "FirstName", r.Form.Get("FirstName"), true)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>Last name:<br>`)
	DisplayConstraintInput(w, "text", MinNameLen, MaxNameLen, "LastName", r.Form.Get("LastName"), true)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>Email:<br>`)
	w.AppendString(`<input type="email" name="Email" value="`)
	w.WriteHTMLString(r.Form.Get("Email"))
	w.AppendString(`" required>`)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>Password:<br>`)
	DisplayConstraintInput(w, "password", MinPasswordLen, MaxPasswordLen, "Password", "", true)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>Repeat password:<br>`)
	DisplayConstraintInput(w, "password", MinPasswordLen, MaxPasswordLen, "RepeatPassword", "", true)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<input type="submit" value="Create">`)

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

	userID, err := GetValidIndex(r.Form.Get("ID"), DB.Users)
	if err != nil {
		return http.ClientError(err)
	}

	if (session.ID != userID) && (session.ID != AdminID) {
		return http.ForbiddenError
	}

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>Edit user</title></head>`)
	w.AppendString(`<body>`)

	w.AppendString(`<h1>User</h1>`)
	w.AppendString(`<h2>Edit</h2>`)

	DisplayErrorMessage(w, r.Form.Get("Error"))

	w.AppendString(`<form method="POST" action="/api/user/edit">`)

	w.AppendString(`<input type="hidden" name="ID" value="`)
	w.WriteInt(userID)
	w.AppendString(`">`)

	w.AppendString(`<label>First name:<br>`)
	DisplayConstraintInput(w, "text", MinNameLen, MaxNameLen, "FirstName", r.Form.Get("FirstName"), true)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>Last name:<br>`)
	DisplayConstraintInput(w, "text", MinNameLen, MaxNameLen, "LastName", r.Form.Get("LastName"), true)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>Email:<br>`)
	w.AppendString(`<input type="email" name="Email" value="`)
	w.WriteHTMLString(r.Form.Get("Email"))
	w.AppendString(`" required>`)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>Password:<br>`)
	DisplayConstraintInput(w, "password", MinPasswordLen, MaxPasswordLen, "Password", "", true)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>Repeat password:<br>`)
	DisplayConstraintInput(w, "password", MinPasswordLen, MaxPasswordLen, "RepeatPassword", "", true)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<input type="submit" value="Save">`)

	w.AppendString(`</form>`)

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func UserSigninPageHandler(w *http.Response, r *http.Request) error {
	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>Sign in</title></head>`)
	w.AppendString(`<body>`)

	w.AppendString(`<h1>Master's degree</h1>`)
	w.AppendString(`<h2>Sign in</h2>`)

	DisplayErrorMessage(w, r.Form.Get("Error"))

	w.AppendString(`<form method="POST" action="/api/user/signin">`)

	w.AppendString(`<label>Email:<br>`)
	w.AppendString(`<input type="email" name="Email" value="`)
	w.WriteHTMLString(r.Form.Get("Email"))
	w.AppendString(`" required>`)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>Password:<br>`)
	w.AppendString(`<input type="password" name="Password" required>`)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<input type="submit" value="Sign in">`)

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
	if err := UserNameValid(firstName); err != nil {
		return WritePage(w, r, UserCreatePageHandler, err)
	}

	lastName := r.Form.Get("LastName")
	if err := UserNameValid(lastName); err != nil {
		return WritePage(w, r, UserCreatePageHandler, err)
	}

	address, err := mail.ParseAddress(r.Form.Get("Email"))
	if err != nil {
		return WritePage(w, r, UserCreatePageHandler, http.BadRequest("provided email is not valid"))
	}
	email := address.Address

	password := r.Form.Get("Password")
	repeatPassword := r.Form.Get("RepeatPassword")
	if !strings.LengthInRange(password, MinPasswordLen, MaxPasswordLen) {
		return WritePage(w, r, UserCreatePageHandler, http.BadRequest("password length must be between %d and %d characters long", MinPasswordLen, MaxPasswordLen))
	}
	if password != repeatPassword {
		return WritePage(w, r, UserCreatePageHandler, http.BadRequest("passwords do not match each other"))
	}

	if GetUserByEmail(email) != nil {
		return WritePage(w, r, UserCreatePageHandler, http.Conflict("user with this email already exists"))
	}

	DB.Users = append(DB.Users, User{ID: len(DB.Users), FirstName: firstName, LastName: lastName, Email: email, Password: password, CreatedOn: time.Now()})

	w.Redirect("/", http.StatusSeeOther)
	return nil

}

func UserEditHandler(w *http.Response, r *http.Request) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}

	if err := r.ParseForm(); err != nil {
		return http.ClientError(err)
	}

	userID, err := GetValidIndex(r.Form.Get("ID"), DB.Users)
	if err != nil {
		return http.ClientError(err)
	}

	if (session.ID != userID) && (session.ID != AdminID) {
		return http.ForbiddenError
	}

	firstName := r.Form.Get("FirstName")
	if err := UserNameValid(firstName); err != nil {
		return WritePage(w, r, UserEditPageHandler, err)
	}

	lastName := r.Form.Get("LastName")
	if err := UserNameValid(lastName); err != nil {
		return WritePage(w, r, UserEditPageHandler, err)
	}

	address, err := mail.ParseAddress(r.Form.Get("Email"))
	if err != nil {
		return WritePage(w, r, UserEditPageHandler, http.BadRequest("provided email is not valid"))
	}
	email := address.Address

	password := r.Form.Get("Password")
	repeatPassword := r.Form.Get("RepeatPassword")
	if !strings.LengthInRange(password, MinPasswordLen, MaxPasswordLen) {
		return WritePage(w, r, UserEditPageHandler, http.BadRequest("password length must be between %d and %d characters long", MinPasswordLen, MaxPasswordLen))
	}
	if password != repeatPassword {
		return WritePage(w, r, UserEditPageHandler, http.BadRequest("passwords do not match each other"))
	}

	user := GetUserByEmail(email)
	if (user != nil) && (user.ID != userID) {
		return WritePage(w, r, UserEditPageHandler, http.Conflict("user with this email already exists"))
	}

	user = &DB.Users[userID]
	user.FirstName = firstName
	user.LastName = lastName
	user.Email = email
	user.Password = password

	w.RedirectID("/user/", userID, http.StatusSeeOther)
	return nil
}

func UserSigninHandler(w *http.Response, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return http.ClientError(err)
	}

	address, err := mail.ParseAddress(r.Form.Get("Email"))
	if err != nil {
		return WritePage(w, r, UserSigninPageHandler, http.BadRequest("provided email is not valid"))
	}
	email := address.Address

	user := GetUserByEmail(email)
	if user == nil {
		return WritePage(w, r, UserSigninPageHandler, http.NotFound("user with this email does not exist"))
	}

	password := r.Form.Get("Password")
	if user.Password != password {
		return WritePage(w, r, UserSigninPageHandler, http.Conflict("provided password is incorrect"))
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
		w.SetCookieUnsafe("Token", token, expiry.Unix())
	} else {
		w.SetCookie("Token", token, expiry.Unix())
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
