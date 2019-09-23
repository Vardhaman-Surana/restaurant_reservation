package jwtTokenGenerate

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/vds/restaurant_reservation/user_service/pkg/models"
	"time"
)

const ExpireDuration=120*time.Minute


func CreateToken(claims *models.Claims) (string,error){
	jwtKey:=[]byte("SecretKey")
	expirationTime:=time.Now().Add(ExpireDuration).Unix()
	claims.ExpiresAt=expirationTime
	//remember to change it later
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err!=nil{
		return "",err
	}
	return tokenString,nil
}
