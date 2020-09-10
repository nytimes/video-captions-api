#!/bin/bash -e

# This script triggers the deployment through a remote Drone server.

DRONE_VERSION=v1.2.2

function install_drone() {
	version=$1
	curl -sL https://github.com/drone/drone-cli/releases/download/${version}/drone_linux_amd64.tar.gz | tar -xzf -
	export PATH=:$PATH
}

function main() {
	env=$1
	image_tag=${TRAVIS_TAG:-${TRAVIS_COMMIT:0:8}}

	if [ "$TRAVIS_EVENT_TYPE" = "cron" ]; then
		echo >&2 "skipping deployment on cron"
		return 0
	fi

	if [ -z "$env" ]; then
		echo >&2 "missing env name"
		return 2
	fi

	install_drone $DRONE_VERSION
	last_build=$(drone build last --format "{{.Number}}" $DRONE_REPO)
	drone build promote -p IMAGE=nytimes/video-captions-api:${image_tag} $DRONE_REPO $last_build $env
}

main "$@"
