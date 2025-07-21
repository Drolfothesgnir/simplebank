package main

import (
	"context"
	"os"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	db "github.com/Drolfothesgnir/simplebank/db/sqlc"
	"github.com/Drolfothesgnir/simplebank/mail"
	"github.com/Drolfothesgnir/simplebank/servers"
	"github.com/Drolfothesgnir/simplebank/util"
	"github.com/Drolfothesgnir/simplebank/worker"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"

	_ "github.com/Drolfothesgnir/simplebank/doc/statik"
)

func main() {
	worker.SetupRedisLogger()

	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot read config file")
	}

	if config.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	conn, err := pgxpool.New(context.Background(), config.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to the database")
	}

	store := db.NewStore(conn)

	runDBMigration(config.MigrationURL, config.DBSource)

	redisOpts := asynq.RedisClientOpt{Addr: config.RedisAddress}

	taskDistributor := worker.NewRedisTaskDistributor(redisOpts)

	emailSender := mail.NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)

	go runTaskProcessor(redisOpts, store, emailSender)

	go servers.RunGatewayServer(config, store, taskDistributor)

	servers.RunGrpcServer(config, store, taskDistributor)
}

func runDBMigration(migrationURL string, dbSource string) {
	mig, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create new migrate instance")
	}

	if err = mig.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal().Err(err).Msg("failed to run migrate up")
	}

	log.Info().Msg("db migrated successfully")
}

func runTaskProcessor(redisOpts asynq.RedisClientOpt, store db.Store, emailSender mail.EmailSender) {
	processor := worker.NewRedisTaskProcessor(redisOpts, store, emailSender)

	err := processor.Start()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start task processor")
	}
}
