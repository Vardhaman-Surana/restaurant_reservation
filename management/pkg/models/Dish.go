package models

type DishOutput struct{
	ID int `json:"id" binding:"required"`
	Name string `json:"name" binding:"required"`
	Price float32 `json:"price" binding:"required"`
}
type Dish struct{
	Name string `json:"name" binding:"required"`
	Price float32 `json:"price" binding:"required"`
}
