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

up: compile run

# $ curl -s https://packagecloud.io/install/repositories/golang-migrate/migrate/script.deb.sh | sudo bash
# $ apt-get update
# $ apt-get install -y migrate