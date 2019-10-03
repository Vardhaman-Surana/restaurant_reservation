package middleware

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/opentracing/opentracing-go"
	"github.com/vds/restaurant_reservation/user_service/pkg/database"
	"github.com/vds/restaurant_reservation/user_service/pkg/tracing"
	"net/http"
)

func TokenValidator(db database.Database) mux.MiddlewareFunc{
	return func(next http.Handler) http.Handler{
		return http.HandlerFunc(func(w http.ResponseWriter, rq *http.Request){
		prevContext:=rq.Context().Value("context")
		prevCtx:=prevContext.(context.Context)
		span,newCtx:=opentracing.StartSpanFromContext(prevCtx,"token_validation")
		defer span.Finish()
		tags:=tracing.TraceTags{FuncName:"TokenValidator",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
		tracing.SetTags(span,tags)

		tokenStr:=rq.Header.Get("token")
		verifyTokenCtx,isValid:=db.VerifyToken(newCtx,tokenStr)
		if !isValid{
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		reqCtx:=context.WithValue(rq.Context(),"context",verifyTokenCtx)
		next.ServeHTTP(w,rq.WithContext(reqCtx))
	})}
}
