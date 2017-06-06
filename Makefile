
default:
	docker build -f Dockerfile.dev -t video-captions-api:dev .

dev:
	docker run -itp 8000:8000 -v $(shell pwd):/go/src/github.com/NYTimes/video-captions-api video-captions-api:dev
