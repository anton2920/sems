package main

import (
	"github.com/anton2920/gofa/errors"
	"github.com/anton2920/gofa/log"
	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/prof"
)

func DisplayErrorMessage(w *http.Response, l Language, message string) {
	if message != "" {
		w.WriteString(`<div><p>`)
		w.WriteString(Ls(l, "Error"))
		w.WriteString(`: `)
		//w.WriteHTMLString(message)
		w.WriteHTMLString(Ls(l, message))
		w.WriteString(`.</p></div>`)
	}
}

func DisplayError(w *http.Response, l Language, err error) {
	var message string

	if err != nil {
		if httpError, ok := err.(http.Error); ok {
			w.StatusCode = httpError.StatusCode
			message = httpError.DisplayMessage
		} else if _, ok := err.(errors.Panic); ok {
			w.StatusCode = http.ServerError(nil).StatusCode
			message = http.ServerError(nil).DisplayMessage
		} else {
			log.Panicf("Unsupported error type %T", err)
		}

		if Debug {
			message = err.Error()
		}
	}

	DisplayErrorMessage(w, l, message)
}

func ErrorPageHandler(w *http.Response, r *http.Request, l Language, err error) {
	defer prof.End(prof.Begin(""))

	const width = WidthMedium

	w.Headers.Set("Connection", "close")
	w.Body = w.Body[:0]

	session, _ := GetSessionFromRequest(r)

	DisplayHTMLStart(w)

	DisplayHeadStart(w)
	{
		w.WriteString(`<title>`)
		w.WriteString(Ls(l, "Error"))
		w.WriteString(`</title>`)
	}
	DisplayHeadEnd(w)

	DisplayBodyStart(w)
	{
		DisplayHeader(w, l)
		if session != nil {
			DisplaySidebar(w, l, session)
		}

		DisplayMainStart(w)

		DisplayPageStart(w, width)
		{
			w.WriteString(`<h2>`)
			w.WriteString(Ls(l, "Error"))
			w.WriteString(`</h2>`)

			DisplayError(w, l, err)
		}
		DisplayPageEnd(w)
		DisplayMainEnd(w)
	}
	DisplayBodyEnd(w)

	DisplayHTMLEnd(w)
}
