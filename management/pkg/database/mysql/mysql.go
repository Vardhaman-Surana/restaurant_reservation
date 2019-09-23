package mysql

import (
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
	"os"
	"strings"
	"time"

	"log"

)

const(
	restaurantDB="restaurant"

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

func(db *MySqlDB)ShowNearBy(location *models.Location)(string,error){
	var result string
	rows,err:=db.Query("select JSON_ARRAYAGG(JSON_OBJECT('name',name)) from restaurants where ST_Distance_Sphere(point(lat,lng),point(?,?))/1000 < 10",location.Lat,location.Lng)
	if err!=nil{
		log.Printf("%v",err)
		return "",database.ErrInternal
	}
	rows.Next()
	rows.Scan(&result)
	return result,nil
}


func (db *MySqlDB)CreateUser(user *models.UserReg) error{
	var tableName string
	switch user.Role{
	case middleware.Admin:
		tableName=AdminTable
	case middleware.SuperAdmin:
		tableName=SuperAdminTable
	}
	pass,err:=encryption.GenerateHash(user.Password)
	if err!=nil{
		fmt.Printf("%v",err)
		return database.ErrInternal
	}
	id:=uuid.New().String()
	_,err=db.Exec(fmt.Sprintf(InsertUser,tableName),id,user.Email,user.Name,pass)
	if err!=nil{
		fmt.Printf("%v",err)
		return database.ErrDupEmail
	}
	return nil
}

func (db *MySqlDB)LogInUser(cred *models.Credentials)(string,error){
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
		return "",database.ErrInternal
	}
	defer rows.Close()
	rows.Next()
	err=rows.Scan(&id,&pass)
	if err!=nil{
		log.Printf("%v",err)
		return "",database.ErrInvalidCredentials
	}
	isValid:=encryption.ComparePasswords(pass,cred.Password)
	if !isValid{
		return "",database.ErrInvalidCredentials
	}
	return id,nil
}

func (db *MySqlDB)ShowOwners(userAuth *models.UserAuth)(string,error){
	if userAuth.Role==middleware.SuperAdmin{
		return showOwnersForSuperAdmin(db)
	}else if userAuth.Role==middleware.Admin{
		return showOwnersForAdmin(db,userAuth.ID)
	}
	return "",database.ErrInternal
}

func (db *MySqlDB)CreateOwner(creatorID string,owner *models.OwnerReg)error{
	pass,err:=encryption.GenerateHash(owner.Password)
	if err!=nil{
		fmt.Printf("%v",err)
		return database.ErrInternal
	}
	id:=uuid.New().String()
	_,err=db.Exec(InsertOwner,id,owner.Email,owner.Name,pass,creatorID)
	if err!=nil {
		log.Printf("%v", err)
		return database.ErrDupEmail
	}
	return nil
}


func(db *MySqlDB)ShowAdmins()(string,error){
	var result sql.NullString
	rows,err:=db.Query("select JSON_ARRAYAGG(JSON_OBJECT('id',id,'email',email_id,'name', name)) from admins")
	if err!=nil{
		log.Printf("%v",err)
		return "",database.ErrInternal
	}
	defer rows.Close()
	rows.Next()
	err=rows.Scan(&result)
	if err!=nil{
		log.Printf("%v",err)
		return "",database.ErrInternal
	}
	return result.String,nil
}

func(db *MySqlDB)CheckAdmin(adminID string)error{
	var count int
	rows,err:=db.Query("select count(*) from admins where id=?",adminID)
	if err!=nil{
		log.Printf("%v",err)
		return database.ErrInternal
	}
	defer rows.Close()
	rows.Next()
	err=rows.Scan(&count)
	if err!=nil{
		log.Printf("%v",err)
		return database.ErrInternal
	}
	if count!=1{
		return database.ErrInternal
	}
	return nil
}
func(db *MySqlDB)UpdateAdmin(admin *models.UserOutput)error{
	_,err:=db.Exec("update admins set email_id=?,name=? where id=?",admin.Email,admin.Name,admin.ID)
	if err!=nil{
		log.Printf("%v",err)
		return database.ErrInternal
	}
	return nil
}
func(db *MySqlDB)RemoveAdmins(adminIDs...string)error{
	var ErrEntries []int
	stmt, err := db.Prepare("delete from admins where id=?")
	if err!=nil{
		log.Printf("%v",err)
		return database.ErrInternal
	}
	for	i,id :=range adminIDs{
		result,err:=stmt.Exec(id)
		if err!=nil{
			log.Printf("%v",err)
			return database.ErrInternal
		}
		numDeletedRows,_:=result.RowsAffected()
		if numDeletedRows==0{
			ErrEntries= append(ErrEntries,i)
		}
	}
	length:=len(ErrEntries)
	if length!=0{
		return sendErrorMessage(ErrEntries,length,"Admins")
	}
	return nil
}







