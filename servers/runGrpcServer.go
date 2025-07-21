package servers

import (
	"context"
	"errors"
	"net"

	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	db "github.com/Drolfothesgnir/simplebank/db/sqlc"
	"github.com/Drolfothesgnir/simplebank/gapi"
	"github.com/Drolfothesgnir/simplebank/pb"
	"github.com/Drolfothesgnir/simplebank/util"
	"github.com/Drolfothesgnir/simplebank/worker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func RunGrpcServer(
	ctx context.Context,
	waitGroup *errgroup.Group,
	config util.Config,
	store db.Store,
	taskDistributor worker.TaskDistributor,
) {
	server, err := gapi.NewServer(config, store, taskDistributor)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create the gRPC server")
	}

	grpcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)

	grpcServer := grpc.NewServer(grpcLogger)
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create listener")
	}

	waitGroup.Go(func() error {
		log.Info().Msgf("start gRPC server at %s", listener.Addr().String())
		err = grpcServer.Serve(listener)

		if err != nil {
			if errors.Is(err, grpc.ErrServerStopped) {
				return nil
			}
			log.Error().Err(err).Msg("cannot start gRPC server")
		}

		return err
	})

	waitGroup.Go(func() error {
		<-ctx.Done()

		log.Info().Msg("gRPC server: graceful shutdown")

		grpcServer.GracefulStop()

		log.Info().Msg("gRPC server is stopped")
		return nil
	})
}
