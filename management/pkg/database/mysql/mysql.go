package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/mysql"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/google/uuid"
	"github.com/vds/restaurant_reservation/management/pkg/database"
	"github.com/vds/restaurant_reservation/management/pkg/encryption"
	"github.com/vds/restaurant_reservation/management/pkg/middleware"
	"github.com/vds/restaurant_reservation/management/pkg/models"
	"github.com/vds/restaurant_reservation/management/pkg/tracing"
	"os"
	"strings"
	"time"

	"log"

)

const(

	SuperAdminTable="super_admins"
 	AdminTable="admins"
	OwnerTable="owners"
	InsertUser="insert into %s(id,email_id,name,password) values(?,?,?,?)"
	GetUserIDPassword="select id,password from %s where email_id=?"
	GetOwnersForSuperAdmin="select JSON_ARRAYAGG(JSON_OBJECT('id',id,'email',email_id,'name', name)) from owners"
	InsertOwner="insert into owners(id,email_id,name,password,creator_id) values(?,?,?,?,?)"
	OwnerUpdate="update owners set email_id=?,name=? where id=?"
	SelectRestaurantsForSuper="select JSON_ARRAYAGG(JSON_OBJECT('id',id,'name',name, 'lat',ROUND(lat,4),'lng',ROUND(lng,4))) from restaurants "
	SelectRestaurantsForAdmin="select JSON_ARRAYAGG(JSON_OBJECT('id',id,'name', name, 'lat',ROUND(lat,4),'lng',ROUND(lng,4))) from restaurants  where creator_id=?"
	SelectRestaurantsForOwner="select JSON_ARRAYAGG(JSON_OBJECT('id',id,'name', name, 'lat',ROUND(lat,4),'lng',ROUND(lng,4))) from restaurants  where owner_id=?"
	InsertRestaurant="insert into restaurants(name,lat,lng,creator_id) values(?,?,?,?)"
	RestaurantUpdate="update restaurants set name=?,lat=?,lng=? where id=?"
	CheckRestaurantOwner="select owner_id from restaurants where id=?"
	CheckRestaurantCreator="select creator_id from restaurants where id=?"
	CheckRestaurantDish="select res_id from dishes where id=?"

	DeleteOwnerBySuperAdmin="delete from owners where id=?"
	DeleteOwnerByAdmin="delete from owners where id=? and creator_id=?"

	DeleteRestaurantsBySuperAdmin="delete from restaurants where id=?"
	DeleteRestaurantsByAdmin="delete from restaurants where id=? and creator_id=?"

	DeleteDishes="delete from dishes where id=?"

)

type MySqlDB struct{
	*sql.DB
}
func NewMySqlDB(dbURL string)(*MySqlDB,error){
	//for docker servicename:3306 e.g. go-resman_database_1:3306
	/*serverName := "localhost:3306"
	user := "root"
	password := "password"*/


	connectionString := dbURL
	db, err := sql.Open("mysql", connectionString)
	if !strings.Contains(dbURL,"restaurant_test") {
		err = migrateDatabase(db)
		if err != nil {
			fmt.Print(err)
			return nil, err
		}
	}
	mySqlDB:=&MySqlDB{db}
	return mySqlDB,err
}

func(db *MySqlDB)ShowNearBy(ctx context.Context,location *models.Location)(context.Context,string,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_show_nearby_restaurants")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"ShowNearBy",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	var result string
	rows,err:=db.Query("select JSON_ARRAYAGG(JSON_OBJECT('name',name)) from restaurants where ST_Distance_Sphere(point(lat,lng),point(?,?))/1000 < 10",location.Lat,location.Lng)
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,"",database.ErrInternal
	}
	rows.Next()
	rows.Scan(&result)
	return newCtx,result,nil
}


func (db *MySqlDB)CreateUser(ctx context.Context,user *models.UserReg)(context.Context,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_create_new_user")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"CreateReservation",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	var tableName string
	switch user.Role{
	case middleware.Admin:
		tableName=AdminTable
	case middleware.SuperAdmin:
		tableName=SuperAdminTable
	}
	genHashCtx,pass,err:=encryption.GenerateHash(newCtx,user.Password)
	if err!=nil{
		fmt.Printf("%v",err)
		return genHashCtx,database.ErrInternal
	}
	id:=uuid.New().String()
	_,err=db.Exec(fmt.Sprintf(InsertUser,tableName),id,user.Email,user.Name,pass)
	if err!=nil{
		fmt.Printf("%v",err)
		return genHashCtx,database.ErrDupEmail
	}
	return genHashCtx,nil
}

