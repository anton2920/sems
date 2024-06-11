package main

import "github.com/anton2920/gofa/net/http"

func DisplayIndexAdminPage(w *http.Response, user *User) error {
	users := make([]User, 10)
	var pos int64

	w.AppendString(`<h2>Users</h2>`)
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
			DisplayUserLink(w, user)
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
			DisplayGroupLink(w, group)
			w.AppendString(`</li>`)
		}
	}
	w.AppendString(`</ul>`)
	w.AppendString(`<form method="POST" action="/group/create">`)
	w.AppendString(`<input type="submit" value="Create group">`)
	w.AppendString(`</form>`)

	courses := make([]Course, 32)
	pos = 0

	w.AppendString(`<h2>Courses</h2>`)
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
			DisplayCourseLink(w, course)
			w.AppendString(`</li>`)
		}
	}
	w.AppendString(`</ul>`)
	w.AppendString(`<form method="POST" action="/course/create">`)
	w.AppendString(`<input type="submit" value="Create course">`)
	w.AppendString(`</form>`)

	subjects := make([]Subject, 32)
	pos = 0

	w.AppendString(`<h2>Subjects</h2>`)
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
			DisplaySubjectLink(w, subject)
			w.AppendString(`</li>`)
		}
	}
	w.AppendString(`</ul>`)
	w.AppendString(`<form method="POST" action="/subject/create">`)
	w.AppendString(`<input type="submit" value="Create subject">`)
	w.AppendString(`</form>`)
	return nil
}

func DisplayIndexUserPage(w *http.Response, user *User) error {
	DisplayUserGroups(w, user.ID)
	DisplayUserCourses(w, user)
	DisplayUserSubjects(w, user.ID)
	return nil
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
		if err := GetUserByID(session.ID, &user); err != nil {
			return http.ServerError(err)
		}

		w.AppendString(`<a href="/user/`)
		w.WriteID(user.ID)
		w.AppendString(`">Profile</a> `)
		w.AppendString(`<a href="/api/user/signout">Sign out</a>`)
		w.AppendString(`<br>`)

		if session.ID == AdminID {
			if err := DisplayIndexAdminPage(w, &user); err != nil {
				return http.ServerError(err)
			}
		} else {
			if err := DisplayIndexUserPage(w, &user); err != nil {
				return http.ServerError(err)
			}
		}
	}

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}
