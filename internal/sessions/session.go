package sessions

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/Vladroon22/Test-Task-BackDev/internal/database"
)

type Session struct {
	db           *database.Storage
	UserID       int
	UserIP       string
	RegTime      time.Time
	ExpireTime   time.Duration
	RefreshToken string
	Email        string
	mu           sync.Mutex
}

func NewSessions(db *database.Storage) *Session {
	return &Session{db: db}
}

func (s *Session) CheckSession(id int, ip string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var res bool
	go func(res *bool) {
		*res = s.expiresTime()
	}(&res)

	sess, err := s.db.GetSession(id)
	if err != nil {
		return "", err
	}
	if sess.UserID != id {
		return "", errors.New("No-such-userID")
	} else {
		if sess.userIP != ip {
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

func (s *Session) expiresTime() bool {
	tk := time.NewTimer(s.ExpireTime)
	<-tk.C
	return tk.Stop()
}

func (s *Session) DeleteSession(id int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.db.DeleteSessionFromDB(id)
}
