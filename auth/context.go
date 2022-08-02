package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/podhmo/quickapi"
	"github.com/podhmo/quickapi/shared"
	"golang.org/x/oauth2"
)

type ctxkey string

const (
	tokenKey ctxkey = "token key"
)

func GetToken(ctx context.Context) (*oauth2.Token, interface {
	error
	shared.StatusCoder
}) {
	v := ctx.Value(tokenKey)
	if v == nil {
		return nil, quickapi.NewAPIError(fmt.Errorf("oauth token is not found"), http.StatusUnauthorized)
	}
	return v.(*oauth2.Token), nil
}
