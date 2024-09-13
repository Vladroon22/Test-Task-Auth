package sessions

import (
	"errors"
	"log"
	"sync"
	"time"

	d "github.com/Vladroon22/Test-Task-BackDev/internal/database"
)

type Session struct {
	db           *d.Storage
	UserID       int
	UserIP       string
	RegTime      time.Time
	ExpireTime   time.Duration
	RefreshToken string
	Email        string
	mu           sync.Mutex
}

func NewSessions(db *d.Storage) *Session {
	return &Session{db: db}
}

func (s *Session) CheckSession(id int, ip string, dur time.Duration) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var res bool
	go func(res *bool) {
		*res = expiresTime(dur)
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

func expiresTime(d time.Duration) bool {
	tk := time.NewTimer(d)
	<-tk.C
	return tk.Stop()
}

func (s *Session) DeleteSession(id int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.db.DeleteSessionFromDB(id)
}
