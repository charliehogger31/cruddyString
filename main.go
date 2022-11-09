package main

import (
	"fmt"
	"net/http"
)
import "sync"

var cache [0]string
var cacheMutext sync.Mutex

func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "hello\n")
}

func main() {
	http.HandleFunc("/hello", hello)

	http.ListenAndServe(":8080", nil)
}
