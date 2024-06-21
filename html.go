package main

import (
	"time"

	"github.com/anton2920/gofa/database"
	"github.com/anton2920/gofa/net/http"
)

func DisplayCSSLink(w *http.Response) {
	w.AppendString(`<link rel="stylesheet" href="/fs/bootstrap.min.css"/>`)
}

func DisplayFormattedTime(w *http.Response, t int64) {
	w.Write(time.Unix(t, 0).AppendFormat(make([]byte, 0, 20), "2006/01/02 15:04:05"))
}

func DisplayConstraintInput(w *http.Response, t string, minLength, maxLength int, name, value string, required bool) {
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

func DisplayConstraintIndexedInput(w *http.Response, t string, minLength, maxLength int, name string, index int, value string, required bool) {
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

func DisplayConstraintTextarea(w *http.Response, cols, rows string, minLength, maxLength int, name, value string, required bool) {
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

func DisplayCommand(w *http.Response, l Language, command string) {
	w.AppendString(` <input type="submit" name="Command" value="`)
	w.AppendString(Ls(l, command))
	w.AppendString(`" formnovalidate>`)
}

func DisplayDeleted(w *http.Response, l Language, deleted bool) {
	if deleted {
		w.AppendString(` [`)
		w.AppendString(Ls(l, "deleted"))
		w.AppendString(`]`)
	}
}

func DisplayDraft(w *http.Response, l Language, draft bool) {
	if draft {
		w.AppendString(` (`)
		w.AppendString(Ls(l, "draft"))
		w.AppendString(`)`)
	}
}

func DisplayIndexedCommand(w *http.Response, l Language, index int, command string) {
	w.AppendString(` <input type="submit" name="Command`)
	w.WriteInt(index)
	w.AppendString(`" value="`)
	w.AppendString(Ls(l, command))
	w.AppendString(`" formnovalidate>`)
}

func DisplayDoublyIndexedCommand(w *http.Response, l Language, pindex, sindex int, command string) {
	w.AppendString(` <input type="submit" name="Command`)
	w.WriteInt(pindex)
	w.AppendString(`.`)
	w.WriteInt(sindex)
	w.AppendString(`" value="`)
	w.AppendString(Ls(l, command))
	w.AppendString(`" formnovalidate>`)
}

func DisplayHiddenID(w *http.Response, name string, id database.ID) {
	w.AppendString(`<input type="hidden" name="`)
	w.AppendString(name)
	w.AppendString(`" value="`)
	w.WriteID(id)
	w.AppendString(`">`)
}

func DisplayHiddenInt(w *http.Response, name string, i int) {
	w.AppendString(`<input type="hidden" name="`)
	w.AppendString(name)
	w.AppendString(`" value="`)
	w.WriteInt(i)
	w.AppendString(`">`)
}

func DisplayHiddenString(w *http.Response, name string, value string) {
	w.AppendString(`<input type="hidden" name="`)
	w.AppendString(name)
	w.AppendString(`" value="`)
	w.WriteHTMLString(value)
	w.AppendString(`">`)
}

func DisplaySubmit(w *http.Response, l Language, name string, value string, verify bool) {
	w.AppendString(` <input type="submit" name="`)
	w.AppendString(name)
	w.AppendString(`" value="`)
	w.AppendString(Ls(l, value))
	w.AppendString(`"`)
	if !verify {
		w.AppendString(` formnoverify`)
	}
	w.AppendString(`>`)
}
