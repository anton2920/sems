package main

import (
	"time"

	"github.com/anton2920/gofa/database"
	"github.com/anton2920/gofa/net/html"
	"github.com/anton2920/gofa/net/http"
)

func DisplayHTMLStart(w *http.Response) {
	w.AppendString(html.Header)
	w.AppendString(`<html lang="en" data-bs-theme="light">`)
}

func DisplayHeadStart(w *http.Response) {
	w.AppendString(`<head>`)
	w.AppendString(`<meta charset="utf-8"/>`)
	w.AppendString(`<meta name="viewport" content="width=device-width, initial-scale=1"/>`)

	w.AppendString(`<link rel="stylesheet" href="/fs/bootstrap.min.css"/>`)
	w.AppendString(`<style>.navbar-custom {position: fixed; z-index: 190; }</style>`)
	// w.AppendString(`<script src="/fs/bootstrap.min.js"></script>`)
}

func DisplayHeadEnd(w *http.Response) {
	w.AppendString(`</head>`)
}

func DisplayBodyStart(w *http.Response) {
	w.AppendString(`<body class="bg-body-secondary">`)
}

func DisplayHeader(w *http.Response, l Language) {
	w.AppendString(`<header class="navbar navbar-dark sticky-top bg-dark flex-md-nowrap p-0 shadow fixed-top">`)

	w.AppendString(`<a class="navbar-brand col-md-3 col-lg-2 me-0 px-3 fs-6 text-center" href="/">`)
	w.AppendString(Ls(l, "Master's degree"))
	w.AppendString(`</a>`)

	w.AppendString(`</header>`)
}

func DisplaySidebarStart(w *http.Response) {
	w.AppendString(`<nav id="sidebarMenu" class="col-md-3 col-lg-2 d-md-block bg-body-tertiary vh-100 sidebar collapse navbar-custom">`)
	w.AppendString(`<div class="position-sticky pt-3 sidebar-sticky">`)
}

func DisplaySidebarListStart(w *http.Response) {
	w.AppendString(`<ul class="nav flex-column">`)
}

func DisplaySidebarListEnd(w *http.Response) {
	w.AppendString(`</ul>`)
}

func DisplaySidebarLink(w *http.Response, l Language, href string, text string) {
	w.AppendString(`<a class="nav-link" href="`)
	w.AppendString(href)
	w.AppendString(`">`)
	w.AppendString(Ls(l, text))
	w.AppendString(`</a>`)
}

func DisplaySidebarUser(w *http.Response, l Language, user *User) {
	w.AppendString(`<div><div class="text-center"><p class="nav-link link-offset-2 link-underline-opacity-25 link-underline-opacity-100-hover">`)
	w.AppendString(`<a class="nav-link" href="/user/`)
	w.WriteID(user.ID)
	w.AppendString(`">`)
	DisplayUserTitle(w, l, user)
	w.AppendString(`</a>`)
	w.AppendString(`</p></div></div>`)
}

func DisplaySidebarEnd(w *http.Response) {
	w.AppendString(`</div></nav>`)
}

func DisplaySidebar(w *http.Response, l Language, userID database.ID) {
	var user User

	if err := GetUserByID(userID, &user); err != nil {
		/* TODO(anton2920): report error. */
	}

	DisplaySidebarStart(w)
	{
		DisplaySidebarUser(w, l, &user)
		w.AppendString(`<hr>`)
		DisplaySidebarListStart(w)
		{
			if userID == AdminID {
				DisplaySidebarLink(w, l, "/users", "Users")
			}
			DisplaySidebarLink(w, l, "/groups", "Groups")
			DisplaySidebarLink(w, l, "/courses", "Courses")
			DisplaySidebarLink(w, l, "/subjects", "Subjects")
			w.AppendString(`<hr>`)
			DisplaySidebarLink(w, l, APIPrefix+"/user/signout", "Sign out")
		}
		DisplaySidebarListEnd(w)
	}
	DisplaySidebarEnd(w)
}

