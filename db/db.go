package db

import (
	"errors"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	DB_USER string
	DB_HOST string
	DB_NAME string
	DB_PWD  string
	DB_PORT string
	DB_SSL  string
)

func ConnectToDB() *gorm.DB {

	dsn := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=%s", DB_USER, DB_PWD, DB_NAME, DB_HOST, DB_PORT, DB_SSL)
	gormDb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Println(err.Error())
		log.Fatal("Error connection to DB: ", err.Error())
	}

	log.Println("DB Connection Successful!")
	return gormDb
}

func AddUser(bankdb *gorm.DB, plaidUser PlaidUser) error {
	if err := bankdb.Create(&plaidUser).Error; err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}

func CreateBankAccount(bankdb *gorm.DB, user PlaidUser) error {
	if err := AddUser(bankdb, user); err != nil {
		log.Println(err.Error())
		return fmt.Errorf("error adding plaid user in db: %v", err.Error())
	}
	return nil

}

func GetRecordUsingTrackId(bankdb *gorm.DB, trackId string) (PlaidUser, error) {
	var user PlaidUser
	result := bankdb.Where("track_id = ?", trackId).First(&user)
	if result.Error != nil {
		log.Println("Error: ", result.Error)
		return PlaidUser{}, errors.New("no records found")
	}
	return user, nil
}

func GetRecordUsingAccountId(bankdb *gorm.DB, accountId string) (PlaidUser, error) {
	var user PlaidUser
	result := bankdb.Where("account_id = ?", accountId).First(&user)
	if result.Error != nil {
		log.Println("Error: ", result.Error)
		return PlaidUser{}, errors.New("no records found")
	}
	return user, nil
}

func GetAllRecordUsingUserId(bankdb *gorm.DB, userId string) ([]PlaidUser, error) {
	var accounts []PlaidUser
	result := bankdb.Where("user_id = ?", userId).Find(&accounts)
	if result.Error != nil {
		log.Println("Error: ", result.Error)
		return []PlaidUser{}, errors.New("no records found")
	}
	return accounts, nil
}

func AddTransaction(bankdb *gorm.DB, transaction Transaction) error {
	if err := bankdb.Create(&transaction).Error; err != nil {
		log.Println(err.Error())
		return fmt.Errorf("error while adding transaction entry in db: %v", err.Error())
	}
	return nil
}

func GetTransactionUsingId(bankdb *gorm.DB, transactionId string) (Transaction, error) {
	var transaction Transaction
	result := bankdb.Where("transaction_id = ?", transactionId).Find(&transaction)
	if result.Error != nil {
		log.Println("Error: ", result.Error)
		return Transaction{}, errors.New("no records found")
	}
	return transaction, nil
}

func GetTransactionUsingSenderBankId(bankdb *gorm.DB, senderBankId string) ([]Transaction, error) {
	var transaction []Transaction
	result := bankdb.Where("sender_bank_id = ?", senderBankId).Find(&transaction)
	if result.Error != nil {
		log.Println("Error: ", result.Error)
		return []Transaction{}, errors.New("no records found")
	}
	return transaction, nil
}

func GetTransactionUsingReceiverBankId(bankdb *gorm.DB, receiverBankId string) ([]Transaction, error) {
	var transaction []Transaction
	result := bankdb.Where("received_bank_id = ?", receiverBankId).Find(&transaction)
	if result.Error != nil {
		log.Println("Error: ", result.Error)
		return []Transaction{}, errors.New("no records found")
	}
	return transaction, nil
}

// func GetTransaction
