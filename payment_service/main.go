package main

import (
	"github.com/gorilla/mux"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"io"
	"math/rand"
	"net/http"
	"time"
)

func main () {
	closer, err := initializeTracer()
	if err != nil {
		panic(err)
	}
	defer closer.Close()

	rand.Seed(time.Now().UnixNano())
	router := mux.NewRouter()

	tracer := opentracing.GlobalTracer()

	router.HandleFunc("/account", func(w http.ResponseWriter, r *http.Request) {
		spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		serverSpan := tracer.StartSpan("create_account", ext.RPCServerOption(spanCtx))
		defer serverSpan.Finish()

		if rand.Int31n(100) % 5 == 0 {
			serverSpan.SetTag("error", true)
			serverSpan.LogKV("error_code", "darn that mod 5")
		}

		time.Sleep(time.Duration(rand.Int31n(50)) * time.Millisecond)
		w.Write([]byte("account created"))
	}).Methods(http.MethodPost)

	http.ListenAndServe(":8003", router)
}

func initializeTracer() (io.Closer, error) {
	cfg :=jaegercfg.Configuration{
		ServiceName: "payment",
		Sampler:     &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter:    &jaegercfg.ReporterConfig{
			LogSpans: true,
		},
	}

	// Example logger and metrics factory. Use github.com/uber/jaeger-client-go/log
	// and github.com/uber/jaeger-lib/metrics respectively to bind to real logging and metrics
	// frameworks.
	jLogger := jaegerlog.StdLogger

	// Initialize tracer with a logger and a metrics factory
	tracer, closer, err := cfg.NewTracer(
		jaegercfg.Logger(jLogger),
	)

	testSpan := tracer.StartSpan("test span")
	testSpan.LogKV("someKey", "someVal")
	testSpan.Finish()
	// Set the singleton opentracing.Tracer with the Jaeger tracer.
	opentracing.SetGlobalTracer(tracer)
	return closer, err
}