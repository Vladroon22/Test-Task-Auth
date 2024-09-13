package database

import (
	"database/sql"
	"log"
	"time"

	s "github.com/Vladroon22/Test-Task-BackDev/internal/sessions"
)

func (db *Storage) GetToken(id int) (string, error) {
	var refreshToken string
	query := "SELECT refresh FROM sessions WHERE id = $1"
	err := db.sql.QueryRow(query, id).Scan(&refreshToken)
	if err == sql.ErrNoRows || err != nil {
		log.Panicln(err)
		return "", err
	}
	return refreshToken, nil
}

func (db *Storage) SaveSession(id int, email, ip string, regTime time.Time, expireTime time.Duration, rt string) error {
	query := "INSERT INTO sessions (userID, email, userIP, regTime, expireTime, refresh) VALUES ($1, $2, $3, $4, $5, $6)"
	if _, err := db.sql.Exec(query, id, email, ip, regTime, expireTime, rt); err != nil {
		log.Panicln(err)
		return err
	}
	return nil
}

func (db *Storage) GetSession(id int) (*s.Session, error) {
	session := &s.Session{}
	query := "SELECT FROM sessions WHERE id = $1"
	err := db.sql.QueryRow(query, id).Scan(&session.UserID, &session.Email, &session.UserIP, &session.RegTime, &session.RefreshToken, &session.ExpireTime)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (db *Storage) DeleteSessionFromDB(id int) error {
	query := "DELETE FROM sessions WHERE id = $1"
	if _, err := db.sql.Exec(query, id); err != nil {
		return err
	}
	return nil
}
