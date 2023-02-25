package common

import (
	"fetchTest/model"
	"fmt"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() *gorm.DB {
	//driverName := "mysql"
	host := viper.GetString("datasource.host")
	port := viper.GetString("datasource.port")
	database := viper.GetString("datasource.database")
	username := viper.GetString("datasource.username")
	password := viper.GetString("datasource.password")
	charset := viper.GetString("datasource.charset")
	args := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=true&loc=Local",
		username,
		password,
		host,
		port,
		database,
		charset)
	db, err := gorm.Open(mysql.Open(args))
	if err != nil {
		panic("fail to connect database,err: " + err.Error())
	}
	db.AutoMigrate(&model.User{})
	db.AutoMigrate(&model.Contest{})
	db.AutoMigrate(&model.Contestant{})
	DB = db
	return db
}
func GetDB() *gorm.DB {
	return DB
}
