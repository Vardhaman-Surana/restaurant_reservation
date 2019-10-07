package controller_test

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/mux"
	"github.com/vds/restaurant_reservation/management/pkg/database/mysql"
	"github.com/vds/restaurant_reservation/management/pkg/server"
	"sync"
)

const TestDbURL=`root:password@tcp(localhost)/restaurant_test?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=true&multiStatements=true`


var (
	globalDB *mysql.MySqlDB
	OnceDB sync.Once
	OnceRouter sync.Once
	router *gin.Engine
)
func InitDB()(*mysql.MySqlDB,error){
	var err error
	OnceDB.Do(func(){
		db,er:=mysql.NewMySqlDB(TestDbURL)
		globalDB = db
		err=er
	})
	return globalDB,err
}

func GetRouter() (*mux.Router,error){
	var err error
	OnceRouter.Do(func(){
		initRouter,er:=server.NewRouter(globalDB,nil)
		router = initRouter.Create()
		err=er
	})
	return router,err
}