func(db *MySqlDB)CheckOwnerCreator(creatorID string,ownerID string)error{
	var creatorIDOut string
	rows,err:=db.Query("select creator_id from owners where id=?",ownerID)
	if err!=nil{
		log.Printf("%v",err)
		return database.ErrInternal
	}
	rows.Next()
	err=rows.Scan(&creatorIDOut)
	if err!=nil{
		log.Printf("%v",err)
		return database.ErrInvalidOwner
	}
	if creatorIDOut!=creatorID{
		return database.ErrInvalidOwnerCreator
	}
	return nil
}

func(db *MySqlDB)UpdateOwner(owner *models.UserOutput)error{
	isValidOwnerID:=CheckOwnerID(db,owner.ID)
	if !isValidOwnerID{
		return database.ErrInvalidOwner
	}
	_,err:=db.Exec(OwnerUpdate,owner.Email,owner.Name,owner.ID)
	if err!=nil{
		log.Printf("%v",err)
		return database.ErrDupEmail
	}
	return nil
}


func(db *MySqlDB)RemoveOwners(userAuth *models.UserAuth,ownerIDs...string)error{
	switch userAuth.Role{
	case middleware.SuperAdmin:
		return removeOwnersBySuperAdmin(db,ownerIDs...)
	case middleware.Admin:
		return removeOwnersByAdmin(db,userAuth.ID,ownerIDs...)
	}
	return database.ErrInternal
}



//restaurants

func(db *MySqlDB)ShowRestaurants(userAuth *models.UserAuth)(string,error){
	switch userAuth.Role{
	case middleware.SuperAdmin:
		return showRestaurantsForSuper(db)
	case middleware.Admin:
		return showRestaurantsForAdmin(db,userAuth.ID)
	case middleware.Owner:
		return showRestaurantsForOwner(db,userAuth.ID)
	}
	return "",database.ErrInternal
}

func(db *MySqlDB)InsertRestaurant(restaurant *models.Restaurant)(int,error){
	var id int
	_,err:=db.Exec(InsertRestaurant,restaurant.Name,restaurant.Lat,restaurant.Lng,restaurant.CreatorID)
	if err!=nil{
		log.Printf("%v",err)
		return 0,database.ErrInternal
	}
	rows,err:=db.Query("select max(id) from restaurants")
	if err!=nil{
		log.Printf("%v",err)
		return 0,database.ErrInternal
	}
	defer rows.Close()
	rows.Next()
	err=rows.Scan(&id)
	if err!=nil{
		log.Printf("%v",err)
		return 0,database.ErrInternal
	}
	return id,nil
}

func(db *MySqlDB)CheckRestaurantCreator(creatorID string,resID int)error{
	var creatorIDOut string
	rows,err:=db.Query(CheckRestaurantCreator,resID)
	if err!=nil{
		log.Printf("%v",err)
		return database.ErrInternal
	}
	rows.Next()
	err=rows.Scan(&creatorIDOut)
	if err!=nil{
		log.Printf("%v",err)
		return database.ErrNonExistingRestaurant
	}
	if creatorIDOut!=creatorID{
		return database.ErrInvalidRestaurantCreator
	}
	return nil
}

func(db *MySqlDB)UpdateRestaurant(restaurant *models.RestaurantOutput)error{
	isValidRestaurant:=CheckRestaurantID(db,restaurant.ID)
	if !isValidRestaurant{
		return database.ErrNonExistingRestaurant
	}
	var err error
	_,err=db.Exec(RestaurantUpdate,restaurant.Name,restaurant.Lat,restaurant.Lng,restaurant.ID)
	if err!=nil{
		log.Printf("%v",err)
		return database.ErrInternal
	}
	return nil
}

