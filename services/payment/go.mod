module github.com/microservices-demo/payment

go 1.16

require (
	github.com/go-kit/kit v0.11.0
	github.com/gorilla/mux v1.8.0
	github.com/prometheus/client_golang v1.11.0
	github.com/streadway/handy v0.0.0-20200128134331-0f66f006fb2e
	go.opentelemetry.io/contrib/instrumentation/github.com/go-kit/kit/otelkit v0.21.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.21.0
	go.opentelemetry.io/otel v1.0.0-RC1
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.0.0-RC1
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.0.0-RC1
	go.opentelemetry.io/otel/sdk v1.0.0-RC1
	go.opentelemetry.io/otel/trace v1.0.0-RC1
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e
	google.golang.org/genproto v0.0.0-20210701191553-46259e63a0a9 // indirect
	google.golang.org/grpc v1.39.0
)
