package main

import (
	"net/mail"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	MinUserNameLen = 1
	MaxUserNameLen = 45

	MinPasswordLen = 5
	MaxPasswordLen = 45
)

func UserNameValid(name string) error {
	if !StringLengthInRange(name, MinUserNameLen, MaxUserNameLen) {
		return BadRequest("length of the name must be between %d and %d characters", MinUserNameLen, MaxUserNameLen)
	}

	/* Fist character must be a letter. */
	r, nbytes := utf8.DecodeRuneInString(name)
	if !unicode.IsLetter(r) {
		return BadRequest("first character of the name must be a letter")
	}

	/* Latter characters may include: letters, spaces, dots, hyphens and apostrophes. */
	for _, r := range name[nbytes:] {
		if (!unicode.IsLetter(r)) && (r != ' ') && (r != '.') && (r != '-') && (r != '\'') {
			return BadRequest("second and latter characters of the name must be letters, spaces, dots, hyphens or apostrophes")
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

func DisplayUserLink(w *HTTPResponse, user *User) {
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

func UserPageHandler(w *HTTPResponse, r *HTTPRequest) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}

	id, err := GetIDFromURL(r.URL, "/user/")
	if err != nil {
		return ClientError(err)
	}
	if (id < 0) || (id >= len(DB.Users)) {
		return NotFound("user with this ID does not exist")
	}
	user := &DB.Users[id]

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

func UserCreatePageHandler(w *HTTPResponse, r *HTTPRequest) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}
	if session.ID != AdminID {
		return ForbiddenError
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

func UserEditPageHandler(w *HTTPResponse, r *HTTPRequest) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}
	if session.ID != AdminID {
		return ForbiddenError
	}

	if err := r.ParseForm(); err != nil {
		return ClientError(err)
	}

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>Edit user</title></head>`)
	w.AppendString(`<body>`)

	w.AppendString(`<h1>User</h1>`)
	w.AppendString(`<h2>Edit</h2>`)

	DisplayErrorMessage(w, r.Form.Get("Error"))

	w.AppendString(`<form method="POST" action="/api/user/edit">`)

	w.AppendString(`<input type="hidden" name="ID" value="`)
	w.WriteHTMLString(r.Form.Get("ID"))
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

func UserSigninPageHandler(w *HTTPResponse, r *HTTPRequest) error {
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

func UserCreateHandler(w *HTTPResponse, r *HTTPRequest) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}
	if session.ID != AdminID {
		return ForbiddenError
	}

	if err := r.ParseForm(); err != nil {
		return ClientError(err)
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
		return WritePage(w, r, UserCreatePageHandler, BadRequest("provided email is not valid"))
	}
	email := address.Address

	password := r.Form.Get("Password")
	repeatPassword := r.Form.Get("RepeatPassword")
	if !StringLengthInRange(password, MinPasswordLen, MaxPasswordLen) {
		return WritePage(w, r, UserCreatePageHandler, BadRequest("password length must be between %d and %d characters long", MinPasswordLen, MaxPasswordLen))
	}
	if password != repeatPassword {
		return WritePage(w, r, UserCreatePageHandler, BadRequest("passwords do not match each other"))
	}

	if GetUserByEmail(email) != nil {
		return WritePage(w, r, UserCreatePageHandler, Conflict("user with this email already exists"))
	}

	DB.Users = append(DB.Users, User{ID: len(DB.Users), FirstName: firstName, LastName: lastName, Email: email, Password: password, CreatedOn: time.Now()})

	w.Redirect("/", HTTPStatusSeeOther)
	return nil

}

func UserEditHandler(w *HTTPResponse, r *HTTPRequest) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}

	if err := r.ParseForm(); err != nil {
		return ClientError(err)
	}

	userID, err := GetValidIndex(r.Form.Get("ID"), DB.Users)
	if err != nil {
		return ClientError(err)
	}

	if (session.ID != userID) && (session.ID != AdminID) {
		return ForbiddenError
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
		return WritePage(w, r, UserEditPageHandler, BadRequest("provided email is not valid"))
	}
	email := address.Address

	password := r.Form.Get("Password")
	repeatPassword := r.Form.Get("RepeatPassword")
	if !StringLengthInRange(password, MinPasswordLen, MaxPasswordLen) {
		return WritePage(w, r, UserEditPageHandler, BadRequest("password length must be between %d and %d characters long", MinPasswordLen, MaxPasswordLen))
	}
	if password != repeatPassword {
		return WritePage(w, r, UserEditPageHandler, BadRequest("passwords do not match each other"))
	}

	user := GetUserByEmail(email)
	if (user != nil) && (user.ID != userID) {
		return WritePage(w, r, UserEditPageHandler, Conflict("user with this email already exists"))
	}

	user = &DB.Users[userID]
	user.FirstName = firstName
	user.LastName = lastName
	user.Email = email
	user.Password = password

	w.RedirectID("/user/", userID, HTTPStatusSeeOther)
	return nil
}

func UserSigninHandler(w *HTTPResponse, r *HTTPRequest) error {
	if err := r.ParseForm(); err != nil {
		return ClientError(err)
	}

	address, err := mail.ParseAddress(r.Form.Get("Email"))
	if err != nil {
		return WritePage(w, r, UserEditPageHandler, BadRequest("provided email is not valid"))
	}
	email := address.Address

	user := GetUserByEmail(email)
	if user == nil {
		return WritePage(w, r, UserSigninPageHandler, NotFound("user with this email does not exist"))
	}

	password := r.Form.Get("Password")
	if user.Password != password {
		return WritePage(w, r, UserSigninPageHandler, Conflict("provided password is incorrect"))
	}

	token, err := GenerateSessionToken()
	if err != nil {
		return ServerError(err)
	}
	expiry := time.Now().Add(OneWeek)

	SessionsLock.Lock()
	Sessions[token] = &Session{
		ID:     user.ID,
		Expiry: expiry,
	}
	SessionsLock.Unlock()

	if Debug {
		w.SetCookieUnsafe("Token", token, expiry)
	} else {
		w.SetCookie("Token", token, expiry)
	}
	w.Redirect("/", HTTPStatusSeeOther)
	return nil
}

func UserSignoutHandler(w *HTTPResponse, r *HTTPRequest) error {
	token := r.Cookie("Token")
	if token == "" {
		return UnauthorizedError
	}

	if _, err := GetSessionFromToken(token); err != nil {
		return UnauthorizedError
	}

	SessionsLock.Lock()
	delete(Sessions, token)
	SessionsLock.Unlock()

	w.DelCookie("Token")
	w.Redirect("/", HTTPStatusSeeOther)
	return nil
}
