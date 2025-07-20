package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

type VerifyEmailTxParams struct {
	EmailID    int64  `json:"email_id"`
	SecretCode string `json:"secret_code"`
}

type VerifyEmailTxResult struct {
	User              User
	VerificationEmail VerificationEmail
}

func (store *SQLStore) VerifyEmailTx(ctx context.Context, arg VerifyEmailTxParams) (VerifyEmailTxResult, error) {
	var result VerifyEmailTxResult
	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		emailParams := UpdateVerificationEmailParams{
			ID:     arg.EmailID,
			IsUsed: pgtype.Bool{Bool: true, Valid: true},
		}

		result.VerificationEmail, err = q.UpdateVerificationEmail(ctx, emailParams)

		if err != nil {
			return err
		}

		userParams := UpdateUserParams{
			Username:        result.VerificationEmail.Username,
			IsEmailVerified: pgtype.Bool{Bool: true, Valid: true},
		}

		result.User, err = q.UpdateUser(ctx, userParams)

		if err != nil {
			return err
		}

		return nil
	})

	return result, err
}
