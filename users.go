package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"sync"
)

const (
	SessionNew      = 1
	SessionExisting = 2
)

type UserStore interface {
	Auth(login, password string) bool
}

type SingleUserStore struct {
	Pass string
}

func NewSingleUserStore(pass string) UserStore {
	return &SingleUserStore{Pass: pass}
}

func (s *SingleUserStore) Auth(login, password string) bool {
	return s.Pass == password
}

type UserAuth interface {
	CheckAuth(r *http.Request) (*Session, error)
}

type Session struct {
	Login, Token string
	Status       int
}

type BasicAuth struct {
	sync.Mutex
	Users    UserStore
	Secret   string
	sessions []*Session
}

func NewBasicAuth(users UserStore, secret string) UserAuth {
	return &BasicAuth{Users: users, Secret: secret, sessions: []*Session{}}
}

//
func (a *BasicAuth) CheckAuth(r *http.Request) (*Session, error) {
	s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)

	// is this basic
	if len(s) == 2 && s[0] == "Basic" {
		return a.checkBasicHeader(s)
	}

	// is this token
	if len(s) == 2 && s[0] == "Token" {
		return a.checkTokenHeader(s)
	}

	return nil, errors.New("failed to decode basic auth header")
}

func (a *BasicAuth) checkBasicHeader(s []string) (*Session, error) {

	b, err := base64.StdEncoding.DecodeString(s[1])

	if err != nil {
		return nil, errors.New("failed to decode basic auth header")
	}

	pair := strings.SplitN(string(b), ":", 2)

	if len(pair) != 2 {
		return nil, errors.New("failed to decode basic auth header")
	}

	if a.Users.Auth(pair[0], pair[1]) {
		session := &Session{pair[0], a.generateToken(pair[0]), SessionNew}
		return a.addSession(session), nil
	}

	return nil, errors.New("failed to decode basic auth header")

}

func (a *BasicAuth) checkTokenHeader(s []string) (*Session, error) {
	return a.lookupSession(s[1])
}

func (a *BasicAuth) generateToken(login string) string {

	mac := hmac.New(sha256.New, []byte(a.Secret))

	mac.Write([]byte(login))

	return hex.EncodeToString(mac.Sum(nil))
}

func (a *BasicAuth) lookupSession(token string) (*Session, error) {
	for _, session := range a.sessions {
		if session.Token == token {
			session.Status = SessionExisting
			return session, nil
		}
	}

	return nil, errors.New("Session not found")
}

func (a *BasicAuth) addSession(session *Session) *Session {
	a.Lock()
	defer a.Unlock()
	a.sessions = append(a.sessions, session)
	return session
}
