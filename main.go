package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/podhmo/flagstruct"
	"github.com/podhmo/gtasks-server/auth"
	"github.com/podhmo/quickapi"
	"github.com/podhmo/quickapi/experimental/define"
	rohandler "github.com/podhmo/reflect-openapi/handler"
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
		DefaultURL:  "http://localhost:8888/api/tasklist",
	}

	ctx := context.Background()
	router := quickapi.DefaultRouter()
	router.Use(middleware.StripSlashes)

	router.Get("/auth/login", auth.Login)
	router.Get("/auth/callback", auth.Callback)

	doc := define.Doc().Server(strings.TrimSuffix(auth.DefaultURL, "/api/tasklist"), "local development").Title("gtask-server")
	bc, err := define.NewBuildContext(doc, router)
	if err != nil {
		return fmt.Errorf("build context: %w", err)
	}

	// mount handler
	{
		{
			path := "/api/tasklist"
			api := &TaskListAPI{Oauth2Config: auth.OauthConfig}
			define.Get(bc, path, api.List,
				auth.WithOauthToken(":default-key:"),
			).OperationID("ListTaskList")
		}
		{
			path := "/api/tasklist/{tasklistId}"
			api := &TaskAPI{Oauth2Config: auth.OauthConfig}
			define.Get(bc, path, api.List,
				auth.WithOauthToken(":default-key:"),
			).OperationID("ListTasksOfTaskList")
		}
	}

	// mount optional handler (not included in openapi.json)
	{
		{
			path := "/api/tasklist"
			api := &TaskListAPI{Oauth2Config: auth.OauthConfig}
			router.Get(path, quickapi.Lift(api.List))
		}
		bc.Router().Mount("/openapi", rohandler.NewHandler(bc.Doc(), "/openapi"))
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
	srv := quickapi.NewServer(fmt.Sprintf(":%s", u.Port()), h, 5*time.Second)
	return srv.ListenAndServe(ctx)
}

type TaskListAPI struct {
	Oauth2Config *oauth2.Config
}

func (api *TaskListAPI) List(ctx context.Context, input quickapi.Empty) ([]*tasks.TaskList, error) {
	conf := api.Oauth2Config

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

type TaskAPI struct {
	Oauth2Config *oauth2.Config
}

type TaskAPIListInput struct {
	TaskListID string `openapi:"path" path:"tasklistId"`
}

func (api *TaskAPI) List(ctx context.Context, input TaskAPIListInput) ([]*tasks.Task, error) {
	conf := api.Oauth2Config

	tok, apiErr := auth.GetToken(ctx)
	if apiErr != nil {
		return nil, apiErr
	}

	client := conf.Client(ctx, tok)
	s, err := tasks.New(client)
	if err != nil {
		return nil, quickapi.NewAPIError(err, http.StatusUnauthorized)
	}
	res, err := s.Tasks.List(input.TaskListID).Do()
	if err != nil {
		return nil, quickapi.NewAPIError(err, http.StatusInternalServerError)
	}
	return res.Items, nil
}