func(db *MySqlDB)RemoveRestaurants(userAuth *models.UserAuth,resIDs...int)error{
	switch userAuth.Role{
	case middleware.SuperAdmin:
		return removeRestaurantsBySuperAdmin(db,resIDs...)
	case middleware.Admin:
		return removeRestaurantsByAdmin(db,userAuth.ID,resIDs...)
	}
	return database.ErrInternal
}

func(db *MySqlDB)ShowAvailableRestaurants(userAuth *models.UserAuth)(string,error){
	switch userAuth.Role{
	case middleware.SuperAdmin:
		return showAvailableRestaurantsForSuper(db)
	case middleware.Admin:
		return showAvailableRestaurantsForAdmin(db,userAuth.ID)
	}
	return "",database.ErrInternal
}

func(db *MySqlDB)InsertOwnerForRestaurants(userAuth *models.UserAuth,ownerID string,resIDs...int)error{
	isValidOwnerID:=CheckOwnerID(db,ownerID)
	if !isValidOwnerID{
		return database.ErrInvalidOwner
	}
	var ErrEntries []int
	stmt, err := db.Prepare("update restaurants set owner_id=? where id=?")
	if err!=nil{
		log.Printf("%v",err)
		return database.ErrInternal
	}
	for	i,id :=range resIDs{
		if userAuth.Role!=middleware.SuperAdmin {
			err=db.CheckRestaurantCreator(userAuth.ID, id)
			if err!=nil{
				ErrEntries= append(ErrEntries,i)
				continue
			}
		}
		_,err:=stmt.Exec(ownerID,id)
		if err!=nil{
			log.Printf("%v",err)
			return database.ErrInternal
		}
	}
	length:=len(ErrEntries)
	fmt.Print(ErrEntries)
	if length!=0{
		return sendErrorMessage(ErrEntries,length,"Restaurants")
	}
	return nil
}




//menu
func(db *MySqlDB)CheckRestaurantOwner(ownerID string,resID int)error{
	var ownerIDOut string
	rows,err:=db.Query(CheckRestaurantOwner,resID)
	if err!=nil{
		log.Printf("%v",err)
		return database.ErrInternal
	}
	rows.Next()
	err=rows.Scan(&ownerIDOut)
	if err!=nil{
		log.Printf("%v",err)
		return database.ErrNonExistingRestaurant
	}
	if ownerIDOut!=ownerID{
		return database.ErrInvalidRestaurantOwner
	}
	return nil
}
func(db *MySqlDB)ShowMenu(resID int)(string,error){
	isValidRestaurant:=CheckRestaurantID(db,resID)
	if !isValidRestaurant{
		return "",database.ErrNonExistingRestaurant
	}
	var result sql.NullString
	rows,err:=db.Query("select JSON_ARRAYAGG(JSON_OBJECT('id',id,'name',name,'price',price)) from dishes where res_id=?",resID)
	if err!=nil{
		log.Printf("%v",err)
		return "",database.ErrInternal
	}
	rows.Next()
	rows.Scan(&result)
	return result.String,nil
}
func(db *MySqlDB)InsertDishes(dishes []models.Dish,resID int)error{
	stmt,err:=db.Prepare("insert into dishes(res_id,name,price) values(?,?,?)")
	if err!=nil{
		log.Printf("%v",err)
		return database.ErrInternal
	}
	for _,dish:=range dishes{
		_,err=stmt.Exec(resID,dish.Name,dish.Price)
		if err!=nil{
			log.Printf("%v",err)
			return database.ErrNonExistingRestaurant
		}
	}
	return nil
}
func(db *MySqlDB)UpdateDish(dish *models.DishOutput)error{
	fmt.Printf("dishe is +%v",dish)
	_,err:=db.Exec("update dishes set name=?,price=? where id=?",dish.Name,dish.Price,dish.ID)
	if err!=nil{
		log.Printf("%v",err)
		return database.ErrInternal
	}
	return nil
}
func(db *MySqlDB)CheckRestaurantDish(resID int,dishID int)error{
	var resIDOut int
	rows,err:=db.Query(CheckRestaurantDish,dishID)
	if err!=nil{
		log.Printf("%v",err)
		return database.ErrInternal
	}
	rows.Next()
	err=rows.Scan(&resIDOut)
	if err!=nil{
		log.Printf("%v",err)
		return database.ErrInvalidDish
	}
	if resIDOut!=resID{
		return database.ErrInvalidRestaurantDish
	}
	return nil
}

