package session

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	UserID    uint64
	SessionID string
	Expiry    time.Time
}

type SessionManager struct {
	Sessions []*Session
}

func NewSessionManager() *SessionManager {
	return &SessionManager{}
}

func (session *SessionManager) CreateSession(user uint64, duration int64) Session {
	var ses = Session{
		SessionID: uuid.NewString(),
		Expiry:    time.Now().Add(time.Second * time.Duration(duration)),
		UserID:    user,
	}
	session.Sessions = append(session.Sessions, &ses)
	return ses
}
func (session *SessionManager) UserSession(user uint64) (string, error) {
	for _, s := range session.Sessions {
		if s.UserID == user {
			return s.SessionID, nil
		}
	}
	return "", fmt.Errorf("no session for user %v", user)
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
