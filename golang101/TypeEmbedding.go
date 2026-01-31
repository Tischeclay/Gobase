package main

import "net/http"

func main() {
	type P = *bool
	type M = map[int]int
	var x struct {
		string
		error
		*int
		P
		M
		http.Header
	}
	x.string = "Go"
	x.error = nil
	x.int = new(int)
	x.P = new(bool)
	x.M = make(M)
	x.Header = http.Header{}
}
