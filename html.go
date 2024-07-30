package main

import (
	"time"
	"unicode/utf8"

	"github.com/anton2920/gofa/database"
	"github.com/anton2920/gofa/net/html"
	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/prof"
	"github.com/anton2920/gofa/strings"
)

const (
	CSSEnabled = true
	JSEnabled  = CSSEnabled && false
)

const (
	WidthSmall  = 4
	WidthMedium = 6
	WidthLarge  = 8
)

func DisplayHTMLStart(w *http.Response) {
	defer prof.End(prof.Begin(""))

	w.WriteString(html.Header)
	w.WriteString(`<html lang="en" data-bs-theme="light">`)
}

func DisplayHeadStart(w *http.Response) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`<head>`)
	w.WriteString(`<meta charset="utf-8"/>`)
	w.WriteString(`<meta name="viewport" content="width=device-width, initial-scale=1"/>`)

	if CSSEnabled {
		w.WriteString(`<link rel="stylesheet" href="/fs/bootstrap.min.css"/>`)
		w.WriteString(`<style>.navbar-custom {position: fixed; z-index: 190; }</style>`)
	}
	if JSEnabled {
		w.WriteString(`<script src="/fs/bootstrap.min.js"></script>`)
	}
}

func DisplayHeadEnd(w *http.Response) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`</head>`)
}

func DisplayBodyStart(w *http.Response) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`<body class="bg-body-secondary">`)
}

func DisplayHeader(w *http.Response, l Language) {
	defer prof.End(prof.Begin(""))

	if CSSEnabled {
		w.WriteString(`<header class="navbar navbar-dark sticky-top bg-dark flex-md-nowrap p-0 shadow fixed-top">`)

		w.WriteString(`<a class="navbar-brand col-md-3 col-lg-2 me-0 px-3 fs-6 text-center" href="/">`)
		w.WriteString(Ls(l, "Master's degree"))
		w.WriteString(`</a>`)

		w.WriteString(`</header>`)
	}
}

func DisplaySidebarStart(w *http.Response) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`<nav id="sidebarMenu" class="col-md-3 col-lg-2 d-md-block bg-body-tertiary vh-100 sidebar collapse navbar-custom">`)
	w.WriteString(`<div class="position-sticky pt-3 sidebar-sticky">`)
}

func DisplaySidebarListStart(w *http.Response) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`<ul class="nav flex-column">`)
}

func DisplaySidebarListEnd(w *http.Response) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`</ul>`)
}

func DisplaySidebarLink(w *http.Response, l Language, href string, text string) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`<a class="nav-link" href="`)
	w.WriteString(href)
	w.WriteString(`">`)
	w.WriteString(Ls(l, text))
	w.WriteString(`</a>`)
}

func DisplaySidebarLinkIDName(w *http.Response, l Language, prefix string, id database.ID, text string, i int, name string) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`<a class="nav-link" href="`)
	w.WriteString(prefix)
	w.WriteString(`/`)
	w.WriteID(id)
	w.WriteString(`">`)
	w.WriteString(Ls(l, text))
	w.WriteString(` #`)
	w.WriteInt(i + 1)
	w.WriteString(`: `)
	DisplayShortenedString(w, name, 25)
	w.WriteString(`</a>`)
}

func DisplaySidebarUser(w *http.Response, l Language, user *User) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`<div><div class="text-center"><p class="nav-link link-offset-2 link-underline-opacity-25 link-underline-opacity-100-hover">`)
	w.WriteString(`<a class="nav-link" href="/user/`)
	w.WriteID(user.ID)
	w.WriteString(`">`)
	DisplayUserTitle(w, l, user)
	w.WriteString(`</a>`)
	w.WriteString(`</p></div></div>`)
}

func DisplaySidebarEnd(w *http.Response) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`</div></nav>`)
}

func DisplaySidebar(w *http.Response, l Language, session *Session) {
	defer prof.End(prof.Begin(""))

	if CSSEnabled {
		DisplaySidebarStart(w)
		{
			session.Lock()
			DisplaySidebarUser(w, l, &session.User)
			session.Unlock()

			w.WriteString(`<hr>`)
			DisplaySidebarListStart(w)
			{
				if session.ID == AdminID {
					DisplaySidebarLink(w, l, "/users", "Users")
				}
				DisplaySidebarLink(w, l, "/groups", "Groups")
				DisplaySidebarLink(w, l, "/courses", "Courses")
				DisplaySidebarLink(w, l, "/subjects", "Subjects")
				if session.ID != AdminID {
					DisplaySidebarLink(w, l, "/steps", "Steps")
				}
				w.WriteString(`<hr>`)
				DisplaySidebarLink(w, l, APIPrefix+"/user/signout", "Sign out")
			}
			DisplaySidebarListEnd(w)
		}
		DisplaySidebarEnd(w)
	}
}

