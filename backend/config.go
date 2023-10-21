package app

import (
	"fmt"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
)

var dbOnce sync.Once
var dbConn *sqlx.DB

var configCache *Conf

type Conf struct {
	BaseURL  string
	HTTPPort string
	DBName   string
	DBHost   string
	DBUser   string
	DBPass   string
}

const (
	BaseURL  = "BASE_URL"
	HTTPPort = "HTTP_PORT"
	DBName   = "DB_NAME"
	DBHost   = "DB_HOST"
	DBUser   = "DB_USER"
	DBPass   = "DB_PASS"
)

func Config() *Conf {
	if configCache != nil {
		return configCache
	}

	viper.SetConfigName("conf")
	viper.SetConfigType("env")
	viper.AddConfigPath("./")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		panic("Viper: failed to find config file")
	}

	configCache = &Conf{
		BaseURL:  viper.GetString(BaseURL),
		HTTPPort: viper.GetString(HTTPPort),
		DBName:   viper.GetString(DBName),
		DBHost:   viper.GetString(DBHost),
		DBUser:   viper.GetString(DBUser),
		DBPass:   viper.GetString(DBPass),
	}
	return configCache
}

func GetDB() *sqlx.DB {
	var err error
	dbOnce.Do(func() {
		dbConn, err = sqlx.Connect(
			"mysql",
			fmt.Sprintf(
				"%s:%s@(%s)/%s?parseTime=true",
				Config().DBUser,
				Config().DBPass,
				Config().DBHost,
				Config().DBName,
			),
		)
		if err != nil {
			panic(err)
		}
	})

	return dbConn
}
