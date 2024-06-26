package main

import "github.com/anton2920/gofa/net/http"

func DisplayIndexButtonsStart(w *http.Response, l Language, title string) {
	w.AppendString(`<div class="container-fluid">`)
	w.AppendString(`<div class="row">`)

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
	w.AppendString(`</div>`)
	w.AppendString(`</div>`)
}

func IndexPageHandler(w *http.Response, r *http.Request) error {
	session, err := GetSessionFromRequest(r)
	if err != nil {
		w.Redirect("/user/signin", http.StatusSeeOther)
		return nil
	}

	DisplayHTMLStart(w)

	DisplayHeadStart(w)
	{
		w.AppendString(`<title>`)
		w.AppendString(Ls(GL, "Master's degree"))
		w.AppendString(`</title>`)
	}
	DisplayHeadEnd(w)

	DisplayBodyStart(w)
	{
		DisplayHeader(w, GL)
		DisplaySidebar(w, GL, session)

		DisplayIndexButtonsStart(w, GL, "Home page")
		{
			if session.ID == AdminID {
				DisplayIndexButton(w, GL, "/users", "Users", "Display information about users, as well as create, edit and delete them")
				DisplayIndexButton(w, GL, "/groups", "Groups", "Display information about groups, as well as create, edit and delete them")
				DisplayIndexButton(w, GL, "/courses", "Courses", "Display information about courses, as well as create, edit and delete them")
				DisplayIndexButton(w, GL, "/subjects", "Subjects", "Display information about subjects, as well as create, edit and delete them")
			} else {
				DisplayIndexButton(w, GL, "/groups", "Groups", "Display information about groups you are a part of")
				DisplayIndexButton(w, GL, "/courses", "Courses", "Display information about your courses, as well as create, edit and delete them")
				DisplayIndexButton(w, GL, "/subjects", "Subjects", "Display information about subjects, that your groups are studying")
				DisplayIndexButton(w, GL, "/steps", "Steps", "Display information about pending and completed steps")
			}
		}
		DisplayIndexButtonsEnd(w)
	}
	DisplayBodyEnd(w)

	DisplayHTMLEnd(w)
	return nil
}
