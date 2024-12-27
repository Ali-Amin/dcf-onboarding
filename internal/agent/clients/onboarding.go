package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"clever.secure-onboard.com/pkg/interfaces"
	"clever.secure-onboard.com/pkg/requests"
	"clever.secure-onboard.com/pkg/responses"
)

type OnboardingServerClient struct {
	baseURL  string
	deviceID string
	logger   interfaces.Logger
}

func NewOnboardingServerClient(
	deviceID, url string,
	logger interfaces.Logger,
) *OnboardingServerClient {
	return &OnboardingServerClient{baseURL: url, deviceID: deviceID, logger: logger}
}

func (c *OnboardingServerClient) SendTPMStatus(hasTPM bool) (statusCode int, err error) {
	url := fmt.Sprintf("%s/device/%s/tpm", c.baseURL, c.deviceID)

	var request requests.UploadDeviceTPMStatus
	request.HasTPM = hasTPM

	body, err := json.Marshal(request)
	if err != nil {
		c.logger.Error("Failed to parse upload device TPM status request: " + err.Error())
		return 0, err
	}
	res, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		c.logger.Error("Failed to send TPM status: " + err.Error())
		return 0, err
	}

	c.logger.Write(slog.LevelDebug, fmt.Sprintf("Received response: %+v", res))

	return res.StatusCode, nil
}

func (c *OnboardingServerClient) RequestChallenge() (challenge string, err error) {
	url := fmt.Sprintf("%s/device/%s/challenge", c.baseURL, c.deviceID)

	res, err := http.Get(url)
	if err != nil {
		c.logger.Error("Failed to send challenge request: " + err.Error())
		return "", err
	}

	c.logger.Write(slog.LevelDebug, fmt.Sprintf("Received response: %+v", res))

	body, _ := io.ReadAll(res.Body)

	var response responses.GenerateChallenge
	err = json.Unmarshal(body, &response)
	if err != nil {
		c.logger.Error("Failed to parse generate challenge response: " + err.Error())
		return "", err
	}

	return response.Challenge, nil
}

func (c *OnboardingServerClient) SendChallengeAnswer(signature string) (passed bool, err error) {
	url := fmt.Sprintf("%s/device/%s/challenge", c.baseURL, c.deviceID)

	var request requests.UploadChallengeAnswer
	request.Signature = signature

	body, err := json.Marshal(request)
	if err != nil {
		c.logger.Error("Failed to parse send challenge answer request: " + err.Error())
		return false, err
	}
	res, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		c.logger.Error("Failed to challenge answer request: " + err.Error())
		return false, err
	}
	c.logger.Write(slog.LevelDebug, fmt.Sprintf("Received response: %+v", res))

	resBody, _ := io.ReadAll(res.Body)

	var response responses.VerifyAnswer
	err = json.Unmarshal(resBody, &response)
	if err != nil {
		c.logger.Error("Failed to parse verify answer response: " + err.Error())
		return false, err
	}

	return response.Passed, nil
}
