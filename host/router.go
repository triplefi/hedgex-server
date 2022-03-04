package host

import (
	"hedgex-public/config"
	"hedgex-public/gl"
	"hedgex-public/model"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/triplefi/go-logger/logger"
)

func StartHttpServer() {
	router := InitRouter()
	go router.Run(":" + strconv.Itoa(config.HttpPort))
}

// Index homepage
func PingPong(c *gin.Context) {
	if err := model.Ping(); err != nil {
		c.String(http.StatusOK, "error")
		gl.OutLogger.Error("connect to mysql error. %v", err)
		return
	}
	c.String(http.StatusOK, "pong")
}

// InitRouter init the router
func InitRouter() *gin.Engine {
	if err := os.MkdirAll("./logs/http", os.ModePerm); err != nil {
		log.Panic("Create dir './logs/http' error. " + err.Error())
	}
	GinLogger, err := logger.New("logs/http/gin.log", 2, 100*1024*1024, 10)
	if err != nil {
		log.Panic("Create GinLogger file error. " + err.Error())
	}

	router := gin.New()
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{Output: GinLogger}), gin.Recovery())

	router.GET("/api/ping", PingPong)

	//con
	con := router.Group("/api/contract")
	{
		con.GET("/trade_pairs", GetTradePairs)
		con.GET("/pair_params", GetPairParams)
		con.GET("/trade", GetTradeRecordsByContract)
		con.GET("/explosive", GetExplosiveRecordsByContract)
		con.GET("/kline", GetKlineData)
		con.GET("/position", GetStatPositions)
	}

	//acc
	acc := router.Group("/api/account")
	{
		acc.GET("/", GetTraders)
		acc.GET("/trade", GetTradeRecords)
		acc.GET("/interest", GetInterest)
		acc.GET("/explosive", GetExplosive)
	}

	//wss
	wss := router.Group("/wss")
	{
		wss.GET("/kline", klineSender)
	}

	other := router.Group("/api/odds")
	{
		other.GET("/add_email", AddEmail)
		other.GET("/emails", GetEmails)
		other.GET("/testcoin", SendTestCoins)
	}

	return router
}
