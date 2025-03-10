package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type User struct {
	Email string `json:"email" binding:"required"`
	Name  string `json:"name" binding:"required"`
}

type BankUser struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Address1    string `json:"address1"`
	City        string `json:"city"`
	State       string `json:"state"`
	PostalCode  string `json:"postal_code"`
	DateOfBirth string `json:"date_of_birth"`
	AadharNo    string `json:"aadhar_no"`
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

	accountId, err := GetAccounts(accessToken)
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

}
