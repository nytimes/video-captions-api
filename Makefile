dev:
	SERVER_HTTP_PORT=8000\
		SERVER_GIZMO_HEALTH_CHECK_PATH=/healthz\
		THREE_PLAY_API_KEY=$(THREE_PLAY_API_KEY) \
		THREE_PLAY_API_SECRET=$(THREE_PLAY_API_SECRET) \
		AMARA_USERNAME=$(AMARA_USERNAME) \
		AMARA_TEAM=$(AMARA_TEAM) \
		AMARA_TOKEN=$(AMARA_TOKEN) \
		PROJECT_ID=nyt-video-dev \
		BUCKET_NAME=video-captions-api-dev \
		go run main.go

install-golangcilint:
	cd /tmp && go get github.com/golangci/golangci-lint/cmd/golangci-lint

run-lint:
	golangci-lint run --fast -D errcheck -E megacheck --deadline 5m ./...

lint: install-golangcilint run-lint

coverage:
	go test -coverprofile=coverage.txt -covermode=atomic ./...

test:
	go test -v ./...
