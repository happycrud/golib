package main

import (
	"fmt"
	"net/http"

	"github.com/rs/cors"

	"github.com/happycrud/golib/app"
)

func main() {
	a := app.New(
		app.WithPort(7890),
		app.WithTLSConfig(&app.TLSConfig{
			Port:     7890,
			KeyPath:  "./key.pem",
			CertPath: "./cert.pem",
		}),
		app.WithCorsOptions(&cors.Options{
			AllowedOrigins: []string{"*"},
			AllowedHeaders: []string{"*"},
			AllowedMethods: []string{"*"},
		}),
	)
	a.GET("/hello", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello")
	}))
	a.GET("/greet", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "greet")
	}))

	a.Run()
}
