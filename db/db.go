package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql" // Necessary for connecting to MySQL.
)

/*
	Structs and variables
*/

// Database connection information.
var (
	// Type of database.
	Type = "mysql"
	// Username to access database.
	Username = "ap"
	// Password to access database.
	Password = os.Getenv("DB_PASSWORD")
	// Protocol of database.
	Protocol = "unix"
	// FileLocation of database socket.
	FileLocation = "/var/run/mysqld/mysqld.sock"
	// Database location.
	Database = "ap"
	// ConnString to connect to database.
	ConnString = Username + ":" + Password + "@" + Protocol + "(" + FileLocation + ")/" + Database
)

var (
	db *sql.DB
)

// InitDB initializes the Database.
func InitDB() (err error) {
	db, err = sql.Open(Type, ConnString)
	return
}

/*
	Helper functions
*/

func rowExists(query string, args ...interface{}) (exists bool, err error) {
	query = fmt.Sprintf("SELECT exists (%s)", query)
	err = db.QueryRow(query, args...).Scan(&exists)
	return
}
