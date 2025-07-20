package gapi

import (
	"context"
	"fmt"
	"strings"

	"github.com/Drolfothesgnir/simplebank/token"
	"google.golang.org/grpc/metadata"
)

const (
	authorizationHeader     = "authorization"
	authorizationTypeBearer = "bearer"
)

func (server *Server) authorizeUser(ctx context.Context) (*token.Payload, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("missing metadata")
	}

	authValues := md.Get(authorizationHeader)

	if len(authValues) == 0 {
		return nil, fmt.Errorf("missing authorization header")
	}

	authHeader := authValues[0]
	fields := strings.Fields(authHeader)
	if len(fields) < 2 {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	authoriztionType := strings.ToLower(fields[0])
	if authoriztionType != authorizationTypeBearer {
		return nil, fmt.Errorf("unsupported authorization type: %s", authoriztionType)
	}

	accessToken := fields[1]

	payload, err := server.tokenMaker.VerifyToken(accessToken)
	if err != nil {

		return nil, fmt.Errorf("invalid access token: %s", err)
	}

	return payload, nil
}
