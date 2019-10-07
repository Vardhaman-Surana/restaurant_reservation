package encryption

import (
	"context"
	"errors"
	"github.com/vds/restaurant_reservation/management/pkg/tracing"
	"golang.org/x/crypto/bcrypt"
)
//errors
var errGenHash = errors.New("error in generating hash for email id")


func GenerateHash(ctx context.Context,value string) (context.Context,string, error) {
	span,newCtx:=tracing.GetSpanFromContext(ctx,"gen_hash_pass")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"GenerateHash",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)
	hash, err := bcrypt.GenerateFromPassword([]byte(value), bcrypt.DefaultCost)
	if err != nil {
		return newCtx,"", errGenHash
	}
	return newCtx,string(hash), nil
}
func ComparePasswords(ctx context.Context,phash ,pass string)(context.Context,bool){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"check_password")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"ComparePasswords",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)
	err:=bcrypt.CompareHashAndPassword([]byte(phash),[]byte(pass))
	if err!=nil{
		return newCtx,false
	}
	return newCtx,true
}