func (db *MySqlDB)LogInUser(ctx context.Context,cred *models.Credentials)(context.Context,string,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_login_user")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"LogInUser",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	var tableName string
	switch cred.Role{
	case middleware.Admin:
		tableName=AdminTable
	case middleware.SuperAdmin:
		tableName=SuperAdminTable
	case middleware.Owner:
		tableName=OwnerTable
	}
	var id string
	var pass string
	rows,err:=db.Query(fmt.Sprintf(GetUserIDPassword,tableName),cred.Email)
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,"",database.ErrInternal
	}
	defer rows.Close()
	rows.Next()
	err=rows.Scan(&id,&pass)
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,"",database.ErrInvalidCredentials
	}
	comparePassCtx,isValid:=encryption.ComparePasswords(newCtx,pass,cred.Password)
	if !isValid{
		return comparePassCtx,"",database.ErrInvalidCredentials
	}
	return comparePassCtx,id,nil
}

func (db *MySqlDB)ShowOwners(ctx context.Context,userAuth *models.UserAuth)(context.Context,string,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_show_owners")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"ShowOwners",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)
	if userAuth.Role==middleware.SuperAdmin{
		return showOwnersForSuperAdmin(newCtx,db)
	}else if userAuth.Role==middleware.Admin{
		return showOwnersForAdmin(newCtx,db,userAuth.ID)
	}
	return newCtx,"",database.ErrInternal
}

func (db *MySqlDB)CreateOwner(ctx context.Context,creatorID string,owner *models.OwnerReg)(context.Context,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_create_owner")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"CreateOwner",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	genHashCtx,pass,err:=encryption.GenerateHash(newCtx,owner.Password)
	if err!=nil{
		fmt.Printf("%v",err)
		return genHashCtx,database.ErrInternal
	}
	id:=uuid.New().String()
	_,err=db.Exec(InsertOwner,id,owner.Email,owner.Name,pass,creatorID)
	if err!=nil {
		log.Printf("%v", err)
		return genHashCtx,database.ErrDupEmail
	}
	return genHashCtx,nil
}


func(db *MySqlDB)ShowAdmins(ctx context.Context)(context.Context,string,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_show_admins")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"ShowAdmins",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	var result sql.NullString
	rows,err:=db.Query("select JSON_ARRAYAGG(JSON_OBJECT('id',id,'email',email_id,'name', name)) from admins")
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,"",database.ErrInternal
	}
	defer rows.Close()
	rows.Next()
	err=rows.Scan(&result)
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,"",database.ErrInternal
	}
	return newCtx,result.String,nil
}

func(db *MySqlDB)CheckAdmin(ctx context.Context,adminID string)(context.Context,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_check_if_admin_exist")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"CheckAdmin",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	var count int
	rows,err:=db.Query("select count(*) from admins where id=?",adminID)
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,database.ErrInternal
	}
	defer rows.Close()
	rows.Next()
	err=rows.Scan(&count)
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,database.ErrInternal
	}
	if count!=1{
		return newCtx,database.ErrInternal
	}
	return newCtx,nil
}
func(db *MySqlDB)UpdateAdmin(ctx context.Context,admin *models.UserOutput)(context.Context,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_update_admin")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"UpdateAdmin",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	_,err:=db.Exec("update admins set email_id=?,name=? where id=?",admin.Email,admin.Name,admin.ID)
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,database.ErrInternal
	}
	return newCtx,nil
}
func(db *MySqlDB)RemoveAdmins(ctx context.Context,adminIDs...string)(context.Context,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_delete_admins")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"RemoveAdmins",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	var ErrEntries []int
	stmt, err := db.Prepare("delete from admins where id=?")
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,database.ErrInternal
	}
	for	i,id :=range adminIDs{
		result,err:=stmt.Exec(id)
		if err!=nil{
			log.Printf("%v",err)
			return newCtx,database.ErrInternal
		}
		numDeletedRows,_:=result.RowsAffected()
		if numDeletedRows==0{
			ErrEntries= append(ErrEntries,i)
		}
	}
	length:=len(ErrEntries)
	if length!=0{
		return sendErrorMessage(newCtx,ErrEntries,length,"Admins")
	}
	return newCtx,nil
}







