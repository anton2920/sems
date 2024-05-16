package main

import "github.com/anton2920/gofa/net/http"

func DisplayErrorMessage(w *http.Response, message string) {
	if message != "" {
		w.AppendString(`<div><p>Error: `)
		w.WriteHTMLString(message)
		w.AppendString(`.</p></div>`)
	}
}

func ErrorPageHandler(w *http.Response, message string) {
	w.Bodies = w.Bodies[:0]

	w.AppendString(`<!DOCTYPE html>`)
	w.AppendString(`<head><title>Error</title></head>`)
	w.AppendString(`<body>`)

	w.AppendString(`<h1>Master's degree</h1>`)
	w.AppendString(`<h2>Error</h2>`)

	DisplayErrorMessage(w, message)

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)
}