func DisplaySidebarWithLessons(w *http.Response, l Language, session *Session, lessons []database.ID) {
	defer prof.End(prof.Begin(""))

	if CSSEnabled {
		DisplaySidebarStart(w)
		{
			session.Lock()
			DisplaySidebarUser(w, l, &session.User)
			session.Unlock()

			w.WriteString(`<hr>`)
			DisplaySidebarListStart(w)
			{
				if session.ID == AdminID {
					DisplaySidebarLink(w, l, "/users", "Users")
				}
				DisplaySidebarLink(w, l, "/groups", "Groups")
				DisplaySidebarLink(w, l, "/courses", "Courses")
				DisplaySidebarLink(w, l, "/subjects", "Subjects")
				if session.ID != AdminID {
					DisplaySidebarLink(w, l, "/steps", "Steps")
				}
				w.WriteString(`<hr>`)
				for i := 0; i < len(lessons); i++ {
					var lesson Lesson
					if err := GetLessonByID(lessons[i], &lesson); err != nil {
						/* TODO(anton2920): report error. */
					}
					DisplaySidebarLinkIDName(w, l, "/lesson", lessons[i], "Lesson", i, lesson.Name)
				}
				w.WriteString(`<hr>`)
				DisplaySidebarLink(w, l, APIPrefix+"/user/signout", "Sign out")
			}
			DisplaySidebarListEnd(w)
		}
		DisplaySidebarEnd(w)
	}
}

func DisplayMainStart(w *http.Response) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`<main class="col-md-9 ms-sm-auto col-lg-10 px-md-2 mt-5">`)
}

func DisplayCrumbsStart(w *http.Response, width int) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`<nav aria-label="breadcrumb" class="col-lg-`)
	w.WriteInt(width)
	w.WriteString(` mx-auto" style="--bs-breadcrumb-divider: url(&#34;data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='8' height='8'%3E%3Cpath d='M2.5 0L1 1.5 3.5 4 1 6.5 2.5 8l4-4-4-4z' fill='%236c757d'/%3E%3C/svg%3E&#34;);">`)
	w.WriteString(`<ol class="breadcrumb breadcrumb-chevron p-3 bg-body-tertiary rounded-2 border">`)
	w.WriteString(`<li class="breadcrumb-item"><a class="link-body-emphasis" href="/"><svg xmlns="http://www.w3.org/2000/svg" x="0px" y="0px" width="16" height="17" viewBox="0 0 24 24"><path d="M 12 2.0996094 L 1 12 L 4 12 L 4 21 L 11 21 L 11 15 L 13 15 L 13 21 L 20 21 L 20 12 L 23 12 L 12 2.0996094 z M 12 4.7910156 L 18 10.191406 L 18 11 L 18 19 L 15 19 L 15 13 L 9 13 L 9 19 L 6 19 L 6 10.191406 L 12 4.7910156 z"></path></svg><span class="visually-hidden">Home</span></a></li>`)
}

func DisplayCrumbsLinkIDStart(w *http.Response, prefix string, id database.ID) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`<li class="breadcrumb-item">`)
	w.WriteString(`<a class="link-body-emphasis text-decoration-none" href="`)
	w.WriteString(prefix)
	w.WriteString(`/`)
	w.WriteID(id)
	w.WriteString(`">`)
}

func DisplayCrumbsLinkEnd(w *http.Response) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`</a>`)
	w.WriteString(`</li>`)
}

func DisplayCrumbsLinkID(w *http.Response, prefix string, id database.ID, title string) {
	defer prof.End(prof.Begin(""))

	DisplayCrumbsLinkIDStart(w, prefix, id)
	w.WriteString(title)
	DisplayCrumbsLinkEnd(w)
}

func DisplayCrumbsLink(w *http.Response, l Language, href string, title string) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`<li class="breadcrumb-item">`)
	w.WriteString(`<a class="link-body-emphasis text-decoration-none" href="`)
	w.WriteString(href)
	w.WriteString(`">`)
	w.WriteString(Ls(l, title))
	DisplayCrumbsLinkEnd(w)
}

