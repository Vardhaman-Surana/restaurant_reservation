package middleware

import (
	"context"
	"github.com/vds/restaurant_reservation/management/pkg/models"
	"github.com/vds/restaurant_reservation/management/pkg/tracing"
	"net/http"
)

func AdminAccessOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, rq *http.Request) {
		prevContext:= rq.Context().Value("context")
		prevCtx := prevContext.(context.Context)
		span, newCtx := tracing.GetSpanFromContext(prevCtx, "check_for_admin_and_sadmin")
		defer span.Finish()
		tags := tracing.TraceTags{FuncName: "AdminAccessOnly", ServiceName: tracing.ServiceName, RequestID: span.BaggageItem("requestID")}
		tracing.SetTags(span, tags)

		value:= rq.Context().Value("userAuth")
		userAuth := value.(*models.UserAuth)
		if userAuth.Role != Admin && userAuth.Role != SuperAdmin {
			w.WriteHeader(http.StatusUnauthorized)
			//c.Abort()
			return
		}
		reqContext:=context.WithValue(rq.Context(),"context", newCtx)
		next.ServeHTTP(w,rq.WithContext(reqContext))
	})
}

func SuperAdminAccessOnly(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, rq *http.Request){

	prevContext:=rq.Context().Value("context")
	prevCtx:=prevContext.(context.Context)
	span,newCtx:=tracing.GetSpanFromContext(prevCtx,"check_for_sadmin")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"SuperAdminAccessOnly",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	value:=rq.Context().Value("userAuth")
	userAuth:=value.(*models.UserAuth)
	if userAuth.Role!=SuperAdmin{
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	reqContext:=context.WithValue(rq.Context(),"context", newCtx)
	next.ServeHTTP(w,rq.WithContext(reqContext))
	})
}