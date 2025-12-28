package main

import (
	"log"
	"net/http"
)

func main() {

	handler := http.NewServeMux()
	const port = "8080"

	handler.HandleFunc("/", func(w http.ResponseWriter, res *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello World! because of course..."))
	})

	server := http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}
	log.Printf("Serving on port: %v\n", port)
	log.Fatalln(server.ListenAndServe())
}
