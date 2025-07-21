package gapi

import (
	"context"
	"testing"
	"time"

	mockdb "github.com/Drolfothesgnir/simplebank/db/mock"
	"github.com/Drolfothesgnir/simplebank/token"
	"github.com/Drolfothesgnir/simplebank/util"
	mockwk "github.com/Drolfothesgnir/simplebank/worker/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAuthMiddleware(t *testing.T) {
	username := util.RandomOwner()

	testCases := []struct {
		name          string
		roles         []string
		setupAuth     func(t *testing.T, tokenMaker token.Maker) context.Context
		checkResponse func(t *testing.T, payload *token.Payload, err error)
	}{
		{
			name:  "OK",
			roles: []string{util.DepositorRole, util.BankerRole},
			setupAuth: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return setAuthorizationHeader(t, tokenMaker, authorizationHeader, authorizationTypeBearer, username, util.DepositorRole, time.Minute)
			},
			checkResponse: func(t *testing.T, payload *token.Payload, err error) {
				require.NoError(t, err)
				require.NotEmpty(t, payload)
				require.Equal(t, username, payload.Username)
			},
		},
		{
			name:  "MissingMetadata",
			roles: []string{util.DepositorRole, util.BankerRole},
			setupAuth: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return context.Background()
			},
			checkResponse: func(t *testing.T, payload *token.Payload, err error) {
				require.Empty(t, payload)
				require.Error(t, err)
			},
		},
		{
			name:  "MissingAuthorizationHeader",
			roles: []string{util.DepositorRole, util.BankerRole},
			setupAuth: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return setAuthorizationHeader(t, tokenMaker, "invalidHeader", authorizationTypeBearer, username, util.DepositorRole, time.Minute)
			},
			checkResponse: func(t *testing.T, payload *token.Payload, err error) {
				require.Empty(t, payload)
				require.Error(t, err)
			},
		},
		{
			name:  "InvalidAuthHeaderFormat",
			roles: []string{util.DepositorRole, util.BankerRole},
			setupAuth: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return setAuthorizationHeader(t, tokenMaker, authorizationHeader, "", username, util.DepositorRole, time.Minute)
			},
			checkResponse: func(t *testing.T, payload *token.Payload, err error) {
				require.Empty(t, payload)
				require.Error(t, err)
			},
		},
		{
			name:  "UnsupportedAuthorizationType",
			roles: []string{util.DepositorRole, util.BankerRole},
			setupAuth: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return setAuthorizationHeader(t, tokenMaker, authorizationHeader, "invalidType", username, util.DepositorRole, time.Minute)
			},
			checkResponse: func(t *testing.T, payload *token.Payload, err error) {
				require.Empty(t, payload)
				require.Error(t, err)
			},
		},
		{
			name:  "ExpiredToken",
			roles: []string{util.DepositorRole, util.BankerRole},
			setupAuth: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return setAuthorizationHeader(t, tokenMaker, authorizationHeader, authorizationTypeBearer, username, util.DepositorRole, -time.Minute)
			},
			checkResponse: func(t *testing.T, payload *token.Payload, err error) {
				require.Empty(t, payload)
				require.Error(t, err)
			},
		},
		{
			name:  "PermissionDenied",
			roles: []string{util.BankerRole},
			setupAuth: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return setAuthorizationHeader(t, tokenMaker, authorizationHeader, authorizationTypeBearer, username, util.DepositorRole, time.Minute)
			},
			checkResponse: func(t *testing.T, payload *token.Payload, err error) {
				require.Empty(t, payload)
				require.Error(t, err)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storeCtrl := gomock.NewController(t)
			defer storeCtrl.Finish()

			store := mockdb.NewMockStore(storeCtrl)

			taskCtrl := gomock.NewController(t)
			defer taskCtrl.Finish()
			taskDistributor := mockwk.NewMockTaskDistributor(taskCtrl)
			server := newTestServer(t, store, taskDistributor)
			ctx := tc.setupAuth(t, server.tokenMaker)
			payload, err := server.authorizeUser(ctx, tc.roles)
			tc.checkResponse(t, payload, err)
		})
	}
}
