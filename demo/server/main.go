package main

import (
	"fmt"

	"net/http"

	"github.com/happycrud/golib/app"
)

func main() {
	a := app.New(app.WithPort(7890))
	a.Get("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello")
	}))
	a.Run()
}
