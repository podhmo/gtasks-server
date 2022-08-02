package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/podhmo/flagstruct"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/tasks/v1"
)

type Config struct {
	ClientID     string `flag:"client-id"`
	ClientSecret string `flag:"client-secret"`
	RedirectURL  string `flag:"redirect-url"`
	APIKey       string `flag:"api-key"`
}

func main() {
	config := Config{RedirectURL: "http://localhost:8888"}
	if err := flagstruct.Parse(&config); err != nil {
		log.Printf("! %+v", err)
		return
	}
	if err := run(config); err != nil {
		log.Printf("!! %+v", err)
	}
}

var SCOPES = []string{
	"https://www.googleapis.com/auth/tasks",
}

func run(config Config) error {
	// Your credentials should be obtained from the Google
	// Developer Console (https://console.developers.google.com).
	conf := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL + "/auth/callback",
		Scopes:       SCOPES,
		Endpoint:     google.Endpoint,
	}

	u, err := url.Parse(config.RedirectURL)
	if err != nil {
		return fmt.Errorf("invalid url: %q -- %w", config.RedirectURL, err)
	}

	state := "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	auth := &Auth{state: state, OauthConfig: conf, Store: NewAPIKeyStore()}

	mux := &http.ServeMux{}
	mux.HandleFunc("/auth/login", auth.Login)
	mux.HandleFunc("/auth/callback", auth.Callback)
	mux.Handle("/", Handler(auth))
	return http.ListenAndServe(fmt.Sprintf(":%s", u.Port()), mux)
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
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.TokenMap[key]
	return v, ok
}
func (s *APIKeyStore) SetToken(key string, token *oauth2.Token) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TokenMap[key] = token
}

type Auth struct {
	OauthConfig *oauth2.Config
	state       string

	Store TokenStore
}

func (h *Auth) State(req *http.Request) string {
	return h.state // TODO: get state from request
}
func (h *Auth) RedirectURL() string {
	return h.OauthConfig.RedirectURL
}
func (h *Auth) Login(w http.ResponseWriter, req *http.Request) {
	conf := h.OauthConfig
	state := h.State(req)

	// Redirect user to Google's consent page to ask for permission
	// for the scopes specified above.
	authURL := conf.AuthCodeURL(state)
	fmt.Printf("Visit the URL for the auth dialog: %v\n", authURL)

	w.Header().Set("Location", authURL)
	w.WriteHeader(http.StatusFound)
}

func (h *Auth) Callback(w http.ResponseWriter, req *http.Request) {
	state := h.State(req) // TODO: get state from store
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
	w.Header().Set("Location", "http://localhost:8888") // TODO: fixme
	w.WriteHeader(http.StatusFound)
}

func Handler(auth *Auth) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		apikey := auth.state
		tok, ok := auth.Store.GetToken(apikey) // TODO: get from request
		if !ok {
			auth.Login(w, req)
			return
		}

		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")

		client := auth.OauthConfig.Client(req.Context(), tok)
		s, err := tasks.New(client)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			enc.Encode(map[string]interface{}{
				"error": fmt.Sprintf("unexpected error: %+v", err),
				"token": tok,
			})
			return
		}
		res, err := s.Tasklists.List().Do()
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			enc.Encode(map[string]interface{}{
				"error": fmt.Sprintf("unexpected error: %+v", err),
				"token": tok,
			})
			return
		}
		enc.Encode(res)
	})
}
