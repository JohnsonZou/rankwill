package common

import (
	"fetchTest/model"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() *gorm.DB {
	//driverName := "mysql"
	host := "localhost"
	port := "3306"
	database := "rankwill"
	username := "root"
	password := "123456"
	charset := "utf8mb4"
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
	DB = db
	return db
}
func GetDB() *gorm.DB {
	return DB
}
