package database

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Vladroon22/Test-Task-BackDev/internal/auth"
	"github.com/Vladroon22/Test-Task-BackDev/internal/database"
	"github.com/Vladroon22/Test-Task-BackDev/internal/sessions"
	"github.com/gorilla/mux"
)

const (
	AccessTTL  = 15 * time.Second
	RefreshTTL = time.Hour
)

type Repo struct {
	db   *database.Storage
	sess *sessions.Session
}

func NewRepo(d *database.Storage) *Repo {
	return &Repo{
		db:   d,
		sess: sessions.NewSessions(d),
	}
}

func SetCookie(w http.ResponseWriter, cookieName string, cookies string) {
	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    cookies,
		Path:     "/",
		Secure:   false,
		HttpOnly: true,
		Expires:  time.Now().Add(AccessTTL),
	}
	http.SetCookie(w, cookie)
}

func (rp *Repo) GetPair(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, _ := strconv.Atoi(vars["id"])

	token, err := auth.GenerateJWT(r.RemoteAddr, userID, AccessTTL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		log.Panicln(err)
		return
	}
	if token == "" {
		http.Error(w, "token is empty", http.StatusUnauthorized)
		log.Panicln("token is empty")
		return
	}

	refreshToken, err := auth.GenerateRT()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Panicln(err)
		return
	}

	if err := rp.db.SaveSession(userID, "1234@mail.ru", r.RemoteAddr, time.Now(), RefreshTTL, refreshToken); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Panicln(err)
		return
	}

	SetCookie(w, "jwt", token)
	SetCookie(w, "refresh", refreshToken)

	WriteJSON(w, http.StatusOK, map[string]interface{}{ // надо убрать
		"access":  token,
		"refresh": refreshToken,
	})
}

func (rp *Repo) MakeRefresh(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ID, _ := strconv.Atoi(vars["id"])

	_, err := rp.db.GetToken(ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Panicln(err)
		return
	}

	var resp string
	go func(resp *string) {
		*resp, err = rp.sess.CheckSession(ID, r.RemoteAddr, RefreshTTL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			log.Panicln(err)
			return
		}
	}(&resp)

	if resp != "OK" {
		token, err := auth.GenerateJWT(r.RemoteAddr, ID, AccessTTL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			log.Panicln(err)
			return
		}
		if token == "" {
			http.Error(w, "token is empty", http.StatusUnauthorized)
			log.Panicln("token is empty")
			return
		}

		refreshToken, err := auth.GenerateRT()

		SetCookie(w, "jwt", token)
		SetCookie(w, "refresh", refreshToken)
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"status": resp,
	})
}

func WriteJSON(w http.ResponseWriter, status int, a interface{}) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(a)
}