func DisplayPageStart(w *http.Response) {
	w.AppendString(`<main class="col-md-9 ms-sm-auto col-lg-10 px-md-2 mt-5">`)
	w.AppendString(`<div class="p-4 p-md-5 border rounded-2 bg-body-tertiary col-md-10 mx-auto col-lg-8">`)
}

func DisplayPageEnd(w *http.Response) {
	w.AppendString(`</div></main>`)
}

func DisplayFormStart(w *http.Response, endpoint string) {
	w.AppendString(`<main class="col-md-9 ms-sm-auto col-lg-10 px-md-2 mt-5">`)
	w.AppendString(`<form class="p-4 p-md-5 border rounded-2 bg-body-tertiary mx-auto col-lg-4" method="POST" action="`)
	w.AppendString(endpoint)
	w.AppendString(`">`)
}

func DisplayFormEnd(w *http.Response) {
	w.AppendString(`</form></main>`)
}

func DisplayBodyEnd(w *http.Response) {
	/*
		w.AppendString(`<div class="dropdown position-fixed bottom-0 end-0 mb-3 me-3 bd-mode-toggle">`)
		w.AppendString(`<input type="checkbox" class="btn-check" id="btn-toggle" onclick="function toggleTheme() { var html = document.querySelector('html'); html.setAttribute('data-bs-theme', html.getAttribute('data-bs-theme') === 'dark' ? 'light' : 'dark'); } toggleTheme()"/>`)
		w.AppendString(`<label style="cursor: pointer" for="btn-toggle">`)
		w.AppendString(`<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-circle-half" viewBox="0 0 16 16"> <path d="M8 15A7 7 0 1 0 8 1v14zm0 1A8 8 0 1 1 8 0a8 8 0 0 1 0 16z"/></svg>`)
		w.AppendString(`</label>`)
		w.AppendString(`</div>`)
	*/

	w.AppendString(`</body>`)
}

func DisplayHTMLEnd(w *http.Response) {
	w.AppendString(`</html>`)
}

func DisplayFormattedTime(w *http.Response, t int64) {
	w.Write(time.Unix(t, 0).AppendFormat(make([]byte, 0, 20), "2006/01/02 15:04:05"))
}

func DisplayDeleted(w *http.Response, l Language, deleted bool) {
	if deleted {
		w.AppendString(` [`)
		w.AppendString(Ls(l, "deleted"))
		w.AppendString(`]`)
	}
}

func DisplayDraft(w *http.Response, l Language, draft bool) {
	if draft {
		w.AppendString(` (`)
		w.AppendString(Ls(l, "draft"))
		w.AppendString(`)`)
	}
}

func DisplayInputLabel(w *http.Response, l Language, text string) {
	w.AppendString(`<label class="form-label">`)
	w.AppendString(Ls(l, text))
	w.AppendString(`:<br>`)
	w.AppendString(`</label>`)
}

func DisplayInput(w *http.Response, t string, name, value string, required bool) {
	w.AppendString(` <input class="form-control" type="`)
	w.AppendString(t)
	w.AppendString(`" name="`)
	w.AppendString(name)
	w.AppendString(`" value="`)
	w.WriteHTMLString(value)
	w.AppendString(`"`)
	if required {
		w.AppendString(` required`)
	}
	w.AppendString(`>`)
}

func DisplayConstraintInput(w *http.Response, t string, minLength, maxLength int, name, value string, required bool) {
	w.AppendString(` <input class="form-control" type="`)
	w.AppendString(t)
	w.AppendString(`" minlength="`)
	w.WriteInt(minLength)
	w.AppendString(`" maxlength="`)
	w.WriteInt(maxLength)
	w.AppendString(`" name="`)
	w.AppendString(name)
	w.AppendString(`" value="`)
	w.WriteHTMLString(value)
	w.AppendString(`"`)
	if required {
		w.AppendString(` required`)
	}
	w.AppendString(`>`)
}

