package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/riandyrn/otelchi"
	"github.com/riandyrn/otelchi/examples/multi-services/utils"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const (
	addr        = ":8091"
	serviceName = "back-svc"
)

func main() {
	// init tracer provider
	tracer, err := utils.NewTracer(serviceName)
	if err != nil {
		log.Fatalf("unable to initialize tracer provider due: %v", err)
	}
	meter, err := utils.NewMeter(serviceName)
	if err != nil {
		log.Fatalf("unable to initialize meter provider due: %v", err)
	}
	apiGetNameCounter, err := meter.Int64Counter("get-name", metric.WithDescription("count api GET /name being hit"))
	if err != nil {
		log.Fatalf("unable to create counter due: %v", err)
	}
	// define router
	r := chi.NewRouter()
	r.Use(otelchi.Middleware(serviceName, otelchi.WithChiRoutes(r)))
	r.Get("/", utils.HealthCheckHandler)
	r.Get("/name", func(w http.ResponseWriter, r *http.Request) {
		apiGetNameCounter.Add(r.Context(), 1)
		w.Write([]byte(generateName(r.Context(), tracer)))
	})
	log.Printf("back service is listening on %v", addr)
	err = http.ListenAndServe(addr, r)
	if err != nil {
		log.Fatalf("unable to execute server due: %v", err)
	}
}

func generateName(ctx context.Context, tracer trace.Tracer) string {
	_, span := tracer.Start(ctx, "generateName")
	defer span.End()

	rndNum := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(100000)
	return fmt.Sprintf("user_%v", rndNum)
}
