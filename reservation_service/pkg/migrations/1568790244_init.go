package migrations

import migrate "github.com/rubenv/sql-migrate"

func init() {
	migrationInstance.add(&migrate.Migration{
		Id: "1568790244",
		//language=SQL
		Up: []string{`
	CREATE TABLE restaurant_tables(
  	ID int AUTO_INCREMENT PRIMARY KEY,
  	Created bigint(20) DEFAULT NULL,
  	Updated bigint(20) DEFAULT NULL,
  	Deleted tinyint(1) DEFAULT 0,
  	Restaurant_ID int NOT NULL
	) ENGINE=InnoDB DEFAULT CHARSET=utf8;`,`
	CREATE TABLE Reservations (
  	ID int AUTO_INCREMENT PRIMARY KEY,
  	Created bigint(20) DEFAULT NULL,
  	Updated bigint(20) DEFAULT NULL,
  	Deleted tinyint(1) DEFAULT 0,
  	Restaurant_ID int NOT NULL,
  	User_ID varchar(255) NOT NULL,
  	Table_ID int NOT NULL,
  	Start_Time int NOT NULL
  	) ENGINE=InnoDB DEFAULT CHARSET=utf8;
	`},
		Down:[]string{
			`DROP TABLE RestaurantTables;`,
			`DROP TABLE Reservations;`,
		},
	})
}