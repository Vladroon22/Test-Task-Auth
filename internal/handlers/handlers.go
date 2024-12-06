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
	"github.com/gorilla/mux"
)

const (
	AccessTTL  = time.Minute * 10
	RefreshTTL = time.Minute * 60
)

type Handlers struct {
	repo *database.Repo
	cnf  *config.Config
}

func NewHandler(r *database.Repo, c *config.Config) *Handlers {
	return &Handlers{
		repo: r,
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

	if err := h.repo.SaveSession(userID, h.cnf.Email, r.RemoteAddr, time.Now().Add(RefreshTTL), tokens.RT); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	SetCookie(w, "jwt", tokens.JWT, AccessTTL)

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"RT": tokens.RT,
	})
}

type input struct {
	Refresh string `json:"refresh"`
}

func (h *Handlers) MakeRefresh(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ID, _ := strconv.Atoi(vars["id"])

	inp := input{}
	if err := json.NewDecoder(r.Body).Decode(&inp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	session, err := h.repo.GetSession(ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	/*
		if session.UserIP != r.RemoteAddr {
			http.Error(w, "Mismatched user's IP", http.StatusForbidden)
			log.Println("Mismatched user's IP")
			return
		}
	*/

	jwt := CheckJWT(w, r)

	ClearCookie(w, "jwt", "")

	tokens, err := auth.RefreshTokens(jwt, inp.Refresh, session.RefreshToken, AccessTTL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		log.Println(err)
		return
	}

	SetCookie(w, "jwt", tokens.JWT, AccessTTL)

	go func() {
		h.sendMail(w, "OK", session)
	}()

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"Access": tokens.JWT,
		"RT":     tokens.RT,
	})
}

func (h *Handlers) sendMail(w http.ResponseWriter, resp string, session *database.MySession) {
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

func CheckJWT(w http.ResponseWriter, r *http.Request) string {
	cookie, err := r.Cookie("jwt")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		log.Println(err)
		return ""
	}
	if cookie.Value == "" {
		log.Println(err)
		return ""
	}
	return cookie.Value
}

func WriteJSON(w http.ResponseWriter, status int, a interface{}) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(a)
}
