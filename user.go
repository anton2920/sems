package main

import (
	"errors"
	"fmt"
	"net/mail"
	"strconv"
	"time"
	"unicode"
	"unicode/utf8"
)

type UserRole int

const (
	UserRoleAdmin = iota
	UserRoleTeacher
	UserRoleStudent
	UserRolePrestudent
)

const (
	MinUserNameLen = 1
	MaxUserNameLen = 45

	MinPasswordLen = 5
	MaxPasswordLen = 45
)

func UserNameValid(name string) error {
	if !StringLengthInRange(name, MinUserNameLen, MaxUserNameLen) {
		return fmt.Errorf("length of the name must be between %d and %d characters", MinUserNameLen, MaxUserNameLen)
	}

	/* Fist character must be a letter. */
	r, nbytes := utf8.DecodeRuneInString(name)
	if !unicode.IsLetter(r) {
		return errors.New("first character of the name must be a letter")
	}

	/* Latter characters may include: letters, spaces, dots, hyphens and apostrophes. */
	for _, r := range name[nbytes:] {
		if (!unicode.IsLetter(r)) && (r != ' ') && (r != '.') && (r != '-') && (r != '\'') {
			return errors.New("second and latter characters of the name must be letters, spaces, dots, hyphens or apostrophes")
		}
	}

	return nil
}

