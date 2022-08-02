package auth

import (
	"log"
	"sync"

	"golang.org/x/oauth2"
)

type InmemoryStore struct {
	mu       sync.Mutex
	TokenMap map[string]*oauth2.Token // apikey -> token
}

func (s *InmemoryStore) GetToken(key string) (*oauth2.Token, bool) {
	if DEBUG {
		log.Printf("get token: %q", key)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.TokenMap[key]
	return v, ok
}
func (s *InmemoryStore) SetToken(key string, token *oauth2.Token) {
	if DEBUG {
		log.Printf("set token: %q", key)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TokenMap[key] = token
}

func NewInmemoryStore() *InmemoryStore {
	return &InmemoryStore{
		TokenMap: map[string]*oauth2.Token{},
	}
}

var _ TokenStore = (*InmemoryStore)(nil)