func DisplayCrumbsSubmitRaw(w *http.Response, l Language, nextPage, title string) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`<li class="breadcrumb-item">`)
	w.WriteString(`<button style="border: 0; vertical-align: top" class="btn btn-link link-body-emphasis text-decoration-none p-0" name="NextPage" value="`)
	w.WriteString(Ls(l, nextPage))
	w.WriteString(`" formnovalidate>`)
	w.WriteString(title)
	w.WriteString(`</button>`)
	w.WriteString(`</li>`)
}

func DisplayCrumbsSubmit(w *http.Response, l Language, nextPage, title string) {
	defer prof.End(prof.Begin(""))

	DisplayCrumbsSubmitRaw(w, l, nextPage, Ls(l, title))
}

func DisplayCrumbsItemStart(w *http.Response) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`<li class="breadcrumb-item fw-semibold" aria-current="page">`)
}

func DisplayCrumbsItemEnd(w *http.Response) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`</li>`)
}

func DisplayCrumbsItemRaw(w *http.Response, title string) {
	defer prof.End(prof.Begin(""))

	DisplayCrumbsItemStart(w)
	w.WriteString(title)
	DisplayCrumbsItemEnd(w)
}

func DisplayCrumbsItem(w *http.Response, l Language, title string) {
	defer prof.End(prof.Begin(""))

	DisplayCrumbsItemRaw(w, Ls(l, title))
}

func DisplayCrumbsEnd(w *http.Response) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`</ol></nav>`)
}

func DisplayPageStart(w *http.Response, width int) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`<div class="p-4 p-md-5 border rounded-2 bg-body-tertiary col-md-10 mx-auto col-lg-`)
	w.WriteInt(width)
	w.WriteString(`">`)
}

func DisplayPageEnd(w *http.Response) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`</div>`)
}

func DisplayFormPageStart(w *http.Response, r *http.Request, l Language, width int, title string, endpoint string, err error) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`<form class="p-4 p-md-5 border rounded-2 bg-body-tertiary mx-auto col-lg-`)
	w.WriteInt(width)
	w.WriteString(`" method="POST" action="`)
	w.WriteString(endpoint)
	w.WriteString(`">`)

	DisplayFormTitle(w, l, title, err)

	DisplayHiddenString(w, "ID", r.Form.Get("ID"))
}

func DisplayFormStart(w *http.Response, r *http.Request, endpoint string) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`<form method="POST" action="`)
	w.WriteString(endpoint)
	w.WriteString(`">`)

	DisplayHiddenString(w, "ID", r.Form.Get("ID"))
}

func DisplayFormTitle(w *http.Response, l Language, title string, err error) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`<h3 class="text-center">`)
	w.WriteString(Ls(l, title))
	w.WriteString(`</h3>`)
	w.WriteString(`<br>`)

	DisplayError(w, l, err)
}

func DisplayFormEnd(w *http.Response) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`</form>`)
}

func DisplayFormPageEnd(w *http.Response) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`</form>`)
}

func DisplayMainEnd(w *http.Response) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`</main>`)
}

func DisplayBodyEnd(w *http.Response) {
	defer prof.End(prof.Begin(""))

	if JSEnabled {
		w.WriteString(`<div class="dropdown position-fixed bottom-0 end-0 mb-3 me-3 bd-mode-toggle">`)
		w.WriteString(`<input type="checkbox" class="btn-check" id="btn-toggle" onclick="function toggleTheme() { var html = document.querySelector('html'); html.setAttribute('data-bs-theme', html.getAttribute('data-bs-theme') === 'dark' ? 'light' : 'dark'); } toggleTheme()"/>`)
		w.WriteString(`<label style="cursor: pointer" for="btn-toggle">`)
		w.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-circle-half" viewBox="0 0 16 16"> <path d="M8 15A7 7 0 1 0 8 1v14zm0 1A8 8 0 1 1 8 0a8 8 0 0 1 0 16z"/></svg>`)
		w.WriteString(`</label>`)
		w.WriteString(`</div>`)
	}

	w.WriteString(`</body>`)
}

func DisplayHTMLEnd(w *http.Response) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`</html>`)
}

func DisplayFormattedTime(w *http.Response, t int64) {
	defer prof.End(prof.Begin(""))

	w.Write(time.Unix(t, 0).AppendFormat(make([]byte, 0, 20), "2006/01/02 15:04:05"))
}

