package encryption

import (
	"errors"
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
func IsCorrectPassword(phash ,pass string)bool{
	err:=bcrypt.CompareHashAndPassword([]byte(phash),[]byte(pass))
	if err!=nil{
		return false
	}
	return true
}