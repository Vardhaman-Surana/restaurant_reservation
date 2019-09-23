package database

import (
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
	ShowNearBy(location *models.Location)(string,error)

	CreateUser(user *models.UserReg)error
	LogInUser(cred *models.Credentials)(string,error)
	ShowAdmins()(string,error)
	CheckAdmin(adminID string)error
	UpdateAdmin(admin *models.UserOutput)error
	RemoveAdmins(adminIDs...string)error

	ShowOwners(userAuth *models.UserAuth)(string,error)
	CreateOwner(creatorID string,owner *models.OwnerReg)error
	CheckOwnerCreator(creatorID string,ownerID string)error
	UpdateOwner(owner *models.UserOutput)error
	RemoveOwners(userAuth *models.UserAuth,ownerIDs...string)error


	ShowRestaurants(userAuth *models.UserAuth)(string,error)
	InsertRestaurant(restaurant *models.Restaurant)(int,error)
	ShowAvailableRestaurants(userAuth *models.UserAuth)(string,error)
	InsertOwnerForRestaurants(userAuth *models.UserAuth,ownerID string,resIDs...int)error
	CheckRestaurantCreator(creatorID string,resID int)error
	UpdateRestaurant(restaurant *models.RestaurantOutput)error
	RemoveRestaurants(userAuth *models.UserAuth,resIDs...int)error

	ShowMenu(resID int)(string,error)
	CheckRestaurantOwner(ownerID string,resID int)error
	InsertDishes(dishes []models.Dish,resID int)error
	UpdateDish(dish *models.DishOutput)error
	CheckRestaurantDish(resID int,dishID int)error
	RemoveDishes(dishIDs...int)error

	StoreToken(token string)error
	VerifyToken(token string)bool
	DeleteExpiredToken(token string,t time.Duration)
}