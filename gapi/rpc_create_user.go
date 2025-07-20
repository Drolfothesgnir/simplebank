package gapi

import (
	"context"
	"errors"
	"time"

	db "github.com/Drolfothesgnir/simplebank/db/sqlc"
	"github.com/Drolfothesgnir/simplebank/pb"
	"github.com/Drolfothesgnir/simplebank/util"
	"github.com/Drolfothesgnir/simplebank/val"
	"github.com/Drolfothesgnir/simplebank/worker"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {

	violations := validateCreateUserRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	hashedPassword, err := util.HashPassword(req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed hash password: %s", err)
	}

	createUserParams := db.CreateUserParams{
		Username:       req.GetUsername(),
		HashedPassword: hashedPassword,
		FullName:       req.GetFullName(),
		Email:          req.GetEmail(),
	}

	cb := func(user db.User) error {
		payload := &worker.PayloadSendVerifyEmail{Username: user.Username}

		opts := []asynq.Option{
			asynq.MaxRetry(10),
			asynq.ProcessIn(10 * time.Second),
			asynq.Queue(worker.QueueCritical),
		}

		return server.taskDistributor.DistributeTaskSendVerifyEmail(ctx, payload, opts...)
	}

	arg := db.CreateUserTxParams{
		CreateUserParams: createUserParams,
		AfterCreate:      cb,
	}

	txResult, err := server.store.CreateUserTx(ctx, arg)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == db.UniqueViolation {
				switch pgErr.ConstraintName {
				case "users_pkey":
					return nil, status.Errorf(codes.AlreadyExists, "user [%s] already exists", arg.Username)
				case "users_email_key":
					return nil, status.Errorf(codes.AlreadyExists, "email [%s] already exists", arg.Email)
				}

			}
		}

		return nil, status.Errorf(codes.Internal, "failed to create user: %s", err)
	}

	return &pb.CreateUserResponse{User: convertUser(txResult.User)}, nil
}

func validateCreateUserRequest(req *pb.CreateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}

	if err := val.ValidatePassword(req.GetPassword()); err != nil {
		violations = append(violations, fieldViolation("password", err))
	}

	if err := val.ValidateFullName(req.GetFullName()); err != nil {
		violations = append(violations, fieldViolation("full_name", err))
	}

	if err := val.ValidateEmail(req.GetEmail()); err != nil {
		violations = append(violations, fieldViolation("email", err))
	}

	return
}
