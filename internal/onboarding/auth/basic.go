package auth

import (
	"errors"
	"fmt"
	"os"

	"clever.secure-onboard.com/pkg/contracts"
	"clever.secure-onboard.com/pkg/interfaces"
)

type FixedBasicAuth struct {
	// These values represent the correct values for which login attempts
	// will be validated against
	validUsername string
	validPassword string
	logger        interfaces.Logger
}

func NewFixedBasicAuth(logger interfaces.Logger) (*FixedBasicAuth, error) {
	username, ok := os.LookupEnv(contracts.TrustedActorUsername)
	if !ok {
		logger.Error(fmt.Sprintf("Value of %s cannot be empty", contracts.TrustedActorUsername))
		return nil, errors.New(
			fmt.Sprintf("Value of %s cannot be empty", contracts.TrustedActorUsername),
		)
	}

	password, ok := os.LookupEnv(contracts.TrustedActorPassowrd)
	if !ok {
		logger.Error(fmt.Sprintf("Value of %s cannot be empty", contracts.TrustedActorPassowrd))
		return nil, errors.New(
			fmt.Sprintf("Value of %s cannot be empty", contracts.TrustedActorPassowrd),
		)
	}

	return &FixedBasicAuth{logger: logger, validUsername: username, validPassword: password}, nil
}

func (a *FixedBasicAuth) Authenticate(username, password string) bool {
	if username == a.validUsername {
		if password == a.validPassword {
			return true
		}
	}
	return false
}
