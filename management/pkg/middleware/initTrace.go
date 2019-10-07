package middleware

import (
	"context"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/opentracing/opentracing-go"
	"github.com/vds/restaurant_reservation/management/pkg/tracing"
	"net/http"
)


func InitTrace(tracer opentracing.Tracer) mux.MiddlewareFunc{
	return func(next http.Handler)http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, rq *http.Request) {

			opentracing.SetGlobalTracer(tracer)
			reqID := uuid.New().String()
			span := tracer.StartSpan("initializing_trace")
			defer span.Finish()
			tags := tracing.TraceTags{FuncName: "InitTrace", ServiceName: tracing.ServiceName, RequestID: reqID}
			tracing.SetTags(span, tags)
			span.SetBaggageItem("requestID", reqID)
			currSpanCtx := context.Background()
			currSpanCtx = opentracing.ContextWithSpan(currSpanCtx, span)

			reqContext:=context.WithValue(rq.Context(),"context",currSpanCtx)
			w.Header().Set("Content-Type","application/json")
			next.ServeHTTP(w, rq.WithContext(reqContext))
		})
	}
}