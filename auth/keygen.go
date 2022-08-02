package auth

import (
	"net/http"

	"golang.org/x/oauth2"
)

type ConstantKeyGenerator struct {
	Key string
}

func (g *ConstantKeyGenerator) GenerateKey(req *http.Request, token *oauth2.Token) string {
	return g.Key
}

var _ KeyGenerator = (*ConstantKeyGenerator)(nil)
