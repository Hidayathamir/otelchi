package utils

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
)

func NewMeter(svcName string) (metric.Meter, error) {
	exporter, err := otlpmetrichttp.New(context.Background(), otlpmetrichttp.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("unable to initialize exporter due: %w", err)
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(
				exporter,
				sdkmetric.WithInterval(time.Second), // default is 60 seconds, for testing we set it 1 second.
			),
		),
		sdkmetric.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(svcName),
		)),
	)

	otel.SetMeterProvider(mp)

	return otel.Meter(svcName), nil
}
