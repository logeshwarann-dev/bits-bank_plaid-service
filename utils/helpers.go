package utils

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/logeshwarann-dev/bits-bank_plaid-service/db"
)

func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println(err.Error())
		log.Fatal("Error loading .env file: ", err.Error())
	}
	db.DB_HOST = os.Getenv("DB_HOST")
	db.DB_PWD = os.Getenv("DB_PWD")
	db.DB_NAME = os.Getenv("DB_NAME")
	db.DB_PORT = os.Getenv("DB_PORT")
	db.DB_USER = os.Getenv("DB_USER")
	db.DB_SSL = os.Getenv("DB_SSL")
}

func EncryptID(id string) string {
	return base64.StdEncoding.EncodeToString([]byte(id))
}

func DecryptID(encoded string) (string, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		log.Println(err.Error())
		return "", err
	}
	return string(decodedBytes), nil
}

func ExtractTimeStamp(value string) string {
	datePart := value[7:15]

	// Parse the date part
	date, err := time.Parse("20060102", datePart)
	if err != nil {
		fmt.Println("Error parsing date:", err)
		return ""
	}

	// Format the date as YYYY-MM-DD
	formattedDate := date.Format("2006-01-02")
	return formattedDate
}
