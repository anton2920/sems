package main

import (
	"fmt"
	"unsafe"
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
	const pageFormat = `
<!DOCTYPE html>
<body>
	<h1>Master's degree</h1>
	<h2>API signin</h2>
	
	<p>%s</p>
</body>
</html>
`

	fmt.Fprintf(w, pageFormat, unsafe.String(unsafe.SliceData(r.Body), len(r.Body)))
	return nil
}
