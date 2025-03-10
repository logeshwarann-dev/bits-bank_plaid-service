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
		log.Fatal("Error connection to DB: ", err.Error())
	}

	fmt.Println("DB Connection Successful!")
	return gormDb
}

func AddUser(bankdb *gorm.DB, plaidUser PlaidUser) error {
	if err := bankdb.Create(&plaidUser).Error; err != nil {
		return err
	}
	return nil
}

func CreateBankAccount(bankdb *gorm.DB, user PlaidUser) error {
	if err := AddUser(bankdb, user); err != nil {
		return fmt.Errorf("error adding plaid user in db: %v", err.Error())
	}
	return nil

}

func GetRecordUsingTrackId(bankdb *gorm.DB, trackId string) (PlaidUser, error) {
	var user PlaidUser
	result := bankdb.Where("trackId = ?", trackId).First(&user)
	if result.Error != nil {
		fmt.Println("Error: ", result.Error)
		return PlaidUser{}, errors.New("no records found")
	}
	return user, nil
}
