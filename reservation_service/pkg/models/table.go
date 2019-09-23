package models

import (
	"gopkg.in/gorp.v1"
	"time"
)

const RestaurantTablesDBTable= "restaurant_tables"

type Table struct{
	BaseModel
	ResID int `json:"restaurantID" db:"Restaurant_ID"`
}

func (t *Table) PreInsert(s gorp.SqlExecutor) error {
	t.Created = time.Now().Unix()
	t.Updated = t.Created
	return nil
}