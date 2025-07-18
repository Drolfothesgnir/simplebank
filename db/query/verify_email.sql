-- name: CreateVerificationEmail :one
INSERT INTO verification_emails (
  username,
  email,
  secter_code
) VALUES (
  $1, $2, $3
) RETURNING *;