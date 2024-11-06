-include .env

ANSI_FG_RED = \033[31m
ANSI_RESET = \033[39m

run:
	go run .

test:
	go test . -failfast -v

build: test
	mkdir ./build -p
	go build -ldflags="-s -w" -o ./build/c .

install: test
	@test -n "${PREFIX}" || (echo '$(ANSI_FG_RED)Error$(ANSI_RESET): missing $$PREFIX' && exit 1)
	go build -ldflags="-s -w" -o ${PREFIX}/c .

clean:
	rm -r ./build

.PHONY: run test build install clean
