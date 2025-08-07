package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello World from Server!")
	})

	port := ":8080"
	log.Printf("Server starting on port %s", port)
	log.Printf("Visit: http://localhost%s", port)

	log.Fatal(http.ListenAndServe(port, nil))
}
