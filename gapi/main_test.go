package gapi

import (
	"context"
	"fmt"
	"testing"
	"time"

	db "github.com/Drolfothesgnir/simplebank/db/sqlc"
	"github.com/Drolfothesgnir/simplebank/token"
	"github.com/Drolfothesgnir/simplebank/util"
	"github.com/Drolfothesgnir/simplebank/worker"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func newTestServer(t *testing.T, store db.Store, taskDistributor worker.TaskDistributor) *Server {
	config := util.Config{
		TokenSymmetricKey:   util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := NewServer(config, store, taskDistributor)
	require.NoError(t, err)
	return server
}

func setAuthorizationHeader(t *testing.T, tokenMaker token.Maker, authKey string, authorizationType string, username string, role string, duration time.Duration) context.Context {
	accessToken, payload, err := tokenMaker.CreateToken(username, role, duration)
	require.NoError(t, err)
	require.NotEmpty(t, payload)
	authorizationToken := fmt.Sprintf("%s %s", authorizationType, accessToken)
	md := metadata.Pairs(authKey, authorizationToken)
	return metadata.NewIncomingContext(context.Background(), md)
}
