GOBUILD=go build -v

LDFLAGS += -X "main.GitHash=$(shell git rev-parse HEAD)"
LDFLAGS += -X "main.BuildTS=$(shell go run script/buildcmd.go time)"

all: build

.PHONY : clean deps build

build: deps 
	$(GOBUILD) -ldflags '-s -w $(LDFLAGS)'
