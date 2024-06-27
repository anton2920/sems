package main

import (
	"time"
	"unicode/utf8"

	"github.com/anton2920/gofa/database"
	"github.com/anton2920/gofa/net/html"
	"github.com/anton2920/gofa/net/http"
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
	w.AppendString(html.Header)
	w.AppendString(`<html lang="en" data-bs-theme="light">`)
}

func DisplayHeadStart(w *http.Response) {
	w.AppendString(`<head>`)
	w.AppendString(`<meta charset="utf-8"/>`)
	w.AppendString(`<meta name="viewport" content="width=device-width, initial-scale=1"/>`)

	if CSSEnabled {
		w.AppendString(`<link rel="stylesheet" href="/fs/bootstrap.min.css"/>`)
		w.AppendString(`<style>.navbar-custom {position: fixed; z-index: 190; }</style>`)
	}
	if JSEnabled {
		w.AppendString(`<script src="/fs/bootstrap.min.js"></script>`)
	}
}

func DisplayHeadEnd(w *http.Response) {
	w.AppendString(`</head>`)
}

func DisplayBodyStart(w *http.Response) {
	w.AppendString(`<body class="bg-body-secondary">`)
}

func DisplayHeader(w *http.Response, l Language) {
	if CSSEnabled {
		w.AppendString(`<header class="navbar navbar-dark sticky-top bg-dark flex-md-nowrap p-0 shadow fixed-top">`)

		w.AppendString(`<a class="navbar-brand col-md-3 col-lg-2 me-0 px-3 fs-6 text-center" href="/">`)
		w.AppendString(Ls(l, "Master's degree"))
		w.AppendString(`</a>`)

		w.AppendString(`</header>`)
	}
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

func DisplaySidebarLinkID(w *http.Response, l Language, prefix string, id database.ID, text string, i int, name string) {
	w.AppendString(`<a class="nav-link" href="`)
	w.AppendString(prefix)
	w.AppendString(`/`)
	w.WriteID(id)
	w.AppendString(`">`)
	w.AppendString(Ls(l, text))
	w.AppendString(` #`)
	w.WriteInt(i + 1)
	w.AppendString(`: `)
	DisplayShortenedString(w, name, 25)
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

func DisplaySidebar(w *http.Response, l Language, session *Session) {
	if CSSEnabled {
		DisplaySidebarStart(w)
		{
			session.Lock()
			DisplaySidebarUser(w, l, &session.User)
			session.Unlock()

			w.AppendString(`<hr>`)
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
				w.AppendString(`<hr>`)
				DisplaySidebarLink(w, l, APIPrefix+"/user/signout", "Sign out")
			}
			DisplaySidebarListEnd(w)
		}
		DisplaySidebarEnd(w)
	}
}

func DisplaySidebarWithLessons(w *http.Response, l Language, session *Session, lessons []database.ID) {
	if CSSEnabled {
		DisplaySidebarStart(w)
		{
			session.Lock()
			DisplaySidebarUser(w, l, &session.User)
			session.Unlock()

			w.AppendString(`<hr>`)
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
				w.AppendString(`<hr>`)
				for i := 0; i < len(lessons); i++ {
					var lesson Lesson
					if err := GetLessonByID(lessons[i], &lesson); err != nil {
						/* TODO(anton2920): report error. */
					}
					DisplaySidebarLinkID(w, l, "/lesson", lessons[i], "Lesson", i, lesson.Name)
				}
				w.AppendString(`<hr>`)
				DisplaySidebarLink(w, l, APIPrefix+"/user/signout", "Sign out")
			}
			DisplaySidebarListEnd(w)
		}
		DisplaySidebarEnd(w)
	}
}

func DisplayMainStart(w *http.Response) {
	w.AppendString(`<main class="col-md-9 ms-sm-auto col-lg-10 px-md-2 mt-5">`)
}

