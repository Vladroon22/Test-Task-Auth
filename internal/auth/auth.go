package auth

import (
	"context"
	"encoding/base64"
	"errors"
	"math/rand"
	"net/http"
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

func AuthMiddleWare(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("jwt")
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		if cookie.Value == "" {
			http.Error(w, "Cookie is empty", http.StatusUnauthorized)
			return
		}
		claims, err := ValidateToken(cookie.Value)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		r = r.WithContext(context.WithValue(r.Context(), "id", claims.Id))

		next(w, r)
	})
}
