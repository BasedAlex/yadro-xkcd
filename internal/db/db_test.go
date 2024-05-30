package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

var database *Postgres

func TestMain(m *testing.M) {
	var err error

	database, err = NewPostgres(context.Background(), "user=postgres password=password host=localhost port=5436 dbname=yadro_test sslmode=disable pool_max_conns=10")

	if err != nil {
		log.Fatal(err)
	}
	os.Exit(m.Run())
}

func Test_GetUserByLogin(t *testing.T) {
	t.Run("valid login", func(t *testing.T) { 
		expectedUser := User{
			Login: "admin",
			Role: "admin",
		}
		user, err := database.GetUserByLogin(context.Background(), "admin")
		require.NoError(t, err)
		assert.Equal(t, expectedUser, user)
	})

	t.Run("invalid login", func(t *testing.T) { 
		_, err := database.GetUserByLogin(context.Background(), "random")

		require.EqualError(t, err, fmt.Sprintf("database: %s", pgx.ErrNoRows))
	})
}