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

func DisplayConstraintInput(w *HTTPResponse, t string, minLength, maxLength int, name, value string, required bool) {
	w.AppendString(` <input type="`)
	w.AppendString(t)
	w.AppendString(`" minlength="`)
	w.WriteInt(minLength)
	w.AppendString(`" maxlength="`)
	w.WriteInt(maxLength)
	w.AppendString(`" name="`)
	w.AppendString(name)
	w.AppendString(`" value="`)
	w.WriteHTMLString(value)
	w.AppendString(`"`)
	if required {
		w.AppendString(` required`)
	}
	w.AppendString(`>`)
}

func DisplayConstraintIndexedInput(w *HTTPResponse, t string, minLength, maxLength int, name string, index int, value string, required bool) {
	w.AppendString(` <input type="`)
	w.AppendString(t)
	w.AppendString(`" minlength="`)
	w.WriteInt(minLength)
	w.AppendString(`" maxlength="`)
	w.WriteInt(maxLength)
	w.AppendString(`" name="`)
	w.AppendString(name)
	w.WriteInt(index)
	w.AppendString(`" value="`)
	w.WriteHTMLString(value)
	w.AppendString(`"`)
	if required {
		w.AppendString(` required`)
	}
	w.AppendString(`>`)
}

func DisplayConstraintTextarea(w *HTTPResponse, cols, rows string, minLength, maxLength int, name, value string, required bool) {
	w.AppendString(` <textarea cols="`)
	w.AppendString(cols)
	w.AppendString(`" rows="`)
	w.AppendString(rows)
	w.AppendString(`" minlength="`)
	w.WriteInt(minLength)
	w.AppendString(`" maxlength="`)
	w.WriteInt(maxLength)
	w.AppendString(`" name="`)
	w.AppendString(name)
	w.AppendString(`"`)
	if required {
		w.AppendString(` required`)
	}
	w.AppendString(`>`)
	w.WriteHTMLString(value)
	w.AppendString(`</textarea>`)
}

func DisplayIndexedCommand(w *HTTPResponse, index int, command string) {
	w.AppendString(` <input type="submit" name="Command`)
	w.WriteInt(index)
	w.AppendString(`" value="`)
	w.AppendString(command)
	w.AppendString(`" formnovalidate>`)
}

func DisplayDoublyIndexedCommand(w *HTTPResponse, pindex, sindex int, command string) {
	w.AppendString(` <input type="submit" name="Command`)
	w.WriteInt(pindex)
	w.AppendString(`.`)
	w.WriteInt(sindex)
	w.AppendString(`" value="`)
	w.AppendString(command)
	w.AppendString(`" formnovalidate>`)
}
