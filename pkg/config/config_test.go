package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)


func TestLoadConfig(t *testing.T) {

	t.Run("correct config", func(t *testing.T) {

	configContent := `
source_url: "http://example.com"
limit: 100
db_file: "test.db"
db_path: "/var/lib/test"
index_file: "index.yaml"
parallel: 5
srv_port: "8080"
until: 10
pg_dsn: "user=postgres password=secret dbname=test sslmode=disable"
concurrency_limit: 10
rate_limit: 1000
token_max_time: 3600
jwt_secret: "supersecretkey"
`
	tempFile, err := os.CreateTemp("", "config-*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write([]byte(configContent))
	assert.NoError(t, err)
	tempFile.Close()

	config, err := Load(tempFile.Name())
	assert.NoError(t, err)

	assert.Equal(t, "http://example.com", config.Path)
	assert.Equal(t, 100, config.Limit)
	assert.Equal(t, "test.db", config.DbFile)
	assert.Equal(t, "/var/lib/test", config.DbPath)
	assert.Equal(t, "index.yaml", config.IndexFile)
	assert.Equal(t, 5, config.Parallel)
	assert.Equal(t, "8080", config.SrvPort)
	assert.Equal(t, 10, config.Until)
	assert.Equal(t, "user=postgres password=secret dbname=test sslmode=disable", config.DSN)
	assert.Equal(t, 10, config.ConcurrencyLimit)
	assert.Equal(t, 1000, config.RateLimit)
	assert.Equal(t, 3600, config.TokenMaxTime)
	assert.Equal(t, "supersecretkey", config.JWTSecret)
	})

	t.Run("incorrect config", func(t *testing.T) {
		configContent := "1234"
		tempFile, err := os.CreateTemp("", "config-*.yaml")
		assert.NoError(t, err)
		defer os.Remove(tempFile.Name())
	
		_, err = tempFile.Write([]byte(configContent))
		assert.NoError(t, err)
		tempFile.Close()
	
		_, err = Load(tempFile.Name())

		assert.EqualError(t, err, "yaml: unmarshal errors:\n  line 1: cannot unmarshal !!int `1234` into config.Config")
	})
	t.Run("no file", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "config-*.cfg")
		assert.NoError(t, err)
		defer os.Remove(tempFile.Name())
	
		_, err = tempFile.Write([]byte(""))
		t.Log(err)
		assert.NoError(t, err)
		tempFile.Close()

		_, err = Load("")

		assert.EqualError(t, err, "open : The system cannot find the file specified.")
	})
}