func(db *MySqlDB)RemoveDishes(dishIDs...int)error{
	var ErrEntries []int
	stmt, err := db.Prepare(DeleteDishes)
	if err!=nil{
		log.Printf("%v",err)
		return database.ErrInternal
	}
	for	i,id :=range dishIDs{
		result,err:=stmt.Exec(id)
		if err!=nil{
			log.Printf("%v",err)
			return database.ErrInternal
		}
		numDeletedRows,_:=result.RowsAffected()
		if numDeletedRows==0{
			ErrEntries= append(ErrEntries,i)
		}
	}
	length:=len(ErrEntries)
	if length!=0{
		return sendErrorMessage(ErrEntries,length,"Dishes")
	}
	return nil
}
















//helpers
func showOwnersForSuperAdmin(db *MySqlDB)(string,error){
	var result string
	rows,err:=db.Query(GetOwnersForSuperAdmin)
	if err!=nil{
		log.Printf("%v",err)
		return "",database.ErrInternal
	}
	rows.Next()
	err=rows.Scan(&result)
	if err!=nil{
		log.Printf("%v",err)
		return "",database.ErrInternal
	}
	return result,nil
}
func showOwnersForAdmin(db *MySqlDB,creatorID string)(string,error){
	var result sql.NullString
	rows,err:=db.Query("select JSON_ARRAYAGG(JSON_OBJECT('id',id,'email',email_id,'name', name)) from owners where creator_id=?",creatorID)
	if err!=nil{
		log.Printf("%v",err)
		return "",database.ErrInternal
	}
	if rows.Next(){
		err=rows.Scan(&result)
	}
	if err!=nil{
		log.Printf("%v",err)
		return "",database.ErrInternal
	}
	return result.String,nil
}

func showRestaurantsForSuper(db *MySqlDB)(string,error){
	var result sql.NullString
	rows,err:=db.Query(SelectRestaurantsForSuper)
	if err!=nil{
		log.Printf("%v",err)
		return "",database.ErrInternal
	}
	defer rows.Close()

	rows.Next()
	rows.Scan(&result)
	return result.String,nil
}
func showRestaurantsForAdmin(db *MySqlDB,adminID string)(string,error){
	var result sql.NullString
	rows,err:=db.Query(SelectRestaurantsForAdmin,adminID)
	if err!=nil{
		log.Printf("%v",err)
		return "",database.ErrInternal
	}
	defer rows.Close()

	rows.Next()
	rows.Scan(&result)
	return result.String,nil
}
func showRestaurantsForOwner(db *MySqlDB,ownerID string)(string,error){
	var result sql.NullString
	rows,err:=db.Query(SelectRestaurantsForOwner,ownerID)
	if err!=nil{
		log.Printf("%v",err)
		return "",database.ErrInternal
	}
	defer rows.Close()

	rows.Next()
	rows.Scan(&result)
	return result.String,nil
}


func showAvailableRestaurantsForSuper(db *MySqlDB)(string,error){
	var result string
	rows,err:=db.Query("select JSON_ARRAYAGG(JSON_OBJECT('id',id,'name',name, 'lat',ROUND(lat,4),'lng',ROUND(lng,4))) from restaurants where owner_id IS NULL")
	if err!=nil{
		log.Printf("%v",err)
		return "",database.ErrInternal
	}
	defer rows.Close()

	rows.Next()
	rows.Scan(&result)
	return result,nil
}
func showAvailableRestaurantsForAdmin(db *MySqlDB,creatorID string)(string,error){
	var result string
	fmt.Print(creatorID)
	rows,err:=db.Query("select JSON_ARRAYAGG(JSON_OBJECT('id',id,'name',name, 'lat',ROUND(lat,4),'lng',ROUND(lng,4))) from restaurants where owner_id IS NULL and creator_id=?",creatorID)
	if err!=nil{
		log.Printf("%v",err)
		return "",database.ErrInternal
	}
	defer rows.Close()

	rows.Next()
	rows.Scan(&result)
	return result,nil
}

