package sessions

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/Vladroon22/Test-Task-BackDev/internal/database"
)

type Session struct {
	UserID       int
	UserIP       string
	RegTime      time.Time
	ExpireTime   time.Duration
	RefreshToken string
	Email        string
	mu           sync.Mutex
	repo         *database.Repo
}

func NewSessions(r *database.Repo) *Session {
	return &Session{
		repo: r,
	}
}

func (s *Session) CheckSession(id int, ip string, dur time.Duration, sess *database.MySession) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var res bool
	go func(res *bool) {
		*res = expiresTime(dur)
	}(&res)

	if sess.UserID != id {
		return "", errors.New("No-such-userID")
	} else {
		if sess.UserIP != ip {
			s.DeleteSession(id)
			log.Println("Session deleted: IP-address was changed")
			return "Session deleted: IP-address was changed", nil
		} else if !res {
			s.DeleteSession(id)
			log.Println("Session deleted: token's time is expired")
			return "Session deleted: token's time is expired", nil
		} else {
			return "OK", nil
		}
	}
}

func expiresTime(d time.Duration) bool {
	tk := time.NewTimer(d)
	<-tk.C
	return tk.Stop()
}

func (s *Session) DeleteSession(id int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.repo.DeleteSessionFromDB(id)
}
