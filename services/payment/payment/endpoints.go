package payment

import (
	"fmt"

	"github.com/go-kit/kit/endpoint"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/net/context"

	oteltrace "go.opentelemetry.io/otel/trace"

	"go.opentelemetry.io/contrib/instrumentation/github.com/go-kit/kit/otelkit"
)

// Endpoints collects the endpoints that comprise the Service.
type Endpoints struct {
	AuthoriseEndpoint endpoint.Endpoint
	HealthEndpoint    endpoint.Endpoint
}

var tracer = otel.Tracer("payment")

// MakeEndpoints returns an Endpoints structure, where each endpoint is
// backed by the given service.
func MakeEndpoints(s Service) Endpoints {
	return Endpoints{
		AuthoriseEndpoint: otelkit.EndpointMiddleware(otelkit.WithOperation("authorize payment"),)(MakeAuthoriseEndpoint(s)),
		HealthEndpoint:    otelkit.EndpointMiddleware(otelkit.WithOperation("health check"),)(MakeHealthEndpoint(s)),
	}
}

// MakeListEndpoint returns an endpoint via the given service.
func MakeAuthoriseEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		span := oteltrace.SpanFromContext(ctx)
		//defer span.End()
		req := request.(AuthoriseRequest)
		span.SetAttributes(attribute.String("amount", fmt.Sprintf("%.2f", req.Amount)),attribute.String("service", "payment"))
		authorisation, err := s.Authorise(req.Amount, span.SpanContext().TraceID().String())
		if (err == ErrGatewayUnavailable) {
			span.RecordError(err)
			span.SetStatus(codes.Error, "payment gateway error")
		}
		
		return AuthoriseResponse{Authorisation: authorisation, Err: err}, nil
	}
}

// MakeHealthEndpoint returns current health of the given service.
func MakeHealthEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		span := oteltrace.SpanFromContext(ctx)
		span.SetAttributes(attribute.String("service", "payment"))
		//defer span.End()
		health := s.Health()
		return healthResponse{Health: health}, nil
	}
}

// AuthoriseRequest represents a request for payment authorisation.
// The Amount is the total amount of the transaction
type AuthoriseRequest struct {
	Amount float32 `json:"amount"`
}

// AuthoriseResponse returns a response of type Authorisation and an error, Err.
type AuthoriseResponse struct {
	Authorisation Authorisation
	Err           error
	TraceID oteltrace.TraceID
}

type healthRequest struct {
	//
}

type healthResponse struct {
	Health []Health `json:"health"`
}