func DisplayDeleted(w *http.Response, l Language, deleted bool) {
	defer prof.End(prof.Begin(""))

	if deleted {
		w.WriteString(` [`)
		w.WriteString(Ls(l, "deleted"))
		w.WriteString(`]`)
	}
}

func DisplayDraft(w *http.Response, l Language, draft bool) {
	defer prof.End(prof.Begin(""))

	if draft {
		w.WriteString(` (`)
		w.WriteString(Ls(l, "draft"))
		w.WriteString(`)`)
	}
}

func DisplayFrameStart(w *http.Response) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`<div class="border round p-4">`)
}

func DisplayFrameEnd(w *http.Response) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`</div>`)
	w.WriteString(`<br>`)
}

func DisplayTableStart(w *http.Response, l Language, cols []string) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`<table class="table table-bordered table-stripped table-hover">`)

	w.WriteString(`<thead>`)
	w.WriteString(`<tr>`)
	for i := 0; i < len(cols); i++ {
		w.WriteString(`<th class="text-center" scope="col">`)
		w.WriteString(Ls(l, cols[i]))
		w.WriteString(`</th>`)
	}
	w.WriteString(`</tr>`)
	w.WriteString(`</thead>`)

	w.WriteString(`<tbody>`)
}

func DisplayTableRowStart(w *http.Response) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`<tr>`)
}

func DisplayTableRowLinkIDStart(w *http.Response, prefix string, id database.ID) {
	defer prof.End(prof.Begin(""))

	DisplayTableRowStart(w)

	w.WriteString(`<th class="text-center align-middle" scope="row">`)
	w.WriteString(`<a href="`)
	w.WriteString(prefix)
	w.WriteString(`/`)
	w.WriteID(id)
	w.WriteString(`">`)
	w.WriteID(id)
	w.WriteString(`</a>`)
	w.WriteString(`</th>`)
}

func DisplayTableItemStart(w *http.Response) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`<td class="text-center align-middle">`)
}

func DisplayTableItemID(w *http.Response, id database.ID) {
	defer prof.End(prof.Begin(""))

	DisplayTableItemStart(w)
	w.WriteID(id)
	DisplayTableItemEnd(w)
}

func DisplayTableItemInt(w *http.Response, x int) {
	defer prof.End(prof.Begin(""))

	DisplayTableItemStart(w)
	w.WriteInt(x)
	DisplayTableItemEnd(w)
}

func DisplayTableItemString(w *http.Response, s string) {
	defer prof.End(prof.Begin(""))

	DisplayTableItemStart(w)
	w.WriteHTMLString(s)
	DisplayTableItemEnd(w)
}

func DisplayTableItemTime(w *http.Response, t int64) {
	defer prof.End(prof.Begin(""))

	DisplayTableItemStart(w)
	DisplayFormattedTime(w, t)
	DisplayTableItemEnd(w)
}

func DisplayTableItemFlags(w *http.Response, l Language, flags int32) {
	defer prof.End(prof.Begin(""))

	DisplayTableItemStart(w)
	switch flags {
	case 0: /* active */
		w.WriteString(`<small class="d-inline-flex px-2 py-1 fw-semibold text-success-emphasis bg-success-subtle border border-success-subtle rounded-2">`)
		w.WriteString(Ls(l, "Active"))
		w.WriteString(`</small>`)
	case 1: /* deleted */
		w.WriteString(`<small class="d-inline-flex px-2 py-1 fw-semibold text-danger-emphasis bg-danger-subtle border border-danger-subtle rounded-2">`)
		w.WriteString(Ls(l, "Deleted"))
		w.WriteString(`</small>`)
	case 2: /* draft */
		w.WriteString(`<small class="d-inline-flex px-2 py-1 fw-semibold text-primary-emphasis bg-primary-subtle border border-primary-subtle rounded-2">`)
		w.WriteString(Ls(l, "Draft"))
		w.WriteString(`</small>`)
	}
	DisplayTableItemEnd(w)
}

func DisplayTableItemEnd(w *http.Response) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`</td>`)
}

func DisplayTableRowEnd(w *http.Response) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`</tr>`)
}

func DisplayTableEnd(w *http.Response) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`</tbody>`)
	w.WriteString(`</table>`)
}

func DisplayLabel(w *http.Response, l Language, text string) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`<label class="form-label">`)
	w.WriteString(Ls(l, text))
	w.WriteString(`:<br>`)
	w.WriteString(`</label>`)
}

