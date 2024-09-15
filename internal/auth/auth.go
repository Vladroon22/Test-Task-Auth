package auth

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

var signKey = os.Getenv("JWT")

type MyClaims struct {
	jwt.StandardClaims
	Id int    `json:"id"`
	IP string `json:"ip"`
}

type MyTokens struct {
	JWT string
	RT  string
}

func GenerateJWT(ip string, id int, TTL time.Duration) (string, error) {
	JWT, err := jwt.NewWithClaims(jwt.SigningMethodHS512, &MyClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(TTL).Unix(), // TTL of token
			IssuedAt:  time.Now().Unix(),
		},
		id,
		ip,
	}).SignedString([]byte(signKey))
	if err != nil {
		return "", err
	}

	return JWT, nil
}

func GenerateRT() (string, error) {
	rt := make([]byte, 32)

	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)

	n, err := r.Read(rt)
	if err != nil {
		return "", err
	}
	str := base64.StdEncoding.EncodeToString(rt[:n])
	HashRT, err := bcrypt.GenerateFromPassword([]byte(str), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(HashRT), nil
}

func ValidateRT(storedHash, providedRT string) error {
	decodedRT, err := base64.StdEncoding.DecodeString(providedRT)
	if err != nil {
		return fmt.Errorf("failed to decode provided refresh token: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHash), decodedRT)
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return errors.New(err.Error())
		}
		return errors.New(err.Error())
	}

	return nil
}
