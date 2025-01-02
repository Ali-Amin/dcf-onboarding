package verifier

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"sync"

	"clever.secure-onboard.com/pkg/contracts"
	"github.com/google/uuid"
	alvarium "github.com/project-alvarium/alvarium-sdk-go/pkg/interfaces"
)

type ChallengeIdentityVerifier struct {
	keyMX         *sync.RWMutex
	inMemKeyStore map[string][]byte

	challengeMX         *sync.RWMutex
	inMemChallengeStore map[string]string

	alvariumSDK alvarium.Sdk
}

func NewChallengeIdentityVerifier(alvariumSDK alvarium.Sdk) *ChallengeIdentityVerifier {
	return &ChallengeIdentityVerifier{
		keyMX:         &sync.RWMutex{},
		inMemKeyStore: make(map[string][]byte),

		challengeMX:         &sync.RWMutex{},
		inMemChallengeStore: make(map[string]string),

		alvariumSDK: alvariumSDK,
	}
}

func (v *ChallengeIdentityVerifier) ReceivePublicKey(deviceID string, key string) error {
	v.keyMX.Lock()
	defer v.keyMX.Unlock()
	pubKeyPEM, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return errors.New("public key is not base64 encoded")
	}
	block, _ := pem.Decode(pubKeyPEM)
	if block == nil {
		return errors.New("Public key is not in PEM format")
	}
	v.inMemKeyStore[deviceID] = block.Bytes
	return nil
}

func (v *ChallengeIdentityVerifier) GenerateChallenge(deviceID string) (string, error) {
	challengeToken := uuid.NewString()
	v.challengeMX.Lock()
	v.inMemChallengeStore[deviceID] = challengeToken
	v.challengeMX.Unlock()

	return challengeToken, nil
}

func (v *ChallengeIdentityVerifier) VerifyAnswer(deviceID, signature string) (bool, error) {
	v.challengeMX.RLock()
	challenge, ok := v.inMemChallengeStore[deviceID]
	v.challengeMX.RUnlock()

	if !ok {
		return false, nil
	}

	challengeDigest := sha256.Sum256([]byte(challenge))

	v.keyMX.RLock()
	pubKeyPEM, ok := v.inMemKeyStore[deviceID]
	v.keyMX.RUnlock()

	if !ok {
		return false, nil
	}

	sig, err := hex.DecodeString(signature)
	if err != nil {
		return false, err
	}

	pubKey, err := x509.ParsePKIXPublicKey(pubKeyPEM)
	if err != nil {
		return false, errors.New("bad rsa public key provided: " + err.Error())
	}
	rsaPub, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return false, errors.New("bad rsa public key provided")
	}
	err = rsa.VerifyPKCS1v15(rsaPub, crypto.SHA256, challengeDigest[:], sig)
	if err != nil {
		ctx := context.WithValue(context.Background(), contracts.HasTPM, false)
		v.alvariumSDK.Create(ctx, []byte(deviceID))
		return false, err
	}

	ctx := context.WithValue(context.Background(), contracts.HasTPM, true)
	v.alvariumSDK.Create(ctx, []byte(deviceID))
	return true, nil
}

func (v *ChallengeIdentityVerifier) HandleTPMStatus(deviceID string, hasTPM bool) error {
	// We publish this annotation in the false scenario only, because if the device has
	// a TPM signature verification is still underway to complete identity verification
	if hasTPM == false {
		ctx := context.WithValue(context.Background(), contracts.HasTPM, hasTPM)
		v.alvariumSDK.Create(ctx, []byte(deviceID))
	}
	return nil
}
