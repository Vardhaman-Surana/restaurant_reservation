package controller

import (
	"context"
	"github.com/vds/restaurant_reservation/management/pkg/database"
	"github.com/vds/restaurant_reservation/management/pkg/models"
	"github.com/vds/restaurant_reservation/management/pkg/tracing"
	"net/http"
	"time"
)

type HelloWorldController struct{
	database.Database
}

func NewHelloWorldController(db database.Database)*HelloWorldController{
	hc:=new(HelloWorldController)
	hc.Database=db
	return hc
}

func(h *HelloWorldController)SayHello(w http.ResponseWriter, rq *http.Request){
	prevContext:=rq.Context().Value("context")
	prevCtx:=prevContext.(context.Context)
	span,_:=tracing.GetSpanFromContext(prevCtx,"say_hello")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"SayHello",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)
	models.WriteToResponse(w,http.StatusOK,&models.DefaultMap{
		"msg":"Hello World",
		"time": time.Now().String(),
		"created_by" : "vardhaman",
		"completed":"23-08-2019",
	})
}