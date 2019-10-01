package jwtTokenGenerate

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"github.com/vds/restaurant_reservation/user_service/pkg/models"
	"github.com/vds/restaurant_reservation/user_service/pkg/tracing"
	"time"
)

const ExpireDuration=120*time.Minute


func CreateToken(ctx context.Context,claims *models.Claims) (context.Context,string,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"generate_token")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"CreateToken",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	jwtKey:=[]byte("SecretKey")
	expirationTime:=time.Now().Add(ExpireDuration).Unix()
	claims.ExpiresAt=expirationTime
	//remember to change it later
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err!=nil{
		return newCtx,"",err
	}
	return newCtx,tokenString,nil
}
