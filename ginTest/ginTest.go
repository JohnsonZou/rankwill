package ginTest

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	_ "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
)

type User struct {
	gorm.Model
	Username string `gorm:"type:varchar(30);not null"`
	Email    string `gorm:"type:varchar(30);not null;unique"`
	Password string `gorm:"size:255;not null"`
}

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
	db.AutoMigrate(&User{})
	return db
}
func isEmailExisted(db *gorm.DB, email string) bool {
	var user User
	db.Where("email=?", email).First(&user)
	return user.ID != 0
}
func GinRun() {
	db := InitDB()
	r := gin.Default()
	r.POST("/api/auth/register", func(c *gin.Context) {
		username := c.PostForm("name")
		email := c.PostForm("email")
		password := c.PostForm("password")
		log.Println(username, email, password)
		if isEmailExisted(db, email) {
			c.JSON(422, gin.H{
				"code":    422,
				"message": "Register failed,email existed.",
			})
			return
		}
		newUser := User{
			Username: username,
			Email:    email,
			Password: password,
		}
		db.Create(&newUser)

		c.JSON(200, gin.H{
			"message": "Successfully register",
		})
	})
	r.Run() // 监听并在 0.0.0.0:8080 上启动服务
}
