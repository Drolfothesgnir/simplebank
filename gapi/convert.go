package gapi

import (
	db "github.com/Drolfothesgnir/simplebank/db/sqlc"
	"github.com/Drolfothesgnir/simplebank/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func convertUser(dbUser db.User) *pb.User {
	return &pb.User{
		Username:          dbUser.Username,
		FullName:          dbUser.FullName,
		Email:             dbUser.Email,
		PasswordChangedAt: timestamppb.New(dbUser.PasswordChangedAt),
		CreatedAt:         timestamppb.New(dbUser.CreatedAt),
		IsEmailVerified:   dbUser.IsEmailVerified,
	}
}
