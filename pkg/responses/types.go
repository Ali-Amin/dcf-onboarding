package responses

type GenerateChallenge struct {
	Challenge string `json:"challenge,omitempty"`
}

type VerifyAnswer struct {
	Passed bool `json:"passed,omitempty"`
}
