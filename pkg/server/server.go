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
	r.HandleFunc("/directions/add", addDirection).Methods(http.MethodPost)
	r.HandleFunc("/direction/{id}/analysis", getDirectionAnalysis).Methods(http.MethodGet)
	r.HandleFunc("/analysis/{analysis}/upload", uploadAnalysisFile).Methods(http.MethodPost)
	r.HandleFunc("/analysis/{analysis}/download", downloadAnalysisFile).Methods(http.MethodGet)
	r.HandleFunc("/status", setDirectionStatus).Methods(http.MethodPost)
	r.HandleFunc("/check", setAnalysisCheck).Methods(http.MethodPost)

	/*CORS pre-flight requests*/
	r.HandleFunc("/registration", corsSkip).Methods(http.MethodOptions)
	r.HandleFunc("/auth", corsSkip).Methods(http.MethodOptions)
	r.HandleFunc("/directions", corsSkip).Methods(http.MethodOptions)
	r.HandleFunc("/directions/add", corsSkip).Methods(http.MethodOptions)
	r.HandleFunc("/direction/{id}/analysis", corsSkip).Methods(http.MethodOptions)
	r.HandleFunc("/analysis/{analysis}/upload", corsSkip).Methods(http.MethodOptions)
	r.HandleFunc("/analysis/{analysis}/download", corsSkip).Methods(http.MethodOptions)
	r.HandleFunc("/status", corsSkip).Methods(http.MethodOptions)
	r.HandleFunc("/check", corsSkip).Methods(http.MethodOptions)

	fmt.Printf("Starting server at localhost:8080\n")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
