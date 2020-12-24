package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/JulianaOsi/medhelp/pkg/config"
	"github.com/JulianaOsi/medhelp/pkg/store"
)

func addDirection(w http.ResponseWriter, r *http.Request) {
	setupCorsResponse(&w) //CORS
	type Direction struct {
		Patient             store.NewPatient `json:"patient"`
		Doctor              store.NewDoctor  `json:"doctor"`
		Date                time.Time        `json:"date"`
		IcdCode             string           `json:"icd_code"`
		MedicalOrganization string           `json:"medical_organization"`
		OrganizationContact string           `json:"organization_contact"`
		Justification       string           `json:"justification"`
	}
	type directions struct {
		Directions []Direction `json:"directions"`
	}

	type status struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	type response struct {
		Status status `json:"status"`
	}

	var resp response
	resp.Status.Status = "ok"

	token, err := jwtMiddleware(r.Header.Get("Authorization"))
	if err != nil {
		logrus.Errorf("failed to parse token: %v\n", err)
		resp.Status.Status = "error"
		resp.Status.Message = err.Error()
		respBytes, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to marshall response: %v\n", err)
			return
		}

		if _, err := w.Write(respBytes); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to write response: %v\n", err)
		}
		return
	}

	var claims = token.Claims.(jwt.MapClaims)
	var isAccess = false

	if claims["role"] == "registrar" {
		isAccess = true
	}

	if isAccess == false {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("WWW-Authenticate", "Bearer realm=\"Access to the direction\", charset=\"UTF-8\"")
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to read body: %v\n", err)
		return
	}
	update := directions{}

	err = json.Unmarshal(body, &update)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to unmarshal json: %v\n", err)
		return
	}

	for _, j := range update.Directions {
		patientId, err := store.DB.AddPatient(context.Background(), j.Patient)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to add patient: %v\n", err)
			return
		}

		doctorId, err := store.DB.AddDoctor(context.Background(), j.Doctor)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to add doctor: %v\n", err)
			return
		}

		direction := store.NewDirection{
			PatientId:           *patientId,
			DoctorId:            *doctorId,
			Date:                j.Date,
			IcdCode:             j.IcdCode,
			MedicalOrganization: j.MedicalOrganization,
			OrganizationContact: j.OrganizationContact,
			Justification:       j.Justification,
		}

		err = store.DB.AddDirection(context.Background(), direction)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to add direction: %v\n", err)
			return
		}
	}
}

func getDirections(w http.ResponseWriter, r *http.Request) {
	setupCorsResponse(&w) //CORS
	type status struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	type response struct {
		Status     status             `json:"status"`
		Directions []*store.Direction `json:"directions"`
	}

	var resp response
	resp.Status.Status = "ok"

	token, err := jwtMiddleware(r.Header.Get("Authorization"))
	if err != nil {
		logrus.Errorf("failed to parse token: %v\n", err)
		resp.Status.Status = "error"
		resp.Status.Message = err.Error()
		respBytes, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to marshall response: %v\n", err)
			return
		}

		w.Header().Set("content-type", "application/json")
		if _, err := w.Write(respBytes); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to write response: %v\n", err)
		}
		return
	}

	var claims = token.Claims.(jwt.MapClaims)

	if claims["role"] == "patient" {
		resp.Directions, err = store.DB.GetDirectionsByPatientId(context.Background(), fmt.Sprintf("%v", claims["patient_id"]))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to get directions by patient id: %v\n", err)
			return
		}
	} else if claims["role"] == "registrar" {
		resp.Directions, err = store.DB.GetDirections(context.Background())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to get directions: %v\n", err)
			return
		}
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to marshall response: %v\n", err)
		return
	}

	w.Header().Set("content-type", "application/json")
	if _, err := w.Write(respBytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to write response: %v\n", err)
	}
	return
}

