package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	db "github.com/Drolfothesgnir/simplebank/db/sqlc"
	"github.com/Drolfothesgnir/simplebank/util"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const (
	TypeVerifyEmail = "email:verify"
)

type PayloadSendVerifyEmail struct {
	Username string `json:"username"`
}

func (distributor *RedisTaskDistributor) DistributeTaskSendVerifyEmail(
	ctx context.Context,
	payload *PayloadSendVerifyEmail,
	opts ...asynq.Option,
) error {

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to serialize email verification payload: %w", err)
	}

	task := asynq.NewTask(TypeVerifyEmail, jsonPayload, opts...)

	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue email verification task: %w", err)
	}

	log.Info().
		Str("type", info.Type).
		Str("id", info.ID).
		Str("queue", info.Queue).
		Bytes("payload", info.Payload).
		Int("max retry", info.MaxRetry).
		Msg("enqueued task")

	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendVerifyEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to deserialize task payload: %v: %w", err, asynq.SkipRetry)
	}

	user, err := processor.store.GetUser(ctx, payload.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user [%s] does not exist: %w", payload.Username, asynq.SkipRetry)
		}

		return fmt.Errorf("failed to retrieve user information: %w", err)
	}

	verificationEmail, err := processor.store.CreateVerificationEmail(ctx, db.CreateVerificationEmailParams{
		Username:   user.Username,
		Email:      user.Email,
		SecterCode: util.RandomString(32),
	})

	if err != nil {
		return fmt.Errorf("failed to create verification email: %w", err)
	}

	err = processor.emailSender.SendEmail(
		"Email verification",
		fmt.Sprintf(`
			<html>
				<body>
					<h1>You need to verify your email address to complete registration</h1>
					<h3>Your code is <strong>%s</strong></h3>
					<p>follow this link to enter the code and finish regisration:</p>
					<p><a href="https://google.com">Enter code here</a></p>
				</body>
			</html>
		`, verificationEmail.SecterCode),
		[]string{verificationEmail.Email},
		nil, nil, nil,
	)

	if err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	log.Info().Str("type", task.Type()).Str("user", payload.Username).Str("email", user.Email).Msg("Sending email")

	return nil
}
