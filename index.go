package main

func IndexPageHandler(w *HTTPResponse, r *HTTPRequest) error {
	const page = `
<!DOCTYPE html>
<head>
	<title>Master's degree</title>
</head>
<body>
	<h1>Master's degree</h1>

	<a href="/user/signin">Sign in</a>
</body>
</html>
`
	w.AppendString(page)
	return nil
}
