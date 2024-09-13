package auth

import (
	"encoding/base64"
	"errors"
	"math/rand"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const (
	signKey = "g78tn8&*^T%RYGY^&T"
)

type MyClaims struct {
	jwt.StandardClaims
	Id int    `json:"id"`
	IP string `json:"ip"`
}

func ValidateToken(tokenStr string) (*MyClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		return signKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return nil, errors.New("Invalid-Signature")
		}
		return nil, errors.New("Bad-Request")
	}

	claims, ok := token.Claims.(*MyClaims)
	if !ok || !token.Valid {
		return nil, errors.New("Unauthorized")
	}

	return claims, nil
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
	return str, nil
}
