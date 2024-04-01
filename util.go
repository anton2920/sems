package main

func StringStartsWith(s, prefix string) bool {
	return (len(s) >= len(prefix)) && (s[:len(prefix)] == prefix)
}
