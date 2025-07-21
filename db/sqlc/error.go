package db

import (
	"github.com/jackc/pgx/v5"
)

var ErrRecordNotFound = pgx.ErrNoRows

const (
	ForeignKeyViolation = "23503"
	UniqueViolation     = "23505"
)
