GOPATH:=$(shell go env GOPATH)
ROOT_DIR:=$(shell dirname $(MAKEFILE_LIST))
ROOT_NAME:=$(shell basename $(ROOT_DIR))

.PHONY: update
# git reset -q --hard HEAD
update:
	@git reset -q --hard HEAD
	@git pull -q
	@git log -3 --format="%C(magenta)%h %C(red)%d %C(yellow)(%cr) %C(green)%s"

.PHONY: build
# go build
build:
	mkdir -p bin/ && \
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -ldflags '-w -s'  -trimpath -o ./bin/$(ROOT_NAME)-linux-amd64 ./ && \
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags '-w -s'  -trimpath -o ./bin/$(ROOT_NAME)-windows-amd64.exe ./

.PHONY: build
# go build MAC相关版本
build-mac:
	mkdir -p bin/ && \
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags '-w -s'  -trimpath -o ./bin/$(ROOT_NAME)-darwin-amd64 ./ && \
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags '-w -s'  -trimpath -o ./bin/$(ROOT_NAME)-darwin-arm64 ./



.PHONY: lint
# golang lint
lint:
	golangci-lint run ./...


.PHONY:
# show help
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\\_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)