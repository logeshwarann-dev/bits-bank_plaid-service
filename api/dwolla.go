package api

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/kolanos/dwolla-v2-go"
)

var (
	DwollaKey     string
	DwollaSecret  string
	DwollaClient  *dwolla.Client
	DwollaBaseUrl string
)

type FundingSourcePayload struct {
	Links      dwolla.Links `json:"_links"`
	Name       string       `json:"name"`
	PlaidToken string       `json:"plaidToken"`
}

type TransferRequestBody struct {
	Links  dwolla.Links  `json:"_links"`
	Amount dwolla.Amount `json:"amount"`
}

func CreateDwollaClient() {
	DwollaClient = dwolla.New(DwollaKey, DwollaSecret, dwolla.Sandbox)
}

func CreateOnDemandAuthorization(client *dwolla.Client) (dwolla.Links, error) {
	ctx := context.Background()
	onDemandAuth, err := client.OnDemandAuthorization.Create(ctx)
	if err != nil {
		log.Println(err.Error())
		return nil, fmt.Errorf("error while creating on-demand Auth: %v", err.Error())
	}

	dwollaAuthLinks := onDemandAuth.Links
	log.Println("DWOLLA AUTH LINKS: ", dwollaAuthLinks)
	return dwollaAuthLinks, nil
}

func CreateDwollaCustomer(dwollaUser BankUser) (string, string, error) {
	ctx := context.Background()
	dwollaCustomerPaylod := dwolla.CustomerRequest{
		FirstName:    dwollaUser.FirstName,
		LastName:     dwollaUser.LastName,
		Email:        dwollaUser.Email,
		BusinessName: fmt.Sprintf("%s %s's Business", dwollaUser.FirstName, dwollaUser.LastName),
	}
	newDwollaCustomer, err := DwollaClient.Customer.Create(ctx, &dwollaCustomerPaylod)
	if err != nil {
		log.Println(err.Error())
		return "", "", fmt.Errorf("error while creating Dwolla customer: %v", err.Error())
	}
	dwollaCustomerUrl := fmt.Sprintf("%s/customers/%s", DwollaBaseUrl, newDwollaCustomer.ID)
	return newDwollaCustomer.ID, dwollaCustomerUrl, nil
}

func CreateFundingSource(dwollaAuthLinks dwolla.Links, ctx context.Context, dwollaCustomerId string, processorToken string, bankName string) (*dwolla.FundingSource, error) {

	resource := dwolla.Resource{
		Links: dwollaAuthLinks,
	}
	dwollaCustomer, err := DwollaClient.Customer.Retrieve(ctx, dwollaCustomerId)
	if err != nil {
		log.Println(err.Error())
		return nil, fmt.Errorf("error while retrieving dwolla customer: %v", err.Error())
	}
	body := dwolla.FundingSourceRequest{
		BankAccountType: dwolla.FundingSourceBankAccountTypeChecking,
		PlaidToken:      processorToken,
		Resource:        resource,
		Name:            bankName,
	}
	fundingSource, err := dwollaCustomer.CreateFundingSource(ctx, &body)
	if err != nil {
		log.Println(err.Error())
		return nil, fmt.Errorf("error while creating funding source: %v", err.Error())
	}
	return fundingSource, nil

}

func CreateFundingSourceUsingPostCall(client *dwolla.Client, ctx context.Context, dwollaCustomerUrl string, processorToken string, bankName string, dwollaAuthLinks dwolla.Links) (map[string]interface{}, error) {

	fundingSourcePayload := FundingSourcePayload{
		Links:      dwollaAuthLinks,
		Name:       bankName,
		PlaidToken: processorToken,
	}
	dwollaCreateFundingSourceUrl := fmt.Sprintf("%s/funding-sources", dwollaCustomerUrl)
	log.Println("Dwolla FS url: ", dwollaCreateFundingSourceUrl)
	headers := &http.Header{}
	var responseContainer map[string]interface{}
	if err := client.Post(ctx, dwollaCreateFundingSourceUrl, fundingSourcePayload, headers, &responseContainer); err != nil {
		log.Println(err.Error())
		return nil, fmt.Errorf("error while creating funding source: %v", err.Error())
	}
	log.Println("Funding Source Response: ", responseContainer)
	return responseContainer, nil
}

func AddFundingSource(dwollaCustomerUrl string, processorToken string, bankName string) (string, error) {
	ctx := context.Background()
	dwollaAuthLinks, err := CreateOnDemandAuthorization(DwollaClient)
	if err != nil {
		log.Println(err.Error())
		return "", err
	}

	fundingSourceResponse, err := CreateFundingSourceUsingPostCall(DwollaClient, ctx, dwollaCustomerUrl, processorToken, bankName, dwollaAuthLinks)
	if err != nil {
		log.Println(err.Error())
		return "", err
	}
	fundingSourceId := fundingSourceResponse["id"]
	fundingSourceUrl := fmt.Sprintf("%s/funding-sources/%s", DwollaBaseUrl, fundingSourceId)
	//return funding source url
	return fundingSourceUrl, nil

	// CreateFundingSource(dwollaAuthLinks, ctx, dwollaCustomerId, processorToken, bankName)
}

func CreateTransfer(ctx context.Context, sourceFundingSourceUrl string, destinationFundingSourceUrl string, amount string) (map[string]interface{}, error) {

	var transferReq TransferRequestBody

	transferLinks := make(dwolla.Links)
	transferLinks["source"] = dwolla.Link{
		Href: sourceFundingSourceUrl,
	}
	transferLinks["destination"] = dwolla.Link{
		Href: destinationFundingSourceUrl,
	}

	transferReq.Links = transferLinks
	transferReq.Amount = dwolla.Amount{
		Currency: dwolla.USD,
		Value:    amount,
	}

	log.Println("Transfer Request: ", transferReq)

	transferUrl := fmt.Sprintf("%s/transfers", DwollaBaseUrl)

	headers := &http.Header{}
	var responseContainer map[string]interface{}
	if err := DwollaClient.Post(ctx, transferUrl, transferReq, headers, &responseContainer); err != nil {
		log.Println(err.Error())
		return nil, fmt.Errorf("error while creating dwolla transfer: %v", err.Error())
	}
	log.Println("Dwolla Transfer Response:  ", responseContainer)
	return responseContainer, nil

}

func RetrieveAccount(client *dwolla.Client) error {
	ctx := context.Background()
	res, err := client.Account.Retrieve(ctx)
	if err != nil {
		log.Println(err.Error())
		log.Println("Error:", err)
		return err
	}

	log.Println("Account ID:", res.ID)
	log.Println("Account Name:", res.Name)
	return nil
}