func(db *MySqlDB)CheckOwnerCreator(ctx context.Context,creatorID string,ownerID string)(context.Context,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_verify_owner_creator")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"CheckOwnerCreator",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	var creatorIDOut string
	rows,err:=db.Query("select creator_id from owners where id=?",ownerID)
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,database.ErrInternal
	}
	rows.Next()
	err=rows.Scan(&creatorIDOut)
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,database.ErrInvalidOwner
	}
	if creatorIDOut!=creatorID{
		return newCtx,database.ErrInvalidOwnerCreator
	}
	return newCtx,nil
}

func(db *MySqlDB)UpdateOwner(ctx context.Context,owner *models.UserOutput)(context.Context,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"update_owner")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"UpdateOwner",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	chkOwnCtx,isValidOwnerID:=CheckOwnerID(newCtx,db,owner.ID)
	if !isValidOwnerID{
		return chkOwnCtx,database.ErrInvalidOwner
	}
	_,err:=db.Exec(OwnerUpdate,owner.Email,owner.Name,owner.ID)
	if err!=nil{
		log.Printf("%v",err)
		return chkOwnCtx,database.ErrDupEmail
	}
	return chkOwnCtx,nil
}


func(db *MySqlDB)RemoveOwners(ctx context.Context,userAuth *models.UserAuth,ownerIDs...string)(context.Context,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_delete_owners")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"RemoveOwners",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	switch userAuth.Role{
	case middleware.SuperAdmin:
		return removeOwnersBySuperAdmin(newCtx,db,ownerIDs...)
	case middleware.Admin:
		return removeOwnersByAdmin(newCtx,db,userAuth.ID,ownerIDs...)
	}
	return newCtx,database.ErrInternal
}



//restaurants

func(db *MySqlDB)ShowRestaurants(ctx context.Context,userAuth *models.UserAuth)(context.Context,string,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_get_restaurants")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"ShowRestaurants",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	switch userAuth.Role{
	case middleware.SuperAdmin:
		return showRestaurantsForSuper(newCtx,db)
	case middleware.Admin:
		return showRestaurantsForAdmin(newCtx,db,userAuth.ID)
	case middleware.Owner:
		return showRestaurantsForOwner(newCtx,db,userAuth.ID)
	}
	return newCtx,"",database.ErrInternal
}

func(db *MySqlDB)InsertRestaurant(ctx context.Context,restaurant *models.Restaurant)(context.Context,int,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_add_restaurant")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"InsertRestaurant",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	var id int
	_,err:=db.Exec(InsertRestaurant,restaurant.Name,restaurant.Lat,restaurant.Lng,restaurant.CreatorID)
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,0,database.ErrInternal
	}
	rows,err:=db.Query("select max(id) from restaurants")
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,0,database.ErrInternal
	}
	defer rows.Close()
	rows.Next()
	err=rows.Scan(&id)
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,0,database.ErrInternal
	}
	return newCtx,id,nil
}

func(db *MySqlDB)CheckRestaurantCreator(ctx context.Context,creatorID string,resID int)(context.Context,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_check_restaurant_creator")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"CheckRestaurantCreator",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	var creatorIDOut string
	rows,err:=db.Query(CheckRestaurantCreator,resID)
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,database.ErrInternal
	}
	rows.Next()
	err=rows.Scan(&creatorIDOut)
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,database.ErrNonExistingRestaurant
	}
	if creatorIDOut!=creatorID{
		return newCtx,database.ErrInvalidRestaurantCreator
	}
	return newCtx,nil
}

func(db *MySqlDB)UpdateRestaurant(ctx context.Context,restaurant *models.RestaurantOutput)(context.Context,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_updateRestaurant")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"UpdateRestaurant",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	chkResCtx,isValidRestaurant:=CheckRestaurantID(newCtx,db,restaurant.ID)
	if !isValidRestaurant{
		return chkResCtx,database.ErrNonExistingRestaurant
	}
	var err error
	_,err=db.Exec(RestaurantUpdate,restaurant.Name,restaurant.Lat,restaurant.Lng,restaurant.ID)
	if err!=nil{
		log.Printf("%v",err)
		return chkResCtx,database.ErrInternal
	}
	return chkResCtx,nil
}

