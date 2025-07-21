package gapi

import (
	"context"
	"database/sql"
	"testing"
	"time"

	mockdb "github.com/Drolfothesgnir/simplebank/db/mock"
	db "github.com/Drolfothesgnir/simplebank/db/sqlc"
	"github.com/Drolfothesgnir/simplebank/pb"
	"github.com/Drolfothesgnir/simplebank/token"
	"github.com/Drolfothesgnir/simplebank/util"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUpdateUser(t *testing.T) {

	user, _ := createRandomUser(t, util.DepositorRole)
	newUser, _ := createRandomUser(t, util.DepositorRole)
	invalidEmail := "invalid-email"
	invalidFullName := "123"

	testCases := []struct {
		name          string
		body          *pb.UpdateUserRequest
		buildStubs    func(store *mockdb.MockStore)
		setupAuth     func(t *testing.T, tokenMaker token.Maker) context.Context
		checkResponse func(t *testing.T, res *pb.UpdateUserResponse, err error)
	}{
		{
			name: "OK",
			body: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newUser.FullName,
				Email:    &newUser.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.UpdateUserParams{
					Username: user.Username,
					FullName: pgtype.Text{String: newUser.FullName, Valid: true},
					Email:    pgtype.Text{String: newUser.Email, Valid: true},
				}

				store.EXPECT().UpdateUser(gomock.Any(), gomock.Eq(arg)).Times(1).Return(newUser, nil)

			},
			setupAuth: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return setAuthorizationHeader(t, tokenMaker, authorizationHeader, authorizationTypeBearer, user.Username, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				createdUser := res.GetUser()
				require.Equal(t, newUser.Username, createdUser.Username)
				require.Equal(t, newUser.Email, createdUser.Email)
				require.Equal(t, newUser.FullName, createdUser.FullName)
			},
		},
		{
			name: "ExpiredToken",
			body: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newUser.FullName,
				Email:    &newUser.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Times(0)

			},
			setupAuth: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return setAuthorizationHeader(t, tokenMaker, authorizationHeader, authorizationTypeBearer, user.Username, user.Role, -time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.PermissionDenied, st.Code())
			},
		},
		{
			name: "WrongUsername",
			body: &pb.UpdateUserRequest{
				Username: "wronguser",
				FullName: &newUser.FullName,
				Email:    &newUser.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Times(0)

			},
			setupAuth: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return setAuthorizationHeader(t, tokenMaker, authorizationHeader, authorizationTypeBearer, user.Username, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.PermissionDenied, st.Code())
			},
		},
		{
			name: "InvalidUsername",
			body: &pb.UpdateUserRequest{
				Username: "/.123",
				FullName: &newUser.FullName,
				Email:    &newUser.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Times(0)

			},
			setupAuth: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return setAuthorizationHeader(t, tokenMaker, authorizationHeader, authorizationTypeBearer, "/.123", user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "InvalidEmail",
			body: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newUser.FullName,
				Email:    &invalidEmail,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Times(0)

			},
			setupAuth: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return setAuthorizationHeader(t, tokenMaker, authorizationHeader, authorizationTypeBearer, user.Username, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "InvalidFullName",
			body: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &invalidFullName,
				Email:    &newUser.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Times(0)

			},
			setupAuth: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return setAuthorizationHeader(t, tokenMaker, authorizationHeader, authorizationTypeBearer, user.Username, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		{
			name: "DuplicateEmail",
			body: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newUser.FullName,
				Email:    &newUser.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				err := &pgconn.PgError{
					Code:           db.UniqueViolation,
					ConstraintName: "users_email_key",
				}

				store.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Times(1).Return(db.User{}, err)

			},
			setupAuth: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return setAuthorizationHeader(t, tokenMaker, authorizationHeader, authorizationTypeBearer, user.Username, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.AlreadyExists, st.Code())
			},
		},
		{
			name: "UserNotFound",
			body: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newUser.FullName,
				Email:    &newUser.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Times(1).Return(db.User{}, db.ErrRecordNotFound)

			},
			setupAuth: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return setAuthorizationHeader(t, tokenMaker, authorizationHeader, authorizationTypeBearer, user.Username, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.NotFound, st.Code())
			},
		},
		{
			name: "InternalError",
			body: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newUser.FullName,
				Email:    &newUser.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Times(1).Return(db.User{}, sql.ErrConnDone)

			},
			setupAuth: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return setAuthorizationHeader(t, tokenMaker, authorizationHeader, authorizationTypeBearer, user.Username, user.Role, time.Minute)
			},
			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, st.Code())
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storeCtrl := gomock.NewController(t)
			defer storeCtrl.Finish()

			store := mockdb.NewMockStore(storeCtrl)

			tc.buildStubs(store)

			server := newTestServer(t, store, nil)

			ctx := tc.setupAuth(t, server.tokenMaker)

			res, err := server.UpdateUser(ctx, tc.body)
			tc.checkResponse(t, res, err)
		})
	}
}
