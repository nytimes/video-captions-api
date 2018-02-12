#!/bin/bash -e

# This script triggers the deployment through a remote Drone server.

DRONE_VERSION=v0.8.1

function install_drone() {
	version=$1
	curl -sL https://github.com/drone/drone-cli/releases/download/${version}/drone_linux_amd64.tar.gz | tar -xzf -
	export PATH=:$PATH
}

function main() {
	env=$1
	image_tag=${TRAVIS_COMMIT:0:8}

	if [ -z "$env" ]; then
		echo >&2 "missing env name"
		exit 2
	fi

	install_drone $DRONE_VERSION
	last_build=$(drone build last --format "{{.Number}}" $DRONE_REPO)
	drone deploy -p IMAGE=nytimes/video-captions-api:${image_tag} $DRONE_REPO $last_build $env
}

main "$@"
