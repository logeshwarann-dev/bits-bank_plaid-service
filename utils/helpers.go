package utils

import (
	"encoding/base64"
	"log"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file: ", err.Error())
	}
}

func EncryptID(id string) string {
	return base64.StdEncoding.EncodeToString([]byte(id))
}

func DecryptID(encoded string) (string, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	return string(decodedBytes), nil
}
