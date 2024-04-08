package main

func IndexPageHandler(w *HTTPResponse, r *HTTPRequest) error {
	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>Master's degree</title></head>`)
	w.AppendString(`<body>`)
	w.AppendString(`<h1>Master's degree</h1>`)

	session, err := GetSessionFromRequest(r)
	if err == nil {
		buffer := make([]byte, 20)
		n := SlicePutInt(buffer, session.ID)

		w.AppendString(`<a href="/user/`)
		w.Write(buffer[:n])
		w.AppendString(`">Profile</a>`)
		w.AppendString("\r\n")
		w.AppendString(`<a href="/api/user/signout">Sign out</a>`)
		w.AppendString(`<br>`)

		user := DB.Users[session.ID]
		switch user.Role {
		default:
			panic("unknown user role")
		case UserRoleAdmin:
			w.AppendString(`<h2>Teachers</h2>`)

			/* TODO(anton2920): display existing teachers. */

			w.AppendString(`<h2>Students</h2>`)

			/* TODO(anton2920): display existing students. */

			/* TODO(anton2920): check dictionary for correct terminology. */
			w.AppendString(`<h2>Pre-students</h2>`)

			/* TODO(anton2920): display existing pre-students. */

			w.AppendString(`<form method="POST" action="/user/create">`)
			w.AppendString(`<input type="submit" value="Create user">`)
			w.AppendString(`</form>`)
		case UserRoleTeacher:
		case UserRoleStudent:
		case UserRolePrestudent:
		}
	} else {
		w.AppendString(`<a href="/user/signin">Sign in</a>`)
	}

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}
