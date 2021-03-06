module github.com/maestre3d/alexandria/media-service

go 1.13

require (
	contrib.go.opencensus.io/exporter/zipkin v0.1.1
	github.com/alexandria-oss/core v0.5.4-beta
	github.com/go-kit/kit v0.10.0
	github.com/go-playground/validator/v10 v10.3.0
	github.com/go-redis/redis/v7 v7.2.0
	github.com/golang/protobuf v1.4.0
	github.com/google/uuid v1.1.1
	github.com/google/wire v0.4.0
	github.com/gorilla/mux v1.7.3
	github.com/lib/pq v1.1.1
	github.com/matoous/go-nanoid v1.4.1
	github.com/oklog/run v1.1.0
	github.com/opentracing/opentracing-go v1.1.0
	github.com/openzipkin/zipkin-go v0.2.2
	github.com/prometheus/client_golang v1.5.1
	github.com/sony/gobreaker v0.4.1
	go.opencensus.io v0.22.3
	go.uber.org/zap v1.14.1 // indirect
	gocloud.dev v0.19.0
	google.golang.org/grpc v1.27.1
	google.golang.org/protobuf v1.21.0
)
