package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/riandyrn/otelchi"
	"github.com/riandyrn/otelchi/examples/multi-services/utils"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const (
	envKeyBackServiceURL = "BACK_SERVICE_URL"
	addr                 = ":8090"
	serviceName          = "front-svc"
)

func main() {
	// initialize tracer
	tracer, err := utils.NewTracer(serviceName)
	if err != nil {
		log.Fatalf("unable to initialize tracer due: %v", err)
	}
	meter, err := utils.NewMeter(serviceName)
	if err != nil {
		log.Fatalf("unable to initialize meter provider due: %v", err)
	}
	apiGetGreetCounter, err := meter.Int64Counter("get-greet", metric.WithDescription("count api GET /greet being hit"))
	if err != nil {
		log.Fatalf("unable to create counter due: %v", err)
	}
	// define router
	r := chi.NewRouter()
	r.Use(otelchi.Middleware(serviceName, otelchi.WithChiRoutes(r)))
	r.Get("/", utils.HealthCheckHandler)
	r.Get("/greet", func(w http.ResponseWriter, r *http.Request) {
		apiGetGreetCounter.Add(r.Context(), 1)
		name, err := getRandomName(r.Context(), tracer)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write([]byte(fmt.Sprintf("Hello, %s!", name)))
	})
	// execute server
	log.Printf("front service is listening on %v", addr)
	err = http.ListenAndServe(addr, r)
	if err != nil {
		log.Fatalf("unable to execute server due: %v", err)
	}
}

func getRandomName(ctx context.Context, tracer trace.Tracer) (string, error) {
	// start span
	ctx, span := tracer.Start(ctx, "getRandomName")
	defer span.End()

	// call back service, notice that here we call the service using instrumented
	// http client
	resp, err := otelhttp.Get(ctx, os.Getenv(envKeyBackServiceURL)+"/name")
	if err != nil {
		err = fmt.Errorf("unable to execute http request due: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}
	defer resp.Body.Close()

	// read response body
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("unable to read response data due: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}

	return string(data), nil
}
