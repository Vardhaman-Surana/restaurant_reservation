package database

import (
	"context"
	"github.com/vds/restaurant_reservation/user_service/pkg/models"
	"time"
)

type Database interface{
	GetUser(ctx context.Context,email string)(context.Context,*models.User,error)
	SelectRestaurants(ctx context.Context)(context.Context,[]models.RestaurantOutput,error)
	CreateUser(ctx context.Context,user *models.User)(context.Context,error)

	StoreToken(ctx context.Context,token string)(context.Context,error)
	VerifyToken(ctx context.Context,token string)(context.Context,bool)

	// go func
	DeleteExpiredToken(ctx context.Context,token string,t time.Duration)

}
