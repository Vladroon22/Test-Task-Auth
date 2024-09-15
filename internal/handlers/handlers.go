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
		SameSite: http.SameSiteDefaultMode,
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

	tokens, err := GenerateTokens(r.RemoteAddr, userID, AccessTTL)
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

func (h *Handlers) MakeRefresh(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ID, _ := strconv.Atoi(vars["id"])

	session, err := h.srv.GetSession(ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	resp := h.sess.CheckSession(ID, r.RemoteAddr, RefreshTTL, session)
	log.Println(resp)

	if resp != "OK" {
		ClearCookie(w, "jwt", "")
		ClearCookie(w, "refresh", "")

		tokens, err := GenerateTokens(r.RemoteAddr, ID, AccessTTL)
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

func GenerateTokens(ip string, id int, TTL time.Duration) (*auth.MyTokens, error) {
	token, err := auth.GenerateJWT(ip, id, TTL)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if token == "" {
		log.Println("token is empty")
		return nil, err
	}

	refreshToken, err := auth.GenerateRT()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &auth.MyTokens{JWT: token, RT: refreshToken}, nil
}

func WriteJSON(w http.ResponseWriter, status int, a interface{}) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(a)
}
