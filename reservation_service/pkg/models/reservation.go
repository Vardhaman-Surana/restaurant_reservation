package models

import (
	"gopkg.in/gorp.v1"
	"time"
)

const ReservationTableName = "Reservations"

type Reservation struct{
	BaseModel
	UserID string `json:"userID" db:"User_ID"`
	ResID int	`json:"restaurantID" db:"Restaurant_ID"`
	TableID int `json:"tableID" db:"Table_ID"`
	StartTime int64 `json:"from" db:"Start_Time"`
}

func (r *Reservation) PreInsert(s gorp.SqlExecutor) error {
	r.Created=time.Now().Unix()
	r.Updated=r.Created
	return nil
}