PACKAGES := \
	github.com/andrewlouis93/SOCKS4/SOCKS
DEPENDENCIES := ""

all: build silent-test

build:
	go build -o proxy proxy.go

test:
	go test -v $(PACKAGES)

silent-test:
	go test $(PACKAGES)

format:
	go fmt $(PACKAGES)

deps:
	go get $(DEPENDENCIES)
