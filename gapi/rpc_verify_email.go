package gapi

import (
	"context"
	"database/sql"
	"time"

	db "github.com/Drolfothesgnir/simplebank/db/sqlc"
	"github.com/Drolfothesgnir/simplebank/pb"
	"github.com/Drolfothesgnir/simplebank/val"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {

	violations := validateVerifyEmailRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	email, err := server.store.GetVerificationEmail(ctx, req.GetEmailId())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "verification email not found")
		}

		return nil, status.Errorf(codes.Internal, "failed to get verification email")
	}

	if email.IsUsed {
		return nil, status.Errorf(codes.Unauthenticated, "secrete code is already used")
	}

	if time.Now().After(email.ExpiredAt) {
		return nil, status.Errorf(codes.Unauthenticated, "secrete code is expired")
	}

	if email.SecretCode != req.GetSecretCode() {
		return nil, status.Errorf(codes.Unauthenticated, "secrete codes mismatch")
	}

	arg := db.VerifyEmailTxParams{
		EmailID: email.ID,
	}

	txResult, err := server.store.VerifyEmailTx(ctx, arg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user's data")
	}

	res := &pb.VerifyEmailResponse{
		IsVerified: txResult.User.IsEmailVerified,
	}

	return res, nil
}

func validateVerifyEmailRequest(req *pb.VerifyEmailRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateEmailID(req.EmailId); err != nil {
		violations = append(violations, fieldViolation("email_id", err))
	}

	if err := val.ValidateSecretCode(req.SecretCode); err != nil {
		violations = append(violations, fieldViolation("secret_code", err))
	}

	return
}