func(db *MySqlDB)RemoveRestaurants(ctx context.Context,userAuth *models.UserAuth,resIDs...int)(context.Context,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_deleteRestaurants")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"RemoveRestaurants",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	switch userAuth.Role{
	case middleware.SuperAdmin:
		return removeRestaurantsBySuperAdmin(newCtx,db,resIDs...)
	case middleware.Admin:
		return removeRestaurantsByAdmin(newCtx,db,userAuth.ID,resIDs...)
	}
	return newCtx,database.ErrInternal
}

func(db *MySqlDB)ShowAvailableRestaurants(ctx context.Context,userAuth *models.UserAuth)(context.Context,string,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_showAvailableRes")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"ShowAvailableRestaurants",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	switch userAuth.Role{
	case middleware.SuperAdmin:
		return showAvailableRestaurantsForSuper(newCtx,db)
	case middleware.Admin:
		return showAvailableRestaurantsForAdmin(newCtx,db,userAuth.ID)
	}
	return newCtx,"",database.ErrInternal
}

func(db *MySqlDB)InsertOwnerForRestaurants(ctx context.Context,userAuth *models.UserAuth,ownerID string,resIDs...int)(context.Context,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_assignOwnerToRestaurants")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"InsertOwnerForRestaurants",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	chkOwnCtx,isValidOwnerID:=CheckOwnerID(newCtx,db,ownerID)
	if !isValidOwnerID{
		return chkOwnCtx,database.ErrInvalidOwner
	}
	var ErrEntries []int
	stmt, err := db.Prepare("update restaurants set owner_id=? where id=?")
	if err!=nil{
		log.Printf("%v",err)
		return chkOwnCtx,database.ErrInternal
	}
	var chkResCtrCtx context.Context
	for	i,id :=range resIDs{
		if userAuth.Role!=middleware.SuperAdmin {
			chkResCtrCtx,err=db.CheckRestaurantCreator(chkOwnCtx,userAuth.ID, id)
			if err!=nil{
				ErrEntries= append(ErrEntries,i)
				continue
			}
		}
		_,err:=stmt.Exec(ownerID,id)
		if err!=nil{
			log.Printf("%v",err)
			return chkResCtrCtx,database.ErrInternal
		}
	}
	length:=len(ErrEntries)
	fmt.Print(ErrEntries)
	if length!=0{
		return sendErrorMessage(chkResCtrCtx,ErrEntries,length,"Restaurants")
	}
	return chkResCtrCtx,nil
}




//menu
func(db *MySqlDB)CheckRestaurantOwner(ctx context.Context,ownerID string,resID int)(context.Context,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_checkResOwner")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"CheckRestaurantOwner",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	var ownerIDOut string
	rows,err:=db.Query(CheckRestaurantOwner,resID)
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,database.ErrInternal
	}
	rows.Next()
	err=rows.Scan(&ownerIDOut)
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,database.ErrNonExistingRestaurant
	}
	if ownerIDOut!=ownerID{
		return newCtx,database.ErrInvalidRestaurantOwner
	}
	return newCtx,nil
}
func(db *MySqlDB)ShowMenu(ctx context.Context,resID int)(context.Context,string,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_getMenu")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"ShowMenu",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	chkResCtx,isValidRestaurant:=CheckRestaurantID(newCtx,db,resID)
	if !isValidRestaurant{
		return chkResCtx,"",database.ErrNonExistingRestaurant
	}
	var result sql.NullString
	rows,err:=db.Query("select JSON_ARRAYAGG(JSON_OBJECT('id',id,'name',name,'price',price)) from dishes where res_id=?",resID)
	if err!=nil{
		log.Printf("%v",err)
		return chkResCtx,"",database.ErrInternal
	}
	rows.Next()
	rows.Scan(&result)
	return chkResCtx,result.String,nil
}
func(db *MySqlDB)InsertDishes(ctx context.Context,dishes []models.Dish,resID int)(context.Context,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_insertDishes")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"InsertDishes",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	stmt,err:=db.Prepare("insert into dishes(res_id,name,price) values(?,?,?)")
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,database.ErrInternal
	}
	for _,dish:=range dishes{
		_,err=stmt.Exec(resID,dish.Name,dish.Price)
		if err!=nil{
			log.Printf("%v",err)
			return newCtx,database.ErrNonExistingRestaurant
		}
	}
	return newCtx,nil
}
func(db *MySqlDB)UpdateDish(ctx context.Context,dish *models.DishOutput)(context.Context,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_updateDish")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"UpdateDish",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	fmt.Printf("dishe is +%v",dish)
	_,err:=db.Exec("update dishes set name=?,price=? where id=?",dish.Name,dish.Price,dish.ID)
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,database.ErrInternal
	}
	return newCtx,nil
}
func(db *MySqlDB)CheckRestaurantDish(ctx context.Context,resID int,dishID int)(context.Context,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_checkResDish")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"CheckRestaurantDish",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	var resIDOut int
	rows,err:=db.Query(CheckRestaurantDish,dishID)
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,database.ErrInternal
	}
	rows.Next()
	err=rows.Scan(&resIDOut)
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,database.ErrInvalidDish
	}
	if resIDOut!=resID{
		return newCtx,database.ErrInvalidRestaurantDish
	}
	return newCtx,nil
}

