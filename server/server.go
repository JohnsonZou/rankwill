package server

import (
	"fetchTest/common"
	_ "fetchTest/model"
	"fetchTest/router"
	"github.com/gin-gonic/gin"
	_ "gorm.io/driver/mysql"
)

func GinRun() {
	_ = common.InitDB()
	r := gin.Default()
	r = router.CollectRoute(r)
	panic(r.Run()) // 监听并在 0.0.0.0:8080 上启动服务
}
