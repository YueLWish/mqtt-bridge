GOPATH:=$(shell go env GOPATH)

.PHONY: update
# git reset -q --hard HEAD
update:
	@git reset -q --hard HEAD
	@git pull -q
	@git log -3 --format="%C(magenta)%h %C(red)%d %C(yellow)(%cr) %C(green)%s"

.PHONY: build
# go build
build:
	mkdir -p bin/ && CGO_ENABLED=0 go build -ldflags '-w -s'  -trimpath -o ./bin/ ./...

# CGO_ENABLED=0
# -tags netgo


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