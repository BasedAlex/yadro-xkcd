all: compile

compile:
	echo "Compiling for every OS and Platform"

	cd cmd/xkcd && GOOS=linux GOARCH=amd64 go build -o ../../xkcd.elf .
	cd cmd/xkcd && GOOS=windows GOARCH=amd64 go build -o ../../xkcd.exe .

run:
	@read -p "Enter a flag (optional):" flag; \
	./xkcd.elf $$flag

docker_up:
	@echo Starting Docker images..
	docker-compose up -d 
	@echo Docker images started

run_migrations:
	go install github.com/pressly/goose/v3/cmd/goose@latest
	cd internal/schema && ${GOOSE_UP}

down_migrations: 
	cd internal/schema && ${GOOSE_DOWN}

up: compile run

first_run: docker_up run_migrations compile run

# $ curl -s https://packagecloud.io/install/repositories/golang-migrate/migrate/script.deb.sh | sudo bash
# $ apt-get update
# $ apt-get install -y migrate