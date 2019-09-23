package mysql

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/rubenv/sql-migrate"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/vds/restaurant_reservation/user_service/pkg/migrations"
	"github.com/vds/restaurant_reservation/user_service/pkg/models"
	"gopkg.in/gorp.v1"
	"log"
	"strings"
	"time"
)

type MysqlDbMap struct{
	DbMap *gorp.DbMap
}

func NewMysqlDbMap(DbURL string)(*MysqlDbMap,error){
	dbMap, err := DBForURL(DbURL)
	if err!=nil{
		return nil,err
	}
	MysqlDbMap:=MysqlDbMap{DbMap:dbMap}
	return &MysqlDbMap,nil
}
// initiating the db instance and running the migrations
func DBForURL(url string)(*gorp.DbMap,error){
	log.Printf("\nCreating DB with url %s \n",url)
	//getting a sql db instance
	db,err:=sql.Open("mysql",url)
	if err!=nil{
		log.Printf("\nUnable To get the db instance:%v\n",err)
		return nil,err
	}
	if !strings.Contains(url,"restaurant_test") {
		_, err = migrate.Exec(db, "mysql", migrations.GetAll(), migrate.Up)
		if err != nil {
			return nil, err
		}
	}
	//setting up db map
	dbMap := &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{Engine: "InnoDB", Encoding: "UTF"}}
	dbMap.AddTableWithName(models.User{}, models.UserTableName).SetKeys(false, "ID").SetKeys(false, "Email")

	return dbMap,nil
}


func (mdb *MysqlDbMap)GetUser(email string)(*models.User,error){
	var userOutput models.User
	err:=mdb.DbMap.SelectOne(&userOutput,`SELECT * FROM users WHERE Email=?`,email)
	if err!=nil{
		return nil,err
	}
	return &userOutput,err
}
func (mdb *MysqlDbMap)GetRestaurants()([]models.RestaurantOutput,error){
	var restaurants []models.RestaurantOutput
	_,err:=mdb.DbMap.Select(&restaurants,"SELECT id,name,lat,lng from restaurants")
	if err!=nil{
		return nil,err
	}
	return restaurants,nil
}

func (mdb *MysqlDbMap)InsertUser(user *models.User)error{
	err:=mdb.DbMap.Insert(user)
	if err!=nil{
		return err
	}
	return nil
}

func(db *MysqlDbMap)StoreToken(token string)error{
	_,err:=db.DbMap.Exec("insert into invalid_tokens(token) values(?)",token)
	if err!=nil{
		log.Println(err)
		return err
	}
	return nil
}
func (db *MysqlDbMap)DeleteExpiredToken(token string,t time.Duration){
	time.Sleep(t)
	_,err:=db.DbMap.Exec("delete from invalid_tokens where token=?",token)
	if err!=nil{
		log.Printf("Error in deleting the invalid token:%v",err)
	}
}
func(db *MysqlDbMap)VerifyToken(tokenIn string)bool{
	var tokenOut string
	err:=db.DbMap.SelectOne(&tokenOut,"select token from invalid_tokens where token=?",tokenIn)
	if err!=nil{
		log.Println(err)
		return false
	}
	if tokenOut==""{
		return true
	}
	return false
}