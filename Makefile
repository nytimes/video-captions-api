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

coverage:
	@ echo "" > coverage.txt; \
		for p in $$(go list ./...); do \
			go test -coverprofile=profile.out -covermode=atomic $$p || export status=2; \
			if [ -f profile.out ]; then cat profile.out >> coverage.txt; rm profile.out; fi; \
		done; \
		exit ${status:-0}


test:
	go test -v ./...
