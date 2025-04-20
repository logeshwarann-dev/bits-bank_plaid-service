package main

import (
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/logeshwarann-dev/bits-bank_plaid-service/api"
	"github.com/logeshwarann-dev/bits-bank_plaid-service/db"
	"github.com/logeshwarann-dev/bits-bank_plaid-service/utils"
)

func init() {
	utils.LoadEnv()
	api.PlaidClientId = os.Getenv("PLAID_CLIENT_ID")
	api.PlaidSecret = os.Getenv("PLAID_SECRET")
	api.DwollaKey = os.Getenv("DWOLLA_KEY")
	api.DwollaSecret = os.Getenv("DWOLLA_SECRET")
	api.DwollaBaseUrl = os.Getenv("DWOLLA_BASE_URL")
	api.PgDb = db.ConnectToDB()
	api.CreatePlaidConfig()
	api.CreateDwollaClient()
}

func main() {

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://www.bitsbank-project.site"},
		// AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	router.POST("/plaid/v1/token/create", api.GenerateLinkToken)
	router.POST("/plaid/v1/token/exchange", api.GenerateAccessToken)
	router.POST("/plaid/v1/dwolla/customer/create", api.CreateDwollaCustomerId)
	router.POST("/plaid/v1/get/accounts", api.GetBankAccounts)
	router.POST("/plaid/v1/get/account", api.GetBankAccount)
	router.POST("/plaid/v1/dwolla/transfer", api.TransferPayment)
	router.Run(":8090")
}
