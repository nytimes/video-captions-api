module github.com/NYTimes/video-captions-api

replace github.com/nytimes-labs/client_golang v1.11.0 => github.com/prometheus/client_golang v1.11.0

require (
	cloud.google.com/go/datastore v1.4.0
	cloud.google.com/go/storage v1.14.0
	contrib.go.opencensus.io/exporter/prometheus v0.1.0
	github.com/NYTimes/gizmo v1.3.6
	github.com/NYTimes/gziphandler v1.1.1
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/google/go-cmp v0.5.5 // indirect
	github.com/google/uuid v1.1.3
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/nytimes/amara v0.3.0
	github.com/nytimes/threeplay v0.3.2
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v0.9.4
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/procfs v0.2.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/tdewolff/parse/v2 v2.4.3
	go.opencensus.io v0.23.0 // indirect
	golang.org/x/net v0.0.0-20210226172049-e18ecbb05110 // indirect
	golang.org/x/sys v0.0.0-20210615035016-665e8c7367d1 // indirect
	golang.org/x/text v0.3.6 // indirect
)

go 1.16
