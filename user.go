package main

import (
	"fmt"
	"strconv"
	"time"
	"unsafe"
)

type UserRole int

const (
	UserRoleAdmin = iota
	UserRoleTeacher
	UserRoleStudent
	UserRolePrestudent
)

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
	w.WriteHTMLString(r.URL.Path[len("/user/"):])
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

	w.AppendString(`<label>First name:<br>`)
	w.AppendString(`<input type="text" name="FirstName" value="`)
	w.WriteHTMLString(r.Form.Get("FirstName"))
	w.AppendString(`" required>`)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>Last name:<br>`)
	w.AppendString(`<input type="text" name="LastName" value="`)
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
	w.AppendString(`<input type="password" name="Password" required>`)
	w.AppendString(`</label>`)
	w.AppendString(`<br><br>`)

	w.AppendString(`<label>Repeat password:<br>`)
	w.AppendString(`<input type="password" name="RepeatPassword" required>`)
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
		r.Form.Set("Error", ReloadPageError.Message)
		w.StatusCode = ReloadPageError.StatusCode
		return UserEditPageHandler(w, r)
	}

	id := r.Form.Get("ID")
	userID, err := strconv.Atoi(id)
	if err != nil {
		r.Form.Set("Error", ReloadPageError.Message)
		w.StatusCode = ReloadPageError.StatusCode
		return UserEditPageHandler(w, r)
	}

	if session.ID != userID {
		r.Form.Set("Error", ForbiddenError.Message)
		w.StatusCode = ForbiddenError.StatusCode
		return UserEditPageHandler(w, r)
	}

	password := r.Form.Get("Password")
	repeatPassword := r.Form.Get("RepeatPassword")

	if password != repeatPassword {
		r.Form.Set("Error", "passwords do not match each other")
		w.StatusCode = HTTPStatusBadRequest
		return UserEditPageHandler(w, r)
	}

	user := DB.Users[userID]
	user.FirstName = r.Form.Get("FirstName")
	user.LastName = r.Form.Get("LastName")
	user.Email = r.Form.Get("Email")
	user.Password = password

	buffer := make([]byte, 0, 20)
	buffer = fmt.Appendf(buffer, "/user/%s", id)
	w.Redirect(unsafe.String(unsafe.SliceData(buffer), len(buffer)), HTTPStatusSeeOther)
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
