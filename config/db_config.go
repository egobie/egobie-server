package config

import (
	"database/sql"
	"time"
	"fmt"

	"github.com/spf13/viper"
	_ "github.com/go-sql-driver/mysql"
)

type db struct {
	*sql.DB
}

var (
	dbHost = ""
	dbUser = ""
	dbPassword = ""
	dbPort = ""
	dbName = ""
	dbProtocol = ""
	// username:password@protocol(address)/dbname?param=value
	dbConfig = ""

	DB db
)

/**
{
    "db_host": "ec2_host",
    "db_port": "3306",
    "db_protocol": "protocol",
    "db_user": "user",
    "db_password": "password",
    "db_name": "database_name"
}
**/
func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("json")

	viper.SetDefault("db_host", "localhost");
	viper.SetDefault("db_user", "root");
	viper.SetDefault("db_password", "");
	viper.SetDefault("db_protocol", "tcp");
	viper.SetDefault("db_port", "3306");
	viper.SetDefault("db_name", "egobie");

	viper.AddConfigPath("/etc/egobie/config/")
	viper.ReadInConfig()

	dbHost = viper.GetString("db_host");
	dbUser = viper.GetString("db_user");
	dbPassword = viper.GetString("db_password");
	dbProtocol = viper.GetString("db_protocol");
	dbPort = viper.GetString("db_port");
	dbName = viper.GetString("db_name");

	if len(dbPassword) != 0 {
		dbConfig = fmt.Sprintf(
			"%v:%v@%v(%v:%v)/%v?charset=utf8&timeout=10m",
			dbUser, dbPassword, dbProtocol, dbHost, dbPort, dbName,
		)
	} else {
		dbConfig = fmt.Sprintf(
			"%v@%v(%v:%v)/%v?charset=utf8&timeout=10m",
			dbUser, dbProtocol, dbHost, dbPort, dbName,
		)
	}

	tmp, _ := sql.Open("mysql", dbConfig)

	DB = db{tmp}

	DB.SetConnMaxLifetime(1 * time.Hour)
}
