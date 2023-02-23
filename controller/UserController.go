package controller

import (
	"fetchTest/common"
	"fetchTest/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"net/http"
	"regexp"
)

func isEmailExisted(db *gorm.DB, email string) bool {
	var user model.User
	db.Where("email=?", email).First(&user)
	return user.ID != 0
}
func getUserByEmail(db *gorm.DB, email string) model.User {
	var user model.User
	db.Where("email=?", email).First(&user)
	return user
}
func validEmail(email string) (bool, error) {
	regex := "^([a-z0-9A-Z]+[-|\\.]?)+[a-z0-9A-Z]@([a-z0-9A-Z]+(-[a-z0-9A-Z]+)?\\.)+[a-zA-Z]{2,}$"
	return regexp.MatchString(regex, email)
}
func Register(c *gin.Context) {
	db := common.GetDB()
	username := c.PostForm("username")
	email := c.PostForm("email")
	password := c.PostForm("password")
	log.Println(username, email, password)

	if res, matchErr := validEmail(email); res == false || matchErr != nil {
		if matchErr != nil {
			c.JSON(500, gin.H{
				"code":    500,
				"message": "Email matching fail",
			})
		}
		c.JSON(422, gin.H{
			"code":    422,
			"message": "Email invalid",
		})
		return
	}

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
		"code":    200,
	})
}
func Login(c *gin.Context) {
	db := common.GetDB()
	email := c.PostForm("email")
	password := c.PostForm("password")
	log.Println(email, password)
	loginUser := getUserByEmail(db, email)

	if loginUser.ID == 0 {
		c.JSON(422, gin.H{
			"code":    422,
			"message": "Login failed,email not exist",
		})
		return
	}
	if loginUser.Password != password {
		c.JSON(400, gin.H{
			"code":    400,
			"message": "Wrong password",
		})
		return
	}
	token, tokenGenErr := common.ReleaseToken(loginUser)
	if tokenGenErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "token generation failed"})
		log.Println("token generation failed", tokenGenErr.Error())
		return
	}
	c.JSON(200, gin.H{
		"code":    200,
		"message": "Successfully login",
		"data":    gin.H{"token": token},
	})
}
