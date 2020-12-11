package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/JulianaOsi/medhelp/pkg/config"
	"github.com/JulianaOsi/medhelp/pkg/store"
)

func getDirections(w http.ResponseWriter, r *http.Request) {
	token, err := jwtMiddleware(r.Header.Get("Authorization"))
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

	if claims["role"] == "patient" {
		directions, err = store.DB.GetDirectionsByPatientId(context.Background(), fmt.Sprintf("%v", claims["user_id"]))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to get directions by patient: %v\n", err)
			return
		}
	} else if claims["role"] == "registrar" {
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
	token, err := jwtMiddleware(r.Header.Get("Authorization"))
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

	if claims["role"] == "patient" {
		directions, err = store.DB.GetDirectionsByPatientId(context.Background(), fmt.Sprintf("%v", claims["user_id"]))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to get directions by patient: %v\n", err)
			return
		}
	} else if claims["role"] == "registrar" {
		directions, err = store.DB.GetDirections(context.Background())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to get directions: %v\n", err)
			return
		}
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to convert string to int: %v\n", err)
		return
	}

	for i := range directions {
		if directions[i].Id == id {
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
				return
			}
			return
		}
	}
	w.WriteHeader(http.StatusUnauthorized)
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

func registrationHandler(w http.ResponseWriter, r *http.Request) {
	type registrar struct {
		Secret string `json:"secret"`
	}
	type patient struct {
		Lastname     string `json:"lastname"`
		PolicyNumber string `json:"policy_number"`
	}
	type form struct {
		Username  string     `json:"username"`
		Password  string     `json:"password"`
		Registrar *registrar `json:"registrar"`
		Patient   *patient   `json:"patient"`
	}
	cred := form{
		Registrar: nil,
		Patient:   nil,
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("registrationHandler(): failed to read body %v\n", err)
		return
	}

	err = json.Unmarshal(body, &cred)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("registrationHandler(): failed to unmarshal json: %v\n", err)
		return
	}

	user, err := store.DB.GetUserByUsername(context.Background(), cred.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("registrationHandler(): %v\n", err)
		return
	}

	if user == nil {
		var claims = jwt.MapClaims{}
		if cred.Registrar != nil {
			if true { // TODO проверка на registrar secret, добавить информации в токен
				err = store.DB.CreateUser(context.Background(), cred.Username, cred.Password, "registrar")
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					logrus.Errorf("registrationHandler(): %v\n", err)
					return
				}

				claims = jwt.MapClaims{
					"role":     "registrar",
					"username": cred.Username,
					"exp":      time.Now().Add(time.Hour * 24).Unix(),
				}
			}
		} else if cred.Patient != nil {
			existingPatient, err := store.DB.GetPatientByLastNameAndPolicyNumber(context.Background(), cred.Patient.Lastname, cred.Patient.PolicyNumber)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				logrus.Errorf("registrationHandler(): %v\n", err)
				return
			}

			if existingPatient == nil {
				if _, err = w.Write([]byte("There is no such patient")); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					logrus.Errorf("registrationHandler(): failed to write response %v\n", err)
					return
				}
				return
			}

			err = store.DB.CreateUser(context.Background(), cred.Username, cred.Password, "patient")
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				logrus.Errorf("registrationHandler(): %v\n", err)
				return
			}

			err = store.DB.AddRelatedIdToUser(context.Background(), cred.Username, existingPatient.Id)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				logrus.Errorf("registrationHandler(): %v\n", err)
				return
			}

			claims = jwt.MapClaims{
				"role":       "patient",
				"username":   cred.Username,
				"patient_id": existingPatient.Id,
				"exp":        time.Now().Add(time.Hour * 24).Unix(),
			}
		} else {
			if _, err = w.Write([]byte("Need more info in request")); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				logrus.Errorf("registrationHandler(): failed to write response %v\n", err)
				return
			}
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(config.SigningKey)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("registrationHandler(): failed to sign jwt%v\n", err)
			return
		}

		if _, err = w.Write([]byte(tokenString)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("registrationHandler(): failed to write token %v\n", err)
			return
		}
		return
	}
	if _, err = w.Write([]byte("User already exists")); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("registrationHandler(): failed to write response %v\n", err)
		return
	}
	return
}

func authenticationHandler(w http.ResponseWriter, r *http.Request) {
	type form struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	cred := form{}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("authenticationHandler(): failed to read body: %v\n", err)
		return
	}

	err = json.Unmarshal(body, &cred)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("authenticationHandler(): failed to unmarshal json: %v\n", err)
		return
	}

	user, err := store.DB.GetUserByUsername(context.Background(), cred.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("authenticationHandler(): %v\n", err)
		return
	}

	if user == nil {
		if _, err = w.Write([]byte("User or password not found")); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("authenticationHandler(): failed to write response %v\n", err)
			return
		}
	}

	cond, err := store.DB.IsPasswordCorrect(context.Background(), cred.Username, cred.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("authenticationHandler(): failed to read body: %v\n", err)
		return
	}

	if *cond {
		var claims = jwt.MapClaims{}
		if user.Role == "registrar" {
			claims = jwt.MapClaims{
				"role":     "registrar",
				"username": user.Username,
				"exp":      time.Now().Add(time.Hour * 24).Unix(),
			}
		}
		if user.Role == "patient" {
			claims = jwt.MapClaims{
				"role":       "patient",
				"username":   user.Username,
				"patient_id": user.RelatedId,
				"exp":        time.Now().Add(time.Hour * 24).Unix(),
			}

		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(config.SigningKey)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("authenticationHandler(): failed to sign jwt%v\n", err)
			return
		}

		if _, err = w.Write([]byte(tokenString)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("authenticationHandler(): failed to write token %v\n", err)
			return
		}
		return
	}
	if _, err = w.Write([]byte("User or password not found")); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("authenticationHandler(): failed to write response %v\n", err)
		return
	}
	return
}

func jwtMiddleware(tokenString string) (*jwt.Token, error) {
	tokenString = strings.Replace(tokenString, "Bearer ", "", 1)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return config.SigningKey, nil
	})
	if err != nil {
		logrus.Errorf("failed to parse token: %v\n", err)
		return nil, err
	}

	return token, nil
}

func uploadAnalysisFile(w http.ResponseWriter, r *http.Request) {
	file, handler, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to get file: %v\n", err)
		return
	}
	defer file.Close()

	vars := mux.Vars(r)
	directionId := vars["id"]

	err = store.DB.UploadAnalysisFile(context.Background(), handler.Filename, file, directionId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to upload analysis file: %v\n", err)
		return
	}
}

func getAnalysisFiles(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	directionId := vars["id"]

	analysisFiles, err := store.DB.GetAnalysisFiles(context.Background(), directionId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to get analysis file: %v\n", err)
		return
	}

	analysisFielsBytes, err := json.MarshalIndent(analysisFiles, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to marshal analysis files: %v\n", err)
		return
	}

	if _, err := w.Write(analysisFielsBytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to write analysis files: %v\n", err)
	}
}