func DisplayCrumbsStart(w *http.Response, width int) {
	w.AppendString(`<nav aria-label="breadcrumb" class="col-lg-`)
	w.WriteInt(width)
	w.AppendString(` mx-auto" style="--bs-breadcrumb-divider: url(&#34;data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='8' height='8'%3E%3Cpath d='M2.5 0L1 1.5 3.5 4 1 6.5 2.5 8l4-4-4-4z' fill='%236c757d'/%3E%3C/svg%3E&#34;);">`)
	w.AppendString(`<ol class="breadcrumb breadcrumb-chevron p-3 bg-body-tertiary rounded-2 border">`)
	w.AppendString(`<li class="breadcrumb-item"><a class="link-body-emphasis" href="/"><svg xmlns="http://www.w3.org/2000/svg" x="0px" y="0px" width="16" height="17" viewBox="0 0 24 24"><path d="M 12 2.0996094 L 1 12 L 4 12 L 4 21 L 11 21 L 11 15 L 13 15 L 13 21 L 20 21 L 20 12 L 23 12 L 12 2.0996094 z M 12 4.7910156 L 18 10.191406 L 18 11 L 18 19 L 15 19 L 15 13 L 9 13 L 9 19 L 6 19 L 6 10.191406 L 12 4.7910156 z"></path></svg><span class="visually-hidden">Home</span></a></li>`)
}

func DisplayCrumbsLinkIDStart(w *http.Response, prefix string, id database.ID) {
	w.AppendString(`<li class="breadcrumb-item">`)
	w.AppendString(`<a class="link-body-emphasis text-decoration-none" href="`)
	w.WriteString(prefix)
	w.AppendString(`/`)
	w.WriteID(id)
	w.AppendString(`">`)
}

func DisplayCrumbsLinkEnd(w *http.Response) {
	w.AppendString(`</a>`)
	w.AppendString(`</li>`)
}

func DisplayCrumbsLinkID(w *http.Response, prefix string, id database.ID, title string) {
	DisplayCrumbsLinkIDStart(w, prefix, id)
	w.WriteString(title)
	DisplayCrumbsLinkEnd(w)
}

func DisplayCrumbsLink(w *http.Response, l Language, href string, title string) {
	w.AppendString(`<li class="breadcrumb-item">`)
	w.AppendString(`<a class="link-body-emphasis text-decoration-none" href="`)
	w.WriteString(href)
	w.AppendString(`">`)
	w.WriteString(Ls(l, title))
	DisplayCrumbsLinkEnd(w)
}

func DisplayCrumbsSubmitRaw(w *http.Response, l Language, nextPage, title string) {
	w.AppendString(`<li class="breadcrumb-item">`)
	w.AppendString(`<button style="border: 0; vertical-align: top" class="btn btn-link link-body-emphasis text-decoration-none p-0" name="NextPage" value="`)
	w.WriteString(Ls(l, nextPage))
	w.AppendString(`" formnovalidate>`)
	w.WriteString(title)
	w.AppendString(`</button>`)
	w.AppendString(`</li>`)
}

func DisplayCrumbsSubmit(w *http.Response, l Language, nextPage, title string) {
	DisplayCrumbsSubmitRaw(w, l, nextPage, Ls(l, title))
}

func DisplayCrumbsItemStart(w *http.Response) {
	w.AppendString(`<li class="breadcrumb-item fw-semibold" aria-current="page">`)
}

func DisplayCrumbsItemEnd(w *http.Response) {
	w.AppendString(`</li>`)
}

func DisplayCrumbsItemRaw(w *http.Response, title string) {
	DisplayCrumbsItemStart(w)
	w.WriteString(title)
	DisplayCrumbsItemEnd(w)
}

func DisplayCrumbsItem(w *http.Response, l Language, title string) {
	DisplayCrumbsItemRaw(w, Ls(l, title))
}

func DisplayCrumbsEnd(w *http.Response) {
	w.AppendString(`</ol></nav>`)
}

func DisplayPageStart(w *http.Response, width int) {
	w.AppendString(`<div class="p-4 p-md-5 border rounded-2 bg-body-tertiary col-md-10 mx-auto col-lg-`)
	w.WriteInt(width)
	w.AppendString(`">`)
}

func DisplayPageEnd(w *http.Response) {
	w.AppendString(`</div>`)
}

