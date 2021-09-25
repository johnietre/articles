package main

import (
	"log"
	"net/http"
)

func main() {
	s := &http.Server{
		Addr:    ":8000",
		Handler: &Articles{},
	}
	log.Fatal(s.ListenAndServe())
}