func removeOwnersBySuperAdmin(db *MySqlDB,ownerIDs...string)error{
	var ErrEntries []int
	stmt, err := db.Prepare(DeleteOwnerBySuperAdmin)
	if err!=nil{
		log.Printf("%v",err)
		return database.ErrInternal
	}
	for	i,id :=range ownerIDs{
		result,err:=stmt.Exec(id)
		if err!=nil{
			log.Printf("%v",err)
			return database.ErrInternal
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
		return sendErrorMessage(ErrEntries,length,"Owners")
	}
	return nil
}
func removeOwnersByAdmin(db *MySqlDB,creatorID string,ownerIDs...string)error{
	var ErrEntries []int
	stmt, err := db.Prepare(DeleteOwnerByAdmin)
	if err!=nil{
		log.Printf("%v",err)
		return database.ErrInternal
	}
	for	i,id :=range ownerIDs{
		result,err:=stmt.Exec(id,creatorID)
		if err!=nil{
			log.Printf("%v",err)
			return database.ErrInternal
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
		return sendErrorMessage(ErrEntries,length,"Owners")
	}
	return nil
}
func sendErrorMessage(ErrEntries []int,length int,data string)error{
	errMsg:=data+" Deleted Except entry no."
	for i,j :=range ErrEntries{
		if i==length-1{
			errMsg=errMsg+fmt.Sprintf(" %v",j+1)
			break
		}
		errMsg=errMsg+fmt.Sprintf(" %v,",j+1)
	}
	return errors.New(errMsg)
}


func removeRestaurantsBySuperAdmin(db *MySqlDB,resIDs...int)error{
	var ErrEntries []int
	stmt, err := db.Prepare(DeleteRestaurantsBySuperAdmin)
	if err!=nil{
		log.Printf("%v",err)
		return database.ErrInternal
	}
	for	i,id :=range resIDs{
		fmt.Printf("id is %v",id)
		result,err:=stmt.Exec(id)
		if err!=nil{
			log.Printf("%v",err)
			return database.ErrInternal
		}
		numDeletedRows,_:=result.RowsAffected()
		if numDeletedRows==0{
			ErrEntries= append(ErrEntries,i)
		}
	}
	length:=len(ErrEntries)
	if length!=0{
		return sendErrorMessage(ErrEntries,length,"Restaurants")
	}
	return nil
}
func removeRestaurantsByAdmin(db *MySqlDB,creatorID string,resIDs...int)error{
	var ErrEntries []int
	stmt, err := db.Prepare(DeleteRestaurantsByAdmin)
	if err!=nil{
		log.Printf("%v",err)
		return database.ErrInternal
	}
	for	i,id :=range resIDs{
		result,err:=stmt.Exec(id,creatorID)
		if err!=nil{
			log.Printf("%v",err)
			return database.ErrInternal
		}
		numDeletedRows,_:=result.RowsAffected()
		if numDeletedRows==0{
			ErrEntries= append(ErrEntries,i)
		}
	}
	length:=len(ErrEntries)
	fmt.Print(ErrEntries)
	if length!=0{
		return sendErrorMessage(ErrEntries,length,"Restaurants")
	}
	return nil
}

func(db *MySqlDB)StoreToken(token string)error{
	_,err:=db.Exec("insert into invalid_tokens(token) values(?)",token)
	if err!=nil{
		fmt.Print(err)
		return database.ErrInternal
	}
	return nil
}
func(db *MySqlDB)VerifyToken(tokenIn string)bool{
	var tokenOut string
	rows,err:=db.Query("select token from invalid_tokens where token=?",tokenIn)
	if err!=nil{
		fmt.Print(err)
		return false
	}
	defer rows.Close()
	rows.Next()
	err=rows.Scan(&tokenOut)
	if err!=nil{
		return true
	}
	return false
}




func CheckOwnerID(db *MySqlDB,ownerID string)bool{
	var count int
	rows,err:=db.Query("select count(*) from owners where id=?",ownerID)
	if err!=nil{
		log.Printf("%v",err)
		return false
	}
	defer rows.Close()
	rows.Next()
	err=rows.Scan(&count)
	if err!=nil{
		log.Printf("%v",err)
		return false
	}
	if count!=1{
		return false
	}
	return true

}
func CheckRestaurantID(db *MySqlDB,resID int)bool{
	var count int
	rows,err:=db.Query("select count(*) from restaurants where id=?",resID)
	if err!=nil{
		log.Printf("%v",err)
		return false
	}
	defer rows.Close()
	rows.Next()
	err=rows.Scan(&count)
	if err!=nil{
		log.Printf("%v",err)
		return false
	}
	if count!=1{
		return false
	}
	return true
}

func (db *MySqlDB)DeleteExpiredToken(token string,t time.Duration){
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


