package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/podhmo/flagstruct"
	"github.com/podhmo/gtasks-server/auth"
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
	auth := &auth.Auth{
		OauthConfig: conf,
		State:       state,
		Store:       auth.NewAPIKeyStore(),
		DefaultURL:  "http://localhost:8888/",
	}

	mux := &http.ServeMux{}
	mux.HandleFunc("/auth/login", auth.Login)
	mux.HandleFunc("/auth/callback", auth.Callback)
	mux.Handle("/", auth.WithToken(Handler(auth.OauthConfig)))
	u.Path = ""
	log.Println("listening ...", u.String())
	return http.ListenAndServe(fmt.Sprintf(":%s", u.Port()), mux)
}

func Handler(conf *oauth2.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")

		tok := auth.GetToken(ctx)
		client := conf.Client(ctx, tok)
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
