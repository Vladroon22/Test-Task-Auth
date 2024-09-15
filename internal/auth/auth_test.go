package auth

import (
	"crypto/rand"
	"encoding/base64"
	"os"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestGenerateJWT(t *testing.T) {
	signKey := os.Getenv("JWT")

	id := 123
	ip := "127.0.0.1"
	TTL := time.Minute * 15

	JWT, err := GenerateJWT(ip, id, TTL)
	assert.NoError(t, err)
	assert.NotEmpty(t, JWT)

	token, err := jwt.ParseWithClaims(JWT, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(signKey), nil
	})
	assert.NoError(t, err)

	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid {
		assert.Equal(t, id, claims.Id)
		assert.Equal(t, ip, claims.IP)
	} else {
		t.Errorf("Invalid JWT")
	}
}

func TestGenerateRT(t *testing.T) {
	rt, err := GenerateRT()
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

	err = ValidateRT(string(hash), encodedRT)
	assert.NoError(t, err)

	invalidRT := "invalid_token"
	err = ValidateRT(string(hash), invalidRT)
	assert.Error(t, err, "ValidateRT should return an error for invalid token")
}
