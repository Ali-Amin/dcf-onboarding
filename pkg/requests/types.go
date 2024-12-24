package requests

type UploadDevicePublicKey struct {
	PublicKey string `json:"publicKey,omitempty"`
}

type UploadDeviceTPMStatus struct {
	HasTPM bool `json:"hasTpm,omitempty"`
}

type UploadChallengeAnswer struct {
	Signature string `json:"signature,omitempty"`
}
