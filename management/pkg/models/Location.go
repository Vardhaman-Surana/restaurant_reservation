package models

type Location struct{
	Lat	float32 `json:"lat" binding:"required"`
	Lng float32 `json:"lng" binding:"required"`
}