func DisplayFormPageStart(w *http.Response, r *http.Request, l Language, width int, title string, endpoint string, err error) {
	w.AppendString(`<form class="p-4 p-md-5 border rounded-2 bg-body-tertiary mx-auto col-lg-`)
	w.WriteInt(width)
	w.AppendString(`" method="POST" action="`)
	w.WriteString(endpoint)
	w.AppendString(`">`)

	DisplayFormTitle(w, l, title, err)

	DisplayHiddenString(w, "ID", r.Form.Get("ID"))
}

func DisplayFormStart(w *http.Response, r *http.Request, endpoint string) {
	w.AppendString(`<form method="POST" action="`)
	w.WriteString(endpoint)
	w.AppendString(`">`)

	DisplayHiddenString(w, "ID", r.Form.Get("ID"))
}

func DisplayFormTitle(w *http.Response, l Language, title string, err error) {
	w.AppendString(`<h3 class="text-center">`)
	w.AppendString(Ls(l, title))
	w.AppendString(`</h3>`)
	w.AppendString(`<br>`)

	DisplayError(w, l, err)
}

func DisplayFormEnd(w *http.Response) {
	w.AppendString(`</form>`)
}

func DisplayFormPageEnd(w *http.Response) {
	w.AppendString(`</form>`)
}

func DisplayMainEnd(w *http.Response) {
	w.AppendString(`</main>`)
}

