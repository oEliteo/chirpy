package main

import (
	"net/http"
)

func main() {
	sMux := http.NewServeMux()

	sMux.Handle("/", http.FileServer(http.Dir(".")))
	httpServer := http.Server{
		Addr:    ":8080",
		Handler: sMux,
	}
	httpServer.ListenAndServe()
}
