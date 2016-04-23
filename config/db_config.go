package config

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type db struct {
	*sql.DB
}

var (
	dbConfig = "root@tcp(localhost:3306)/egobie?charset=utf8&timeout=10m"
	DB db
)

func init() {
	tmp, _ := sql.Open("mysql", dbConfig)

	DB = db{tmp}

	DB.SetConnMaxLifetime(1 * time.Hour)
}
