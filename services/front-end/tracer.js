/* tracing.js */

// Require dependencies
const { NodeTracerProvider } = require("@opentelemetry/node");
const { diag, DiagConsoleLogger, DiagLogLevel } = require("@opentelemetry/api");
const { BatchSpanProcessor, ConsoleSpanExporter } = require("@opentelemetry/tracing");
const { CollectorTraceExporter } =  require('@opentelemetry/exporter-collector-grpc');
const { getNodeAutoInstrumentations } = require('@opentelemetry/auto-instrumentations-node');
const { registerInstrumentations } = require('@opentelemetry/instrumentation');
const { Resource } = require('@opentelemetry/resources');
const { ResourceAttributes } = require('@opentelemetry/semantic-conventions');

const collectorOptions = {
  // url is optional and can be omitted - default is localhost:4317
  url: process.env.OTEL_EXPORTER_OTLP_ENDPOINT,
};

// Create a tracer provider
const provider = new NodeTracerProvider({
  resource: new Resource({
    [ResourceAttributes.SERVICE_NAME]: 'front-end',
  }),
});

diag.setLogger(new DiagConsoleLogger(), DiagLogLevel.INFO);

// The exporter handles sending spans to your tracing backend
//const exporter = new ConsoleSpanExporter();
const exporter = new CollectorTraceExporter(collectorOptions);

// The simple span processor sends spans to the exporter as soon as they are ended.
const processor = new BatchSpanProcessor(exporter);
provider.addSpanProcessor(processor);

// The provider must be registered in order to
// be used by the OpenTelemetry API and instrumentations
provider.register();

['SIGINT', 'SIGTERM'].forEach(signal => {
  process.on(signal, () => provider.shutdown().catch(console.error));
});

// This will automatically enable all instrumentations
registerInstrumentations({
  instrumentations: [getNodeAutoInstrumentations()],
});

console.log('tracing initialized using ${exporter.url}');

