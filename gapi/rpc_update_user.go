package gapi

import (
	"context"
	"errors"
	"time"

	db "github.com/Drolfothesgnir/simplebank/db/sqlc"
	"github.com/Drolfothesgnir/simplebank/pb"
	"github.com/Drolfothesgnir/simplebank/util"
	"github.com/Drolfothesgnir/simplebank/val"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {

	accessibleRoles := []string{util.DepositorRole, util.BankerRole}
	// TODO: proper role handling
	authPayload, err := server.authorizeUser(ctx, accessibleRoles)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	violations := validateUpdateUserRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	if authPayload.Role != util.BankerRole && authPayload.Username != req.GetUsername() {
		return nil, status.Errorf(codes.PermissionDenied, "cannot update other user's info")
	}

	arg := db.UpdateUserParams{
		Username: req.Username,
		FullName: pgtype.Text{String: req.GetFullName(), Valid: req.FullName != nil},
		Email:    pgtype.Text{String: req.GetEmail(), Valid: req.Email != nil},
	}

	if req.Password != nil {
		hashedPassword, err := util.HashPassword(*req.Password)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed hash password: %s", err)
		}

		arg.HashedPassword = pgtype.Text{String: hashedPassword, Valid: true}

		arg.PasswordChangedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
	}

	user, err := server.store.UpdateUser(ctx, arg)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == db.UniqueViolation && pgErr.ConstraintName == "users_email_key" {
				return nil, status.Errorf(codes.AlreadyExists, "email [%s] already exists", arg.Email.String)
			}
		}

		if err == db.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "user [%s] does not exist", req.Username)
		}

		return nil, status.Errorf(codes.Internal, "failed to update user: %s", err)
	}

	return &pb.UpdateUserResponse{User: convertUser(user)}, nil
}

func validateUpdateUserRequest(req *pb.UpdateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}

	if req.Password != nil {
		if err := val.ValidatePassword(*req.Password); err != nil {
			violations = append(violations, fieldViolation("password", err))
		}
	}

	if req.FullName != nil {
		if err := val.ValidateFullName(*req.FullName); err != nil {
			violations = append(violations, fieldViolation("full_name", err))
		}
	}

	if req.Email != nil {
		if err := val.ValidateEmail(*req.Email); err != nil {
			violations = append(violations, fieldViolation("email", err))
		}
	}

	return
}
