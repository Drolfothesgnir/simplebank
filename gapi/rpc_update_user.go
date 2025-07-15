package gapi

import (
	"context"
	"database/sql"

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

	violations, parsedParams := parseUserUpdateParams(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	_, err := server.store.GetUser(ctx, parsedParams.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user [%s] does not exist", parsedParams.Username)
		}

		return nil, status.Error(codes.Internal, "failed to get user")
	}

	arg := db.UpdateUserParams{
		Username: parsedParams.Username,
		FullName: *parsedParams.FullName,
		Email:    *parsedParams.Email,
	}

	if parsedParams.Password.Valid {
		var err error
		arg.HashedPassword.String, err = util.HashPassword(parsedParams.Password.String)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed hash password: %s", err)
		}
	}

	user, err := server.store.UpdateUser(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code.Name() == "unique_violation" && pqErr.Constraint == "users_email_key" {
				return nil, status.Errorf(codes.AlreadyExists, "email [%s] already exists", arg.Email.String)
			}
		}

		return nil, status.Errorf(codes.Internal, "failed to update user: %s", err)
	}

	return &pb.UpdateUserResponse{User: convertUser(user)}, nil
}

type parsedUserParams struct {
	Username string
	Password *sql.NullString
	FullName *sql.NullString
	Email    *sql.NullString
}

func parseUserUpdateParams(req *pb.UpdateUserRequest) (violations []*errdetails.BadRequest_FieldViolation, parsedParams parsedUserParams) {

	if err := val.ValidUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	} else {
		parsedParams.Username = req.Username
	}

	if password := req.GetPassword(); len(password) > 0 {
		if err := val.ValidPassword(password); err != nil {
			violations = append(violations, fieldViolation("password", err))
		} else {
			parsedParams.Password = &sql.NullString{String: password, Valid: true}
		}
	} else {
		parsedParams.Password = &sql.NullString{Valid: false}
	}

	if fullName := req.GetFullName(); len(fullName) > 0 {
		if err := val.ValidFullName(fullName); err != nil {
			violations = append(violations, fieldViolation("full_name", err))
		} else {
			parsedParams.FullName = &sql.NullString{String: fullName, Valid: true}
		}
	} else {
		parsedParams.FullName = &sql.NullString{Valid: false}
	}

	if email := req.GetEmail(); len(email) > 0 {
		if err := val.ValidEmail(email); err != nil {
			violations = append(violations, fieldViolation("email", err))
		} else {
			parsedParams.Email = &sql.NullString{String: email, Valid: true}
		}
	} else {
		parsedParams.Email = &sql.NullString{Valid: false}
	}

	return
}
