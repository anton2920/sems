package main

func IndexPageHandler(w *HTTPResponse, r *HTTPRequest) error {
	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>Master's degree</title></head>`)
	w.AppendString(`<body>`)
	w.AppendString(`<h1>Master's degree</h1>`)

	session, err := GetSessionFromRequest(r)
	if err == nil {
		user := DB.Users[session.ID]

		w.AppendString(`<a href="/user/`)
		w.WriteString(user.StringID)
		w.AppendString(`">Profile</a>`)
		w.AppendString("\r\n")
		w.AppendString(`<a href="/api/user/signout">Sign out</a>`)
		w.AppendString(`<br>`)

		switch session.ID {
		default:
		case AdminID:
			w.AppendString(`<h2>Users</h2>`)
			w.AppendString(`<ul>`)
			for _, user := range DB.Users[:min(len(DB.Users), 10)] {
				w.AppendString(`<li>`)
				w.AppendString(`<a href="/user/`)
				w.WriteString(user.StringID)
				w.AppendString(`">`)
				w.WriteHTMLString(user.LastName)
				w.AppendString(` `)
				w.WriteHTMLString(user.FirstName)
				w.AppendString(` (ID: `)
				w.WriteString(user.StringID)
				w.AppendString(`)`)
				w.AppendString(`</a>`)
				w.AppendString(`</li>`)
			}
			w.AppendString(`</ul>`)
			w.AppendString(`<form method="POST" action="/user/create">`)
			w.AppendString(`<input type="submit" value="Create user">`)
			w.AppendString(`</form>`)

			w.AppendString(`<h2>Groups</h2>`)
			w.AppendString(`<ul>`)
			for _, group := range DB.Groups {
				w.AppendString(`<li>`)
				w.AppendString(`<a href="/group/`)
				w.WriteString(group.StringID)
				w.AppendString(`">`)
				w.WriteHTMLString(group.Name)
				w.AppendString(` (ID: `)
				w.WriteString(group.StringID)
				w.AppendString(`)`)
				w.AppendString(`</a>`)
				w.AppendString(`</li>`)
			}
			w.AppendString(`</ul>`)
			w.AppendString(`<form method="POST" action="/group/create">`)
			w.AppendString(`<input type="submit" value="Create group">`)
			w.AppendString(`</form>`)
		}
	} else {
		w.AppendString(`<a href="/user/signin">Sign in</a>`)
	}

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}
