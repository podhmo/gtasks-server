package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/podhmo/flagstruct"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/tasks/v1"
)

type Config struct {
	ClientID     string `flag:"client-id"`
	ClientSecret string `flag:"client-secret"`
	RedirectURL  string `flag:"redirect-url"`
}

func main() {
	config := Config{}
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
		RedirectURL:  config.RedirectURL,
		Scopes:       SCOPES,
		Endpoint:     google.Endpoint,
	}

	u, err := url.Parse(config.RedirectURL)
	if err != nil {
		return fmt.Errorf("invalid url: %q -- %w", config.RedirectURL, err)
	}

	// Redirect user to Google's consent page to ask for permission
	// for the scopes specified above.
	state := "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	authURL := conf.AuthCodeURL(state)
	fmt.Printf("Visit the URL for the auth dialog: %v\n", authURL)

	return http.ListenAndServe(fmt.Sprintf(":%s", u.Port()), Handler(conf, state))
}

func Handler(conf *oauth2.Config, state string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
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

		client := conf.Client(oauth2.NoContext, tok)
		s, err := tasks.New(client)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			enc.Encode(map[string]interface{}{
				"error": fmt.Sprintf("unexpected error: %+v", err),
				"scope": q.Get("scope"),
				"token": tok,
			})
			return
		}
		res, err := s.Tasklists.List().Do()
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			enc.Encode(map[string]interface{}{
				"error": fmt.Sprintf("unexpected error: %+v", err),
				"scope": q.Get("scope"),
				"token": tok,
			})
			return
		}
		enc.Encode(res)
	})
}
