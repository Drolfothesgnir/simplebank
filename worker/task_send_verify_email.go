package worker

import (
	"context"
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
		if err == db.ErrRecordNotFound {
			return fmt.Errorf("user [%s] does not exist: %w", payload.Username, asynq.SkipRetry)
		}

		return fmt.Errorf("failed to retrieve user information: %w", err)
	}

	verificationEmail, err := processor.store.CreateVerificationEmail(ctx, db.CreateVerificationEmailParams{
		Username:   user.Username,
		Email:      user.Email,
		SecretCode: util.RandomString(32),
	})

	if err != nil {
		return fmt.Errorf("failed to create verification email: %w", err)
	}

	subject := "Welcome to Simple Bank!"
	verificationUrl := fmt.Sprintf(
		"http://localhost:8080/v1/verify_email?email_id=%d&secret_code=%s",
		verificationEmail.ID,
		verificationEmail.SecretCode,
	)

	err = processor.emailSender.SendEmail(
		subject,
		fmt.Sprintf(`
			Hello %s, <br/>
			Thank you for joining us!<br/>
			Please <a href="%s">click here</a> to verify your email address.<br/>
		`, user.FullName, verificationUrl),
		[]string{verificationEmail.Email},
		nil, nil, nil,
	)

	if err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	log.Info().Str("type", task.Type()).Str("user", payload.Username).Str("email", user.Email).Msg("Sending email")

	return nil
}
