package main

import (
	"github.com/anton2920/gofa/errors"
	"github.com/anton2920/gofa/log"
	"github.com/anton2920/gofa/net/http"
)

func DisplayErrorMessage(w *http.Response, l Language, message string) {
	if message != "" {
		w.AppendString(`<div><p>`)
		w.AppendString(Ls(l, "Error"))
		w.AppendString(`: `)
		//w.WriteHTMLString(message)
		w.WriteHTMLString(Ls(l, message))
		w.AppendString(`.</p></div>`)
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
	const width = WidthMedium

	w.SetHeaderUnsafe("Connection", "close")
	w.Bodies = w.Bodies[:0]

	session, _ := GetSessionFromRequest(r)

	DisplayHTMLStart(w)

	DisplayHeadStart(w)
	{
		w.AppendString(`<title>`)
		w.AppendString(Ls(l, "Error"))
		w.AppendString(`</title>`)
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
			w.AppendString(`<h2>`)
			w.AppendString(Ls(l, "Error"))
			w.AppendString(`</h2>`)

			DisplayError(w, l, err)
		}
		DisplayPageEnd(w)
		DisplayMainEnd(w)
	}
	DisplayBodyEnd(w)

	DisplayHTMLEnd(w)
}
