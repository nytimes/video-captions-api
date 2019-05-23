module github.com/NYTimes/video-captions-api

require (
	cloud.google.com/go v0.38.0
	github.com/NYTimes/gizmo v1.2.7
	github.com/NYTimes/gziphandler v1.1.1
	github.com/google/uuid v1.1.1
	github.com/kelseyhightower/envconfig v1.3.0
	github.com/nytimes/amara v0.3.0
	github.com/nytimes/threeplay v0.2.0
	github.com/prometheus/common v0.0.0-20181218105931-67670fe90761 // indirect
	github.com/sirupsen/logrus v1.4.1
	github.com/stretchr/testify v1.3.0
)

replace github.com/nytimes/threeplay v0.2.0 => ../threeplay

go 1.13
