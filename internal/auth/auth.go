package auth

import (
	"encoding/base64"
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
