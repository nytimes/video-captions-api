dev:
	SERVER_HTTP_PORT=8000\
		SERVER_GIZMO_HEALTH_CHECK_PATH=/healthz\
		THREE_PLAY_API_KEY=$(THREE_PLAY_API_KEY) \
		THREE_PLAY_API_SECRET=$(THREE_PLAY_API_SECRET) \
		PROJECT_ID=nyt-video-dev \
		BUCKET_NAME=video-captions-api-dev \
		go run main.go

test:
	go test -v $$(go list ./... |grep -v vendor)
