package mysql

import (
	"database/sql"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/rubenv/sql-migrate"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/vds/restaurant_reservation/reservation_service/pkg/controller"
	"github.com/vds/restaurant_reservation/reservation_service/pkg/migrations"
	"github.com/vds/restaurant_reservation/reservation_service/pkg/models"
	"gopkg.in/gorp.v1"
	"log"
	"strings"
	"sync"
	"time"
)

var(
	m=sync.Mutex{}
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
	log.Printf("Creating DB with url %s ",url)
	//getting a sql db instance
	db,err:=sql.Open("mysql",url)
	if err!=nil{
		log.Printf("Unable To get the db instance:%v",err)
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
	dbMap.AddTableWithName(models.Table{}, models.RestaurantTablesDBTable).SetKeys(true, "ID")
	dbMap.AddTableWithName(models.Reservation{}, models.ReservationTableName).SetKeys(true, "ID")

	return dbMap,nil
}

func(mdb *MysqlDbMap) CreateTablesForRestaurant(resID int,numTables int)error{
	trans, err := mdb.DbMap.Begin()
	if err != nil {
		return err
	}
	for i:=0;i < numTables;i++{
		var restaurantTable  models.Table
		restaurantTable.ResID=resID
		err:=trans.Insert(&restaurantTable)
		if err!=nil{
			log.Printf("error in inserting the table %v",err)
			er:=trans.Rollback()
			if er!=nil{
				log.Printf("Error in rolling back transaction:%v",err)
			}
			return err
		}
	}
	err=trans.Commit()
	if err!=nil{
		log.Printf("error in commiting the transaction:%v",err)
		return err
	}
	return nil
}

func(mdb *MysqlDbMap)GetNumAvailableTables(resID int,startTime int64)(numTables int,err error) {
	err=mdb.DbMap.SelectOne(&numTables,`SELECT COUNT(ID) FROM restaurant_tables WHERE Restaurant_ID=? AND ID NOT IN (SELECT Table_ID FROM Reservations WHERE Restaurant_ID=? AND ABS(Start_Time-?)<3600 AND Deleted<>1)`,resID,resID,startTime)
	if err!=nil{
		log.Printf("\nerror in selectin the number of tables : %v\n",err)
		return 0,err
	}
	return numTables,err
}

func(mdb *MysqlDbMap)CreateReservation(resID int,startTime int64,userID string)(resvID int,err error){
	m.Lock()
	numTables,err:=mdb.GetNumAvailableTables(resID,startTime)
	if err!=nil{
		m.Unlock()
		return resvID,err
	}
	if numTables==0{
		m.Unlock()
		return  resvID,errors.New(controller.ReservationNotAvailableMessage)
	}
	var tableIdToReserve int
	err=mdb.DbMap.SelectOne(&tableIdToReserve,	`SELECT ID FROM restaurant_tables WHERE Restaurant_ID=?  AND ID NOT IN (SELECT Table_ID FROM Reservations WHERE Restaurant_ID=? AND ABS(Start_Time-?)<3600 AND Deleted<>1) limit 1`,resID,resID,startTime)
	if err!=nil{
		log.Printf("\nError finding a table to reserve:%v\n",err)
		m.Unlock()
		return resvID,err
	}
	reservation:=models.Reservation{StartTime:startTime,ResID:resID,TableID:tableIdToReserve,UserID:userID}
	err=mdb.DbMap.Insert(&reservation)
	if err!=nil{
		log.Printf("\nerror inserting the reservation instance:%v\n",err)
		m.Unlock()
		return resvID,err
	}
	err=mdb.DbMap.SelectOne(&resvID,"SELECT MAX(ID) FROM Reservations")
	if err!=nil{
		log.Printf("\nerror retrieving the reservation id:%v\n",err)
		m.Unlock()
		return resvID,err
	}
	m.Unlock()
	return resvID,nil
}

func(mdb *MysqlDbMap)MarkReservationAsDeleted(){
	currentTime:=time.Now().Unix()
	_,err:=mdb.DbMap.Exec("Update reservations set Updated=?, Deleted=1 where (Start_Time+3600)<=?",currentTime,currentTime)
	if err!=nil{
		log.Printf("Can not mark reservations as deleted")
		return
	}
}



