func DisplayInput(w *http.Response, t string, name, value string, required bool) {
	defer prof.End(prof.Begin(""))

	w.WriteString(` <input class="form-control" type="`)
	w.WriteString(t)
	w.WriteString(`" name="`)
	w.WriteString(name)
	w.WriteString(`" value="`)
	w.WriteHTMLString(value)
	w.WriteString(`"`)
	if required {
		w.WriteString(` required`)
	}
	w.WriteString(`>`)
}

func DisplayConstraintInput(w *http.Response, t string, minLength, maxLength int, name, value string, required bool) {
	defer prof.End(prof.Begin(""))

	w.WriteString(` <input class="form-control" type="`)
	w.WriteString(t)
	w.WriteString(`" minlength="`)
	w.WriteInt(minLength)
	w.WriteString(`" maxlength="`)
	w.WriteInt(maxLength)
	w.WriteString(`" name="`)
	w.WriteString(name)
	w.WriteString(`" value="`)
	w.WriteHTMLString(value)
	w.WriteString(`"`)
	if required {
		w.WriteString(` required`)
	}
	w.WriteString(`>`)
}

func DisplayConstraintIndexedInput(w *http.Response, t string, minLength, maxLength int, name string, index int, value string, required bool) {
	defer prof.End(prof.Begin(""))

	w.WriteString(` <input class="input-field" type="`)
	w.WriteString(t)
	w.WriteString(`" minlength="`)
	w.WriteInt(minLength)
	w.WriteString(`" maxlength="`)
	w.WriteInt(maxLength)
	w.WriteString(`" name="`)
	w.WriteString(name)
	w.WriteInt(index)
	w.WriteString(`" value="`)
	w.WriteHTMLString(value)
	w.WriteString(`"`)
	if required {
		w.WriteString(` required`)
	}
	w.WriteString(`>`)
}

func DisplayConstraintInlineTextarea(w *http.Response, minLength, maxLength int, name, value string, required bool) {
	defer prof.End(prof.Begin(""))

	w.WriteString(` <textarea class="btn btn-outline-dark" rows="1" minlength="`)
	w.WriteInt(minLength)
	w.WriteString(`" maxlength="`)
	w.WriteInt(maxLength)
	w.WriteString(`" name="`)
	w.WriteString(name)
	w.WriteString(`"`)
	if required {
		w.WriteString(` required`)
	}
	w.WriteString(`>`)
	w.WriteHTMLString(value)
	w.WriteString(`</textarea>`)
}

func DisplayConstraintTextarea(w *http.Response, minLength, maxLength int, name, value string, required bool) {
	defer prof.End(prof.Begin(""))

	w.WriteString(` <textarea class="form-control" rows="10" minlength="`)
	w.WriteInt(minLength)
	w.WriteString(`" maxlength="`)
	w.WriteInt(maxLength)
	w.WriteString(`" name="`)
	w.WriteString(name)
	w.WriteString(`"`)
	if required {
		w.WriteString(` required`)
	}
	w.WriteString(`>`)
	w.WriteHTMLString(value)
	w.WriteString(`</textarea>`)
}

func DisplayHiddenID(w *http.Response, name string, id database.ID) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`<input type="hidden" name="`)
	w.WriteString(name)
	w.WriteString(`" value="`)
	w.WriteID(id)
	w.WriteString(`">`)
}

func DisplayHiddenInt(w *http.Response, name string, i int) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`<input type="hidden" name="`)
	w.WriteString(name)
	w.WriteString(`" value="`)
	w.WriteInt(i)
	w.WriteString(`">`)
}

func DisplayHiddenString(w *http.Response, name string, value string) {
	defer prof.End(prof.Begin(""))

	w.WriteString(`<input type="hidden" name="`)
	w.WriteString(name)
	w.WriteString(`" value="`)
	w.WriteHTMLString(value)
	w.WriteString(`">`)
}

func DisplayCommand(w *http.Response, l Language, command string) {
	defer prof.End(prof.Begin(""))

	w.WriteString(` <input class="btn btn-outline-dark" type="submit" name="Command" value="`)
	w.WriteString(Ls(l, command))
	w.WriteString(`" formnovalidate>`)
}

func DisplayIndexedCommand(w *http.Response, l Language, index int, command string) {
	defer prof.End(prof.Begin(""))

	w.WriteString(` <input class="btn btn-outline-dark" type="submit" name="Command`)
	w.WriteInt(index)
	w.WriteString(`" value="`)
	w.WriteString(Ls(l, command))
	w.WriteString(`" formnovalidate>`)
}

