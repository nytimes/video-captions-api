
default:
	docker build -f Dockerfile.dev -t video-captions-api:dev .

dev:
	docker run -itp 8000:8000 -v $(shell pwd):/go/src/github.com/NYTimes/video-captions-api -v ~/.config/gcloud/:/root/.config/gcloud \
		-e THREE_PLAY_API_KEY=$(THREE_PLAY_API_KEY) \
		-e THREE_PLAY_API_SECRET=$(THREE_PLAY_API_SECRET) \
		-e BUCKET_NAME=video-captions-api-dev \
		-e PROJECT_ID=nyt-video-dev \
		video-captions-api:dev

test:
	docker run -it -v $(shell pwd):/go/src/github.com/NYTimes/video-captions-api video-captions-api:dev  go test -v $$(go list ./... |grep -v vendor)
