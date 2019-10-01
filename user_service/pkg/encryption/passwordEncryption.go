package encryption

import (
	"context"
	"errors"
	"github.com/vds/restaurant_reservation/user_service/pkg/tracing"
	"golang.org/x/crypto/bcrypt"
)

var errGenHash = errors.New("error in generating hash for email id")


func GenerateHash(value string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(value), bcrypt.DefaultCost)
	if err != nil {
		return "", errGenHash
	}
	return string(hash), nil
}
func IsCorrectPassword(ctx context.Context,phash ,pass string)(context.Context,bool){
	span, newCtx :=tracing.GetSpanFromContext(ctx,"match_password")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"IsCorrectPassword",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	err:=bcrypt.CompareHashAndPassword([]byte(phash),[]byte(pass))
	if err!=nil{
		return newCtx,false
	}
	return newCtx,true
}