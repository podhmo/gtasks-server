//go:generate go run ./ --gendoc --docfile openapi.json
package main

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/gomarkdown/markdown"
	"github.com/podhmo/flagstruct"
	"github.com/podhmo/gtasks-server/auth"
	"github.com/podhmo/quickapi"
	"github.com/podhmo/quickapi/qopenapi/define"
	"github.com/podhmo/reflect-openapi/dochandler"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/tasks/v1"
)

//go:embed openapi.json
var openapiDocData []byte

type Options struct {
	ClientID     string `flag:"client-id" required:"true"`
	ClientSecret string `flag:"client-secret" required:"true"`
	RedirectURL  string `flag:"redirect-url" required:"true"`

	GenDoc  bool   `flag:"gendoc" help:"generate openapi.json to stdout"`
	Docfile string `flag:"docfile" help:"write name of openapi.json"`
}

func main() {
	config := Options{RedirectURL: "http://localhost:8888"}
	flagstruct.Parse(&config)
	if err := run(config); err != nil {
		log.Printf("!! %+v", err)
	}
}

var SCOPES = []string{
	"https://www.googleapis.com/auth/tasks",
}

func run(options Options) error {
	// Your credentials should be obtained from the Google
	// Developer Console (https://console.developers.google.com).
	conf := &oauth2.Config{
		ClientID:     options.ClientID,
		ClientSecret: options.ClientSecret,
		RedirectURL:  options.RedirectURL + "/auth/callback",
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

	doc := define.Doc().
		Server(strings.TrimSuffix(auth.DefaultURL, "/api/tasklist"), "local development").
		Title("gtask-server")

	if !options.GenDoc {
		doc = doc.LoadFromData(openapiDocData)
	}

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
			)
		}
		{
			api := &TaskAPI{Oauth2Config: auth.OauthConfig}
			{
				path := "/api/tasklist/{tasklistId}"

				define.Get(bc, path, api.List,
					auth.WithOauthToken(":default-key:"),
				)
			}
		}
		{
			api := &MarkdownAPI{Oauth2Config: auth.OauthConfig}
			{
				path := "/"
				define.GetHTML(bc, path, api.ListTaskList, dumpMarkdown,
					auth.WithOauthToken(":default-key:"),
				)
			}
			{
				path := "/{tasklistId}"
				define.GetHTML(bc, path, api.DetailTaskList, dumpMarkdown,
					auth.WithOauthToken(":default-key:"),
				)
			}

		}
	}

	// mount optional handler (not included in openapi.json)
	{
		bc.Router().Mount("/openapi", dochandler.New(bc.Doc(), "/openapi"))
	}

	if options.GenDoc {
		var w io.Writer = os.Stdout
		if options.Docfile != "" {
			f, err := os.Create(options.Docfile)
			if err != nil {
				return fmt.Errorf("write file: %w", err)
			}
			defer f.Close()
			w = f
		}
		if err := bc.EmitDoc(ctx, w); err != nil {
			return err
		}
		return nil
	}

	h, err := bc.BuildHandler(ctx)
	if err != nil {
		return fmt.Errorf("build handler: %w", err)
	}

	u, err := url.Parse(options.RedirectURL)
	if err != nil {
		return fmt.Errorf("invalid url: %q -- %w", options.RedirectURL, err)
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
	res, err := s.Tasklists.List().MaxResults(100).Do()
	if err != nil {
		return nil, quickapi.NewAPIError(err, http.StatusInternalServerError)
	}
	return res.Items, nil
}

type TaskAPI struct {
	Oauth2Config *oauth2.Config
}

type TaskAPIListInput struct {
	TaskListID string `in:"path" path:"tasklistId"`
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
	var items []*tasks.Task
	{
		pageSize := 100 // 20~100
		res, err := s.Tasks.List(input.TaskListID).MaxResults(int64(pageSize)).Do()
		if err != nil {
			return nil, quickapi.NewAPIError(err, http.StatusInternalServerError) // xxx
		}
		items = res.Items
		for res.NextPageToken != "" {
			pageToken := res.NextPageToken
			res, err = s.Tasks.List(input.TaskListID).PageToken(pageToken).MaxResults(int64(pageSize)).Do()
			if err != nil {
				log.Printf("unexpected error: %+v", err)
				break
			}
			items = append(items, res.Items...)
		}
	}
	return items, nil
}

type MarkdownAPI struct {
	Oauth2Config *oauth2.Config
}

