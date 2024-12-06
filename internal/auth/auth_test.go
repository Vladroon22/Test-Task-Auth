package auth

import (
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

const (
	AccessTTL = time.Minute * 10
)

func TestGenerateJWT(t *testing.T) {
	id := 123
	ip := "127.0.0.1"

	JWT, err := GenerateJWT(ip, id, AccessTTL)
	assert.NoError(t, err)
	assert.NotEmpty(t, JWT)

	if _, err := ValidateToken(JWT); err != nil {
		t.Errorf("Invalid JWT")
	}
}

func TestGenerateRT(t *testing.T) {
	id := 123
	ip := "127.0.0.1"

	rt, err := GenerateRT(id, ip)
	assert.NoError(t, err)
	assert.NotEmpty(t, rt)
}

func TestValidateRT(t *testing.T) {
	id := 123
	ip := "127.0.0.1"

	payload := refreshClaims{ID: id, IP: ip}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Error(err)
	}
	encodedRT := base64.StdEncoding.EncodeToString(payloadBytes)

	hash, err := bcrypt.GenerateFromPassword([]byte(encodedRT), bcrypt.DefaultCost)
	assert.NoError(t, err)

	_, err = ValidateRT(string(hash), encodedRT)
	assert.NoError(t, err)
}
