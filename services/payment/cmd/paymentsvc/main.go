package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"golang.org/x/net/context"

	"google.golang.org/grpc"

	"github.com/microservices-demo/payment/payment"

	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const (
	ServiceName = "payment"
)

func main() {
	var (
		port          = flag.String("port", "8080", "Port to bind HTTP listener")
		otelurl       = flag.String("otel", os.Getenv("OTEL"), "OTLP address")
		declineAmount = flag.Float64("decline", 100, "Decline payments over certain amount")
	)
	flag.Parse()

	// Log domain.
	var logger log.Logger
	ctx := context.Background()
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
	}

	//var bsp sdktrace.SpanProcessor

	if *otelurl == "" {
		traceExporter, err := stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
		)
		if err != nil {
			level.Error(logger).Log("failed to initialize stdouttrace export pipeline: %v", err)
			os.Exit(1)
		}
		bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
		tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(bsp))
		otel.SetTracerProvider(tp)
	} else {
		//logger := log.NewContext(logger).With("tracer", "Otel")
		otelurlValue := *otelurl
		logger.Log("addr", otelurl)

		ctx := context.Background()
		traceExporter, err := otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(otelurlValue),
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithDialOption(grpc.WithBlock()),
		)
		if err != nil {
			level.Error(logger).Log("Failed to create the collector exporter: %v", err)
			os.Exit(1)
		}
		defer func() {
			ctx, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel()
			if err := traceExporter.Shutdown(ctx); err != nil {
				otel.Handle(err)
			}
		}()

		res, err := resource.New(ctx,
			resource.WithAttributes(
				// the service name used to display traces in backends
				semconv.ServiceNameKey.String("payment"),
			),
		)

		tp := sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithResource(res),
			sdktrace.WithBatcher(
				traceExporter,
				// add following two options to ensure flush
				sdktrace.WithBatchTimeout(5*time.Second),
				sdktrace.WithMaxExportBatchSize(10),
			),
		)
		defer func() {
			ctx, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel()
			if err := tp.Shutdown(ctx); err != nil {
				otel.Handle(err)
			}
		}()
		otel.SetTracerProvider(tp)
		propagator := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
		otel.SetTextMapPropagator(propagator)
	}
	// Mechanical stuff.
	errc := make(chan error)

	handler, logger := payment.WireUp(ctx, float32(*declineAmount), ServiceName)

	// Create and launch the HTTP server.
	go func() {
		logger.Log("transport", "HTTP", "port", *port)
		errc <- http.ListenAndServe(":"+*port, otelhttp.NewHandler(handler, "payment",
			otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents),
	))
	}()

	// Capture interrupts.
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	logger.Log("exit", <-errc)
}
