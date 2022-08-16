package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/podhmo/flagstruct"
	"github.com/podhmo/gtasks-server/auth"
	"github.com/podhmo/quickapi"
	"github.com/podhmo/quickapi/experimental/define"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/tasks/v1"
)

type Config struct {
	ClientID     string `flag:"client-id"`
	ClientSecret string `flag:"client-secret"`
	RedirectURL  string `flag:"redirect-url"`

	GenDoc bool `flag:"gendoc" help:"generate openapi.json to stdout"`
}

func main() {
	config := Config{RedirectURL: "http://localhost:8888"}
	flagstruct.Parse(&config)
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

	auth := &auth.Auth{
		OauthConfig: conf,
		Salt:        ":me:",
		Store:       auth.NewInmemoryStore(),
		KeyGen:      &auth.UUIDKeyGenerator{},
		DefaultURL:  "http://localhost:8888/",
	}

	ctx := context.Background()
	router := quickapi.DefaultRouter()
	router.Get("/auth/login", auth.Login)
	router.Get("/auth/callback", auth.Callback)

	bc, err := define.NewBuildContext(define.Doc().Server(auth.DefaultURL, "local development"), router)
	if err != nil {
		return fmt.Errorf("build context: %w", err)
	}

	{
		define.Get(bc, "/", ListTaskList(auth.OauthConfig),
			auth.WithOauthToken(":default-key:"),
		).OperationID("ListTaskList")
	}

	if config.GenDoc {
		return bc.EmitDoc(ctx, os.Stdout)
	}

	h, err := bc.BuildHandler(ctx)
	if err != nil {
		return fmt.Errorf("build handler: %w", err)
	}

	u, err := url.Parse(config.RedirectURL)
	if err != nil {
		return fmt.Errorf("invalid url: %q -- %w", config.RedirectURL, err)
	}
	u.Path = ""
	log.Println("listening ...", u.String())
	srv := quickapi.NewServer(fmt.Sprintf(":%s", u.Port()), h, 5*time.Second)
	return srv.ListenAndServe(ctx)
}

func ListTaskList(conf *oauth2.Config) quickapi.Action[quickapi.Empty, []*tasks.TaskList] {
	return func(ctx context.Context, input quickapi.Empty) ([]*tasks.TaskList, error) {
		tok, apiErr := auth.GetToken(ctx)
		if apiErr != nil {
			return nil, apiErr
		}

		client := conf.Client(ctx, tok)
		s, err := tasks.New(client)
		if err != nil {
			return nil, quickapi.NewAPIError(err, http.StatusUnauthorized)
		}
		res, err := s.Tasklists.List().Do()
		if err != nil {
			return nil, quickapi.NewAPIError(err, http.StatusInternalServerError)
		}
		return res.Items, nil
	}
}
