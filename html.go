package main

import "unsafe"

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
