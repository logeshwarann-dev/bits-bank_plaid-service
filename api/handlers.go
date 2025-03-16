package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/logeshwarann-dev/bits-bank_plaid-service/db"
	"github.com/logeshwarann-dev/bits-bank_plaid-service/utils"
	"gorm.io/gorm"
)

var PgDb *gorm.DB

type User struct {
	UserId string `json:"userId" binding:"required"`
	Email  string `json:"email" binding:"required"`
	Name   string `json:"name" binding:"required"`
}

type TrackIdRequest struct {
	TrackId string `json:"trackId" binding:"required"`
}

type BankUserId struct {
	UserId string `json:"userId" binding:"required"`
}

type Account struct {
	Id               string `json:"id"`
	AvailableBalance string `json:"availableBalance"`
	CurrentBalance   string `json:"currentBalance"`
	InstitutionId    string `json:"institutionId"`
	Name             string `json:"name"`
	OfficialName     string `json:"officialName"`
	Mask             string `json:"mask"`
	Type             string `json:"type"`
	SubType          string `json:"subType"`
	PlaidTrackId     string `json:"plaidTrackId"`
	ShareableId      string `json:"shareableId"`
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
	UserId            string `json:"userId"`
}

type PlaidAccount struct {
	PublicToken string   `json:"publicToken"`
	PlaidUser   BankUser `json:"user"`
}

func GenerateLinkToken(c *gin.Context) {
	var plaidUser User
	if err := c.ShouldBindJSON(&plaidUser); err != nil {
		// log.Println(err.Error())
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request - " + err.Error()})
		return
	}

	linkToken, err := CreatePlaidLinkToken(plaidUser)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Println("Link Token: ", linkToken)
	c.JSON(http.StatusOK, gin.H{"link_token": linkToken})

}

func GenerateAccessToken(c *gin.Context) {
	var plaidAccount PlaidAccount
	if err := c.ShouldBindJSON(&plaidAccount); err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request - " + err.Error()})
		return
	}
	log.Println("Public Token: ", plaidAccount.PublicToken, "| Plaid User: ", plaidAccount.PlaidUser)

	accessToken, itemId, err := ExchangePublicToken(plaidAccount.PublicToken)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Println("Access Token: ", accessToken, "| Item ID: ", itemId)

	accountData, _, err := GetAccounts(accessToken)
	accountId := accountData.GetAccountId()
	bankName := accountData.GetName()
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Println("Account ID: ", accountId, "| Bank Name: ", bankName)

	processorToken, err := CreataDwollaAccount(accessToken, accountId)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Println("Processor Token: ", processorToken)

	fundingSrcUrl, err := AddFundingSource(plaidAccount.PlaidUser.DwollaCustomerUrl, processorToken, bankName)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	shareableId := utils.EncryptID(accountId)
	trackId := fmt.Sprintf("PLAID%s%v", strings.ToUpper(plaidAccount.PlaidUser.FirstName[:3]), time.Now().Format("20060102150405"))

	newPlaidUser := db.PlaidUser{
		TrackId:          trackId,
		AccountId:        accountId,
		BankId:           itemId,
		AccessToken:      accessToken,
		FundingSourceUrl: fundingSrcUrl,
		ShareableId:      shareableId,
		UserId:           plaidAccount.PlaidUser.UserId,
	}

	if err = db.CreateBankAccount(PgDb, newPlaidUser); err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	plaidUserFromDb, err := db.GetRecordUsingTrackId(PgDb, trackId)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Plaid Account Linked Successfully", "plaidUser": plaidUserFromDb})

}

func CreateDwollaCustomerId(c *gin.Context) {
	var dwollaUser BankUser
	if err := c.ShouldBindJSON(&dwollaUser); err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request - " + err.Error()})
		return
	}

	customerId, customerUrl, err := CreateDwollaCustomer(dwollaUser)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"customer_id": customerId, "customer_url": customerUrl})
}

func GetBankAccounts(c *gin.Context) {
	var userData BankUserId
	if err := c.ShouldBindJSON(&userData); err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request - " + err.Error()})
		return
	}

	plaidDBRecords, err := db.GetAllRecordUsingUserId(PgDb, userData.UserId)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to fetch accounts - " + err.Error()})
		return
	}

	log.Println("Plaid DB records: ", plaidDBRecords)

	var accounts []Account

	for _, eachRecord := range plaidDBRecords {
		accountData, accountItem, err := GetAccounts(eachRecord.AccessToken)
		// log.Println("AccountData: ", accountData, "| accountItem: ", accountItem)
		if err != nil {
			log.Println(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var availableBal string
		if accountData.Balances.Available.IsSet() {
			availableBal = strconv.FormatFloat(float64(*accountData.Balances.Available.Get()), 'f', -1, 32)
		} else {
			availableBal = ""
			log.Println("Account Available Balance nil")
		}
		log.Println("Avail Bal: ", availableBal)
		currentBal := strconv.FormatFloat(float64(accountData.Balances.GetCurrent()), 'f', -1, 32)

		instId := accountItem.GetInstitutionId()
		if len(instId) == 0 {
			instId = "ins_1"
		}
		institutionId, err := GetAccountInstituionId(instId)
		if err != nil {
			log.Println(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if len(institutionId) == 0 {
			institutionId = instId
		}
		account := Account{
			Id:               accountData.GetAccountId(),
			AvailableBalance: availableBal,
			CurrentBalance:   currentBal,
			InstitutionId:    institutionId,
			Name:             accountData.Name,
			OfficialName:     accountData.GetOfficialName(),
			Mask:             accountData.GetMask(),
			Type:             string(accountData.GetType()),
			SubType:          string(*accountData.Subtype.Get()),
			PlaidTrackId:     eachRecord.TrackId,
			ShareableId:      eachRecord.ShareableId,
		}

		log.Println("Account: ", account)

		accounts = append(accounts, account)
	}

	totalBanks := len(accounts)
	var totalCurrentBalance float64
	for _, eachAccount := range accounts {
		currentBal, err := strconv.ParseFloat(eachAccount.CurrentBalance, 64)
		if err != nil {
			log.Println("unable to convert current balance str to float64: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to convert current balance str to float64: " + err.Error()})
			return
		}
		totalCurrentBalance += currentBal
	}

	log.Println("accounts: ", accounts, "| totalBanks: ", totalBanks, "| totalcurrentbal: ", totalCurrentBalance)

	c.JSON(http.StatusOK, gin.H{"accounts": accounts, "totalBanks": strconv.Itoa(totalBanks), "totalCurrentBalance": totalCurrentBalance})

}

func GetBankAccount(c *gin.Context) {
	// var request TrackIdRequest
	// if err := c.ShouldBindJSON(&request); err != nil {
	// 	log.Println("invalid request body: " + err.Error())
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
	// 	return
	// }

	// bankDetails, err := db.GetRecordUsingTrackId(PgDb, request.TrackId)
	// if err != nil {
	// 	log.Println(err.Error())
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to fetch record using track id: " + err.Error()})
	// 	return
	// }

}
