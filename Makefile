UNAME := $(shell uname)
ifeq ($(UNAME), Linux)
	TARGET_PLATFORM := "linux"
else ifeq ($(UNAME), Darwin)
	TARGET_PLATFORM := "darwin"
else
	TARGET_PLATFORM := "unknown"
endif

aws2fa-linux-amd64: .version
	env GOOS=linux GOARCH=amd64 go build -ldflags="-X 'gyd/cmd.version=$(shell cat .version)'" -o build/aws2fa-linux-amd64

aws2fa-darwin-amd64: .version
	env GOOS=darwin GOARCH=amd64 go build -ldflags="-X 'gyd/cmd.version=$(shell cat .version)'" -o build/aws2fa-darwin-amd64

build: aws2fa-linux-amd64 aws2fa-darwin-amd64

.version:
	echo "dev" > .version

clean:
	rm -rf ./build

install:
	cp build/aws2fa-$(TARGET_PLATFORM)-amd64 $(HOME)/.bin/aws2fa

default: build
