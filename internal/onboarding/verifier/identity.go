package verifier

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
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
	pubKey, err := hex.DecodeString(key)
	if err != nil {
		return err
	}
	v.inMemKeyStore[deviceID] = pubKey
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

	v.keyMX.RLock()
	pubKey, ok := v.inMemKeyStore[deviceID]
	v.keyMX.RUnlock()

	if !ok {
		return false, nil
	}

	sig, err := hex.DecodeString(signature)
	if err != nil {
		return false, err
	}

	passed := ed25519.Verify(pubKey, []byte(challenge), sig)
	if !passed {
		ctx := context.WithValue(context.Background(), contracts.HasTPM, false)
		v.alvariumSDK.Create(ctx, []byte(deviceID))
		return false, nil
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
