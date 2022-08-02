package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"golang.org/x/oauth2"
)

type Auth struct {
	OauthConfig *oauth2.Config
	State       string

	Store      TokenStore
	DefaultURL string
}

func (h *Auth) GetState(req *http.Request) string {
	return h.State // TODO: get state from request
}
func (h *Auth) RedirectURL() string {
	return h.OauthConfig.RedirectURL
}
func (h *Auth) Login(w http.ResponseWriter, req *http.Request) {
	conf := h.OauthConfig
	state := h.GetState(req)

	// Redirect user to Google's consent page to ask for permission
	// for the scopes specified above.
	authURL := conf.AuthCodeURL(state)
	fmt.Printf("Visit the URL for the auth dialog: %v\n", authURL)

	w.Header().Set("Location", authURL)
	w.WriteHeader(http.StatusFound)
}

func (h *Auth) Callback(w http.ResponseWriter, req *http.Request) {
	state := h.GetState(req) // TODO: get state from store
	conf := h.OauthConfig

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	q := req.URL.Query()
	if state != q.Get("state") {
		w.WriteHeader(http.StatusUnauthorized)
		enc.Encode(map[string]interface{}{
			"error":  "invalid state",
			"scope":  q.Get("scope"),
			"header": req.Header,
			"query":  req.URL.Query(),
		})
		return
	}

	code := q.Get("code")
	tok, err := conf.Exchange(req.Context(), code)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		enc.Encode(map[string]interface{}{
			"error":  fmt.Sprintf("invalid code: %+v", err),
			"scope":  q.Get("scope"),
			"header": req.Header,
			"query":  req.URL.Query(),
		})
		return
	}

	h.Store.SetToken(q.Get("state"), tok)
	w.Header().Set("Location", h.DefaultURL)
	w.WriteHeader(http.StatusFound)
}

func (h *Auth) WithToken(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		apikey := h.State
		tok, ok := h.Store.GetToken(apikey) // TODO: get from request
		if !ok {
			h.Login(w, req)
			return
		}
		req = req.WithContext(context.WithValue(req.Context(), tokenKey, tok))
		handler.ServeHTTP(w, req)
	})
}

// ----------------------------------------

type ctxkey string

const (
	tokenKey ctxkey = "token key"
)

func GetToken(ctx context.Context) *oauth2.Token {
	return ctx.Value(tokenKey).(*oauth2.Token)
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

var DEBUG bool

func init() {
	if ok, _ := strconv.ParseBool(os.Getenv("DEBUG")); ok {
		DEBUG = ok
	}
}
