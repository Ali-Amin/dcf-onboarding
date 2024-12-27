package clients

import "os"

type TPMClient struct{}

func NewTPMClient() *TPMClient {
	return &TPMClient{}
}

func (c *TPMClient) HasTPM() bool {
	hasTPM := true
	fi, err := os.Stat("/dev/tpm0")
	if err == nil {
		// TPM mounted at default path
		if fi.Mode()&os.ModeDevice != 0 || fi.Mode()&os.ModeSocket != 0 {
			hasTPM = true
		}
	}
	return hasTPM
}
