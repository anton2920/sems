package main

import "github.com/anton2920/gofa/net/http"

func DisplayIndexAdminPage(w *http.Response, l Language, user *User) error {
	users := make([]User, 10)
	pos := int64(0)

	w.AppendString(`<h2>`)
	w.AppendString(Ls(l, "Users"))
	w.AppendString(`</h2>`)
	w.AppendString(`<ul>`)
	for {
		n, err := GetUsers(&pos, users)
		if err != nil {
			/* TODO(anton2920): report error. */
		}
		if n == 0 {
			break
		}
		for i := 0; i < n; i++ {
			user := &users[i]
			if user.Flags == UserDeleted {
				continue
			}

			w.AppendString(`<li>`)
			DisplayUserLink(w, l, user)
			w.AppendString(`</li>`)
		}
	}
	w.AppendString(`</ul>`)
	w.AppendString(`<form method="POST" action="/user/create">`)
	DisplaySubmit(w, l, "", "Create user", true)
	w.AppendString(`</form>`)

	groups := make([]Group, 32)
	pos = 0

	w.AppendString(`<h2>`)
	w.AppendString(Ls(l, "Groups"))
	w.AppendString(`</h2>`)
	w.AppendString(`<ul>`)
	for {
		n, err := GetGroups(&pos, groups)
		if err != nil {
			/* TODO(anton2920): report error. */
		}
		if n == 0 {
			break
		}
		for i := 0; i < n; i++ {
			group := &groups[i]
			if group.Flags == GroupDeleted {
				continue
			}

			w.AppendString(`<li>`)
			DisplayGroupLink(w, l, group)
			w.AppendString(`</li>`)
		}
	}
	w.AppendString(`</ul>`)
	w.AppendString(`<form method="POST" action="/group/create">`)
	DisplaySubmit(w, l, "", "Create group", true)
	w.AppendString(`</form>`)

	courses := make([]Course, 32)
	pos = 0

	w.AppendString(`<h2>`)
	w.AppendString(Ls(l, "Courses"))
	w.AppendString(`</h2>`)
	w.AppendString(`<ul>`)
	for {
		n, err := GetCourses(&pos, courses)
		if err != nil {
			/* TODO(anton2920): report error. */
		}
		if n == 0 {
			break
		}
		for i := 0; i < n; i++ {
			course := &courses[i]
			if course.Flags == CourseDeleted {
				continue
			}

			w.AppendString(`<li>`)
			DisplayCourseLink(w, l, course)
			w.AppendString(`</li>`)
		}
	}
	w.AppendString(`</ul>`)
	w.AppendString(`<form method="POST" action="/course/create">`)
	DisplaySubmit(w, l, "", "Create course", true)
	w.AppendString(`</form>`)

	subjects := make([]Subject, 32)
	pos = 0

	w.AppendString(`<h2>`)
	w.AppendString(Ls(l, "Subjects"))
	w.AppendString(`</h2>`)
	w.AppendString(`<ul>`)
	for {
		n, err := GetSubjects(&pos, subjects)
		if err != nil {
			/* TODO(anton2920): report error. */
		}
		if n == 0 {
			break
		}
		for i := 0; i < n; i++ {
			subject := &subjects[i]
			if subject.Flags == SubjectDeleted {
				continue
			}

			w.AppendString(`<li>`)
			DisplaySubjectLink(w, l, subject)
			w.AppendString(`</li>`)
		}
	}
	w.AppendString(`</ul>`)
	w.AppendString(`<form method="POST" action="/subject/create">`)
	DisplaySubmit(w, l, "", "Create subject", true)
	w.AppendString(`</form>`)
	return nil
}

func DisplayIndexUserPage(w *http.Response, l Language, user *User) error {
	DisplayUserGroups(w, l, user.ID)
	DisplayUserCourses(w, l, user)
	DisplayUserSubjects(w, l, user.ID)
	return nil
}

func DisplayIndexTitle(w *http.Response, l Language) {
	w.AppendString(Ls(l, "Master's degree"))
}

func IndexPageHandler(w *http.Response, r *http.Request) error {
	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head>`)
	w.AppendString(`<title>`)
	DisplayIndexTitle(w, GL)
	w.AppendString(`</title>`)
	DisplayCSSLink(w)
	w.AppendString(`</head>`)

	w.AppendString(`<body>`)

	w.AppendString(`<h1>`)
	DisplayIndexTitle(w, GL)
	w.AppendString(`</h1>`)

	session, err := GetSessionFromRequest(r)
	if err != nil {
		w.AppendString(`<a href="/user/signin">`)
		w.AppendString(Ls(GL, "Sign in"))
		w.AppendString(`</a>`)
	} else {
		var user User
		if err := GetUserByID(session.ID, &user); err != nil {
			return http.ServerError(err)
		}

		w.AppendString(`<a href="/user/`)
		w.WriteID(user.ID)
		w.AppendString(`">`)
		w.AppendString(Ls(GL, "Profile"))
		w.AppendString(`</a> `)
		w.AppendString(`<a href="/api/user/signout">`)
		w.AppendString(Ls(GL, "Sign out"))
		w.AppendString(`</a>`)
		w.AppendString(`<br>`)

		if session.ID == AdminID {
			if err := DisplayIndexAdminPage(w, GL, &user); err != nil {
				return http.ServerError(err)
			}
		} else {
			if err := DisplayIndexUserPage(w, GL, &user); err != nil {
				return http.ServerError(err)
			}
		}
	}

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}
