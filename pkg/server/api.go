package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/JulianaOsi/medhelp/pkg/config"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/antonlindstrom/pgstore"
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

func setDirectionStatus(w http.ResponseWriter, r *http.Request) {
	type directionUpdate struct {
		DirectionId int `json:"directionId"`
		Status      int `json:"status"`
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatalf("failed to read body: %v\n", err)
	}
	update := directionUpdate{}

	err = json.Unmarshal(body, &update)
	if err != nil {
		log.Fatalf("failed to unmarshal json: %v\n", err)
	}

	err = store.DB.SetDirectionStatus(context.Background(), update.DirectionId, update.Status)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to set direction status: %v\n", err)
		return
	}
}

func setAnalysisCheck(w http.ResponseWriter, r *http.Request) {
	type analysisUpdate struct {
		AnalysisId int  `json:"analysisId"`
		Checked    bool `json:"checked"`
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatalf("failed to read body: %v\n", err)
	}
	update := analysisUpdate{}

	err = json.Unmarshal(body, &update)
	if err != nil {
		log.Fatalf("failed to unmarshal json: %v\n", err)
	}

	err = store.DB.SetAnalysisState(context.Background(), update.AnalysisId, update.Checked)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to set analysis state: %v\n", err)
		return
	}
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	type login struct {
		LastName     string `json:"lastName"`
		PolicyNumber string `json:"policyNumber"`
	}
	authData := login{}

	type token struct {
		Content string `json:"content"`
	}
	newToken := token{}

	c := config.ReadConfig().DB
	sessionStore, err := pgstore.NewPGStore(fmt.Sprintf("postgres://%s:@%s:%s/%s?sslmode=disable",
		c.User, c.Host, c.Port, c.Name), []byte("secretly-secret")) // TODO Разобраться с секретом
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to connect to session storage: %v\n", err)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to read body: %v\n", err)
		return
	}

	err = json.Unmarshal(body, &authData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to unmarshal json: %v\n", err)
		return
	}

	patient, err := store.DB.FindPatient(context.Background(), authData.LastName, authData.PolicyNumber)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to find patient: %v\n", err)
		return
	}

	if patient == nil {
		w.WriteHeader(http.StatusExpectationFailed) // TODO Подсмотреть нормальный статус userNotFound
		return
	}

	newToken.Content = "session-key" // TODO сделать генерацию токена для новой сессии
	session, err := sessionStore.Get(r, newToken.Content)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to create new session: %v\n", err)
		return
	}

	session.Values["patient_id"] = patient.Id
	tokenBytes, err := json.Marshal(newToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to create json: %v\n", err)
		return
	}

	if _, err := w.Write(tokenBytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to write token: %v\n", err)
		return
	}

	if err := session.Save(r, w); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to save session: %v\n", err)
		return
	}

}
