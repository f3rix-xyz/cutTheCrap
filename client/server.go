package main

import (
	"log"
	"net/http"
)

func main() {
	fileServer := http.FileServer(http.Dir("./"))
	http.Handle("/", fileServer)
	log.Printf("Starting server on http://localhost:8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("Error starting server: ", err)
	}
}
