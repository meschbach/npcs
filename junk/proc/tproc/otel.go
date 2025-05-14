package tproc

import (
	"context"
	"errors"
	slogmulti "github.com/samber/slog-multi"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
)

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func setupOTelSDK(ctx context.Context, name string) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up logger provider.
	loggerProvider, replaceDefaultLogger, err := newLoggerProvider(ctx)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	global.SetLoggerProvider(loggerProvider)

	if replaceDefaultLogger {
		otelLogger := otelslog.NewHandler(name)
		defaultHandler := slog.Default().Handler()
		combined := slogmulti.Fanout(otelLogger, defaultHandler)
		slog.SetDefault(slog.New(combined))
	}

	// Set up trace provider.
	tracerProvider, err := newTracerProvider(ctx)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	// Set up meter provider.
	meterProvider, err := newMeterProvider(ctx)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	return
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTracerProvider(ctx context.Context) (*trace.TracerProvider, error) {
	var opts []trace.TracerProviderOption

	if hasOneEnv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "OTEL_EXPORTER_OTLP_ENDPOINT") {
		slog.InfoContext(ctx, "OTLP Traces exported via GRPC")
		traceExporter, err := otlptracegrpc.New(ctx)
		if err != nil {
			return nil, err
		}
		opts = append(opts, trace.WithBatcher(traceExporter))
	}

	tracerProvider := trace.NewTracerProvider(opts...)
	return tracerProvider, nil
}

func newMeterProvider(ctx context.Context) (*metric.MeterProvider, error) {
	var opts []metric.Option
	if hasOneEnv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", "OTEL_EXPORTER_OTLP_ENDPOINT") {
		slog.InfoContext(ctx, "OTLP Metrics exported via GRPC")
		metricExporter, err := otlpmetricgrpc.New(ctx)
		if err != nil {
			return nil, err
		}
		opts = append(opts, metric.WithReader(metric.NewPeriodicReader(metricExporter)))
	}

	meterProvider := metric.NewMeterProvider(opts...)
	return meterProvider, nil
}

func newLoggerProvider(ctx context.Context) (*log.LoggerProvider, bool, error) {
	var opts []log.LoggerProviderOption
	replaceDefaultLogger := false

	if hasOneEnv("OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", "OTEL_EXPORTER_OTLP_ENDPOINT") {
		slog.InfoContext(ctx, "OTLP Logs exported via GRPC")
		logExporter, err := otlploggrpc.New(ctx)
		if err != nil {
			return nil, false, err
		}
		opts = append(opts, log.WithProcessor(log.NewBatchProcessor(logExporter)))
		replaceDefaultLogger = true
	}

	loggerProvider := log.NewLoggerProvider(opts...)
	return loggerProvider, replaceDefaultLogger, nil
}

func hasOneEnv(key ...string) bool {
	for _, k := range key {
		if _, ok := os.LookupEnv(k); ok {
			return true
		}
	}
	return false
}
