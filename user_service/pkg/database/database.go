package database

import (
	"github.com/vds/restaurant_reservation/user_service/pkg/models"
	"time"
)

type Database interface{
	GetUser(email string)(*models.User,error)
	GetRestaurants()([]models.RestaurantOutput,error)
	InsertUser(user *models.User)error

	StoreToken(token string)error
	DeleteExpiredToken(token string,t time.Duration)
	VerifyToken(token string)bool

}
