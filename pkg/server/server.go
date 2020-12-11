package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func LaunchServer() {
	r := mux.NewRouter()

	r.HandleFunc("/registration", registrationHandler).Methods(http.MethodPost)
	r.HandleFunc("/auth", authenticationHandler).Methods(http.MethodPost)
	r.HandleFunc("/directions", getDirections).Methods(http.MethodGet)
	r.HandleFunc("/direction/{id}/analysis", getDirectionAnalysis).Methods(http.MethodGet)
	r.HandleFunc("/status", setDirectionStatus).Methods(http.MethodPost)
	r.HandleFunc("/check", setAnalysisCheck).Methods(http.MethodPost)
	r.HandleFunc("/direction/{id}/analysisFile", uploadAnalysisFile).Methods(http.MethodPost)
	r.HandleFunc("/direction/{id}/analysisFile", getAnalysisFiles).Methods(http.MethodGet)

	fmt.Printf("Starting server at localhost:8080\n")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
