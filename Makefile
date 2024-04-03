all: compile

compile:
	echo "Compiling for every OS and Platform"

	GOOS=linux GOARCH=amd64 cd cmd/xkcd && go build -o ../../xkcd.elf .

	GOOS=windows GOARCH=amd64 cd cmd/xkcd && go build -o ../../xkcd.exe .

run:
	@read -p "Enter a flag (optional):" flag; \
	./xkcd.elf $$flag

run_windows:
	@read -p "Enter a flag (optional):" flag; \
	./xkcd.exe $$flag

up: compile run