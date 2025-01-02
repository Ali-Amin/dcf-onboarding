package clients

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"clever.secure-onboard.com/internal/config"
	"clever.secure-onboard.com/pkg/interfaces"
	"github.com/google/uuid"
)

type TPMCLIClient struct {
	cfg    config.TCPCLIConfig
	logger interfaces.Logger
}

func NewTPMCLIClient(cfg config.TCPCLIConfig, logger interfaces.Logger) *TPMCLIClient {
	return &TPMCLIClient{cfg: cfg, logger: logger}
}

func (c *TPMCLIClient) Sign(data string) (string, error) {
	// Flush TPM context before starting
	flushCMD := "tpm2_flushcontext"
	result := exec.Command(flushCMD, "-t")

	stdout, err := result.CombinedOutput()
	if err != nil {
		c.logger.Error(
			fmt.Sprintf(
				"failed to flush context: %s : %s",
				err.Error(),
				stdout,
			),
		)
		return "", err
	}

	digestOutPath := "/tmp/" + uuid.NewString()
	signatureOutPath := "/tmp/" + uuid.NewString()

	// TPM tool expects digest in a file
	digest := sha256.Sum256([]byte(data))
	err = os.WriteFile(digestOutPath, digest[:], os.ModeAppend)
	if err != nil {
		c.logger.Error(err.Error())
		return "", err
	}

	// Use tpm2_sign to sign the data and write to signatureOutPath
	result = exec.Command(
		"tpm2_sign",
		"-Q",
		"-c", c.cfg.PublicKey,
		"-g", "sha256",
		"-d",
		"-f", "plain",
		"-o", signatureOutPath,
		digestOutPath,
	)
	c.logger.Write(slog.LevelDebug, "Running command: "+result.String())
	stdout, err = result.CombinedOutput()
	if err != nil {
		c.logger.Error(
			fmt.Sprintf(
				"failed to sign data: %s : %s",
				err.Error(),
				stdout,
			),
		)
		return "", err
	}

	signature, err := os.ReadFile(signatureOutPath)
	if err != nil {
		c.logger.Error(err.Error())
		return "", err
	}

	signatureHEX := hex.EncodeToString(signature)
	c.logger.Write(slog.LevelDebug, "signature: "+signatureHEX)

	// Clean up
	defer func() {
		exec.Command(flushCMD, "-t")
		err := os.Remove(digestOutPath)
		if err != nil {
			c.logger.Error("could not remove digest file: " + err.Error())
		}
		err = os.Remove(signatureOutPath)
		if err != nil {
			c.logger.Error("could not remove signature file: " + err.Error())
		}
	}()

	return signatureHEX, nil
}

func (c *TPMCLIClient) HasTPM() (bool, error) {
	hasTPM := true
	result := exec.Command("tpm2_pcrread")
	if _, err := result.Output(); err != nil {
		hasTPM = false
	}
	return hasTPM, nil
}
