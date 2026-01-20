package main

import (
	"github.com/anton2920/gofa/net/html"
	"github.com/anton2920/gofa/net/http"
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

func NewPage(h *html.HTML) error {
	defer trace.End(trace.Begin(""))

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

		h.FormBegin("POST", html.Action(APIPrefix+"/new"), html.Class("mb-3"))
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

		h.FormBegin("POST", html.Action(APIPrefix+"/new"))
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

func NewHandler(w *http.Response, r *http.Request) error {
	h := html.New(w, r, &Styles)
	return NewPage(&h)
}

func NewPage2(h *html.HTML) error {
	h.Begin2()

	h.HeadBegin2()
	{
		h.String(`<link rel="stylesheet" href="/fs/bootstrap.min.css"/>`)
		h.String(`<style>.navbar-custom { position: fixed; z-index: 190; }</style>`)
		h.String(`<style>.bttn {}</style>`)
		h.String(`<style>input.bttn[type="submit"]:hover { transform: translate(1000px, 0px); }</style>`)
		h.Title2("New page")
	}
	h.HeadEnd2()

	h.BodyBegin2()
	{
		h.H12("This is the new test page")
		h.H22("This page is indended to test new HTML component system")

		h.H3Begin2()
		{
			h.LString("Powered by")
			h.A2("https://github.com/anton2920/gofa", "GOFA library")
		}
		h.H3End2()

		h.WithoutTheme().Button2("Are you satisfied with your salary?").Class("bttn")

		h.FormBegin2("POST").Action(APIPrefix + "/new").Class("mb-3")
		{
			h.H42("Edit user").Class("text-center mb-3")

			h.Label2("First name")
			h.Input2("text").MinLength(MinNameLen).MaxLength(MaxNameLen).Name("FirstName").Value(h.Form.Get("FirstName")).Required(true)

			h.Label2("Last name")
			h.Input2("text").MinLength(MinNameLen).MaxLength(MaxNameLen).Name("LastName").Value(h.Form.Get("LastName")).Required(true)

			h.Label2("Email")
			h.Input2("email").Name("Email").Value(h.Form.Get("Email")).Required(true)

			h.Label2("Password")
			h.Input2("password").Name("Password").Required(true)

			h.Label2("Repeat password")
			h.Input2("password").MinLength(MinPasswordLen).MaxLength(MaxPasswordLen).Name("RepeatRepeat").Required(true)

			h.Button2("Save").Name("ButtonName")
		}
		h.FormEnd2()

		h.FormBegin2("POST").Action(APIPrefix + "/new")
		{
			h.H4Begin2().Class("text-center mb-3").LString("Sign in").H4End2()

			h.Input2("email").Name("Email").Value(h.Form.Get("Email")).Placeholder("Email").Required(true)
			h.Input2("password").Name("Password").Placeholder("Password").Required(true)

			h.DivBegin2().Class("form-check mb-2")
			{
				h.LabelBegin2()
				h.Checkbox2().Name("Remember").Checked(true)
				h.LString("Remember me")
				h.LabelEnd2()
			}
			h.DivEnd2()

			h.Button2("Sign in").Name("ButtonName")
		}
		h.FormEnd2()
	}
	h.BodyEnd2()

	h.End2()
	return nil
}

func NewHandler2(w *http.Response, r *http.Request) error {
	h := html.New(w, r, &Styles)
	return NewPage2(&h)
}
