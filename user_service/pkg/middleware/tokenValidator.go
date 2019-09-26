package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/vds/restaurant_reservation/user_service/pkg/database"
	"github.com/vds/restaurant_reservation/user_service/pkg/tracing"
	"net/http"
	"time"
)

func TokenValidator(tracer opentracing.Tracer,db database.Database) gin.HandlerFunc{
	return func(c *gin.Context){
		opentracing.SetGlobalTracer(tracer)
		span, newCtx := opentracing.StartSpanFromContext(c, "user_authentication")
		span.SetBaggageItem("requestID", uuid.New().String())
		span.SetBaggageItem("requestUrl",c.Request.URL.String())
		span.SetTag("funcName","AuthMiddleware")
		span.SetTag("serviceName",tracing.ServiceName)
		span.SetTag("startTime",time.Now().String())
		defer span.Finish()

		tokenStr:=c.Request.Header.Get("token")
		isValid:=db.VerifyToken(newCtx,tokenStr)
		if !isValid{
			c.AbortWithStatus(http.StatusUnauthorized)
			c.Abort()
		}
		c.Set("context",newCtx)
		c.Next()
	}
}
