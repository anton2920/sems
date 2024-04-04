package main

import "time"

func UserPageHandler(w *HTTPResponse, r *HTTPRequest) error {
	buffer := make([]byte, 20)

	if _, err := GetSessionFromRequest(r); err != nil {
		return UnauthorizedError
	}

	id, err := GetIDFromURL(r.URL, "/user/")
	if err != nil {
		return err
	}

	user, ok := UsersDB[id]
	if !ok {
		return NotFoundError
	}

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>`)
	w.WriteString(user.LastName)
	w.AppendString(` `)
	w.WriteString(user.FirstName)
	w.AppendString(`</title></head>`)

	w.AppendString(`<body>`)

	w.AppendString(`<h1>`)
	w.WriteString(user.LastName)
	w.AppendString(` `)
	w.WriteString(user.FirstName)
	w.AppendString(`</h1>`)

	w.AppendString(`<h2>Info</h2>`)

	w.AppendString(`<p>ID: `)
	n := SlicePutInt(buffer, id)
	w.Write(buffer[:n])
	w.AppendString(`</p>`)

	w.AppendString(`<p>Email: `)
	w.WriteString(user.Email)
	w.AppendString(`</p>`)

	w.AppendString(`<p>Created on: `)
	w.Write(user.CreatedOn.AppendFormat(buffer[:0], "2006/01/02 15:04:05"))
	w.AppendString(`</p>`)

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

func UserSigninHandler(w *HTTPResponse, r *HTTPRequest) error {
	if err := r.ParseForm(); err != nil {
		return ReloadPageError
	}

	/* TODO(anton2920): replace with actual user checks. */
	var id int
	email := r.Form.Get("Email")
	password := r.Form.Get("Password")

	for i, user := range UsersDB {
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
