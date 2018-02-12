#!/bin/bash -e

COMMIT_TAG=${TRAVIS_COMMIT:0:8}
IMAGE_NAME=nytimes/video-captions-api

docker login -u "${DOCKER_USERNAME}" -p "${DOCKER_PASSWORD}"
docker build -t ${IMAGE_NAME}:latest .

docker tag ${IMAGE_NAME}:latest ${IMAGE_NAME}:${COMMIT_TAG}

if [ -n "${TRAVIS_TAG}" ]; then
	docker tag ${IMAGE_NAME}:${TRAVIS_TAG}
fi

docker push ${IMAGE_NAME}
