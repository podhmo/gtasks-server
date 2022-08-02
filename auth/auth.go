package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

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

func (h *Auth) WithOauthToken(handler http.Handler) http.Handler {
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

var DEBUG bool

func init() {
	if ok, _ := strconv.ParseBool(os.Getenv("DEBUG")); ok {
		DEBUG = ok
	}
}
