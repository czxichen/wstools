GOBUILD=go build -v

LDFLAGS += -X "main.GitHash=$(shell git rev-parse HEAD)"
LDFLAGS += -X "main.BuildTS=$(shell go run buildcmd.go time)"

all: build

.PHONY : clean deps build build-linux build-windows test

build: deps 
	$(GOBUILD) -ldflags '-s -w $(LDFLAGS)' -i wstools.go

test: deps
	go test -v ./...
