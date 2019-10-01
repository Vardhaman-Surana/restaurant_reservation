package middleware

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/vds/restaurant_reservation/reservation_service/pkg/tracing"
)

func InitTrace(tracer opentracing.Tracer) gin.HandlerFunc{
	return func(c *gin.Context){
		opentracing.SetGlobalTracer(tracer)
		span:= tracer.StartSpan("initializing_trace")
		defer span.Finish()
		tags:=tracing.TraceTags{FuncName:"InitTrace",ServiceName:tracing.ServiceName}
		tracing.SetTags(span,tags)
		currSpanCtx := context.Background()
		currSpanCtx=opentracing.ContextWithSpan(currSpanCtx,span)
		c.Set("context",currSpanCtx)
		c.Next()
	}
}