func DisplayDoublyIndexedCommand(w *http.Response, l Language, pindex, sindex int, command string) {
	defer prof.End(prof.Begin(""))

	w.WriteString(` <input class="btn btn-outline-dark" type="submit" name="Command`)
	w.WriteInt(pindex)
	w.WriteString(`.`)
	w.WriteInt(sindex)
	w.WriteString(`" value="`)
	w.WriteString(Ls(l, command))
	w.WriteString(`" formnovalidate>`)
}

func DisplayButton(w *http.Response, l Language, name string, value string) {
	defer prof.End(prof.Begin(""))

	w.WriteString(` <input class="btn btn-outline-dark" type="submit" name="`)
	w.WriteString(name)
	w.WriteString(`" value="`)
	w.WriteString(Ls(l, value))
	w.WriteString(`" formnovalidate>`)
}

func DisplayNextPage(w *http.Response, l Language, value string) {
	defer prof.End(prof.Begin(""))

	w.WriteString(` <input class="w-100 btn btn-outline-primary mt-2" type="submit" name="NextPage" value="`)
	w.WriteString(Ls(l, value))
	w.WriteString(`" formnovalidate>`)
}

func DisplaySubmit(w *http.Response, l Language, name string, value string, verify bool) {
	defer prof.End(prof.Begin(""))

	w.WriteString(` <input class="w-100 mb-2 btn btn-lg rounded-3 btn-primary" type="submit" name="`)
	w.WriteString(name)
	w.WriteString(`" value="`)
	w.WriteString(Ls(l, value))
	w.WriteString(`"`)
	if !verify {
		w.WriteString(` formnovalidate`)
	}
	w.WriteString(`>`)
}

func DisplayShortenedString(w *http.Response, s string, maxVisibleLen int) {
	defer prof.End(prof.Begin(""))

	if utf8.RuneCountInString(s) < maxVisibleLen {
		w.WriteHTMLString(s)
	} else {
		space := strings.FindCharReverse(s[:maxVisibleLen], ' ')
		if space == -1 {
			w.WriteHTMLString(s[:maxVisibleLen])
		} else {
			w.WriteHTMLString(s[:space])
		}
		w.WriteString(`...`)
	}
}

func DisplayMarkdown(w *http.Response, md string) {
	defer prof.End(prof.Begin(""))

	type Token struct {
		Type   int
		Start  string
		End    string
		RStart string
		REnd   string
	}

	const (
		None = iota
		Code
		H3
		H2
		H1
		List
		Mono
		Bold
		Italics
		Break
	)

	tokens := [...]Token{
		{Code, "```\r\n", "\r\n```\r\n", "<pre><code>", "</code></pre>"},
		{H3, "###", "\r\n", "<h6>", "</h6>"},
		{H2, "##", "\r\n", "<h5>", "</h5>"},
		{H1, "#", "\r\n", "<h4>", "</h4>"},
		{List, "\n-", "\r", `<li class="ms-4">`, "</li>"},
		{Mono, "`", "`", "<tt>", "</tt>"},
		{Bold, "**", "**", "<b>", "</b>"},
		{Bold, "__", "__", "<b>", "</b>"},
		{Italics, "*", "*", "<i>", "</i>"},
		{Italics, "_", "_", "<i>", "</i>"},
		{Break, "\r", "\n", "<br>", ""},
	}

	for len(md) > 0 {
		var replaced bool

		for i := 0; i < len(tokens); i++ {
			tok := &tokens[i]

			start := strings.FindSubstring(md, tok.Start)
			if start == -1 {
				continue
			}

			end := strings.FindSubstring(md[start+len(tok.Start):], tok.End)
			if end == -1 {
				continue
			}
			end += start + len(tok.Start)

			if end-start == 0 {
				continue
			}
			inside := md[start+len(tok.Start) : end]

			DisplayMarkdown(w, md[:start])

			w.WriteString(tok.RStart)
			switch tok.Type {
			default:
				DisplayMarkdown(w, inside)
			case Code, Mono:
				w.WriteHTMLString(inside)
			}
			w.WriteString(tok.REnd)

			md = md[end+len(tok.End):]
			replaced = true
			break
		}

		if !replaced {
			w.WriteHTMLString(md)
			break
		}
	}
}
