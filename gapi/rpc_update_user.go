package gapi

import (
	"context"
	"database/sql"
	"time"

	db "github.com/Drolfothesgnir/simplebank/db/sqlc"
	"github.com/Drolfothesgnir/simplebank/pb"
	"github.com/Drolfothesgnir/simplebank/util"
	"github.com/Drolfothesgnir/simplebank/val"
	"github.com/lib/pq"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {

	violations := validateUpdateUserRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	arg := db.UpdateUserParams{
		Username: req.Username,
		FullName: sql.NullString{String: req.GetFullName(), Valid: req.FullName != nil},
		Email:    sql.NullString{String: req.GetEmail(), Valid: req.Email != nil},
	}

	if req.Password != nil {
		hashedPassword, err := util.HashPassword(*req.Password)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed hash password: %s", err)
		}

		arg.HashedPassword = sql.NullString{String: hashedPassword, Valid: true}

		arg.PasswordChangedAt = sql.NullTime{Time: time.Now(), Valid: true}
	}

	user, err := server.store.UpdateUser(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code.Name() == "unique_violation" && pqErr.Constraint == "users_email_key" {
				return nil, status.Errorf(codes.AlreadyExists, "email [%s] already exists", arg.Email.String)
			}
		}

		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user [%s] does not exist", req.Username)
		}

		return nil, status.Errorf(codes.Internal, "failed to update user: %s", err)
	}

	return &pb.UpdateUserResponse{User: convertUser(user)}, nil
}

func validateUpdateUserRequest(req *pb.UpdateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}

	if req.Password != nil {
		if err := val.ValidPassword(*req.Password); err != nil {
			violations = append(violations, fieldViolation("password", err))
		}
	}

	if req.FullName != nil {
		if err := val.ValidFullName(*req.FullName); err != nil {
			violations = append(violations, fieldViolation("full_name", err))
		}
	}

	if req.Email != nil {
		if err := val.ValidEmail(*req.Email); err != nil {
			violations = append(violations, fieldViolation("email", err))
		}
	}

	return
}
