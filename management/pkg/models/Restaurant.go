package models

type RestaurantOutput struct{
	ID int `json:"id"`
	Name string `json:"name" binding:"required"`
	Lat float64 `json:"lat"`
	Lng	float64 `json:"lng"`
}
type Restaurant struct{
	Name string `json:"name" binding:"required"`
	Lat float64 `json:"lat"`
	Lng	float64 `json:"lng"`
	NumTables int `json:"numTables"`
	CreatorID string
}
