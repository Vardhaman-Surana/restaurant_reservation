package database

import "context"

type Database interface {
	CreateTablesForRestaurant(ctx context.Context,resID int,numTables int)(context.Context,error)

	GetNumAvailableTables(ctx context.Context,resID int,startTime int64)(ctx2 context.Context,numTables int,err error)
	CreateReservation(ctx context.Context,resID int,startTime int64,userID string)(ctx2 context.Context,resvID int,err error)

	//go function
	MarkReservationAsDeleted(ctx context.Context)
}
