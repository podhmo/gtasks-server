package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"golang.org/x/oauth2"
)

type TokenStore interface {
	GetToken(key string) (*oauth2.Token, bool)
	SetToken(key string, token *oauth2.Token)
}

type KeyGenerator interface {
	GenerateKey(req *http.Request, token *oauth2.Token, salt string) (string, error)
}

type Auth struct {
	OauthConfig *oauth2.Config
	Salt        string

	Store      TokenStore
	KeyGen     KeyGenerator
	DefaultURL string
}

func (h *Auth) RedirectURL() string {
	return h.OauthConfig.RedirectURL
}
func (h *Auth) Login(w http.ResponseWriter, req *http.Request) {
	conf := h.OauthConfig

	// Redirect user to Google's consent page to ask for permission
	// for the scopes specified above.
	authURL := conf.AuthCodeURL(h.Salt)
	fmt.Printf("Visit the URL for the auth dialog: %v\n", authURL)

	w.Header().Set("Location", authURL)
	w.WriteHeader(http.StatusFound)
}

func (h *Auth) Callback(w http.ResponseWriter, req *http.Request) {
	conf := h.OauthConfig

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	q := req.URL.Query()
	if h.Salt != q.Get("state") {
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

	apikey, err := h.KeyGen.GenerateKey(req, tok, h.Salt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(map[string]interface{}{
			"error": fmt.Sprintf("something wrong in generate key: %+v", err),
			"scope": q.Get("scope"),
			"query": req.URL.Query(),
			"token": tok,
		})
		return
	}

	h.Store.SetToken(apikey, tok)

	qs := url.Values{}
	qs.Add("apikey", apikey)
	w.Header().Set("Location", h.DefaultURL+"?"+qs.Encode())
	w.WriteHeader(http.StatusFound)
}

func (h *Auth) WithOauthToken(defaultKey string) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			apikey := req.URL.Query().Get("apikey")
			if apikey == "" {
				apikey = defaultKey
			}

			tok, ok := h.Store.GetToken(apikey) // TODO: get from request
			if !ok {
				h.Login(w, req)
				return
			}
			req = req.WithContext(context.WithValue(req.Context(), tokenKey, tok))
			handler.ServeHTTP(w, req)
		})
	}
}

var DEBUG bool

func init() {
	if ok, _ := strconv.ParseBool(os.Getenv("DEBUG")); ok {
		DEBUG = ok
	}
}
