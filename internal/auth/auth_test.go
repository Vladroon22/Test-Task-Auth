package auth

import (
	"crypto/rand"
	"encoding/base64"
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

	err = bcrypt.CompareHashAndPassword([]byte(rt), []byte(rt))
	assert.Error(t, err, "bcrypt.CompareHashAndPassword should return an error")
}

func TestValidateRT(t *testing.T) {
	rt := make([]byte, 32)
	_, err := rand.Read(rt)
	assert.NoError(t, err)

	encodedRT := base64.StdEncoding.EncodeToString(rt)
	hash, err := bcrypt.GenerateFromPassword(rt, bcrypt.DefaultCost)
	assert.NoError(t, err)

	_, err = ValidateRT(string(hash), encodedRT)
	assert.Error(t, err, "ValidateRT should return an error for invalid token")
}
