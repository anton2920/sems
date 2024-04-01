package main

func Fatal(msg string) {
	println(msg)
	Exit(1)
}

func FatalWithCode(msg string, code int) {
	println(msg, code)
	Exit(1)
}

func FatalError(msg string, err error) {
	println(msg, err.Error())
	Exit(1)
}
