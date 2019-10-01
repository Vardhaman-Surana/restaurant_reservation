package middleware

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/vds/restaurant_reservation/user_service/pkg/tracing"
)

func InitTrace(tracer opentracing.Tracer) gin.HandlerFunc{
	return func(c *gin.Context){
		opentracing.SetGlobalTracer(tracer)
		reqID:=uuid.New().String()
		span:= tracer.StartSpan("initializing_trace")
		span.SetBaggageItem("requestID", reqID)

		defer span.Finish()
		tags:=tracing.TraceTags{FuncName:"InitTrace",ServiceName:tracing.ServiceName,RequestID:reqID}
		tracing.SetTags(span,tags)
		span.SetBaggageItem("requestID",reqID)
		currSpanCtx := context.Background()
		currSpanCtx=opentracing.ContextWithSpan(currSpanCtx,span)
		c.Set("context",currSpanCtx)
		c.Next()
	}
}