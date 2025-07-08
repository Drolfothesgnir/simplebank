package api

import (
	"testing"

	db "github.com/Drolfothesgnir/simplebank/db/sqlc"
	"github.com/Drolfothesgnir/simplebank/util"
	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) (user db.User, password string) {
	password = util.RandomString(10)
	hashedPasword, err := util.HashPassword(password)
	require.NoError(t, err)

	user = db.User{
		Username:       util.RandomOwner(),
		HashedPassword: hashedPasword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}

	return
}
