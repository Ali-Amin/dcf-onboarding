.PHONY: build

build:
	go build -o ./cmd/agent/dcfagent ./cmd/agent/main.go
	go build -o ./cmd/onboarder/onboarder ./cmd/onboarder/main.go


run:
	./cmd/onboarder/onboarder -cfg ./cmd/onboarder/res/config.json
