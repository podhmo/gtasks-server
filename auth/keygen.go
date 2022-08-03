package auth

import (
	"net/http"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type ConstantKeyGenerator struct {
	Key string
}

func (g *ConstantKeyGenerator) GenerateKey(req *http.Request, token *oauth2.Token, salt string) (string, error) {
	return g.Key, nil
}

var _ KeyGenerator = (*ConstantKeyGenerator)(nil)

type UUIDKeyGenerator struct {
}

func (g *UUIDKeyGenerator) GenerateKey(req *http.Request, token *oauth2.Token, salt string) (string, error) {
	u, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return salt + u.String(), nil
}

var _ KeyGenerator = (*ConstantKeyGenerator)(nil)
