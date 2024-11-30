//go:build sasl
// +build sasl

package mongo

import (
	"github.com/lzw5399/go-common-public/library/database/mongo/internal/sasl"
)

func saslNew(cred Credential, host string) (saslStepper, error) {
	return sasl.New(cred.Username, cred.Password, cred.Mechanism, cred.Service, host)
}
