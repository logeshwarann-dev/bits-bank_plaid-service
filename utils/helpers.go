package utils

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/logeshwarann-dev/bits-bank_plaid-service/api"
)

func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file: ", err.Error())
	}

	api.PlaidClientId = os.Getenv("PLAID_CLIENT_ID")
	api.PlaidSecret = os.Getenv("PLAID_SECRET")

}