func(db *MySqlDB)RemoveDishes(ctx context.Context,dishIDs...int)(context.Context,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_deleteDishes")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"RemoveDishes",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	var ErrEntries []int
	stmt, err := db.Prepare(DeleteDishes)
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,database.ErrInternal
	}
	for	i,id :=range dishIDs{
		result,err:=stmt.Exec(id)
		if err!=nil{
			log.Printf("%v",err)
			return newCtx,database.ErrInternal
		}
		numDeletedRows,_:=result.RowsAffected()
		if numDeletedRows==0{
			ErrEntries= append(ErrEntries,i)
		}
	}
	length:=len(ErrEntries)
	if length!=0{
		return sendErrorMessage(newCtx,ErrEntries,length,"Dishes")
	}
	return newCtx,nil
}
















//helpers
func showOwnersForSuperAdmin(ctx context.Context,db *MySqlDB)(context.Context,string,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_show_owners_sadmin")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"showOwnersForSuperAdmin",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	var result string
	rows,err:=db.Query(GetOwnersForSuperAdmin)
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,"",database.ErrInternal
	}
	rows.Next()
	err=rows.Scan(&result)
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,"",database.ErrInternal
	}
	return newCtx,result,nil
}
func showOwnersForAdmin(ctx context.Context,db *MySqlDB,creatorID string)(context.Context,string,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_show_owners_admin")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"showOwnersForAdmin",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)
	var result sql.NullString
	rows,err:=db.Query("select JSON_ARRAYAGG(JSON_OBJECT('id',id,'email',email_id,'name', name)) from owners where creator_id=?",creatorID)
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,"",database.ErrInternal
	}
	if rows.Next(){
		err=rows.Scan(&result)
	}
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,"",database.ErrInternal
	}
	return newCtx,result.String,nil
}

func showRestaurantsForSuper(ctx context.Context,db *MySqlDB)(context.Context,string,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_showResForSA")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"showRestaurantsForSuper",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	var result sql.NullString
	rows,err:=db.Query(SelectRestaurantsForSuper)
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,"",database.ErrInternal
	}
	defer rows.Close()

	rows.Next()
	rows.Scan(&result)
	return newCtx,result.String,nil
}
func showRestaurantsForAdmin(ctx context.Context,db *MySqlDB,adminID string)(context.Context,string,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_showResForAdmin")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"showRestaurantsForAdmin",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	var result sql.NullString
	rows,err:=db.Query(SelectRestaurantsForAdmin,adminID)
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,"",database.ErrInternal
	}
	defer rows.Close()

	rows.Next()
	rows.Scan(&result)
	return newCtx,result.String,nil
}
func showRestaurantsForOwner(ctx context.Context,db *MySqlDB,ownerID string)(context.Context,string,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_showOwnerRestaurants")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"showRestaurantsForOwner",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	var result sql.NullString
	rows,err:=db.Query(SelectRestaurantsForOwner,ownerID)
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,"",database.ErrInternal
	}
	defer rows.Close()

	rows.Next()
	rows.Scan(&result)
	return newCtx,result.String,nil
}


