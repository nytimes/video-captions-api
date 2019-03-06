# video-captions-api

[![Build Status](https://travis-ci.org/nytimes/video-captions-api.svg?branch=master)](https://travis-ci.org/nytimes/video-captions-api)
[![codecov](https://codecov.io/gh/nytimes/video-captions-api/branch/master/graph/badge.svg)](https://codecov.io/gh/nytimes/video-captions-api)
[![Go Report Card](https://goreportcard.com/badge/github.com/nytimes/video-captions-api)](https://goreportcard.com/report/github.com/nytimes/video-captions-api)

Agnostic API to generate captions for media assets across different cloud services.

## Development

GCP credentials are required to access Google Datastore.

```
$ gcloud auth application-default login
```

Environment variables required:

```
THREE_PLAY_API_KEY
THREE_PLAY_API_SECRET
```

Run:

```
$ make dev
```

Test:

```
$ make test
```

Go to http://localhost:8000

## Docker image

A pre-built image is available on Docker Hub: https://hub.docker.com/r/nytimes/video-captions-api

## Documentation

For more info check the [docs](https://github.com/nytimes/video-captions-api/wiki/Home)
