module github.com/maestre3d/alexandria/category-service

go 1.14

require (
	contrib.go.opencensus.io/exporter/prometheus v0.2.0
	contrib.go.opencensus.io/exporter/zipkin v0.1.1
	github.com/afex/hystrix-go v0.0.0-20180502004556-fa1af6a1f4f5
	github.com/alexandria-oss/core v0.5.4-beta
	github.com/eapache/go-resiliency v1.2.0
	github.com/go-kit/kit v0.10.0
	github.com/go-playground/validator/v10 v10.3.0
	github.com/go-redis/redis/v7 v7.2.0
	github.com/gocql/gocql v0.0.0-20200624222514-34081eda590e
	github.com/google/uuid v1.1.1
	github.com/google/wire v0.3.0
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.3
	github.com/matoous/go-nanoid v1.4.1
	github.com/oklog/run v1.1.0 // indirect
	github.com/openzipkin/zipkin-go v0.2.2
	github.com/prometheus/client_golang v1.3.0
	github.com/spf13/viper v1.6.3
	go.opencensus.io v0.22.3
	go.uber.org/ratelimit v0.1.0
	gocloud.dev v0.19.0
)
