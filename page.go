package main

import "github.com/anton2920/gofa/net/http"

func WritePage(w *http.Response, r *http.Request, page func(*http.Response, *http.Request) error, err error) error {
	e := err.(http.Error)
	r.Form.Set("Error", e.DisplayMessage)
	w.StatusCode = e.StatusCode
	return page(w, r)
}

func WritePageEx[T any](w *http.Response, r *http.Request, session *Session, page func(*http.Response, *http.Request, *Session, *T) error, extra *T, err error) error {
	e := err.(http.Error)
	r.Form.Set("Error", e.DisplayMessage)
	w.StatusCode = e.StatusCode
	return page(w, r, session, extra)
}
