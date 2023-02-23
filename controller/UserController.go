package controller

import (
	"fetchTest/common"
	"fetchTest/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
)

func isEmailExisted(db *gorm.DB, email string) bool {
	var user model.User
	db.Where("email=?", email).First(&user)
	return user.ID != 0
}
func Register(c *gin.Context) {
	db := common.GetDB()
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
	newUser := model.User{
		Username: username,
		Email:    email,
		Password: password,
	}
	db.Create(&newUser)

	c.JSON(200, gin.H{
		"message": "Successfully register",
	})
}
