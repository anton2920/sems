package main

import (
	"time"
	"unsafe"
)

const HTMLHeader = `<!DOCTYPE html>`

var (
	HTMLQuot = "&#34;" // shorter than "&quot;"
	HTMLApos = "&#39;" // shorter than "&apos;" and apos was not in HTML until HTML5
	HTMLAmp  = "&amp;"
	HTMLLt   = "&lt;"
	HTMLGt   = "&gt;"
	HTMLNull = "\uFFFD"
)

func ContentTypeHTML(bodies []Iovec) bool {
	if len(bodies) == 0 {
		return false
	}

	header := unsafe.String((*byte)(bodies[0].Base), int(bodies[0].Len))
	return header == HTMLHeader
}

func DisplayFormattedTime(w *HTTPResponse, t time.Time) {
	w.Write(t.AppendFormat(make([]byte, 0, 20), "2006/01/02 15:04:05"))
}