func getDirectionAnalysis(w http.ResponseWriter, r *http.Request) {
	setupCorsResponse(&w) //CORS
	type status struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	type response struct {
		Status   status            `json:"status"`
		Analysis []*store.Analysis `json:"analysis"`
	}

	var resp response
	resp.Status.Status = "ok"

	token, err := jwtMiddleware(r.Header.Get("Authorization"))
	if err != nil {
		logrus.Errorf("failed to parse token: %v\n", err)
		resp.Status.Status = "error"
		resp.Status.Message = err.Error()
		respBytes, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to marshall response: %v\n", err)
			return
		}

		w.Header().Set("content-type", "application/json")
		if _, err := w.Write(respBytes); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to write response: %v\n", err)
		}
		return
	}

	var claims = token.Claims.(jwt.MapClaims)
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to convert string to int: %v\n", err)
		return
	}

	if claims["role"] == "registrar" {
		resp.Analysis, err = store.DB.GetAnalysisByDirectionId(context.Background(), id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to get analysis by direction id: %v\n", err)
			return
		}
	} else if claims["role"] == "patient" {
		directions, err := store.DB.GetDirectionsByPatientId(context.Background(), fmt.Sprintf("%v", claims["patient_id"]))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to get directions by patient id: %v\n", err)
			return
		}

		for i := range directions {
			if directions[i].Id == id {
				resp.Analysis, err = store.DB.GetAnalysisByDirectionId(context.Background(), id)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					logrus.Errorf("failed to get analysis by direction id: %v\n", err)
					return
				}
			}
		}
	}

	if resp.Analysis == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to marshall response: %v\n", err)
		return
	}

	w.Header().Set("content-type", "application/json")
	if _, err := w.Write(respBytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to write response: %v\n", err)
	}
	return
}

func setDirectionStatus(w http.ResponseWriter, r *http.Request) {
	setupCorsResponse(&w) //CORS
	type directionUpdate struct {
		DirectionId int `json:"directionId"`
		Status      int `json:"status"`
	}

	type status struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	type response struct {
		Status status `json:"status"`
	}

	var resp response
	resp.Status.Status = "ok"
	token, err := jwtMiddleware(r.Header.Get("Authorization"))
	if err != nil {
		logrus.Errorf("failed to parse token: %v\n", err)
		resp.Status.Status = "error"
		resp.Status.Message = err.Error()
		respBytes, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to marshall response: %v\n", err)
			return
		}

		w.Header().Set("content-type", "application/json")
		if _, err := w.Write(respBytes); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to write response: %v\n", err)
		}
		return
	}

	var claims = token.Claims.(jwt.MapClaims)
	var isAccess = false

	if claims["role"] == "registrar" {
		isAccess = true
	}

	if isAccess == false {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("WWW-Authenticate", "Bearer realm=\"Access to the direction\", charset=\"UTF-8\"")
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to read body: %v\n", err)
		return
	}
	update := directionUpdate{}

	err = json.Unmarshal(body, &update)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to unmarshal json: %v\n", err)
		return
	}

	err = store.DB.SetDirectionStatus(context.Background(), update.DirectionId, update.Status)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to set direction status: %v\n", err)
		return
	}
}

func setAnalysisCheck(w http.ResponseWriter, r *http.Request) {
	setupCorsResponse(&w) //CORS
	type analysisUpdate struct {
		AnalysisId int  `json:"analysisId"`
		Checked    bool `json:"checked"`
	}

	type status struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	type response struct {
		Status status `json:"status"`
	}

	var resp response
	resp.Status.Status = "ok"
	token, err := jwtMiddleware(r.Header.Get("Authorization"))
	if err != nil {
		logrus.Errorf("failed to parse token: %v\n", err)
		resp.Status.Status = "error"
		resp.Status.Message = err.Error()
		respBytes, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to marshall response: %v\n", err)
			return
		}

		w.Header().Set("content-type", "application/json")
		if _, err := w.Write(respBytes); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to write response: %v\n", err)
		}
		return
	}

	var claims = token.Claims.(jwt.MapClaims)
	var isAccess = false

	if claims["role"] == "registrar" {
		isAccess = true
	}

	if isAccess == false {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("WWW-Authenticate", "Bearer realm=\"Access to the analysis\", charset=\"UTF-8\"")
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to read body: %v\n", err)
		return
	}
	update := analysisUpdate{}

	err = json.Unmarshal(body, &update)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to unmarshal json: %v\n", err)
		return
	}

	err = store.DB.SetAnalysisState(context.Background(), update.AnalysisId, update.Checked)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to set analysis state: %v\n", err)
		return
	}
}

