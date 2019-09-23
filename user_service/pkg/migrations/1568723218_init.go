package migrations

import migrate "github.com/rubenv/sql-migrate"

func init() {
	migrationInstance.add(&migrate.Migration{
		Id: "1568723218",
		//language=SQL
		Up: []string{`
	CREATE TABLE users (
  	ID varchar(255) NOT NULL,
  	Created bigint(20) DEFAULT NULL,
  	Updated bigint(20) DEFAULT NULL,
  	Deleted tinyint(1) DEFAULT 0,
  	Email VARCHAR(255) NOT NULL,
  	Name varchar(255) DEFAULT NULL,
  	PasswordHash VARCHAR(255) NOT NULL,
  	UNIQUE (Email),
  	PRIMARY KEY (ID)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8;`},
		Down:[]string{
			`DROP TABLE users;`,
		},
	})
}


