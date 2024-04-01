package main

import "unsafe"

const PageSize = 4096

var DebugMode = "off"

func IndexPageHandler(w *HTTPResponse, r *HTTPRequest) {
	const indexPage = `
<!DOCTYPE html>
<head>
	<title>Master's degree</title>
</head>
<body>
	<h1>Master's degree</h1>
</body>
</html>
`
	w.WriteResponseNoCopy("text/html", unsafe.Slice(unsafe.StringData(indexPage), len(indexPage)))
}

func Router(w *HTTPResponse, r *HTTPRequest) {
	if r.URL.Path == "/" {
		IndexPageHandler(w, r)
	}
}

func main() {
	if err := ListenAndServe(7072, Router); err != nil {
		FatalError("Failed to listen on port:", err)
	}
}
