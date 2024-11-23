package integtest

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
	"strings"
	"sync"
	"testing"
)

type FakeAuth struct {
	changes      sync.Mutex
	tokensToUser map[string]*fakeUser
}

func NewFakeAuth() *FakeAuth {
	return &FakeAuth{
		changes:      sync.Mutex{},
		tokensToUser: make(map[string]*fakeUser),
	}
}

type fakerUserProfile struct {
	Token string `faker:"jwt"`
	ID    string `faker:"uuid_hyphenated"`
}

var contextKey = "fake-auth"

func (f *FakeAuth) NewUser(parent context.Context, t *testing.T, name string) (token string, ctx context.Context) {
	f.changes.Lock()
	defer f.changes.Unlock()

	p := &fakerUserProfile{}
	require.NoError(t, faker.FakeData(p))
	u := &fakeUser{
		ID: p.ID,
	}
	f.tokensToUser[p.Token] = u

	userContext := context.WithValue(parent, contextKey, u)
	return p.Token, userContext
}

func (f *FakeAuth) GetUserID(ctx context.Context) (string, error) {
	userContext, ok := ctx.Value(contextKey).(*fakeUser)
	if !ok {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return "", errors.New("unauthenticated")
		}
		authorization := md["authorization"]
		if len(authorization) != 1 {
			return "", errors.New("authentication error")
		}
		full := authorization[0]
		parts := strings.SplitN(full, " ", 2)
		token := parts[1]

		fmt.Printf("Checking token %q against %#v\n", token, f.tokensToUser)
		f.changes.Lock()
		defer f.changes.Unlock()
		if u, ok := f.tokensToUser[token]; !ok {
			return "", errors.New("unauthorized")
		} else {
			userContext = u
		}
	}
	return userContext.ID, nil
}

type fakeUser struct {
	ID string
}
