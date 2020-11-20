package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/JulianaOsi/medhelp/pkg/store"
)

func getDirections(w http.ResponseWriter, r *http.Request) {
	directions, err := store.DB.GetDirections(context.Background())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to get directions: %v\n", err)
		return
	}

	directionsBytes, err := json.MarshalIndent(directions, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to marshal directions: %v\n", err)
		return
	}

	if _, err := w.Write(directionsBytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to write directions: %v\n", err)
	}
}

func getDirectionAnalysis(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	analysis, err := store.DB.GetAnalysisByDirectionId(context.Background(), id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to get direction analysis: %v\n", err)
		return
	}

	analysisBytes, err := json.MarshalIndent(analysis, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to marshal direction analysis: %v\n", err)
		return
	}

	if _, err := w.Write(analysisBytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to write direction analysis: %v\n", err)
	}
}
