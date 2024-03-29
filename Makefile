UNAME := $(shell uname)
ifeq ($(UNAME), Linux)
	TARGET_PLATFORM := "linux"
else ifeq ($(UNAME), Darwin)
	TARGET_PLATFORM := "darwin"
else
	TARGET_PLATFORM := "unknown"
endif

build: .version
	env GOOS=linux GOARCH=amd64 go build -ldflags="-X 'aws2fa/cmd.version=$(shell cat .version)'" -o build/aws2fa-linux-amd64
	env GOOS=darwin GOARCH=amd64 go build -ldflags="-X 'aws2fa/cmd.version=$(shell cat .version)'" -o build/aws2fa-darwin-amd64
	env GOOS=windows GOARCH=amd64 go build -ldflags="-X 'aws2fa/cmd.version=$(shell cat .version)'" -o build/aws2fa-windows-amd64

.version:
	echo "0.0+dev-build" > .version

clean:
	rm -rf ./build

install:
	cp build/aws2fa-$(TARGET_PLATFORM)-amd64 $(HOME)/.bin/aws2fa

test:
	go test -v ./...

default: build
