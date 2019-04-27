.PHONY: all check build clean

CLEAN=

all: build

build:
	go build -v -race

check:
	go test -vet "" -bench .

clean:
	$(RM) $(CLEAN)
