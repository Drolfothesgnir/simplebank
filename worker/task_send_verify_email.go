package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

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

	// TODO: email sending
	log.Info().Str("type", task.Type()).Str("user", payload.Username).Str("email", user.Email).Msg("Sending email")

	return nil
}
