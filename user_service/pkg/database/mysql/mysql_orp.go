package mysql

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	logot "github.com/opentracing/opentracing-go/log"
	_ "github.com/rubenv/sql-migrate"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/vds/restaurant_reservation/user_service/pkg/migrations"
	"github.com/vds/restaurant_reservation/user_service/pkg/models"
	"github.com/vds/restaurant_reservation/user_service/pkg/tracing"
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


func (mdb *MysqlDbMap)GetUser(ctx context.Context,email string)(context.Context,*models.User,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"get_user_from_db")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"GetUser",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)
	var userOutput models.User
	err:=mdb.DbMap.SelectOne(&userOutput,`SELECT * FROM users WHERE Email=?`,email)
	if err!=nil{
		return newCtx,nil,err
	}
	return newCtx,&userOutput,err
}
func (mdb *MysqlDbMap)SelectRestaurants(ctx context.Context)(context.Context,[]models.RestaurantOutput,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"get_restaurants_db")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"SelectRestaurants",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)
	var restaurants []models.RestaurantOutput
	_,err:=mdb.DbMap.Select(&restaurants,"SELECT id,name,lat,lng from restaurants")
	if err!=nil{
		return newCtx,nil,err
	}
	return newCtx,restaurants,nil
}

func (mdb *MysqlDbMap)CreateUser(ctx context.Context,user *models.User)(context.Context,error){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"InsertUserDb")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"CreateUser",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)
	err:=mdb.DbMap.Insert(user)
	if err!=nil{
		return newCtx,err
	}
	return newCtx,nil
}

func(db *MysqlDbMap)StoreToken(ctx context.Context,token string)(context.Context,error){
	span, newCtx := tracing.GetSpanFromContext(ctx, "db_insert_logout_token")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"StoreToken",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)
	_,err:=db.DbMap.Exec("insert into invalid_tokens(token) values(?)",token)
	if err!=nil{
		log.Println(err)
		return newCtx,err
	}
	return newCtx,nil
}
func (db *MysqlDbMap)DeleteExpiredToken(ctx context.Context,token string,t time.Duration){
	span, _ := tracing.GetSpanFromContext(ctx, "db_delete_expire_token")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"DeleteExpiredToken",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)
	time.Sleep(t)
	_,err:=db.DbMap.Exec("delete from invalid_tokens where token=?",token)
	if err!=nil{
		log.Printf("Error in deleting the invalid token:%v",err)
	}
}
func(db *MysqlDbMap)VerifyToken(ctx context.Context,tokenIn string)(context.Context,bool){
	span,newCtx:=tracing.GetSpanFromContext(ctx,"checkInvalidToken")
	defer span.Finish()
	tags:=tracing.TraceTags{FuncName:"VerifyToken",ServiceName:tracing.ServiceName,RequestID:span.BaggageItem("requestID")}
	tracing.SetTags(span,tags)
	count,err:=db.DbMap.SelectInt(fmt.Sprintf(`select count(*) from invalid_tokens where token="%s"`,tokenIn))
	if err!=nil{
		log.Println(err)
		span.LogFields(
			logot.String("error",err.Error()),
			)
		return newCtx,false
	}
	if count==0{
		return newCtx,true
	}
	return newCtx,false
}