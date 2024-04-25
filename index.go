package main

func IndexPageHandler(w *HTTPResponse, r *HTTPRequest) error {
	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>Master's degree</title></head>`)
	w.AppendString(`<body>`)

	w.AppendString(`<h1>Master's degree</h1>`)

	session, err := GetSessionFromRequest(r)
	if err == nil {
		user := &DB.Users[session.ID]

		w.AppendString(`<a href="/user/`)
		w.WriteInt(user.ID)
		w.AppendString(`">Profile</a>`)
		w.AppendString("\r\n")
		w.AppendString(`<a href="/api/user/signout">Sign out</a>`)
		w.AppendString(`<br>`)

		if session.ID == AdminID {
			w.AppendString(`<h2>Users</h2>`)
			w.AppendString(`<ul>`)
			for i := 0; i < min(len(DB.Users), 10); i++ {
				user := &DB.Users[i]

				w.AppendString(`<li>`)
				w.AppendString(`<a href="/user/`)
				w.WriteInt(user.ID)
				w.AppendString(`">`)
				w.WriteHTMLString(user.LastName)
				w.AppendString(` `)
				w.WriteHTMLString(user.FirstName)
				w.AppendString(` (ID: `)
				w.WriteInt(user.ID)
				w.AppendString(`)`)
				w.AppendString(`</a>`)
				w.AppendString(`</li>`)
			}
			w.AppendString(`</ul>`)
			w.AppendString(`<form method="POST" action="/user/create">`)
			w.AppendString(`<input type="submit" value="Create user">`)
			w.AppendString(`</form>`)
		}

		if session.ID == AdminID {
			w.AppendString(`<h2>Groups</h2>`)
			w.AppendString(`<ul>`)
			for i := 0; i < len(DB.Groups); i++ {
				group := &DB.Groups[i]

				w.AppendString(`<li>`)
				w.AppendString(`<a href="/group/`)
				w.WriteInt(group.ID)
				w.AppendString(`">`)
				w.WriteHTMLString(group.Name)
				w.AppendString(` (ID: `)
				w.WriteInt(group.ID)
				w.AppendString(`)`)
				w.AppendString(`</a>`)
				w.AppendString(`</li>`)
			}
			w.AppendString(`</ul>`)
			w.AppendString(`<form method="POST" action="/group/create">`)
			w.AppendString(`<input type="submit" value="Create group">`)
			w.AppendString(`</form>`)
		} else {
			w.AppendString(`<h2>Groups</h2>`)
			w.AppendString(`<ul>`)
			for i := 0; i < len(DB.Groups); i++ {
				group := &DB.Groups[i]

				var member bool
				for j := 0; j < len(group.Users); j++ {
					user := group.Users[j]

					if session.ID == user.ID {
						member = true
						break
					}
				}
				if !member {
					continue
				}

				w.AppendString(`<li>`)
				w.AppendString(`<a href="/group/`)
				w.WriteInt(group.ID)
				w.AppendString(`">`)
				w.WriteHTMLString(group.Name)
				w.AppendString(` (ID: `)
				w.WriteInt(group.ID)
				w.AppendString(`)`)
				w.AppendString(`</a>`)
				w.AppendString(`</li>`)
			}
			w.AppendString(`</ul>`)
		}

		w.AppendString(`<h2>Courses</h2>`)
		w.AppendString(`<ul>`)
		for i := 0; i < len(user.Courses); i++ {
			course := user.Courses[i]
			w.AppendString(`<li>`)
			w.AppendString(`<a href="/course/`)
			w.WriteInt(i)
			w.AppendString(`">`)
			w.WriteHTMLString(course.Name)
			if course.Draft {
				w.AppendString(` (draft)`)
			}
			w.AppendString(`</a>`)
			w.AppendString(`</li>`)
		}
		w.AppendString(`</ul>`)
		w.AppendString(`<form method="POST" action="/course/create">`)
		w.AppendString(`<input type="submit" value="Create course">`)
		w.AppendString(`</form>`)

		/* TODO(anton2920): don't display header if no subjects available. */
		w.AppendString(`<h2>Subjects</h2>`)
		w.AppendString(`<ul>`)
		for i := 0; i < len(DB.Subjects); i++ {
			subject := &DB.Subjects[i]

			if WhoIsUserInSubject(session.ID, subject) != SubjectUserNone {
				w.AppendString(`<li>`)
				w.AppendString(`<a href="/subject/`)
				w.WriteInt(subject.ID)
				w.AppendString(`">`)
				w.WriteHTMLString(subject.Name)
				w.AppendString(` with `)
				w.WriteHTMLString(subject.Teacher.LastName)
				w.AppendString(` `)
				w.WriteHTMLString(subject.Teacher.FirstName)
				w.AppendString(` (ID: `)
				w.WriteInt(subject.ID)
				w.AppendString(`)`)
				w.AppendString(`</a>`)
				w.AppendString(`</li>`)
			}
		}
		w.AppendString(`</ul>`)

		if session.ID == AdminID {
			w.AppendString(`<form method="POST" action="/subject/create">`)
			w.AppendString(`<input type="submit" value="Create subject">`)
			w.AppendString(`</form>`)
		}
	} else {
		w.AppendString(`<a href="/user/signin">Sign in</a>`)
	}

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}
