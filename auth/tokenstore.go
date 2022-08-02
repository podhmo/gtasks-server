package auth

import (
	"context"
	"log"
	"sync"

	"golang.org/x/oauth2"
)

type ctxkey string

const (
	tokenKey ctxkey = "token key"
)

func GetToken(ctx context.Context) *oauth2.Token {
	v := ctx.Value(tokenKey)
	if v == nil {
		return nil
	}
	return v.(*oauth2.Token)
}

type TokenStore interface {
	GetToken(key string) (*oauth2.Token, bool)
	SetToken(key string, token *oauth2.Token)
}

type APIKeyStore struct {
	mu       sync.Mutex
	TokenMap map[string]*oauth2.Token // apikey -> token
}

func NewAPIKeyStore() *APIKeyStore {
	return &APIKeyStore{TokenMap: map[string]*oauth2.Token{}}
}
func (s *APIKeyStore) GetToken(key string) (*oauth2.Token, bool) {
	if DEBUG {
		log.Printf("get token: %q", key)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.TokenMap[key]
	return v, ok
}
func (s *APIKeyStore) SetToken(key string, token *oauth2.Token) {
	if DEBUG {
		log.Printf("set token: %q", key)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TokenMap[key] = token
}
