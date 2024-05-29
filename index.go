package main

import "github.com/anton2920/gofa/net/http"

func DisplayIndexAdminPage(w *http.Response, user *User) {
	users := make([]User, 10)
	var pos int64

	w.AppendString(`<h2>Users</h2>`)
	w.AppendString(`<ul>`)
	for {
		n, err := GetUsers(DB2, &pos, users)
		if err != nil {
			/* TODO(anton2920): report error. */
		}
		if n == 0 {
			break
		}
		for i := 0; i < n; i++ {
			w.AppendString(`<li>`)
			DisplayUserLink(w, &users[i])
			w.AppendString(`</li>`)
		}
	}
	w.AppendString(`</ul>`)
	w.AppendString(`<form method="POST" action="/user/create">`)
	w.AppendString(`<input type="submit" value="Create user">`)
	w.AppendString(`</form>`)

	groups := make([]Group, 32)
	pos = 0

	w.AppendString(`<h2>Groups</h2>`)
	w.AppendString(`<ul>`)
	for {
		n, err := GetGroups(DB2, &pos, groups)
		if err != nil {
			/* TODO(anton2920): report error. */
		}
		if n == 0 {
			break
		}
		for i := 0; i < n; i++ {
			w.AppendString(`<li>`)
			DisplayGroupLink(w, &groups[i])
			w.AppendString(`</li>`)
		}
	}
	w.AppendString(`</ul>`)
	w.AppendString(`<form method="POST" action="/group/create">`)
	w.AppendString(`<input type="submit" value="Create group">`)
	w.AppendString(`</form>`)

	w.AppendString(`<h2>Courses</h2>`)
	w.AppendString(`<ul>`)
	for i := 0; i < len(DB.Courses); i++ {
		course := &DB.Courses[i]

		w.AppendString(`<li>`)
		DisplayCourseLink(w, i, course)
		w.AppendString(`</li>`)
	}
	w.AppendString(`</ul>`)
	w.AppendString(`<form method="POST" action="/course/create">`)
	w.AppendString(`<input type="submit" value="Create course">`)
	w.AppendString(`</form>`)

	w.AppendString(`<h2>Subjects</h2>`)
	w.AppendString(`<ul>`)
	for i := 0; i < len(DB.Subjects); i++ {
		subject := &DB.Subjects[i]

		w.AppendString(`<li>`)
		DisplaySubjectLink(w, subject)
		w.AppendString(`</li>`)
	}
	w.AppendString(`</ul>`)
	w.AppendString(`<form method="POST" action="/subject/create">`)
	w.AppendString(`<input type="submit" value="Create subject">`)
	w.AppendString(`</form>`)
}

func DisplayIndexUserPage(w *http.Response, user *User) {
	DisplayUserGroups(w, user.ID)
	DisplayUserCourses(w, user)
	DisplayUserSubjects(w, user.ID)
}

func IndexPageHandler(w *http.Response, r *http.Request) error {
	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>Master's degree</title></head>`)
	w.AppendString(`<body>`)

	w.AppendString(`<h1>Master's degree</h1>`)

	session, err := GetSessionFromRequest(r)
	if err != nil {
		w.AppendString(`<a href="/user/signin">Sign in</a>`)
	} else {
		var user User
		if err := GetUserByID(DB2, session.ID, &user); err != nil {
			return http.ServerError(err)
		}

		w.AppendString(`<a href="/user/`)
		w.WriteInt(int(user.ID))
		w.AppendString(`">Profile</a> `)
		w.AppendString(`<a href="/api/user/signout">Sign out</a>`)
		w.AppendString(`<br>`)

		if session.ID == AdminID {
			DisplayIndexAdminPage(w, &user)
		} else {
			DisplayIndexUserPage(w, &user)
		}
	}

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}
