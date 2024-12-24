package onboarding

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"clever.secure-onboard.com/pkg/interfaces"
	"clever.secure-onboard.com/pkg/requests"
	"clever.secure-onboard.com/pkg/responses"
	"github.com/go-chi/chi"
)

type Router struct {
	Mux              *chi.Mux
	logger           interfaces.Logger
	deviceIdentifier interfaces.DeviceIdentityVerifier
}

func newRouter(
	identityVerifier interfaces.DeviceIdentityVerifier,
	logger interfaces.Logger,
) *chi.Mux {
	r := &Router{
		Mux:              chi.NewRouter(),
		logger:           logger,
		deviceIdentifier: identityVerifier,
	}

	r.Mux.Post("/device/{deviceID}/key", r.storeDevicePublicKey)

	r.Mux.Post("/device/{deviceID}/tpm", r.handleTPMStatus)

	r.Mux.Get("/device/{deviceID}/challenge", r.getChallengeToken)

	r.Mux.Post("/device/{deviceID}/challenge", r.verifyChallengeAnswer)
	return r.Mux
}

func (r *Router) storeDevicePublicKey(w http.ResponseWriter, req *http.Request) {
	deviceID := chi.URLParam(req, "deviceID")
	r.logger.Write(slog.LevelDebug, "received request to store public key for device "+deviceID)

	body, _ := io.ReadAll(req.Body)

	var request requests.UploadDevicePublicKey

	err := json.Unmarshal(body, &request)
	if err != nil {
		r.logger.Error("failed to parse request: " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = r.deviceIdentifier.ReceivePublicKey(deviceID, request.PublicKey)
	if err != nil {
		r.logger.Error("failed to receive public key: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (r *Router) handleTPMStatus(w http.ResponseWriter, req *http.Request) {
	deviceID := chi.URLParam(req, "deviceID")
	r.logger.Write(
		slog.LevelDebug,
		"received request to handle device TPM status for device "+deviceID,
	)

	body, _ := io.ReadAll(req.Body)

	var request requests.UploadDeviceTPMStatus

	err := json.Unmarshal(body, &request)
	if err != nil {
		r.logger.Error("failed to parse request: " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = r.deviceIdentifier.HandleTPMStatus(deviceID, request.HasTPM)
	if err != nil {
		r.logger.Error("failed to handle tpm status" + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (r *Router) getChallengeToken(w http.ResponseWriter, req *http.Request) {
	deviceID := chi.URLParam(req, "deviceID")
	r.logger.Write(
		slog.LevelDebug,
		"received request to get challenge token for device "+deviceID,
	)

	challenge, err := r.deviceIdentifier.GenerateChallenge(deviceID)
	if err != nil {
		r.logger.Error("failed to generate challenge for device " + deviceID + ": " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var res responses.GenerateChallenge
	res.Challenge = challenge

	response, err := json.Marshal(res)
	if err != nil {
		r.logger.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(response)
	w.WriteHeader(http.StatusOK)
}

func (r *Router) verifyChallengeAnswer(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusAccepted)
	deviceID := chi.URLParam(req, "deviceID")
	r.logger.Write(
		slog.LevelDebug,
		"received request to verify challenge answer for device "+deviceID,
	)

	body, _ := io.ReadAll(req.Body)

	var request requests.UploadChallengeAnswer

	err := json.Unmarshal(body, &request)
	if err != nil {
		r.logger.Error("failed to parse request: " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	passed, err := r.deviceIdentifier.VerifyAnswer(deviceID, request.Signature)
	if err != nil {
		r.logger.Error("failed to receive public key: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	r.logger.Write(slog.LevelDebug, fmt.Sprintf("Device %s signature passed: %v", deviceID, passed))

	res := responses.VerifyAnswer{Passed: passed}
	resData, err := json.Marshal(res)
	if err != nil {
		r.logger.Error("failed to parse response: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(resData)
}