func (api *MarkdownAPI) ListTaskList(ctx context.Context, input quickapi.Empty) (string, error) {
	conf := api.Oauth2Config

	tok, apiErr := auth.GetToken(ctx)
	if apiErr != nil {
		return "", apiErr
	}

	client := conf.Client(ctx, tok)
	s, err := tasks.New(client)
	if err != nil {
		return "", quickapi.NewAPIError(err, http.StatusUnauthorized)
	}
	res, err := s.Tasklists.List().MaxResults(100).Do()
	if err != nil {
		return "", quickapi.NewAPIError(err, http.StatusInternalServerError)
	}

	buf := new(strings.Builder)
	qs := quickapi.GetRequest(ctx).URL.RawQuery

	fmt.Fprintln(buf, "## api")
	fmt.Fprintln(buf, "- [doc](/openapi/doc) [redoc](/openapi/redoc)")
	fmt.Fprintf(buf, "- [list](/api/tasklist?%s) proxy of https://developers.google.com/tasks\n", qs)

	fmt.Fprintln(buf, "## list")
	fmt.Fprintln(buf, "")
	tasks := res.Items
	sort.SliceStable(tasks, func(i, j int) bool { return tasks[i].Updated > tasks[j].Updated })
	for _, tl := range tasks {
		fmt.Fprintf(buf, "- [%s](/%s?%s) (updated: %s)\n", tl.Title, tl.Id, qs, tl.Updated)
	}
	return buf.String(), nil
}

type MarkdownAPIDetailTaskListInput struct {
	TaskListID string `in:"path" path:"tasklistId"`
}

func (api *MarkdownAPI) DetailTaskList(ctx context.Context, input MarkdownAPIDetailTaskListInput) (string, error) {
	conf := api.Oauth2Config

	tok, apiErr := auth.GetToken(ctx)
	if apiErr != nil {
		return "", apiErr
	}

	client := conf.Client(ctx, tok)
	s, err := tasks.New(client)
	if err != nil {
		return "", quickapi.NewAPIError(err, http.StatusUnauthorized)
	}

	var tasklist *tasks.TaskList
	{

		res, err := s.Tasklists.Get(input.TaskListID).Do()
		if err != nil {
			return "", quickapi.NewAPIError(err, http.StatusNotFound)
		}
		tasklist = res
	}

	var items []*tasks.Task
	{
		pageSize := 100 // 20~100
		res, err := s.Tasks.List(input.TaskListID).MaxResults(int64(pageSize)).Do()
		if err != nil {
			return "", quickapi.NewAPIError(err, http.StatusInternalServerError) // xxx
		}
		items = res.Items
		for res.NextPageToken != "" {
			pageToken := res.NextPageToken
			res, err = s.Tasks.List(input.TaskListID).PageToken(pageToken).MaxResults(int64(pageSize)).Do()
			if err != nil {
				log.Printf("unexpected error: %+v", err)
				break
			}
			items = append(items, res.Items...)
		}
	}

	buf := new(strings.Builder)
	fmt.Fprintf(buf, "# %s", tasklist.Title)
	fmt.Fprintln(buf, "")

	for _, task := range items {
		fmt.Fprintf(buf, "- %s\n", task.Title)
		if task.Notes != "" {
			fmt.Fprintln(buf, "")
			fmt.Fprintf(buf, "  - %s\n", task.Notes)
			fmt.Fprintln(buf, "")
		}
	}

	fmt.Fprintln(buf, "")
	fmt.Fprintln(buf, "----------------------------------------")
	fmt.Fprintln(buf, "")
	fmt.Fprintln(buf, "```")
	for _, task := range items {
		fmt.Fprintf(buf, "- %s\n", task.Title)
		if task.Notes != "" {
			fmt.Fprintln(buf, "")
			fmt.Fprintf(buf, "  - %s\n", task.Notes)
			fmt.Fprintln(buf, "")
		}
	}
	fmt.Fprintln(buf, "```")
	return buf.String(), nil
}

func dumpMarkdown(ctx context.Context, w http.ResponseWriter, req *http.Request, text string, err error) {
	template := `<!DOCTYPE html>
<html lang="en">
<meta charset="UTF-8">
<title>-</title>
<meta name="viewport" content="width=device-width, initial-scale=1">
<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/github-markdown-css/5.1.0/github-markdown.min.css">

<style>
	.markdown-body {
		box-sizing: border-box;
		min-width: 200px;
		max-width: 980px;
		margin: 0 auto;
		padding: 45px;
	}

	@media (max-width: 767px) {
		.markdown-body {
			padding: 15px;
		}
	}
</style>
<body>
<article class="markdown-body">
%s
</article>
</body>
<html>
`
	block := markdown.ToHTML([]byte(text), nil, nil)
	render.HTML(w, req, fmt.Sprintf(template, block))
}
