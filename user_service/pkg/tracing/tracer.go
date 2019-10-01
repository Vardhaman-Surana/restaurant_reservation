package tracing

import (
	"context"
	"fmt"
	"github.com/fatih/structs"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"io"
)

type TraceTags struct{
	RequestID string
	FuncName string
	ServiceName string
}


const ServiceName = "userSvc"

func NewTracer(service string) (opentracing.Tracer, io.Closer) {
	cfg := &config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans: true,
		},
	}
	tracer, closer, err := cfg.New(service, config.Logger(jaeger.StdLogger))
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}
	return tracer, closer
}

func GetSpan(tracer opentracing.Tracer,operationName string)opentracing.Span{
	span:=tracer.StartSpan(operationName)
	return span
}
func GetSpanFromContext(ctx context.Context,operationName string)(opentracing.Span,context.Context) {
	span,newCtx:=opentracing.StartSpanFromContext(ctx,operationName)
	return span,newCtx
}

func SetTags(span opentracing.Span,tags TraceTags){
	tagMap:=structs.Map(tags)
	for k,v:=range tagMap{
		span.SetTag(k,v)
	}
}

/*
func SetBaggageItems(span opentracing.Span,baggageMap map[string]string){
	for k,v:=range baggageMap{
		span.SetBaggageItem(k,v)
	}
}
*/





