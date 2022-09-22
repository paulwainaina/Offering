package session

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	SessionID string
	Expiry    time.Time
}

type SessionManager struct {
	Sessions []*Session
}

func NewSessionManager() *SessionManager {
	return &SessionManager{}
}

func (session *SessionManager) CreateSession(duration int64) *Session {
	return &Session{
		SessionID: uuid.NewString(),
		Expiry:    time.Now().Add(time.Second * time.Duration(duration)),
	}
}

func (session *SessionManager) SessionExist(sessionid string) bool {
	for _, s := range session.Sessions {
		if s.SessionID == sessionid {
			return true
		}
	}
	return false
}

func (session *SessionManager) DeleteSession() {
	for {
		time.Sleep(time.Second * 1)
		if len(session.Sessions) == 0 {
			continue
		} else {
			for i, s := range session.Sessions {
				if s.Expiry.Before(time.Now()) {
					session.Sessions = append(session.Sessions[:i], session.Sessions[i+1:]...)
				}
			}
		}
	}
}