func showAvailableRestaurantsForSuper(ctx context.Context,db *MySqlDB)(context.Context,string,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_showAvlResSA")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"showAvailableRestaurantsForSuper",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	var result string
	rows,err:=db.Query("select JSON_ARRAYAGG(JSON_OBJECT('id',id,'name',name, 'lat',ROUND(lat,4),'lng',ROUND(lng,4))) from restaurants where owner_id IS NULL")
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,"",database.ErrInternal
	}
	defer rows.Close()

	rows.Next()
	rows.Scan(&result)
	return newCtx,result,nil
}
func showAvailableRestaurantsForAdmin(ctx context.Context,db *MySqlDB,creatorID string)(context.Context,string,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_showAvlResAdmin")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"showAvailableRestaurantsForAdmin",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	var result string
	fmt.Print(creatorID)
	rows,err:=db.Query("select JSON_ARRAYAGG(JSON_OBJECT('id',id,'name',name, 'lat',ROUND(lat,4),'lng',ROUND(lng,4))) from restaurants where owner_id IS NULL and creator_id=?",creatorID)
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,"",database.ErrInternal
	}
	defer rows.Close()

	rows.Next()
	rows.Scan(&result)
	return newCtx,result,nil
}

func removeOwnersBySuperAdmin(ctx context.Context,db *MySqlDB,ownerIDs...string)(context.Context,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_removeOwnersBySA")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"removeOwnersBySuperAdmin",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	var ErrEntries []int
	stmt, err := db.Prepare(DeleteOwnerBySuperAdmin)
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,database.ErrInternal
	}
	for	i,id :=range ownerIDs{
		result,err:=stmt.Exec(id)
		if err!=nil{
			log.Printf("%v",err)
			return newCtx,database.ErrInternal
		}
		numDeletedRows,_:=result.RowsAffected()
		if numDeletedRows==0{
			ErrEntries= append(ErrEntries,i)
		}else{
			_, _ = db.Query("Update Restaurants set owner_id=null where owner_id=?",id)
		}
	}
	length:=len(ErrEntries)
	if length!=0{
		return sendErrorMessage(newCtx,ErrEntries,length,"Owners")
	}
	return newCtx,nil
}
func removeOwnersByAdmin(ctx context.Context,db *MySqlDB,creatorID string,ownerIDs...string)(context.Context,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_removeOwnerByAdmin")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"removeOwnersByAdmin",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	var ErrEntries []int
	stmt, err := db.Prepare(DeleteOwnerByAdmin)
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,database.ErrInternal
	}
	for	i,id :=range ownerIDs{
		result,err:=stmt.Exec(id,creatorID)
		if err!=nil{
			log.Printf("%v",err)
			return newCtx,database.ErrInternal
		}
		numDeletedRows,_:=result.RowsAffected()
		if numDeletedRows==0{
			ErrEntries= append(ErrEntries,i)
		}else{
			_, _ = db.Query("Update Restaurants set owner_id=null where owner_id=?",id)
		}
	}
	length:=len(ErrEntries)
	if length!=0{
		return sendErrorMessage(newCtx,ErrEntries,length,"Owners")
	}
	return newCtx,nil
}
func sendErrorMessage(ctx context.Context,ErrEntries []int,length int,data string)(context.Context,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"sendErrorMessage")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"sendErrorMessage",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	errMsg:=data+" Deleted Except entry no."
	for i,j :=range ErrEntries{
		if i==length-1{
			errMsg=errMsg+fmt.Sprintf(" %v",j+1)
			break
		}
		errMsg=errMsg+fmt.Sprintf(" %v,",j+1)
	}
	return newCtx,errors.New(errMsg)
}


