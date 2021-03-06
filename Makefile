.PHONY: all check build clean

CLEAN=

all: build

build:
	go build -v -race

check:
	go test -vet "" -bench . -v

clean:
	$(RM) $(CLEAN)
