package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

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

var interruptSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
	syscall.SIGINT,
}

func main() {
	worker.SetupRedisLogger()

	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot read config file")
	}

	if config.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	ctx, stop := signal.NotifyContext(context.Background(), interruptSignals...)
	defer stop()

	conn, err := pgxpool.New(ctx, config.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to the database")
	}

	store := db.NewStore(conn)

	runDBMigration(config.MigrationURL, config.DBSource)

	redisOpts := asynq.RedisClientOpt{Addr: config.RedisAddress}

	taskDistributor := worker.NewRedisTaskDistributor(redisOpts)

	emailSender := mail.NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)

	waitGroup, ctx := errgroup.WithContext(ctx)

	runTaskProcessor(ctx, waitGroup, redisOpts, store, emailSender)

	servers.RunGatewayServer(ctx, waitGroup, config, store, taskDistributor)

	servers.RunGrpcServer(ctx, waitGroup, config, store, taskDistributor)

	err = waitGroup.Wait()
	if err != nil {
		log.Fatal().Err(err).Msg("error from wait group")
	}
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

func runTaskProcessor(
	ctx context.Context,
	waitGroup *errgroup.Group,
	redisOpts asynq.RedisClientOpt,
	store db.Store,
	emailSender mail.EmailSender,
) {
	processor := worker.NewRedisTaskProcessor(redisOpts, store, emailSender)

	err := processor.Start()
	if err != nil {
		log.Error().Err(err).Msg("failed to start task processor")
	}

	waitGroup.Go(func() error {
		<-ctx.Done()

		log.Info().Msg("task processor shutdown")

		processor.Shutdown()

		log.Info().Msg("task processor is stopped")

		return nil
	})
}
