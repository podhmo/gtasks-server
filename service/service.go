package service

import (
	"context"
	"net/http"

	"github.com/podhmo/gtasks-server/auth"
	"github.com/podhmo/quickapi"
	"github.com/podhmo/quickapi/shared"
	"golang.org/x/oauth2"
)

type APIError interface {
	shared.StatusCoder
	error
}

func New[T any](ctx context.Context, conf *oauth2.Config, f func(*http.Client) (*T, error)) (*T, APIError) {
	tok, apiErr := auth.GetToken(ctx)
	if apiErr != nil {
		return nil, apiErr
	}

	client := conf.Client(ctx, tok)
	s, err := f(client)
	if err != nil {
		return nil, quickapi.NewAPIError(err, http.StatusUnauthorized)
	}
	return s, nil
}

func OrAPIError[T any](ob T, err error) (T, APIError) {
	if err != nil {
		return ob, quickapi.NewAPIError(err, http.StatusInternalServerError)
	}
	return ob, nil
}
