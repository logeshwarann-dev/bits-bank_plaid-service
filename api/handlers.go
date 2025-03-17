package api

import (
	"fmt"
	"log"
	"net/http"
	"sort"
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
	TrackId string `json:"plaidTrackId" binding:"required"`
}

type BankUserId struct {
	UserId string `json:"userId" binding:"required"`
}

type TransactionRequest struct {
	Name           string `json:"name"`
	Amount         string `json:"amount"`
	SenderId       string `json:"senderId"`
	SenderBankId   string `json:"senderBankId"`
	ReceiverId     string `json:"receiverId"`
	ReceiverBankId string `json:"receiverBankId"`
	Email          string `json:"email"`
}

type TransactionsUsingBankId struct {
	total     int              `json:"total"`
	documents []db.Transaction `json:"documents"`
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

		institutionId, _ := GetDefaultInstitutionId(accountItem)

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
	var request TrackIdRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Println("invalid request body: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}

	bankDetails, err := db.GetRecordUsingTrackId(PgDb, request.TrackId)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to fetch record using track id: " + err.Error()})
		return
	}

	accountData, accountItem, err := GetAccounts(bankDetails.AccessToken)
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
	currentBal := strconv.FormatFloat(float64(accountData.Balances.GetCurrent()), 'f', -1, 32)

	institutionId, _ := GetDefaultInstitutionId(accountItem)

	transferTransactionsData, err := GetTransactionsByBankId(PgDb, bankDetails.TrackId)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var transferTransactions []PlaidTransaction
	for _, eachTransaction := range transferTransactionsData.documents {
		transferTransactions = append(transferTransactions, PlaidTransaction{
			Id:             eachTransaction.TransactionId,
			Name:           eachTransaction.Name,
			Amount:         eachTransaction.Amount,
			Date:           utils.ExtractTimeStamp(eachTransaction.TransactionId),
			PaymentChannel: eachTransaction.Channel,
			Category:       eachTransaction.Category,
			Type: func() string {
				if eachTransaction.SenderBankId == bankDetails.TrackId {
					return "debit"
				}
				return "credit"
			}(),
		})
	}

	transactions, err := GetTransactionsFromPlaid(bankDetails.AccessToken)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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
		PlaidTrackId:     bankDetails.TrackId,
		ShareableId:      bankDetails.ShareableId,
	}

	allTransactions := append(transactions, transferTransactions...)

	// Sort the combined slice by date in descending order
	sort.Slice(allTransactions, func(i, j int) bool {
		dateFormat := "2006-01-02"
		dateI, errI := time.Parse(dateFormat, allTransactions[i].Date)
		dateJ, errJ := time.Parse(dateFormat, allTransactions[j].Date)
		if errI != nil || errJ != nil {
			return false
		}
		return dateJ.Before(dateI)
	})

	log.Println("Account: ", account, "| transactions: ", allTransactions)
	c.JSON(http.StatusOK, gin.H{"data": account, "transactions": allTransactions})

}

func CreateTransaction(bankdb *gorm.DB, transactionReq TransactionRequest) (db.Transaction, error) {

	transactionId := fmt.Sprintf("TRANSCT%v", time.Now().Format("20060102150405"))
	transactionRecord := db.Transaction{
		TransactionId:  transactionId,
		Amount:         transactionReq.Amount,
		Channel:        "online",
		Category:       "Transfer",
		SenderId:       transactionReq.SenderId,
		ReceiverId:     transactionReq.ReceiverId,
		SenderBankId:   transactionReq.SenderBankId,
		ReceiverBankId: transactionReq.ReceiverBankId,
	}

	if err := db.AddTransaction(bankdb, transactionRecord); err != nil {
		return db.Transaction{}, err
	}

	transaction, err := db.GetTransactionUsingId(bankdb, transactionId)
	if err != nil {
		return db.Transaction{}, err
	}

	return transaction, nil
}

func GetTransactionsByBankId(bankdb *gorm.DB, bankId string) (TransactionsUsingBankId, error) {

	senderBankDocs, err := db.GetTransactionUsingSenderBankId(bankdb, bankId)
	if err != nil {
		return TransactionsUsingBankId{}, fmt.Errorf("error while fetching transactions using sender bank id: %v ", err.Error())
	}

	receiverBankDocs, err := db.GetTransactionUsingReceiverBankId(bankdb, bankId)
	if err != nil {
		return TransactionsUsingBankId{}, fmt.Errorf("error while fetching transactions using receiver bank id: %v ", err.Error())
	}

	docs := append(senderBankDocs, receiverBankDocs...)

	transactions := TransactionsUsingBankId{
		total:     len(senderBankDocs) + len(receiverBankDocs),
		documents: docs,
	}

	return transactions, nil
}
