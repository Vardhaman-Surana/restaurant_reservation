package database

type Database interface {
	CreateTablesForRestaurant(resID int,numTables int)error

	GetNumAvailableTables(resID int,startTime int64)(numTables int,err error)
	CreateReservation(resID int,startTime int64,userID string)(resvID int,err error)

	MarkReservationAsDeleted()
}
