package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Vladroon22/Test-Task-BackDev/internal/auth"
	"github.com/Vladroon22/Test-Task-BackDev/internal/database"
	"github.com/Vladroon22/Test-Task-BackDev/internal/service"
	"github.com/Vladroon22/Test-Task-BackDev/internal/sessions"
	"github.com/gorilla/mux"
)

const (
	AccessTTL  = time.Minute * 15
	RefreshTTL = time.Minute * 60
)

type Handlers struct {
	db   *database.Repo
	srv  *service.Service
	sess *sessions.Session
}

func NewHandler(d *database.Repo, s *service.Service) *Handlers {
	return &Handlers{
		db:   d,
		srv:  s,
		sess: sessions.NewSessions(),
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

func (h *Handlers) GetPair(w http.ResponseWriter, r *http.Request) {
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

	if err := h.srv.SaveSession(userID, "1234@mail.ru", r.RemoteAddr, time.Now(), RefreshTTL, refreshToken); err != nil {
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

func (h *Handlers) MakeRefresh(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ID, _ := strconv.Atoi(vars["id"])

	_, err := h.srv.GetToken(ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Panicln(err)
		return
	}

	session, err := h.srv.GetSession(ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Panicln(err)
		return
	}

	var resp string
	go func(resp *string) {
		*resp, err = h.sess.CheckSession(ID, r.RemoteAddr, RefreshTTL, session)
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
