package main

import (
	"net/http"
)

func main() {
	sMux := http.NewServeMux()
	httpServer := http.Server{
		Addr:    ":8080",
		Handler: sMux,
	}
	httpServer.ListenAndServe()
}
