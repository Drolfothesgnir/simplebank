package servers

import (
	"github.com/rs/zerolog/log"

	"github.com/Drolfothesgnir/simplebank/api"
	db "github.com/Drolfothesgnir/simplebank/db/sqlc"
	"github.com/Drolfothesgnir/simplebank/util"
)

func RunGinServer(config util.Config, store db.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create the server")
	}

	if err := server.Start(config.HTTPServerAddress); err != nil {
		log.Fatal().Err(err).Msg("cannot start the server")
	}
}
