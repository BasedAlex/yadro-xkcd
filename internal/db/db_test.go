package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/basedalex/yadro-xkcd/pkg/config"
	"github.com/go-playground/assert/v2"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

var database *Postgres

func Test_NewPostgres(t *testing.T) {
	t.Run("Invalid connection string", func(t *testing.T) { 
		_, err := NewPostgres(context.Background(), "")
		if err == nil {
			t.Log("no error where should be") 
		}
	})
}

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

func Test_GetUserPasswordByLogin(t *testing.T) {
	t.Run("valid login", func(t *testing.T) { 
		expectedPassword := "admin"
		password, err := database.GetUserPasswordByLogin(context.Background(), "admin")
		require.NoError(t, err)
		assert.Equal(t, expectedPassword, password)
	})

	t.Run("invalid login", func(t *testing.T) { 
		_, err := database.GetUserPasswordByLogin(context.Background(), "random")
		require.EqualError(t, err, fmt.Sprintf("database: %s", pgx.ErrNoRows))
	})
}

func Test_SaveComics(t *testing.T) {

	cfg := &config.Config{
	}
	ctx := context.Background()
	t.Run("comic doesn't exist yet", func(t *testing.T) { 
		var comic Page
		comic.Img = "example.com/test"
		comic.Index = "2000000"
		comic.Keywords = []string{"testing", "software"}

		database.SaveComics(context.Background(), cfg, comic)

		query := `SELECT id FROM comics WHERE image = $1;`
		row := database.db.QueryRow(ctx, query, comic.Img)
		var id int 
		err := row.Scan(&id)
		require.NoError(t, err)
		stmt := `DELETE FROM comics WHERE image = $1;`
		_, err = database.db.Exec(ctx, stmt, comic.Img)
		require.NoError(t, err)
	})
}

func Test_InvertSearch(t *testing.T) {

	cfg := &config.Config{
	}
	t.Run("valid search", func(t *testing.T) { 
		database.Reverse(context.Background(), cfg)
		expectedResult := make(map[string][]int, 0)
		expectedResult["test"] = []int{135} 

		result, err := database.InvertSearch(context.Background(), cfg, "test")
		require.NoError(t, err)
		assert.Equal(t, expectedResult, result)
	})

	t.Run("invalid search", func(t *testing.T) { 
		_, err := database.InvertSearch(context.Background(), cfg, "")
		t.Log(err)
		expErr := errors.New("please provide a string to be stemmed")
		t.Log(expErr)
		assert.Equal(t, expErr, err)
	})
}