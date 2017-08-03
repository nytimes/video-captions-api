# video-captions-api

[![Build Status](https://travis-ci.org/NYTimes/video-captions-api.svg?branch=master)](https://travis-ci.org/NYTimes/video-captions-api)
[![codecov](https://codecov.io/gh/NYTimes/video-captions-api/branch/master/graph/badge.svg)](https://codecov.io/gh/NYTimes/video-captions-api)
[![Go Report Card](https://goreportcard.com/badge/github.com/NYTimes/video-captions-api)](https://goreportcard.com/report/github.com/NYTimes/video-captions-api)

Agnostic API to generate captions for media assets across different cloud services.

## Development

Install [Docker](https://www.docker.com/).

GCP credentials are required to access Google Datastore.  
`gcloud auth application-default login`

Build the dev image:

`$ make`

Run:

`$ make dev`

Go to http://localhost:8000

## Documentation

For more info check the [docs](https://github.com/NYTimes/video-captions-api/wiki/Home)


