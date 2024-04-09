package main

func IndexPageDisplayUsers(w *HTTPResponse, users []*User) {}

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
		switch user.RoleID {
		default:
			panic("unknown user role")
		case UserRoleAdmin:
			admins := make([]*User, 0, 20)
			teachers := make([]*User, 0, 200)
			students := make([]*User, 0, 2000)
			prestudents := make([]*User, 0, 20000)

			for _, user := range DB.Users {
				switch user.RoleID {
				case UserRoleAdmin:
					admins = append(admins, user)
				case UserRoleTeacher:
					teachers = append(teachers, user)
				case UserRoleStudent:
					students = append(students, user)
				case UserRolePrestudent:
					prestudents = append(prestudents, user)
				}
			}

			w.AppendString(`<h2>Users</h2>`)
			UserDisplayList(w, "h3", admins)
			UserDisplayList(w, "h3", teachers)
			UserDisplayList(w, "h3", students)
			UserDisplayList(w, "h3", prestudents)
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
