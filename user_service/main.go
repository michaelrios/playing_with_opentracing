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

	router.HandleFunc("/create", func(w http.ResponseWriter, r *http.Request) {
		// initialize tracer
		spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		serverSpan := tracer.StartSpan("create_user", ext.RPCServerOption(spanCtx))
		defer serverSpan.Finish()

		// do something that takes time
		time.Sleep(time.Duration(rand.Int31n(50)) * time.Millisecond)

		// create child span
		clientSpan := tracer.StartSpan("create_user_event", opentracing.ChildOf(serverSpan.Context()))
		defer clientSpan.Finish()

		go func() {
			// prep call to notification service
			url := "http://localhost:8001/notify"
			req, _ := http.NewRequest("GET", url, nil)
			// inject tracer
			tracer.Inject(clientSpan.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
			// make call to notification service
			http.DefaultClient.Do(req)
		}()

		// prep call to notification service
		url := "http://localhost:8003/account"
		req, _ := http.NewRequest("POST", url, nil)
		// inject tracer
		tracer.Inject(clientSpan.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
		// make call to notification service
		http.DefaultClient.Do(req)

		w.Write([]byte("user created"))
	}).Methods(http.MethodGet)

	http.ListenAndServe(":8002", router)
}

func initializeTracer() (io.Closer, error) {
	cfg :=jaegercfg.Configuration{
		ServiceName: "user",
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
	testSpan.SetTag("someTag", "someVal")

	// do more thing
	testSpan.Finish()
	// Set the singleton opentracing.Tracer with the Jaeger tracer.
	opentracing.SetGlobalTracer(tracer)
	return closer, err
}