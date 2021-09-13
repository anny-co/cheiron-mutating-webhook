package main

import (
	"fmt"
	"net/http"
)

// ping returns immediately
var ping http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
})
