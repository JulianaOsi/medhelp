package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/JulianaOsi/medhelp/pkg/config"
	"github.com/JulianaOsi/medhelp/pkg/store"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func getDirections(w http.ResponseWriter, r *http.Request) {

	token, err := jwtMiddleware(r.Header.Get("Bearer"))
	if err != nil {
		_, err = w.Write([]byte("Invalid token. Unauthorized user have no access. Please log in."))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to write response message: %v\n", err)
			return
		}
		return
	}

	var directions []*store.Direction
	var claims = token.Claims.(jwt.MapClaims)

	if claims["type"] == "patient" {
		directions, err = store.DB.GetDirectionsByPatientId(context.Background(), fmt.Sprintf("%v", claims["user_id"]))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to get directions by patient: %v\n", err)
			return
		}
	} else if claims["type"] == "admin" {
		directions, err = store.DB.GetDirections(context.Background())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to get directions: %v\n", err)
			return
		}
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

	var claims = jwt.MapClaims{ // TODO заносится только тип patient. Нужно подумать
		"type":    "patient",
		"user_id": patient.Id,
		"name":    patient.LastName + " " + patient.FirstName,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(config.SigningKey)

	if _, err = w.Write([]byte(tokenString)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to write token: %v\n", err)
		return
	}

}

func jwtMiddleware(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return config.SigningKey, nil
	})
	if err != nil {
		logrus.Errorf("failed to parse token: %v\n", err)
		return nil, err
	}

	return token, nil
}
