#!/bin/sh
docker-compose exec postgres psql -c 'CREATE DATABASE yadro_test;' -U postgres

goose -dir "internal/schema" postgres "user=postgres password=password host=localhost port=5436 dbname=yadro_test sslmode=disable" up