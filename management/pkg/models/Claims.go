package models

import "github.com/dgrijalva/jwt-go"

type Claims struct{
	ID string
	Role string
	jwt.StandardClaims
}