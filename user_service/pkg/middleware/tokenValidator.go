package middleware

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/vds/restaurant_reservation/user_service/pkg/database"
	"github.com/vds/restaurant_reservation/user_service/pkg/tracing"
	"net/http"
)

func TokenValidator(db database.Database) gin.HandlerFunc{
	return func(c *gin.Context){
		prevContext,_:=c.Get("context")
		prevCtx:=prevContext.(context.Context)
		span,newCtx:=opentracing.StartSpanFromContext(prevCtx,"token_validation")
		defer span.Finish()
		tags:=tracing.TraceTags{FuncName:"TokenValidator",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
		tracing.SetTags(span,tags)

		tokenStr:=c.Request.Header.Get("token")
		verifyTokenCtx,isValid:=db.VerifyToken(newCtx,tokenStr)
		if !isValid{
			c.AbortWithStatus(http.StatusUnauthorized)
			c.Abort()
		}
		c.Set("context",verifyTokenCtx)
		c.Next()
	}
}
