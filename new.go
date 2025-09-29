package main

import (
	"github.com/anton2920/gofa/net/html"
	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/session"
	"github.com/anton2920/gofa/trace"
)

var (
	Styles = html.Theme{
		A:        html.Class("text-decoration-none"),
		Body:     html.Class("p-md-5"),
		Form:     html.Class("p-4 p-md-5 border rounded-2 bg-body-tertiary mx-auto col-lg-4"),
		Label:    html.Class("form-label"),
		Input:    html.Class("form-control mb-3"),
		Button:   html.Class("w-100 btn btn-lg rounded-3 btn-primary"),
		Checkbox: html.Class("form-check-input"),
	}

	StyleTextBlack = html.Class("text-dark")
	StyleTextGray  = html.Class("text-muted")
	StyleTextSmall = html.Class("small")

	/*
		StyleButtonSubmit
		StyleButtonLarge
		StyleButtonSmall
		StyleButtonCommand
	*/

)

func NewPage(w *http.Response, r *http.Request) error {
	defer trace.End(trace.Begin(""))

	session := session.Get(w, r)
	h := html.New(w, r, session, Styles)

	h.Begin()

	h.HeadBegin()
	{
		h.String(`<link rel="stylesheet" href="/fs/bootstrap.min.css"/>`)
		h.String(`<style>.navbar-custom { position: fixed; z-index: 190; }</style>`)
		h.String(`<style>.bttn {}</style>`)
		h.String(`<style>input.bttn[type="submit"]:hover { transform: translate(1000px, 0px); }</style>`)
		h.Title("New page")
	}
	h.HeadEnd()

	h.BodyBegin()
	{
		h.H1("This is the new test page")
		h.H2("This page is indended to test new HTML component system")

		h.H3Begin()
		{
			h.LString("Powered by")
			h.A("https://github.com/anton2920/gofa", "GOFA library")
		}
		h.H3End()

		h.WithoutTheme().Button("Are you satisfied with your salary?", html.Class("bttn"))

		h.FormBegin("POST", APIPrefix+"/new")
		{
			h.H4("Edit user", html.Class("text-center mb-3"))

			h.Label("First name")
			h.Input("text", html.Attributes{MinLength: MinNameLen, MaxLength: MaxNameLen, Name: "FirstName", Value: h.Form.Get("FirstName"), Required: true})

			h.Label("Last name")
			h.Input("text", html.Attributes{MinLength: MinNameLen, MaxLength: MaxNameLen, Name: "LastName", Value: h.Form.Get("LastName"), Required: true})

			h.Label("Email")
			h.Input("email", html.Attributes{Name: "Email", Value: h.Form.Get("Email"), Required: true})

			h.Label("Password")
			h.Input("password", html.Attributes{Name: "Password", Required: true})

			h.Label("Repeat password")
			h.Input("password", html.Attributes{MinLength: MinPasswordLen, MaxLength: MaxPasswordLen, Name: "RepeatPassword", Required: true})

			h.Button("Save", html.Name("ButtonName"))
		}
		h.FormEnd()
		h.BR()

		h.FormBegin("POST", APIPrefix+"/new")
		{
			h.H4("Sign in", html.Class("text-center mb-3"))

			h.Input("email", html.Attributes{Name: "Email", Value: h.Form.Get("Email"), Placeholder: "Email", Required: true})

			h.Input("password", html.Attributes{Name: "Password", Placeholder: "Password", Required: true})

			h.DivBegin(html.Class("form-check mb-2"))
			{
				h.LabelBegin()
				h.Checkbox(html.Attributes{Name: "Remember", Checked: true})
				h.LString("Remember me")
				h.LabelEnd()
			}
			h.DivEnd()

			h.Button("Sign in", html.Name("ButtonName"))
		}
		h.FormEnd()
	}
	h.BodyEnd()

	h.End()
	return nil
}