func DisplayConstraintIndexedInput(w *http.Response, t string, minLength, maxLength int, name string, index int, value string, required bool) {
	w.AppendString(` <input class="form-control" type="`)
	w.AppendString(t)
	w.AppendString(`" minlength="`)
	w.WriteInt(minLength)
	w.AppendString(`" maxlength="`)
	w.WriteInt(maxLength)
	w.AppendString(`" name="`)
	w.AppendString(name)
	w.WriteInt(index)
	w.AppendString(`" value="`)
	w.WriteHTMLString(value)
	w.AppendString(`"`)
	if required {
		w.AppendString(` required`)
	}
	w.AppendString(`>`)
}

func DisplayConstraintTextarea(w *http.Response, cols, rows string, minLength, maxLength int, name, value string, required bool) {
	w.AppendString(` <textarea cols="`)
	w.AppendString(cols)
	w.AppendString(`" rows="`)
	w.AppendString(rows)
	w.AppendString(`" minlength="`)
	w.WriteInt(minLength)
	w.AppendString(`" maxlength="`)
	w.WriteInt(maxLength)
	w.AppendString(`" name="`)
	w.AppendString(name)
	w.AppendString(`"`)
	if required {
		w.AppendString(` required`)
	}
	w.AppendString(`>`)
	w.WriteHTMLString(value)
	w.AppendString(`</textarea>`)
}

func DisplayHiddenID(w *http.Response, name string, id database.ID) {
	w.AppendString(`<input type="hidden" name="`)
	w.AppendString(name)
	w.AppendString(`" value="`)
	w.WriteID(id)
	w.AppendString(`">`)
}

func DisplayHiddenInt(w *http.Response, name string, i int) {
	w.AppendString(`<input type="hidden" name="`)
	w.AppendString(name)
	w.AppendString(`" value="`)
	w.WriteInt(i)
	w.AppendString(`">`)
}

func DisplayHiddenString(w *http.Response, name string, value string) {
	w.AppendString(`<input type="hidden" name="`)
	w.AppendString(name)
	w.AppendString(`" value="`)
	w.WriteHTMLString(value)
	w.AppendString(`">`)
}

func DisplayCommand(w *http.Response, l Language, command string) {
	w.AppendString(` <input class="btn btn-outline-dark" type="submit" name="Command" value="`)
	w.AppendString(Ls(l, command))
	w.AppendString(`" formnovalidate>`)
}

func DisplayIndexedCommand(w *http.Response, l Language, index int, command string) {
	w.AppendString(` <input class="btn btn-outline-dark" type="submit" name="Command`)
	w.WriteInt(index)
	w.AppendString(`" value="`)
	w.AppendString(Ls(l, command))
	w.AppendString(`" formnovalidate>`)
}

func DisplayDoublyIndexedCommand(w *http.Response, l Language, pindex, sindex int, command string) {
	w.AppendString(` <input class="btn btn-outline-dark" type="submit" name="Command`)
	w.WriteInt(pindex)
	w.AppendString(`.`)
	w.WriteInt(sindex)
	w.AppendString(`" value="`)
	w.AppendString(Ls(l, command))
	w.AppendString(`" formnovalidate>`)
}

func DisplayButton(w *http.Response, l Language, name string, value string) {
	w.AppendString(` <input class="btn btn-outline-dark" type="submit" name="`)
	w.AppendString(name)
	w.AppendString(`" value="`)
	w.AppendString(Ls(l, value))
	w.AppendString(`" formnovalidate>`)
}

func DisplaySubmit(w *http.Response, l Language, name string, value string, verify bool) {
	w.AppendString(` <input class="w-100 mb-2 btn btn-lg rounded-3 btn-primary" type="submit" name="`)
	w.AppendString(name)
	w.AppendString(`" value="`)
	w.AppendString(Ls(l, value))
	w.AppendString(`"`)
	if !verify {
		w.AppendString(` formnovalidate`)
	}
	w.AppendString(`>`)
}
