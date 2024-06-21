package main

import "github.com/anton2920/gofa/net/http"

func DisplayIndexStart(w *http.Response) {
	w.AppendString(`<div class="container-fluid">`)
	w.AppendString(`<div class="row">`)
}

func DisplayIndexButtonsStart(w *http.Response, l Language, title string) {
	w.AppendString(`<div aria-live="polite" class="position-relative">`)
	w.AppendString(`<div class="top-0 end-0 p-3">`)
	w.AppendString(`<main class="col-md-9 ms-sm-auto col-lg-10 px-md-4">`)
	w.AppendString(`<div class="container py-4">`)

	w.AppendString(`<h2 class="text-center w-100 mb-3">`)
	w.AppendString(Ls(l, title))
	w.AppendString(`</h2>`)
	w.AppendString(`<br>`)

	w.AppendString(`<div class="row align-items-md-stretch mb-4">`)
}

func DisplayIndexButton(w *http.Response, l Language, href string, h2 string, p string) {
	w.AppendString(`<div class="col-md-6 mb-2">`)
	w.AppendString(`<div class="h-100 p-5 bg-body-tertiary border rounded-3">`)

	w.AppendString(`<h2>`)
	w.AppendString(Ls(l, h2))
	w.AppendString(`</h2>`)

	w.AppendString(`<p>`)
	w.AppendString(Ls(l, p))
	w.AppendString(`</p>`)

	w.AppendString(`<a class="btn btn-outline-primary" type="button" href="`)
	w.AppendString(href)
	w.AppendString(`">`)
	w.AppendString(Ls(l, "Open"))
	w.AppendString(`</a>`)

	w.AppendString(`</div>`)
	w.AppendString(`</div>`)
}

func DisplayIndexButtonsEnd(w *http.Response) {
	w.AppendString(`</div></div></main></div></div>`)
}

func DisplayIndexEnd(w *http.Response) {
	w.AppendString(`</div></div>`)
}

func DisplayIndexAdminPage(w *http.Response, l Language, user *User) {
	DisplayIndexStart(w)

	DisplaySidebarStart(w)
	{
		DisplaySidebarUser(w, l, user)
		w.AppendString(`<hr>`)
		DisplaySidebarListStart(w)
		{
			DisplaySidebarLink(w, l, "/users", "Users")
			DisplaySidebarLink(w, l, "/groups", "Groups")
			DisplaySidebarLink(w, l, "/courses", "Courses")
			DisplaySidebarLink(w, l, "/subjects", "Subjects")
			w.AppendString(`<hr>`)
			DisplaySidebarLink(w, l, APIPrefix+"/user/signout", "Sign out")
		}
		DisplaySidebarListEnd(w)
	}
	DisplaySidebarEnd(w)

	DisplayIndexButtonsStart(w, l, "Administration")
	{
		DisplayIndexButton(w, l, "/users", "Users", "Create, edit and delete users")
		DisplayIndexButton(w, l, "/groups", "Groups", "Create, edit and delete groups")
		DisplayIndexButton(w, l, "/courses", "Courses", "Create, edit and delete courses")
		DisplayIndexButton(w, l, "/subjects", "Subjects", "Create, edit and delete subjects")
	}
	DisplayIndexButtonsEnd(w)

	DisplayIndexEnd(w)
}

func DisplayIndexUserPage(w *http.Response, l Language, user *User) {
	DisplayUserGroups(w, l, user.ID)
	DisplayUserCourses(w, l, user)
	DisplayUserSubjects(w, l, user.ID)
}

func DisplayIndexTitle(w *http.Response, l Language) {
	w.AppendString(Ls(l, "Master's degree"))
}

func IndexPageHandler(w *http.Response, r *http.Request) error {
	var user User

	session, err := GetSessionFromRequest(r)
	if err != nil {
		w.Redirect("/user/signin", http.StatusSeeOther)
		return nil
	}

	DisplayHTMLHeader(w)
	w.AppendString(`<head>`)
	w.AppendString(`<title>`)
	DisplayIndexTitle(w, GL)
	w.AppendString(`</title>`)
	DisplayCSS(w)
	DisplayJS(w)
	w.AppendString(`</head>`)

	w.AppendString(`<body class="bg-body-secondary">`)

	DisplayHeader(w, GL)

	if err := GetUserByID(session.ID, &user); err != nil {
		return http.ServerError(err)
	}

	if session.ID == AdminID {
		DisplayIndexAdminPage(w, GL, &user)
	} else {
		DisplayIndexUserPage(w, GL, &user)
	}

	DisplayThemeToggle(w)
	w.AppendString(`</body>`)
	w.AppendString(`</html>`)

	return nil
}
