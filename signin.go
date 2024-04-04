package main

import "time"

func SigninPageHandler(w *HTTPResponse, r *HTTPRequest) error {
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

func SigninHandler(w *HTTPResponse, r *HTTPRequest) error {
	if err := r.ParseForm(); err != nil {
		return ReloadPageError
	}

	const email = "anton2920@gmail.com"
	const pass = "pass&word"

	if email != r.Form.Get("Email") {
		r.Form.Set("Error", "user with this email does not exist")
		w.StatusCode = HTTPStatusNotFound
		return SigninPageHandler(w, r)
	}

	if pass != r.Form.Get("Password") {
		r.Form.Set("Error", "provided password is incorrect")
		w.StatusCode = HTTPStatusConflict
		return SigninPageHandler(w, r)
	}

	token, _ := GenerateSessionToken()
	expiry := time.Now().Add(OneWeek)

	SessionsLock.Lock()
	Sessions[token] = &Session{
		ID:     1,
		Expiry: expiry,
	}
	SessionsLock.Unlock()

	w.SetCookie("Token", token, expiry)
	w.Redirect("/", HTTPStatusSeeOther)
	return nil
}
