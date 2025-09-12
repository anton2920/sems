package main

import (
	"github.com/anton2920/gofa/l10n"
	"github.com/anton2920/gofa/net/html"
	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/trace"
)

var (
	StyleA        = html.Class("text-decoration-none")
	StyleBody     = html.Class("p-md-5")
	StyleForm     = html.Class("p-4 p-md-5 border rounded-2 bg-body-tertiary mx-auto col-lg-4")
	StyleLabel    = html.Class("form-label")
	StyleInput    = html.Class("form-control mb-3")
	StyleButton   = html.Class("w-100 btn btn-lg rounded-3 btn-primary")
	StyleCheckbox = html.Class("")
)

func NewPage(w *http.Response, r *http.Request) error {
	defer trace.End(trace.Begin(""))

	h := html.New(w, l10n.LanguageRussian, 3)
	h.Theme.Button = StyleButton

	h.Begin()

	h.HeadBegin()
	{
		h.String(`<link rel="stylesheet" href="/fs/bootstrap.min.css"/>`)
		h.String(`<style>.navbar-custom { position: fixed; z-index: 190; }</style>`)
		h.Title("New page")
	}
	h.HeadEnd()

	h.BodyBegin(StyleBody)
	{
		h.H1("This is the new test page")
		h.H2("This page is indended to test new HTML component system")

		h.H3Begin()
		{
			h.LString("Powered by")
			h.A("https://github.com/anton2920/gofa", "GOFA library", StyleA)
		}
		h.H3End()

		h.FormBegin("POST", APIPrefix+"/new", StyleForm)
		{
			h.H4("Edit user", html.Class("text-center mb-3"))

			h.Label("First name", StyleLabel)
			h.Input("text", StyleInput, html.Attributes{MinLength: MinNameLen, MaxLength: MaxNameLen, Name: "FirstName", Value: r.Form.Get("FirstName"), Required: true})

			h.Label("Last name", StyleLabel)
			h.Input("text", StyleInput, html.Attributes{MinLength: MinNameLen, MaxLength: MaxNameLen, Name: "LastName", Value: r.Form.Get("LastName"), Required: true})

			h.Label("Email", StyleLabel)
			h.Input("email", StyleInput, html.Attributes{Name: "Email", Value: r.Form.Get("Email"), Required: true})

			h.Label("Password", StyleLabel)
			h.Input("password", StyleInput, html.Attributes{Name: "Password", Required: true})

			h.Label("Repeat password", StyleLabel)
			h.Input("password", StyleInput, html.Attributes{MinLength: MinPasswordLen, MaxLength: MaxPasswordLen, Name: "RepeatPassword", Required: true})

			h.Button("Save", html.Name("ButtonName"))
		}
		h.FormEnd()
		h.BR()

		h.FormBegin("POST", APIPrefix+"/new", StyleForm)
		{
			h.H4("Sign in", html.Class("text-center mb-3"))

			h.Input("email", StyleInput, html.Attributes{Name: "Email", Value: r.Form.Get("Email"), Placeholder: "Email", Required: true})

			h.Input("password", StyleInput, html.Attributes{Name: "Password", Placeholder: "Password", Required: true})

			h.LabelBegin(StyleLabel)
			h.Checkbox(StyleCheckbox, html.Name("Remember"), html.Value(r.Form.Get("Remember")))
			h.LString("Remember me")
			h.LabelEnd()

			h.Button("Sign in", html.Name("ButtonName"))
		}
		h.FormEnd()
	}
	h.BodyEnd()

	h.End()
	return nil
}
