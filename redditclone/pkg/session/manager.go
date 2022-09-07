package session

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"strings"
	"sync"
)

var (
	ErrInvalidToken = errors.New("invalid jwt token")
	ErrNoPayload    = errors.New("no payload")
	ErrBadSign      = errors.New("bad sign method")
)

type SessionsManager struct {
	data map[string]*Session
	mu   *sync.RWMutex
}

func NewSessionsManager() *SessionsManager {
	return &SessionsManager{
		data: make(map[string]*Session, 5),
		mu:   &sync.RWMutex{},
	}
}

func (sm *SessionsManager) Check(r *http.Request) (*Session, error) {
	inToken := strings.TrimPrefix(r.Header.Get("authorization"), "Bearer ")

	hashSecretGetter := func(token *jwt.Token) (interface{}, error) {
		method, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok || method.Alg() != "HS256" {
			return nil, ErrBadSign
		}
		return secretToken, nil
	}

	token, err := jwt.Parse(inToken, hashSecretGetter)
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	payload, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrNoPayload
	}

	sessID := payload["sessID"]

	sm.mu.RLock()
	sess, ok := sm.data[sessID.(string)]
	sm.mu.RUnlock()

	if !ok {
		return nil, ErrNoAuth
	}
	return sess, nil
}

func (sm *SessionsManager) Create(w http.ResponseWriter, userID uint32, login string) (*Session, error) {
	sess, err := NewSession(userID, login)
	if err != nil {
		return nil, err
	}

	sm.mu.Lock()
	sm.data[sess.ID] = sess
	sm.mu.Unlock()

	return sess, nil
}
