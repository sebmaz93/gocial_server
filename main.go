package main

import (
	"log"
	"net/http"
)

func main() {
	const port = "8080"
	const rootPath = "."

	mux := http.NewServeMux()
	dir := http.Dir(rootPath)
	fileServer := http.FileServer(dir)
	mux.Handle("/", fileServer)
	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("failed to start the server!")
	}
	log.Printf("Server listening on port :%s\n", port)
}
