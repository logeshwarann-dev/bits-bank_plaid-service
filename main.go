package main

import (
	"github.com/gin-gonic/gin"
	"github.com/logeshwarann-dev/bits-bank_plaid-service/api"
	"github.com/logeshwarann-dev/bits-bank_plaid-service/utils"
)

func init() {
	utils.LoadEnv()
	api.CreateConfig()
}

func main() {

	router := gin.Default()
	router.POST("/plaid/v1/token/create", api.GenerateLinkToken)
	router.POST("/plaid/v1/token/exchange", api.GenerateAccessToken)
	router.Run(":8090")
}