func registrationHandler(w http.ResponseWriter, r *http.Request) {
	setupCorsResponse(&w) //CORS
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

	type status struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	type response struct {
		Status status `json:"status"`
		Token  string `json:"token"`
	}

	cred := form{
		Registrar: nil,
		Patient:   nil,
	}

	var resp response
	resp.Status.Status = "ok"
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to read body: %v\n", err)
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
		logrus.Errorf("failed to get user by username: %v\n", err)
		return
	}

	if user == nil {
		var claims = jwt.MapClaims{}
		if cred.Registrar != nil {
			if true { // TODO проверка на registrar secret, добавить информации в токен
				err = store.DB.CreateUser(context.Background(), cred.Username, cred.Password, "registrar")
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					logrus.Errorf("failed to create user: %v\n", err)
					return
				}

				claims = jwt.MapClaims{
					"role":     "registrar",
					"username": cred.Username,
					"exp":      time.Now().Add(time.Hour * 24).Unix(),
				}
			}
		} else if cred.Patient != nil {
			existingPatient, err := store.DB.GetPatient(context.Background(), cred.Patient.Lastname, cred.Patient.PolicyNumber)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				logrus.Errorf("failed to get patient: %v\n", err)
				return
			}

			if existingPatient == nil {
				resp.Status.Status = "info"
				resp.Status.Message = "There is no such patient"
				respBytes, err := json.Marshal(resp)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					logrus.Errorf("failed to marshall response: %v\n", err)
					return
				}

				w.Header().Set("content-type", "application/json")
				if _, err := w.Write(respBytes); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					logrus.Errorf("failed to write response: %v\n", err)
				}
				return
			}

			cond, err := store.DB.IsRelatedIdSet(context.Background(), existingPatient.Id)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				logrus.Errorf("failed to check related patient: %v\n", err)
				return
			}

			if *cond {
				resp.Status.Status = "info"
				resp.Status.Message = "This patient already registered"
				respBytes, err := json.Marshal(resp)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					logrus.Errorf("failed to marshall response: %v\n", err)
					return
				}

				w.Header().Set("content-type", "application/json")
				if _, err := w.Write(respBytes); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					logrus.Errorf("failed to write response: %v\n", err)
				}
				return
			}

			err = store.DB.CreateUser(context.Background(), cred.Username, cred.Password, "patient")
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				logrus.Errorf("failed to create user: %v\n", err)
				return
			}

			err = store.DB.AddRelatedIdToUser(context.Background(), cred.Username, existingPatient.Id)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				logrus.Errorf("failed to add related id to user: %v\n", err)
				return
			}

			claims = jwt.MapClaims{
				"role":       "patient",
				"username":   cred.Username,
				"patient_id": existingPatient.Id,
				"exp":        time.Now().Add(time.Hour * 24).Unix(),
			}
		} else {
			resp.Status.Status = "info"
			resp.Status.Message = "Need more info in request"
			respBytes, err := json.Marshal(resp)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				logrus.Errorf("failed to marshall response: %v\n", err)
				return
			}

			w.Header().Set("content-type", "application/json")
			if _, err := w.Write(respBytes); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				logrus.Errorf("failed to write response: %v\n", err)
			}
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		resp.Token, err = token.SignedString(config.SigningKey)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to sign JWT: %v\n", err)
			return
		}

		respBytes, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to marshall response: %v\n", err)
			return
		}

		w.Header().Set("content-type", "application/json")
		if _, err := w.Write(respBytes); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to write response: %v\n", err)
		}
		return
	}
	resp.Status.Status = "info"
	resp.Status.Message = "User already exists"
	respBytes, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to marshall response: %v\n", err)
		return
	}

	w.Header().Set("content-type", "application/json")
	if _, err := w.Write(respBytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to write response: %v\n", err)
	}
	return
}

