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

func ErrorPageHandler(w *http.Response, l Language, err error) {
	w.Bodies = w.Bodies[:0]

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>`)
	w.AppendString(Ls(l, "Error"))
	w.AppendString(`</title></head>`)

	w.AppendString(`<body>`)

	w.AppendString(`<h1>`)
	w.AppendString(Ls(l, "Master's degree"))
	w.AppendString(`</h1>`)
	w.AppendString(`<h2>`)
	w.AppendString(Ls(l, "Error"))
	w.AppendString(`</h2>`)

	DisplayError(w, l, err)

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)
}
