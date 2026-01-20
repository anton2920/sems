package main

import (
	"testing"

	"github.com/anton2920/gofa/net/html"
	"github.com/anton2920/gofa/net/http"
)

func BenchmarkNewPageReference(b *testing.B) {
	var w http.Response

	for i := 0; i < b.N; i++ {
		w.WriteString(`<!DOCTYPE html><html lang="en"> <head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"> <link> <script></script><link rel="stylesheet" href="/fs/bootstrap.min.css"/><style>.navbar-custom { position: fixed; z-index: 190; }</style><style>.bttn {}</style><style>input.bttn[type="submit"]:hover { transform: translate(1000px, 0px); }</style> <title>New page</title></head> <body class="p-md-5"> <h1>This is the new test page</h1> <h2>This page is indended to test new HTML component system</h2> <h3>Powered by <a class="text-decoration-none" href="https://github.com/anton2920/gofa">GOFA library</a></h3> <input class="bttn" type="submit" value="Are you satisfied with your salary?"> <form class="p-4 p-md-5 border rounded-2 bg-body-tertiary mx-auto col-lg-4 mb-3" action="/api/new" method="POST"> <h4 class="text-center mb-3">Edit user</h4> <label class="form-label">First name</label> <input class="form-control mb-3" name="FirstName" type="text" maxlength="45" minlength="1" required> <label class="form-label">Last name</label> <input class="form-control mb-3" name="LastName" type="text" maxlength="45" minlength="1" required> <label class="form-label">Email</label> <input class="form-control mb-3" name="Email" type="email" required> <label class="form-label">Password</label> <input class="form-control mb-3" name="Password" type="password" required> <label class="form-label">Repeat password</label> <input class="form-control mb-3" name="RepeatPassword" type="password" maxlength="45" minlength="5" required> <input class="w-100 btn btn-lg rounded-3 btn-primary" name="ButtonName" type="submit" value="Save"></form> <form class="p-4 p-md-5 border rounded-2 bg-body-tertiary mx-auto col-lg-4" action="/api/new" method="POST"> <h4 class="text-center mb-3">Sign in</h4> <input class="form-control mb-3" name="Email" type="email" placeholder="Email" required> <input class="form-control mb-3" name="Password" type="password" placeholder="Password" required> <div class="form-check mb-2"> <label class="form-label"> <input class="form-check-input" name="Remember" type="checkbox" checked>Remember me</label></div> <input class="w-100 btn btn-lg rounded-3 btn-primary" name="ButtonName" type="submit" value="Sign in"></form></body></html>`)
		w.Body = w.Body[:0]
	}
}

func BenchmarkNewPage(b *testing.B) {
	var w http.Response
	var r http.Request

	h := html.New(&w, &r, &Styles)

	for i := 0; i < b.N; i++ {
		NewPage(&h)
		w.Body = w.Body[:0]
	}
}

func BenchmarkNewPage2(b *testing.B) {
	var w http.Response
	var r http.Request

	h := html.New(&w, &r, &Styles)

	for i := 0; i < b.N; i++ {
		NewPage2(&h)
		w.Body = w.Body[:0]
	}
}

func BenchmarkUserSigninPage(b *testing.B) {
	var w http.Response
	var r http.Request

	for i := 0; i < b.N; i++ {
		UserSigninPageHandler(&w, &r, nil)
		w.Body = w.Body[:0]
	}
}
