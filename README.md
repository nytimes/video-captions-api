# video-captions-api

[![Build Status](https://travis-ci.org/NYTimes/video-captions-api.svg?branch=master)](https://travis-ci.org/NYTimes/video-captions-api)
[![codecov](https://codecov.io/gh/NYTimes/video-captions-api/branch/master/graph/badge.svg)](https://codecov.io/gh/NYTimes/video-captions-api)
[![Go Report Card](https://goreportcard.com/badge/github.com/NYTimes/video-captions-api)](https://goreportcard.com/report/github.com/NYTimes/video-captions-api)

Agnostic API to generate captions for media assets across different cloud services.

## Development

Use `dep` to install dependencies:  
`go get -u github.com/golang/dep/cmd/dep`  
`dep ensure`

GCP credentials are required to access Google Datastore.  
`gcloud auth application-default login`

Environment variables required:

```
THREE_PLAY_API_KEY
THREE_PLAY_API_SECRET
```

Run:

`$ make dev`

Test:

`$ make test`

Go to http://localhost:8000

## Documentation

For more info check the [docs](https://github.com/NYTimes/video-captions-api/wiki/Home)


