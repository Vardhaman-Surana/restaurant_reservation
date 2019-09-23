package migrations

import migrate "github.com/rubenv/sql-migrate"

func init() {
	migrationInstance.add(&migrate.Migration{
		Id: "1568969826",
		//language=SQL
		Up: []string{`
	ALTER TABLE Reservations  ADD FOREIGN KEY (TABLE_ID)  REFERENCES restaurant_tables(ID)`},
	})
}

