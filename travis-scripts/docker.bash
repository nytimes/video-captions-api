#!/bin/bash -e

if [ "${TRAVIS_PULL_REQUEST}" != "false" ]; then
	echo >&2 "Skipping image build on pull requests..."
	exit 0
fi

if [ "${TRAVIS_EVENT_TYPE}" = "cron" ]; then
	echo >&2 "Skipping image build on cron..."
	exit 0
fi

if [ "${TRAVIS_GO_VERSION}" != "${GO_FOR_RELEASE}" ]; then
	echo >&2 "Skipping image build on Go ${TRAVIS_GO_VERSION}"
	exit 0
fi

if [ "${TRAVIS_BRANCH}" != "main" ] && [ -z "${TRAVIS_TAG}" ]; then
	echo >&2 "Skipping image build on branch ${TRAVIS_BRANCH}"
	exit 0
fi

COMMIT_TAG=${TRAVIS_COMMIT:0:8}
IMAGE_NAME=nytimes/video-captions-api

docker login -u "${DOCKER_USERNAME}" -p "${DOCKER_PASSWORD}"
docker build -t ${IMAGE_NAME}:latest -f Dockerfile.ci .

docker tag ${IMAGE_NAME}:latest ${IMAGE_NAME}:${COMMIT_TAG}

if [ -n "${TRAVIS_TAG}" ]; then
	docker tag ${IMAGE_NAME}:latest ${IMAGE_NAME}:${TRAVIS_TAG}
fi

docker push ${IMAGE_NAME}
