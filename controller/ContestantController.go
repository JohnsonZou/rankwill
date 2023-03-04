package controller

import (
	"fetchTest/common"
	"fetchTest/model"
	"fetchTest/response"
	"github.com/gin-gonic/gin"
	"strconv"
)

func Getpage(c *gin.Context) {
	db := common.GetDB()
	contestname := c.PostForm("contestname")
	page := c.PostForm("page")
	if contestname == "" || page == "" {
		response.Fail(c, gin.H{}, "invalid form data")
		return
	}
	var con []model.Contestant
	p, err := strconv.Atoi(page)
	if err != nil {
		panic(err)
	}
	db.Where("rank>?", (p-1)*25).Where("rank<=?", p*25).Where("contestname=?", contestname).Find(&con)
	if con == nil {
		response.Fail(c, gin.H{}, "page empty")
		return

	}
	response.Success(c, gin.H{"result": con}, "Successfully query page")
}
func Getbyname(c *gin.Context) {
	db := common.GetDB()
	contestname := c.PostForm("contestname")
	contestantname := c.PostForm("contestantname")
	if contestname == "" || contestantname == "" {
		response.Fail(c, gin.H{}, "invalid form data")
		return
	}
	var con model.Contestant
	db.Where("username=?", contestantname).Where("contestname=?", contestname).First(&con)
	if con.ID == 0 {
		response.Fail(c, gin.H{}, "no such user")
		return
	}
	response.Success(c, gin.H{"result": con}, "Successfully query by name")
}
