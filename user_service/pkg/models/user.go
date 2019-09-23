package models

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/vds/restaurant_reservation/user_service/pkg/encryption"
	"gopkg.in/gorp.v1"
	"time"
)

const UserTableName = "users"


type User struct{
	BaseModel
	Email string `json:"email"`
	Name string `json:"name"`
	Password  string `db:"-" json:"password,omitempty"`
	PasswordHash string `json:"-"`
}

func (u *User) PreInsert(s gorp.SqlExecutor) error {
	u.Created=time.Now().Unix()
	u.Updated=time.Now().Unix()
	u.ID=uuid.New().String()
	pass,err:=encryption.GenerateHash(u.Password)
	if err!=nil{
		fmt.Printf("%v",err)
		return err
	}
	u.PasswordHash=pass
	return nil
}