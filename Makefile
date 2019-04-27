.PHONY: all check build

all: build

build:
	go build -v -race

check:
	go test -bench .
