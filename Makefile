dev:
	@go build -v -o captions-api .

	@SERVER_HTTP_PORT=8000\
		SERVER_GIZMO_HEALTH_CHECK_PATH=/healthz\
		THREE_PLAY_API_KEY=$(THREE_PLAY_API_KEY) \
		THREE_PLAY_API_SECRET=$(THREE_PLAY_API_SECRET) \
		AMARA_USERNAME=$(AMARA_USERNAME) \
		AMARA_TEAM=$(AMARA_TEAM) \
		AMARA_TOKEN=$(AMARA_TOKEN) \
		PROJECT_ID=$(CAPTIONS_PROJECT_ID) \
		BUCKET_NAME=$(CAPTIONS_BUCKET_NAME) \
		CALLBACK_URL=$(CALLBACK_URL) \
		CALLBACK_API_KEY=$(CALLBACK_API_KEY) \
		./captions-api

install-golangcilint:
	go get github.com/golangci/golangci-lint/cmd/golangci-lint

run-lint:
	golangci-lint run -D lll -D funlen --deadline 5m ./...

PHONY:
lint: install-golangcilint run-lint

coverage:
	go test -coverprofile=coverage.txt -covermode=atomic ./...

test:
	go test -race ./...
