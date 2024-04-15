package main

import "errors"

func WritePage(w *HTTPResponse, r *HTTPRequest, page func(*HTTPResponse, *HTTPRequest) error, err error) error {
	r.Form.Set("Error", err.Error())

	var httpError HTTPError
	if errors.As(err, &httpError) {
		w.StatusCode = httpError.StatusCode
	} else {
		w.StatusCode = HTTPStatusInternalServerError
	}

	return page(w, r)
}

func WritePageEx[T any](w *HTTPResponse, r *HTTPRequest, page func(*HTTPResponse, *HTTPRequest, *T) error, extra *T, err error) error {
	r.Form.Set("Error", err.Error())

	var httpError HTTPError
	if errors.As(err, &httpError) {
		w.StatusCode = httpError.StatusCode
	} else {
		w.StatusCode = HTTPStatusInternalServerError
	}

	return page(w, r, extra)
}
