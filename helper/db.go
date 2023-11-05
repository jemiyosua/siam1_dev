package helper

import (
	"database/sql"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func Connect(c *gin.Context) *sql.DB {

	Conn := os.Getenv("DBROOT")
	db, err := sql.Open("mysql", Conn)
	db.SetMaxIdleConns(0)
	db.SetMaxOpenConns(25)
	db.SetConnMaxLifetime(3 * time.Second)

	if err != nil {
		errormessage := "Got error in mysql connector : " + err.Error()
		SendLogError("", "Database", errormessage, "", "", "3", "", "", "", "", c)
	}

	err = db.Ping()
	if db.Ping() != nil {
		errormessage := "Could not connect to database : " + err.Error()
		SendLogError("", "Database", errormessage, "", "", "3", "", "", "", "", c)
	}

	return db
}
