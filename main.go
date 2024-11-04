package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	routes := router()

	server := &http.Server{
		Addr:    ":8080",
		Handler: routes,
	}

	ESClientConnection()
	ESCreateIndexIfNotExist()

	fmt.Println("Server listening on port :8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func router() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Test API"))
	})
	mux.HandleFunc("GET /user", GetUser)
	mux.HandleFunc("POST /user", PostUser)
	mux.HandleFunc("PUT /user/{id}", UpdateUser)
	mux.HandleFunc("DELETE /user/{id}", DeleteUser)
	mux.HandleFunc("GET /info", Info)

	return mux
}
