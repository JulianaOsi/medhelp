package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func LaunchServer() {
	r := mux.NewRouter()

	r.HandleFunc("/directions", getDirections).Methods(http.MethodGet)
	r.HandleFunc("/direction/{id}/analysis", getDirectionAnalysis).Methods(http.MethodGet)

	fmt.Printf("Starting server at localhost:8080\n")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
