package main

func IndexPageHandler(w *HTTPResponse, r *HTTPRequest) error {
	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>Master's degree</title></head>`)
	w.AppendString(`<body>`)
	w.AppendString(`<h1>Master's degree</h1>`)

	session, err := GetSessionFromRequest(r)
	if err != nil {
		w.AppendString(`<a href="/user/signin">Sign in</a>`)
	} else {
		buffer := make([]byte, 20)
		n := SlicePutInt(buffer, session.ID)

		w.AppendString(`<a href="/user/`)
		w.Write(buffer[:n])
		w.AppendString(`">Profile</a>`)
	}

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}
