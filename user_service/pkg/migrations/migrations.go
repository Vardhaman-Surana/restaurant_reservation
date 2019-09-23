package migrations

import (
	migrate "github.com/rubenv/sql-migrate"
	"sync"
)

type dbMigrations struct{
	m sync.Mutex
	migrations []*migrate.Migration
}

var migrationInstance=&dbMigrations{
	m: sync.Mutex{},
	migrations: make([]*migrate.Migration,0),
}

func (dm *dbMigrations)add(migration *migrate.Migration){
	dm.m.Lock()
	dm.migrations = append(dm.migrations, migration)
	dm.m.Unlock()
}
func GetAll() *migrate.MemoryMigrationSource {
	return &migrate.MemoryMigrationSource{
		Migrations: migrationInstance.migrations,
	}
}