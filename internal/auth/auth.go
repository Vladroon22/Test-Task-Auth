package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

var signKey = []byte(os.Getenv("JWT"))

type MyClaims struct {
	jwt.StandardClaims
	ID int    `json:"id"`
	IP string `json:"ip"`
}

type MyTokens struct {
	JWT string
	RT  string
}

// generate new jwt token
func GenerateJWT(ip string, id int, TTL time.Duration) (string, error) {
	JWT, err := jwt.NewWithClaims(jwt.SigningMethodHS512, &MyClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(TTL).Unix(), // TTL of token
			IssuedAt:  time.Now().Unix(),
		},
		id,
		ip,
	}).SignedString(signKey)
	if err != nil {
		return "", err
	}

	return JWT, nil
}

type refreshClaims struct {
	ID int    `json:"id"`
	IP string `json:"ip"`
}

// generate new refresh token
func GenerateRT(id int, ip string) (string, error) {
	payload := refreshClaims{ID: id, IP: ip}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", errors.New("failed to marshal refresh payload: %w" + err.Error())
	}
	encodedRT := base64.StdEncoding.EncodeToString(payloadBytes)
	hash, err := bcrypt.GenerateFromPassword([]byte(encodedRT), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New(err.Error())
	}

	return string(hash), nil
}

// validate stored refresh token
func ValidateRT(storedHash, providedRT string) (*refreshClaims, error) {
	decodedPayload, err := base64.StdEncoding.DecodeString(providedRT)
	if err != nil {
		return nil, fmt.Errorf("failed to decode token payload: %w", err)
	}

	var payload refreshClaims
	if err := json.Unmarshal(decodedPayload, &payload); err != nil {
		return nil, errors.New("failed to unmarshal token payload: %w" + err.Error())
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(providedRT))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return nil, errors.New("mismatched hash and password")
		}
		return nil, errors.New("failed to compare hash and password: %w" + err.Error())
	}

	return &payload, nil
}

// generate new tokens - jwt and refresh
func GenerateTokens(ip string, id int, ttlJWT time.Duration) (*MyTokens, error) {
	jwtToken, err := GenerateJWT(ip, id, ttlJWT)
	if err != nil {
		return nil, errors.New("failed to generate JWT: %w" + err.Error())
	}

	rtToken, err := GenerateRT(id, ip)
	if err != nil {
		return nil, errors.New("failed to generate RT: %w" + err.Error())
	}

	return &MyTokens{
		JWT: jwtToken,
		RT:  rtToken,
	}, nil
}

// refresh-token make refresh pair of token based on JWT и RT
func RefreshTokens(jwt, rtStr, storedHash string, ttl time.Duration) (*MyTokens, error) {
	claims, err := ValidateToken(jwt)
	if err != nil {
		return nil, err
	}

	payload, err := ValidateRT(storedHash, rtStr)
	if err != nil {
		return nil, errors.New("refresh token validation failed: %w" + err.Error())
	}

	newTokens, err := GenerateTokens(payload.IP, payload.ID, ttl)
	if err != nil {
		return nil, errors.New("failed to generate new tokens: %w" + err.Error())
	}

	if claims.IP == payload.IP {
		return nil, errors.New("tokens are unlinked")
	}

	return newTokens, nil
}

// validate jwt
func ValidateToken(tokenStr string) (*MyClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		return signKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return nil, errors.New("Unauthorized")
		}
		return nil, errors.New("Bad-Request")
	}

	claims, ok := token.Claims.(*MyClaims)
	if !ok || !token.Valid {
		return nil, errors.New("Unauthorized")
	}

	return claims, nil
}
