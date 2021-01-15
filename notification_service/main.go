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

	router.HandleFunc("/notify", func(w http.ResponseWriter, r *http.Request) {
		spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		serverSpan := tracer.StartSpan("notify", ext.RPCServerOption(spanCtx))
		defer serverSpan.Finish()

		serverSpan.LogKV("notification_id", rand.Int31n(1000))

		time.Sleep(time.Duration(rand.Int31n(50)) * time.Millisecond)
		w.Write([]byte("notify"))
	}).Methods(http.MethodGet)

	http.ListenAndServe(":8001", router)
}

func initializeTracer() (io.Closer, error) {
	cfg :=jaegercfg.Configuration{
		ServiceName: "notification",
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