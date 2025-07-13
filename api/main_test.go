package api

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	db "github.com/Drolfothesgnir/simplebank/db/sqlc"
	"github.com/Drolfothesgnir/simplebank/token"
	"github.com/Drolfothesgnir/simplebank/util"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

func newTestServer(t *testing.T, store db.Store) *Server {
	config := util.Config{
		TokenSymmetricKey:   util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := NewServer(config, store)
	require.NoError(t, err)
	return server
}

func setAuthorizationHeader(t *testing.T, tokenMaker token.Maker, authorizationType string, username string, duration time.Duration, request *http.Request) {
	accessToken, payload, err := tokenMaker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, payload)
	authorizationToken := fmt.Sprintf("%s %s", authorizationType, accessToken)
	request.Header.Set(authorizationheaderKey, authorizationToken)
}
