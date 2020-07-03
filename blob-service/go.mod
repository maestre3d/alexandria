module github.com/maestre3d/alexandria/blob-service

go 1.14

require (
	contrib.go.opencensus.io/exporter/zipkin v0.1.1
	github.com/alexandria-oss/core v0.5.2-beta
	github.com/go-kit/kit v0.10.0
	github.com/go-playground/validator/v10 v10.3.0
	github.com/google/uuid v1.1.1
	github.com/google/wire v0.4.0
	github.com/gorilla/mux v1.7.4
	github.com/maestre3d/alexandria/author-service v0.0.0-20200622084343-25bca5313f9f // indirect
	github.com/oklog/run v1.1.0
	github.com/opentracing/opentracing-go v1.1.0
	github.com/openzipkin/zipkin-go v0.2.2
	github.com/prometheus/client_golang v1.5.1
	github.com/sony/gobreaker v0.4.1
	go.opencensus.io v0.22.3
	gocloud.dev v0.20.0
	gocloud.dev/pubsub/kafkapubsub v0.20.0 // indirect
)
