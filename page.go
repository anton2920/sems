package main

func WritePage(w *HTTPResponse, r *HTTPRequest, page func(*HTTPResponse, *HTTPRequest) error, err error) error {
	e := err.(HTTPError)
	r.Form.Set("Error", e.DisplayMessage)
	w.StatusCode = e.StatusCode
	return page(w, r)
}

func WritePageEx[T any](w *HTTPResponse, r *HTTPRequest, page func(*HTTPResponse, *HTTPRequest, *T) error, extra *T, err error) error {
	e := err.(HTTPError)
	r.Form.Set("Error", e.DisplayMessage)
	w.StatusCode = e.StatusCode
	return page(w, r, extra)
}
