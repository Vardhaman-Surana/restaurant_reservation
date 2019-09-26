package database

import (
	"context"
	"github.com/vds/restaurant_reservation/user_service/pkg/models"
	"time"
)

type Database interface{
	GetUser(ctx context.Context,email string)(*models.User,error)
	GetRestaurants(ctx context.Context)([]models.RestaurantOutput,error)
	InsertUser(ctx context.Context,user *models.User)error

	StoreToken(ctx context.Context,token string)error
	DeleteExpiredToken(ctx context.Context,token string,t time.Duration)
	VerifyToken(ctx context.Context,token string)bool

}
