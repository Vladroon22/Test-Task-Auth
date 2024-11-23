package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Vladroon22/Test-Task-BackDev/config"
	"github.com/Vladroon22/Test-Task-BackDev/internal/auth"
	"github.com/Vladroon22/Test-Task-BackDev/internal/database"
	"github.com/Vladroon22/Test-Task-BackDev/internal/mailer"
	"github.com/Vladroon22/Test-Task-BackDev/internal/service"
	"github.com/Vladroon22/Test-Task-BackDev/internal/sessions"
	"github.com/gorilla/mux"
)

const (
	AccessTTL  = time.Minute * 10
	RefreshTTL = time.Minute * 60
)

type Handlers struct {
	db   *database.Repo
	srv  *service.Service
	sess *sessions.Session
	cnf  *config.Config
}

func NewHandler(d *database.Repo, s *service.Service, session *sessions.Session, c *config.Config) *Handlers {
	return &Handlers{
		db:   d,
		srv:  s,
		sess: session,
		cnf:  c,
	}
}

func SetCookie(w http.ResponseWriter, cookieName string, cookies string, exp time.Duration) {
	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    cookies,
		Path:     "/",
		Secure:   false,
		HttpOnly: true,
		Expires:  time.Now().Add(exp),
	}
	http.SetCookie(w, cookie)
}

func ClearCookie(w http.ResponseWriter, cookieName string, cookies string) {
	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    cookies,
		Path:     "/",
		Expires:  time.Unix(0, 0),
		Secure:   false,
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
}

func (h *Handlers) GetPair(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, _ := strconv.Atoi(vars["id"])

	tokens, err := auth.GenerateTokens(r.RemoteAddr, userID, AccessTTL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		log.Println(err)
		return
	}

	if err := h.srv.SaveSession(userID, h.cnf.Email, r.RemoteAddr, time.Now().Add(RefreshTTL), tokens.RT); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	SetCookie(w, "jwt", tokens.JWT, AccessTTL)
	SetCookie(w, "refresh", tokens.RT, RefreshTTL)

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"Session": "OK",
	})
}

type input struct {
	refresh string `json:"refresh"`
}

func (h *Handlers) MakeRefresh(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ID, _ := strconv.Atoi(vars["id"])

	inp := input{}
	if err := json.NewDecoder(r.Body).Decode(&inp.refresh); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	session, err := h.srv.GetSession(ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	resp := h.sess.CheckSession(ID, r.RemoteAddr, RefreshTTL, session)
	log.Println(resp)

	currJWT := CheckJWT(w, r)

	if resp != "OK" {
		ClearCookie(w, "jwt", "")
		ClearCookie(w, "refresh", "")

		tokens, err := auth.RefreshTokens(currJWT, inp.refresh, session.RefreshToken, AccessTTL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			log.Println(err)
			return
		}

		SetCookie(w, "jwt", tokens.JWT, AccessTTL)
		SetCookie(w, "refresh", tokens.RT, RefreshTTL)

		go func() {
			h.SendMail(w, resp, session)
		}()
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"status": resp,
	})
}

func CheckJWT(w http.ResponseWriter, r *http.Request) string {
	cookie, err := r.Cookie("jwt")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return ""
	}
	if cookie.Value == "" {
		http.Error(w, "Cookie is empty", http.StatusUnauthorized)
		return ""
	}
	return cookie.Value
}

func (h *Handlers) SendMail(w http.ResponseWriter, resp string, session *database.MySession) {
	sender, err := mailer.NewSender(session.Email, h.cnf.AppPass, "smtp.mail.ru", 587)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	err = sender.Send(&mailer.EmailInput{
		To:      session.Email,
		Subject: "WarningMessage",
		Body:    resp,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
}

func WriteJSON(w http.ResponseWriter, status int, a interface{}) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(a)
}