func authenticationHandler(w http.ResponseWriter, r *http.Request) {
	setupCorsResponse(&w) //CORS
	type form struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	type status struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	type response struct {
		Status status `json:"status"`
		Token  string `json:"token"`
	}

	cred := form{}
	var resp response
	resp.Status.Status = "ok"

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to read body: %v\n", err)
		return
	}

	err = json.Unmarshal(body, &cred)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to unmarshall json: %v\n", err)
		return
	}

	user, err := store.DB.GetUserByUsername(context.Background(), cred.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to get user by username: %v\n", err)
		return
	}

	if user == nil {
		resp.Status.Status = "info"
		resp.Status.Message = "User or password not found"
		respBytes, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to marshall response: %v\n", err)
			return
		}

		w.Header().Set("content-type", "application/json")
		if _, err := w.Write(respBytes); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to write response: %v\n", err)
		}
		return
	}

	cond, err := store.DB.IsPasswordCorrect(context.Background(), cred.Username, cred.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to check password: %v\n", err)
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
		resp.Token, err = token.SignedString(config.SigningKey)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to sign JWT: %v\n", err)
			return
		}

		respBytes, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to marshall response: %v\n", err)
			return
		}

		w.Header().Set("content-type", "application/json")
		if _, err := w.Write(respBytes); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to write response: %v\n", err)
		}
		return
	}
	resp.Status.Status = "info"
	resp.Status.Message = "User or password not found"
	respBytes, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to marshall response: %v\n", err)
		return
	}

	w.Header().Set("content-type", "application/json")
	if _, err := w.Write(respBytes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to write response: %v\n", err)
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

func corsSkip(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	return
}

func setupCorsResponse(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Authorization")
}

func uploadAnalysisFile(w http.ResponseWriter, r *http.Request) {
	setupCorsResponse(&w) //CORS
	type status struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	type response struct {
		Status status `json:"status"`
	}

	var resp response
	resp.Status.Status = "ok"
	token, err := jwtMiddleware(r.Header.Get("Authorization"))
	if err != nil {
		logrus.Errorf("failed to parse token: %v\n", err)
		resp.Status.Status = "error"
		resp.Status.Message = err.Error()
		respBytes, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to marshall response: %v\n", err)
			return
		}

		w.Header().Set("content-type", "application/json")
		if _, err := w.Write(respBytes); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to write response: %v\n", err)
		}
		return
	}

	vars := mux.Vars(r)
	analysisId, err := strconv.Atoi(vars["analysis"])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to convert string to int: %v\n", err)
		return
	}

	var claims = token.Claims.(jwt.MapClaims)
	var isAccess = false

	if claims["role"] == "patient" {
		directions, err := store.DB.GetDirectionsByPatientId(context.Background(), fmt.Sprintf("%v", claims["patient_id"]))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to get directions by patient id: %v\n", err)
			return
		}

		for _, j := range directions {
			analysis, err := store.DB.GetAnalysisByDirectionId(context.Background(), j.Id)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				logrus.Errorf("failed to get analysis by direction id: %v\n", err)
				return
			}

			for _, n := range analysis {
				if n.Id == analysisId {
					isAccess = true
				}
			}
		}
	}

	if claims["role"] == "registrar" {
		isAccess = true
	}

	if isAccess == false {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("WWW-Authenticate", "Bearer realm=\"Access to the patient analysis\", charset=\"UTF-8\"")
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to get file: %v\n", err)
		return
	}
	defer file.Close()

	fileId, err := store.DB.SaveFile(context.Background(), file, handler.Filename)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to save analysis file: %v\n", err)
		return
	}

	err = store.DB.SetAnalysisFile(context.Background(), analysisId, *fileId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to set analysis file: %v\n", err)
		return
	}

	return
}

func downloadAnalysisFile(w http.ResponseWriter, r *http.Request) {
	setupCorsResponse(&w) //CORS
	type status struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	type response struct {
		Status status `json:"status"`
	}

	var resp response
	resp.Status.Status = "ok"
	token, err := jwtMiddleware(r.Header.Get("Authorization"))
	if err != nil {
		logrus.Errorf("failed to parse token: %v\n", err)
		resp.Status.Status = "error"
		resp.Status.Message = err.Error()
		respBytes, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to marshall response: %v\n", err)
			return
		}

		w.Header().Set("content-type", "application/json")
		if _, err := w.Write(respBytes); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to write response: %v\n", err)
		}
		return
	}

	vars := mux.Vars(r)
	analysisId, err := strconv.Atoi(vars["analysis"])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to convert string to int: %v\n", err)
		return
	}

	var claims = token.Claims.(jwt.MapClaims)
	var isAccess = false

	if claims["role"] == "patient" {
		directions, err := store.DB.GetDirectionsByPatientId(context.Background(), fmt.Sprintf("%v", claims["patient_id"]))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to get directions by patient id: %v\n", err)
			return
		}

		for _, j := range directions {
			analysis, err := store.DB.GetAnalysisByDirectionId(context.Background(), j.Id)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				logrus.Errorf("failed to get analysis by direction id: %v\n", err)
				return
			}

			for _, n := range analysis {
				if n.Id == analysisId {
					isAccess = true
				}
			}
		}
	}

	if claims["role"] == "registrar" {
		isAccess = true
	}

	if isAccess == false {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("WWW-Authenticate", "Bearer realm=\"Access to the patient analysis\", charset=\"UTF-8\"")
		return
	}

	analysis, err := store.DB.GetAnalysisById(context.Background(), analysisId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to get analysis by id: %v\n", err)
		return
	}

	if analysis == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if analysis.FileId == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	filePath, err := store.DB.GetFilepath(context.Background(), *analysis.FileId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logrus.Errorf("failed to get filepath: %v\n", err)
		return
	}

	if filePath != nil {
		streamBytes, err := ioutil.ReadFile(*filePath)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to read file: %v\n", err)
			return
		}
		ext := filepath.Ext(*filePath)
		b := bytes.NewBuffer(streamBytes)
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", analysis.Name+ext))
		w.Header().Set("Content-Type", "multipart/form-data")

		_, err = w.Write(b.Bytes())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Errorf("failed to response with file: %v\n", err)
			return
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
	return
}
