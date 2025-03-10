package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/logeshwarann-dev/bits-bank_plaid-service/db"
	"github.com/logeshwarann-dev/bits-bank_plaid-service/utils"
	"gorm.io/gorm"
)

var PgDb *gorm.DB

type User struct {
	Email string `json:"email" binding:"required"`
	Name  string `json:"name" binding:"required"`
}

type BankUser struct {
	Email             string `json:"email"`
	Password          string `json:"password"`
	FirstName         string `json:"firstName"`
	LastName          string `json:"lastName"`
	DwollaCustomerUrl string `json:"dwollaCustomerUrl"`
	DwollaCustomerId  string `json:"dwollaCustomerId"`
	Address1          string `json:"address1"`
	City              string `json:"city"`
	State             string `json:"state"`
	PostalCode        string `json:"postalCode"`
	DateOfBirth       string `json:"dateOfBbirth"`
	AadharNo          string `json:"aadharNo"`
}

type PlaidAccount struct {
	PublicToken string
	PlaidUser   BankUser
}

func GenerateLinkToken(c *gin.Context) {
	var plaidUser User
	if err := c.ShouldBindJSON(&plaidUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request - " + err.Error()})
		return
	}

	linkToken, err := CreatePlaidLinkToken(plaidUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fmt.Println("Link Token: ", linkToken)
	c.JSON(http.StatusOK, gin.H{"link_token": linkToken})

}

func GenerateAccessToken(c *gin.Context) {
	var plaidAccount PlaidAccount
	if err := c.ShouldBindJSON(&plaidAccount); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request - " + err.Error()})
		return
	}
	fmt.Println("Public Token: ", plaidAccount.PublicToken)

	accessToken, itemId, err := ExchangePublicToken(plaidAccount.PublicToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fmt.Println("Access Token: ", accessToken, "| Item ID: ", itemId)

	accountData, err := GetAccounts(accessToken)
	accountId := accountData.GetAccountId()
	bankName := accountData.GetName()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fmt.Println("Account ID: ", accountId)

	processorToken, err := CreataDwollaAccount(accessToken, accountId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fmt.Println("Processor Token: ", processorToken)

	fundingSrcUrl, err := AddFundingSource(plaidAccount.PlaidUser.DwollaCustomerId, processorToken, bankName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	shareableId := utils.EncryptID(accountId)
	trackId := fmt.Sprintf("PLAID%s%v", plaidAccount.PlaidUser.FirstName, time.Now().Format("20060102150405"))

	newPlaidUser := db.PlaidUser{
		TrackId:          trackId,
		AccountId:        accountId,
		BankId:           itemId,
		AccessToken:      accessToken,
		FundingSourceUrl: fundingSrcUrl,
		ShareableId:      shareableId,
		UserId:           plaidAccount.PlaidUser.Email,
	}

	if err = db.CreateBankAccount(PgDb, newPlaidUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Plaid Account Linked Successfully"})

}

func CreateDwollaCustomerId(c *gin.Context) {
	var dwollaUser BankUser
	if err := c.ShouldBindJSON(&dwollaUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request - " + err.Error()})
		return
	}

	customerId, customerUrl, err := CreateDwollaCustomer(dwollaUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"customer_id": customerId, "customer_url": customerUrl})
}
