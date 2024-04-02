package main

import (
	"fmt"
)

func SigninPageHandler(w *HTTPResponse, r *HTTPRequest) error {
	const pageFormat = `
<!DOCTYPE html>
<head>
	<title>Sign in</title>
</head>
<body>
	<h1>Master's degree</h1>
	<h2>Sign in</h2>

	<form method="POST" action="/api/user/signin">
		<label>Email:
			<input type="email" name="Email" value="%s" required>
		</label>
		<br><br>

		<label>Password:
			<input type="password" name="Password" value="%s" required>
		</label>
		<br><br>

		<input type="submit" value="Sign in">
	</form>
</body>
</html>
`
	const email = ""
	const password = ""

	/* TODO(anton2920): sanitize. */
	fmt.Fprintf(w, pageFormat, email, password)
	return nil
}

func SigninHandler(w *HTTPResponse, r *HTTPRequest) error {
	if err := r.ParseForm(); err != nil {
		return ReloadPageError
	}

	const email = "anton2920@gmail.com"
	const pass = "pass&word"

	if (email == r.Form.Get("Email")) && (pass == r.Form.Get("Password")) {
		w.AppendString("<h1>Success!</h1>")
	} else {
		w.AppendString("<h1>Failure</h1>")
	}

	return nil
}
