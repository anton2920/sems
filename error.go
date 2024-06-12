package main

import "github.com/anton2920/gofa/net/http"

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

func ErrorPageHandler(w *http.Response, l Language, message string) {
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

	DisplayErrorMessage(w, l, message)

	w.AppendString(`</body>`)
	w.AppendString(`</html>`)
}
