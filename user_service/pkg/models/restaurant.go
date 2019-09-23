package models

type RestaurantOutput struct{
	ID int `json:"id" db:"id"`
	Name string `json:"name" binding:"required" db:"name"`
	Lat float64 `json:"lat" db:"lat"`
	Lng	float64 `json:"lng" db:"lng"`
}