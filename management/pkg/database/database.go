package database

import (
	"context"
	"errors"
	"github.com/vds/restaurant_reservation/management/pkg/models"
	"time"
)


var(
	ErrInternal = errors.New("internal server error")
	ErrDupEmail=errors.New("email already used try a different one")
	ErrInvalidCredentials = errors.New("incorrect login details")
	ErrInvalidOwner = errors.New("owner does not exist")
	ErrInvalidOwnerCreator=errors.New("can not update owner created by other admin")
	ErrInvalidRestaurantCreator=errors.New("can not update restaurant created by other admin")
	ErrNonExistingRestaurant=errors.New("restaurant does not exist")
	ErrInvalidRestaurantOwner=errors.New("can not update restaurant owned by others")
	ErrInvalidDish=errors.New("dish does not exist")
	ErrInvalidRestaurantDish=errors.New("can not update dish of other restaurant")
	)


type Database interface {
	ShowNearBy(ctx context.Context,location *models.Location)(context.Context,string,error)

	CreateUser(ctx context.Context,user *models.UserReg)(context.Context,error)
	LogInUser(ctx context.Context,cred *models.Credentials)(context.Context,string,error)
	ShowAdmins(ctx context.Context)(context.Context,string,error)
	CheckAdmin(ctx context.Context,adminID string)(context.Context,error)
	UpdateAdmin(ctx context.Context,admin *models.UserOutput)(context.Context,error)
	RemoveAdmins(ctx context.Context,adminIDs...string)(context.Context,error)

	ShowOwners(ctx context.Context,userAuth *models.UserAuth)(context.Context,string,error)
	CreateOwner(ctx context.Context,creatorID string,owner *models.OwnerReg)(context.Context,error)
	CheckOwnerCreator(ctx context.Context,creatorID string,ownerID string)(context.Context,error)
	UpdateOwner(ctx context.Context,owner *models.UserOutput)(context.Context,error)
	RemoveOwners(ctx context.Context,userAuth *models.UserAuth,ownerIDs...string)(context.Context,error)


	ShowRestaurants(ctx context.Context,userAuth *models.UserAuth)(context.Context,string,error)
	InsertRestaurant(ctx context.Context,restaurant *models.Restaurant)(context.Context,int,error)
	ShowAvailableRestaurants(ctx context.Context,userAuth *models.UserAuth)(context.Context,string,error)
	InsertOwnerForRestaurants(ctx context.Context,userAuth *models.UserAuth,ownerID string,resIDs...int)(context.Context,error)
	CheckRestaurantCreator(ctx context.Context,creatorID string,resID int)(context.Context,error)
	UpdateRestaurant(ctx context.Context,restaurant *models.RestaurantOutput)(context.Context,error)
	RemoveRestaurants(ctx context.Context,userAuth *models.UserAuth,resIDs...int)(context.Context,error)

	ShowMenu(ctx context.Context,resID int)(context.Context,string,error)
	CheckRestaurantOwner(ctx context.Context,ownerID string,resID int)(context.Context,error)
	InsertDishes(ctx context.Context,dishes []models.Dish,resID int)(context.Context,error)
	UpdateDish(ctx context.Context,dish *models.DishOutput)(context.Context,error)
	CheckRestaurantDish(ctx context.Context,resID int,dishID int)(context.Context,error)
	RemoveDishes(ctx context.Context,dishIDs...int)(context.Context,error)

	StoreToken(ctx context.Context,token string)(context.Context,error)
	VerifyToken(ctx context.Context,token string)(context.Context,bool)
	DeleteExpiredToken(ctx context.Context,token string,t time.Duration)
}