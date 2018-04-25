GOBUILD=go build -v
BuildTS=$(shell go run script/buildcmd.go time)

LDFLAGS += -X "main.GitHash=$(shell git rev-parse HEAD)"
LDFLAGS += -X "main.BuildTS=$(BuildTS)"

all: build

.PHONY : clean build

build:  
	$(GOBUILD) -ldflags '-s -w $(LDFLAGS)'
