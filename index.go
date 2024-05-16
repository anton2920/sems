package main

import "github.com/anton2920/gofa/net/http"

func DisplayIndexAdminPage(w *http.Response, user *User) {
	w.AppendString(`<h2>Users</h2>`)
	w.AppendString(`<ul>`)
	for i := 0; i < min(len(DB.Users), 10); i++ {
		user := &DB.Users[i]

		w.AppendString(`<li>`)
		DisplayUserLink(w, user)
		w.AppendString(`</li>`)
	}
	w.AppendString(`</ul>`)
	w.AppendString(`<form method="POST" action="/user/create">`)
	w.AppendString(`<input type="submit" value="Create user">`)
	w.AppendString(`</form>`)

	w.AppendString(`<h2>Groups</h2>`)
	w.AppendString(`<ul>`)
	for i := 0; i < len(DB.Groups); i++ {
		group := &DB.Groups[i]

		w.AppendString(`<li>`)
		DisplayGroupLink(w, group)
		w.AppendString(`</li>`)
	}
	w.AppendString(`</ul>`)
	w.AppendString(`<form method="POST" action="/group/create">`)
	w.AppendString(`<input type="submit" value="Create group">`)
	w.AppendString(`</form>`)

	w.AppendString(`<h2>Courses</h2>`)
	w.AppendString(`<ul>`)
	for i := 0; i < len(user.Courses); i++ {
		course := user.Courses[i]

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
	userID := user.ID

	var displayGroups bool
	for i := 0; i < len(DB.Groups); i++ {
		group := &DB.Groups[i]

		if UserInGroup(userID, group) {
			displayGroups = true
			break
		}
	}
	if displayGroups {
		w.AppendString(`<h2>Groups</h2>`)
		w.AppendString(`<ul>`)
		for i := 0; i < len(DB.Groups); i++ {
			group := &DB.Groups[i]

			if UserInGroup(userID, group) {
				w.AppendString(`<li>`)
				DisplayGroupLink(w, group)
				w.AppendString(`</li>`)
			}
		}
		w.AppendString(`</ul>`)
	}

	w.AppendString(`<h2>Courses</h2>`)
	w.AppendString(`<ul>`)
	for i := 0; i < len(user.Courses); i++ {
		course := user.Courses[i]

		w.AppendString(`<li>`)
		DisplayCourseLink(w, i, course)
		w.AppendString(`</li>`)
	}
	w.AppendString(`</ul>`)
	w.AppendString(`<form method="POST" action="/course/create">`)
	w.AppendString(`<input type="submit" value="Create course">`)
	w.AppendString(`</form>`)

	var displaySubjects bool
	for i := 0; i < len(DB.Subjects); i++ {
		subject := &DB.Subjects[i]

		if WhoIsUserInSubject(userID, subject) != SubjectUserNone {
			displaySubjects = true
			break
		}
	}
	if displaySubjects {
		w.AppendString(`<h2>Subjects</h2>`)
		w.AppendString(`<ul>`)
		for i := 0; i < len(DB.Subjects); i++ {
			subject := &DB.Subjects[i]

			if WhoIsUserInSubject(userID, subject) != SubjectUserNone {
				w.AppendString(`<li>`)
				DisplaySubjectLink(w, subject)
				w.AppendString(`</li>`)
			}
		}
		w.AppendString(`</ul>`)
	}
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
		user := &DB.Users[session.ID]

		w.AppendString(`<a href="/user/`)
		w.WriteInt(user.ID)
		w.AppendString(`">Profile</a> `)
		w.AppendString(`<a href="/api/user/signout">Sign out</a>`)
		w.AppendString(`<br>`)

		if session.ID == AdminID {
			DisplayIndexAdminPage(w, user)
		} else {
			DisplayIndexUserPage(w, user)
		}
	}

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}