func UserPageHandler(w *HTTPResponse, r *HTTPRequest) error {
	buffer := make([]byte, 20)

	if _, err := GetSessionFromRequest(r); err != nil {
		return UnauthorizedError
	}

	id, err := GetIDFromURL(r.URL, "/user/")
	if err != nil {
		return err
	}

	user, ok := DB.Users[id]
	if !ok {
		return NotFoundError
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
	n := SlicePutInt(buffer, id)
	w.Write(buffer[:n])
	w.AppendString(`</p>`)

	w.AppendString(`<p>Email: `)
	w.WriteHTMLString(user.Email)
	w.AppendString(`</p>`)

	w.AppendString(`<p>Created on: `)
	w.Write(user.CreatedOn.AppendFormat(buffer[:0], "2006/01/02 15:04:05"))
	w.AppendString(`</p>`)

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

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func UserEditPageHandler(w *HTTPResponse, r *HTTPRequest) error {
	if _, err := GetSessionFromRequest(r); err != nil {
		return UnauthorizedError
	}

	if err := r.ParseForm(); err != nil {
		return ReloadPageError
	}

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>Edit user</title></head>`)
	w.AppendString(`<body>`)
	w.AppendString(`<h1>User</h1>`)
	w.AppendString(`<h2>Edit</h2>`)

	ErrorDiv(w, r.Form.Get("Error"))

	w.AppendString(`<form method="POST" action="/api/user/edit">`)

	w.AppendString(`<input type="hidden" name="ID" value="`)
	w.WriteHTMLString(r.Form.Get("ID"))
	w.AppendString(`">`)

	/* TODO(anton2920): replace with dynamically inserted length checks. */
	w.AppendString(`<label>First name:<br>`)
	w.AppendString(`<input type="text" minlength="1" maxlength="45" name="FirstName" value="`)
	w.WriteHTMLString(r.Form.Get("FirstName"))
	w.AppendString(`" required>`)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>Last name:<br>`)
	w.AppendString(`<input type="text" minlength="1" maxlength="45" name="LastName" value="`)
	w.WriteHTMLString(r.Form.Get("LastName"))
	w.AppendString(`" required>`)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>Email:<br>`)
	w.AppendString(`<input type="email" name="Email" value="`)
	w.WriteHTMLString(r.Form.Get("Email"))
	w.AppendString(`" required>`)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>Password:<br>`)
	w.AppendString(`<input type="password" minlength="5" maxlength="45" name="Password" required>`)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>Repeat password:<br>`)
	w.AppendString(`<input type="password" minlength="5" maxlength="45" name="RepeatPassword" required>`)
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

	ErrorDiv(w, r.Form.Get("Error"))

	w.AppendString(`<form method="POST" action="/api/user/signin">`)

	w.AppendString(`<label>Email:<br>`)
	w.AppendString(`<input type="email" name="Email" value="`)
	w.WriteHTMLString(r.Form.Get("Email"))
	w.AppendString(`" required>`)
	w.AppendString(`</label><br><br>`)

	w.AppendString(`<label>Password:<br>`)
	w.AppendString(`<input type="password" name="Password" required>`)
	w.AppendString(`</label><br><br>`)

	w.AppendString(`<input type="submit" value="Sign in">`)

	w.AppendString(`</form>`)

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}

func UserEditHandler(w *HTTPResponse, r *HTTPRequest) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		return UnauthorizedError
	}

	if err := r.ParseForm(); err != nil {
		return WritePage(w, r, UserEditPageHandler, ReloadPageError)
	}

	id := r.Form.Get("ID")
	userID, err := strconv.Atoi(id)
	if err != nil {
		return WritePage(w, r, UserEditPageHandler, ReloadPageError)
	}

	if session.ID != userID {
		return WritePage(w, r, UserEditPageHandler, ForbiddenError)
	}

	firstName := r.Form.Get("FirstName")
	if err := UserNameValid(firstName); err != nil {
		return WritePage(w, r, UserEditPageHandler, NewHTTPError(HTTPStatusBadRequest, err.Error()))
	}

	lastName := r.Form.Get("LastName")
	if err := UserNameValid(lastName); err != nil {
		return WritePage(w, r, UserEditPageHandler, NewHTTPError(HTTPStatusBadRequest, err.Error()))
	}

	address, err := mail.ParseAddress(r.Form.Get("Email"))
	if err != nil {
		return WritePage(w, r, UserEditPageHandler, NewHTTPError(HTTPStatusBadRequest, "provided email is not valid"))
	}
	email := address.Address

	password := r.Form.Get("Password")
	repeatPassword := r.Form.Get("RepeatPassword")
	if !StringLengthInRange(password, MinPasswordLen, MaxPasswordLen) {
		return WritePage(w, r, UserEditPageHandler, NewHTTPError(HTTPStatusBadRequest, fmt.Sprintf("password length must be between %d and %d characters long", MinPasswordLen, MaxPasswordLen)))
	}
	if password != repeatPassword {
		return WritePage(w, r, UserEditPageHandler, NewHTTPError(HTTPStatusBadRequest, "passwords do not match each other"))
	}

	user := DB.Users[userID]
	user.FirstName = firstName
	user.LastName = lastName
	user.Email = email
	user.Password = password

	w.Redirect(fmt.Appendf(make([]byte, 0, 20), "/user/%s", id), HTTPStatusSeeOther)
	return nil
}

func UserSigninHandler(w *HTTPResponse, r *HTTPRequest) error {
	if err := r.ParseForm(); err != nil {
		return ReloadPageError
	}

	/* TODO(anton2920): replace with actual user checks. */
	var id int
	email := r.Form.Get("Email")
	password := r.Form.Get("Password")

	for i, user := range DB.Users {
		if email == user.Email {
			if password != user.Password {
				r.Form.Set("Error", "provided password is incorrect")
				w.StatusCode = HTTPStatusConflict
				return UserSigninPageHandler(w, r)
			}

			id = i
			break
		}
	}
	if id == 0 {
		r.Form.Set("Error", "user with this email does not exist")
		w.StatusCode = HTTPStatusNotFound
		return UserSigninPageHandler(w, r)
	}

	token, err := GenerateSessionToken()
	if err != nil {
		return err
	}
	expiry := time.Now().Add(OneWeek)

	SessionsLock.Lock()
	Sessions[token] = &Session{
		ID:     id,
		Expiry: expiry,
	}
	SessionsLock.Unlock()

	w.SetCookie("Token", token, expiry)
	w.RedirectString("/", HTTPStatusSeeOther)
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
	w.RedirectString("/", HTTPStatusSeeOther)
	return nil
}