func DisplayBodyEnd(w *http.Response) {
	if JSEnabled {
		w.AppendString(`<div class="dropdown position-fixed bottom-0 end-0 mb-3 me-3 bd-mode-toggle">`)
		w.AppendString(`<input type="checkbox" class="btn-check" id="btn-toggle" onclick="function toggleTheme() { var html = document.querySelector('html'); html.setAttribute('data-bs-theme', html.getAttribute('data-bs-theme') === 'dark' ? 'light' : 'dark'); } toggleTheme()"/>`)
		w.AppendString(`<label style="cursor: pointer" for="btn-toggle">`)
		w.AppendString(`<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-circle-half" viewBox="0 0 16 16"> <path d="M8 15A7 7 0 1 0 8 1v14zm0 1A8 8 0 1 1 8 0a8 8 0 0 1 0 16z"/></svg>`)
		w.AppendString(`</label>`)
		w.AppendString(`</div>`)
	}

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

func DisplayFrameStart(w *http.Response) {
	w.AppendString(`<div class="border round p-4">`)
}

func DisplayFrameEnd(w *http.Response) {
	w.AppendString(`</div>`)
	w.AppendString(`<br>`)
}

func DisplayTableStart(w *http.Response, l Language, cols []string) {
	w.AppendString(`<table class="table table-bordered table-stripped table-hover">`)

	w.AppendString(`<thead>`)
	w.AppendString(`<tr>`)
	for i := 0; i < len(cols); i++ {
		w.AppendString(`<th class="text-center" scope="col">`)
		w.AppendString(Ls(l, cols[i]))
		w.AppendString(`</th>`)
	}
	w.AppendString(`</tr>`)
	w.AppendString(`</thead>`)

	w.AppendString(`<tbody>`)
}

func DisplayTableRowStart(w *http.Response) {
	w.AppendString(`<tr>`)
}

func DisplayTableRowLinkIDStart(w *http.Response, prefix string, id database.ID) {
	DisplayTableRowStart(w)

	w.AppendString(`<th class="text-center align-middle" scope="row">`)
	w.AppendString(`<a href="`)
	w.AppendString(prefix)
	w.AppendString(`/`)
	w.WriteID(id)
	w.AppendString(`">`)
	w.WriteID(id)
	w.AppendString(`</a>`)
	w.AppendString(`</th>`)
}

func DisplayTableItemStart(w *http.Response) {
	w.AppendString(`<td class="text-center align-middle">`)
}

func DisplayTableItemID(w *http.Response, id database.ID) {
	DisplayTableItemStart(w)
	w.WriteID(id)
	DisplayTableItemEnd(w)
}

func DisplayTableItemInt(w *http.Response, x int) {
	DisplayTableItemStart(w)
	w.WriteInt(x)
	DisplayTableItemEnd(w)
}

func DisplayTableItemString(w *http.Response, s string) {
	DisplayTableItemStart(w)
	w.WriteHTMLString(s)
	DisplayTableItemEnd(w)
}

func DisplayTableItemTime(w *http.Response, t int64) {
	DisplayTableItemStart(w)
	DisplayFormattedTime(w, t)
	DisplayTableItemEnd(w)
}

func DisplayTableItemFlags(w *http.Response, l Language, flags int32) {
	DisplayTableItemStart(w)
	switch flags {
	case 0: /* active */
		w.AppendString(`<small class="d-inline-flex px-2 py-1 fw-semibold text-success-emphasis bg-success-subtle border border-success-subtle rounded-2">`)
		w.AppendString(Ls(l, "Active"))
		w.AppendString(`</small>`)
	case 1: /* deleted */
		w.AppendString(`<small class="d-inline-flex px-2 py-1 fw-semibold text-danger-emphasis bg-danger-subtle border border-danger-subtle rounded-2">`)
		w.AppendString(Ls(l, "Deleted"))
		w.AppendString(`</small>`)
	case 2: /* draft */
		w.AppendString(`<small class="d-inline-flex px-2 py-1 fw-semibold text-primary-emphasis bg-primary-subtle border border-primary-subtle rounded-2">`)
		w.AppendString(Ls(l, "Draft"))
		w.AppendString(`</small>`)
	}
	DisplayTableItemEnd(w)
}

func DisplayTableItemEnd(w *http.Response) {
	w.AppendString(`</td>`)
}

func DisplayTableRowEnd(w *http.Response) {
	w.AppendString(`</tr>`)
}

func DisplayTableEnd(w *http.Response) {
	w.AppendString(`</tbody>`)
	w.AppendString(`</table>`)
}

func DisplayLabel(w *http.Response, l Language, text string) {
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
	w.AppendString(` <input class="input-field" type="`)
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

func DisplayConstraintInlineTextarea(w *http.Response, minLength, maxLength int, name, value string, required bool) {
	w.AppendString(` <textarea class="btn btn-outline-dark" rows="1" minlength="`)
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

func DisplayConstraintTextarea(w *http.Response, minLength, maxLength int, name, value string, required bool) {
	w.AppendString(` <textarea class="form-control" rows="10" minlength="`)
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

func DisplayNextPage(w *http.Response, l Language, value string) {
	w.AppendString(` <input class="w-100 btn btn-outline-primary mt-2" type="submit" name="NextPage" value="`)
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

func DisplayShortenedString(w *http.Response, s string, maxVisibleLen int) {
	if utf8.RuneCountInString(s) < maxVisibleLen {
		w.WriteHTMLString(s)
	} else {
		space := strings.FindCharReverse(s[:maxVisibleLen], ' ')
		if space == -1 {
			w.WriteHTMLString(s[:maxVisibleLen])
		} else {
			w.WriteHTMLString(s[:space])
		}
		w.AppendString(`...`)
	}
}

func DisplayMarkdown(w *http.Response, md string) {
	type Token struct {
		Type   int
		Start  string
		End    string
		RStart string
		REnd   string
	}

	const (
		None = iota
		H1
		H2
		H3
		Code
		Mono
		Bold
		Italics
		List
		Break
	)

	tokens := [...]Token{
		{Code, "```\r\n", "\r\n```", "<pre><code>", "</code></pre>"},
		{Mono, "`", "`", "<tt>", "</tt>"},
		{H1, "#", "\r\n", "<h4>", "</h4>"},
		{H2, "##", "\r\n", "<h5>", "</h5>"},
		{H3, "###", "\r\n", "<h6>", "</h6>"},
		{Bold, "**", "**", "<b>", "</b>"},
		{Bold, "__", "__", "<b>", "</b>"},
		{Italics, "*", "*", "<i>", "</i>"},
		{Italics, "_", "_", "<i>", "</i>"},
		{List, "\n-", "\r", `<li class="ms-4">`, "</li>"},
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

			w.AppendString(tok.RStart)
			switch tok.Type {
			default:
				DisplayMarkdown(w, inside)
			case Code, Mono:
				w.WriteHTMLString(inside)
			}
			w.AppendString(tok.REnd)

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