func removeRestaurantsBySuperAdmin(ctx context.Context,db *MySqlDB,resIDs...int)(context.Context,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"removeResBySA")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"removeRestaurantsBySuperAdmin",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	var ErrEntries []int
	stmt, err := db.Prepare(DeleteRestaurantsBySuperAdmin)
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,database.ErrInternal
	}
	for	i,id :=range resIDs{
		fmt.Printf("id is %v",id)
		result,err:=stmt.Exec(id)
		if err!=nil{
			log.Printf("%v",err)
			return newCtx,database.ErrInternal
		}
		numDeletedRows,_:=result.RowsAffected()
		if numDeletedRows==0{
			ErrEntries= append(ErrEntries,i)
		}
	}
	length:=len(ErrEntries)
	if length!=0{
		return sendErrorMessage(newCtx,ErrEntries,length,"Restaurants")
	}
	return newCtx,nil
}
func removeRestaurantsByAdmin(ctx context.Context,db *MySqlDB,creatorID string,resIDs...int)(context.Context,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_deleteResByAdmin")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"removeRestaurantsByAdmin",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	var ErrEntries []int
	stmt, err := db.Prepare(DeleteRestaurantsByAdmin)
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,database.ErrInternal
	}
	for	i,id :=range resIDs{
		result,err:=stmt.Exec(id,creatorID)
		if err!=nil{
			log.Printf("%v",err)
			return newCtx,database.ErrInternal
		}
		numDeletedRows,_:=result.RowsAffected()
		if numDeletedRows==0{
			ErrEntries= append(ErrEntries,i)
		}
	}
	length:=len(ErrEntries)
	fmt.Print(ErrEntries)
	if length!=0{
		return sendErrorMessage(newCtx,ErrEntries,length,"Restaurants")
	}
	return newCtx,nil
}

func(db *MySqlDB)StoreToken(ctx context.Context,token string)(context.Context,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_storeToken")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"StoreToken",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	_,err:=db.Exec("insert into invalid_tokens(token) values(?)",token)
	if err!=nil{
		fmt.Print(err)
		return newCtx,database.ErrInternal
	}
	return newCtx,nil
}
func(db *MySqlDB)VerifyToken(ctx context.Context,tokenIn string)(context.Context,bool){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_tokenVerification")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"VerifyToken",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	var tokenOut string
	rows,err:=db.Query("select token from invalid_tokens where token=?",tokenIn)
	if err!=nil{
		fmt.Print(err)
		return newCtx,false
	}
	defer rows.Close()
	rows.Next()
	err=rows.Scan(&tokenOut)
	if err!=nil{
		return newCtx,true
	}
	return newCtx,false
}




func CheckOwnerID(ctx context.Context,db *MySqlDB,ownerID string)(context.Context,bool){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_checkExistOwner")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"CheckOwnerID",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	var count int
	rows,err:=db.Query("select count(*) from owners where id=?",ownerID)
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,false
	}
	defer rows.Close()
	rows.Next()
	err=rows.Scan(&count)
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,false
	}
	if count!=1{
		return newCtx,false
	}
	return newCtx,true

}
func CheckRestaurantID(ctx context.Context,db *MySqlDB,resID int)(context.Context,bool){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"db_checkExistRestaurant")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"CheckRestaurantID",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	var count int
	rows,err:=db.Query("select count(*) from restaurants where id=?",resID)
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,false
	}
	defer rows.Close()
	rows.Next()
	err=rows.Scan(&count)
	if err!=nil{
		log.Printf("%v",err)
		return newCtx,false
	}
	if count!=1{
		return newCtx,false
	}
	return newCtx,true
}

func (db *MySqlDB)DeleteExpiredToken(ctx context.Context,token string,t time.Duration){
	span,_:=tracing.GetSpanFromContext(ctx,"delete_expire_token")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"DeleteExpiredToken",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)

	time.Sleep(t)
	_,err:=db.Exec("delete from invalid_tokens where token=?",token)
	if err!=nil{
		log.Fatalf("Error in deleting the invalid token:%v",err)
	}
}





////////////////////
//Database migration


func migrateDatabase(db *sql.DB) error {
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return err
	}

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	migration, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s/database", dir),
		"restaurant",
		driver,
	)
	if err != nil {
		return err
	}

	migration.Log = &models.MigrationLogger{}

	migration.Log.Printf("Applying database migrations")
	err = migration.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	version, _, err := migration.Version()
	if err != nil {
		return err
	}

	migration.Log.Printf("Active database version: %d", version)

	return nil
}


