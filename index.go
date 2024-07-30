package main

import (
	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/prof"
)

func DisplayIndexButtonsStart(w *http.Response, l Language, title string) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`<div class="container-fluid">`)
	w.WriteString(`<div class="row">`)

	w.WriteString(`<div aria-live="polite" class="position-relative">`)
	w.WriteString(`<div class="top-0 end-0 p-3">`)
	w.WriteString(`<main class="col-md-9 ms-sm-auto col-lg-10 px-md-4">`)
	w.WriteString(`<div class="container py-4">`)

	w.WriteString(`<h2 class="text-center w-100 mb-3">`)
	w.WriteString(Ls(l, title))
	w.WriteString(`</h2>`)
	w.WriteString(`<br>`)

	w.WriteString(`<div class="row align-items-md-stretch mb-4">`)
}

func DisplayIndexButton(w *http.Response, l Language, href string, h2 string, p string) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`<div class="col-md-6 mb-2">`)
	w.WriteString(`<div class="h-100 p-5 bg-body-tertiary border rounded-3">`)

	w.WriteString(`<h2>`)
	w.WriteString(Ls(l, h2))
	w.WriteString(`</h2>`)

	w.WriteString(`<p>`)
	w.WriteString(Ls(l, p))
	w.WriteString(`</p>`)

	w.WriteString(`<a class="btn btn-outline-primary" type="button" href="`)
	w.WriteString(href)
	w.WriteString(`">`)
	w.WriteString(Ls(l, "Open"))
	w.WriteString(`</a>`)

	w.WriteString(`</div>`)
	w.WriteString(`</div>`)
}

func DisplayIndexButtonsEnd(w *http.Response) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`</div></div></main></div></div>`)
	w.WriteString(`</div>`)
	w.WriteString(`</div>`)
}

func IndexPageHandler(w *http.Response, r *http.Request) error {
	defer prof.End(prof.Begin(""))

	session, err := GetSessionFromRequest(r)
	if err != nil {
		w.Redirect("/user/signin", http.StatusSeeOther)
		return nil
	}

	DisplayHTMLStart(w)

	DisplayHeadStart(w)
	{
		w.WriteString(`<title>`)
		w.WriteString(Ls(GL, "Master's degree"))
		w.WriteString(`</title>`)
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
