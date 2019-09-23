package models

type UserReg struct{
	Role string `json:"role" binding:"required"`
	Email string `json:"email"  binding:"required"`
	Name string `json:"name"  binding:"required"`
	Password string	`json:"password"  binding:"required"`
}
type Credentials struct{
	Role string `json:"role" binding:"required"`
	Email string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}
type UserAuth struct{
	ID string `json:"id" binding:"required"`
	Role string `json:"role" binding:"required"`
}
type UserOutput struct{
	ID string `json:"id"`
	Email string `json:"email" binding:"required"`
	Name string `json:"name" binding:"required"`
}
type OwnerReg struct{
	Email string `json:"email" binding:"required"`
	Name string `json:"name" binding:"required"`
	Password string	`json:"password"  binding:"required"`
}


/*type LoginInput struct {
	Role string
	Email string
	Password string
}

type User struct {
	ID int
	name string
	email string
	password string
}

func deleteUsers(ids ...int) {

}*/