-- name: CreateVerificationEmail :one
INSERT INTO verification_emails (
  username,
  email,
  secret_code
) VALUES (
  $1, $2, $3
) RETURNING *;

-- name: GetVerificationEmail :one
SELECT * FROM verification_emails
WHERE id = $1 LIMIT 1;

-- name: UpdateVerificationEmail :one
UPDATE verification_emails
SET 
  email = COALESCE(sqlc.narg('email'), email),
  is_used = COALESCE(sqlc.narg('is_used'), is_used)
WHERE id = $1
RETURNING *;
