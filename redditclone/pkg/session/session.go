package session

import (
	"errors"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"time"
)

var (
	ErrNoAuth = errors.New("no session found")
)

var (
	secretToken = []byte("1A2aa2slG92ass1iW9vj18aLv10B8a1")
)

const (
	SessKey = "sessionKey"
)

type Session struct {
	ID          string
	UserID      uint32
	UserName    string
	AccessToken string
}

func NewSession(userID uint32, login string) (*Session, error) {
	sessID := uuid.New().String()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": map[string]interface{}{
			"username": login,
			"id":       userID},
		"sessID": sessID,
		"iat":    time.Now().Unix(),
		"exp":    time.Now().Add(time.Hour * 24 * 7).Unix(),
	})
	tokenString, err := token.SignedString(secretToken)
	if err != nil {
		return nil, err
	}

	return &Session{
		ID:          sessID,
		UserID:      userID,
		UserName:    login,
		AccessToken: tokenString,
	}, nil
}
