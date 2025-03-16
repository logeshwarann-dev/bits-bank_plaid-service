package api

import (
	"context"
	"fmt"
	"log"

	"github.com/plaid/plaid-go/plaid"
)

var (
	PlaidClientId      string
	PlaidSecret        string
	PlaidAPIClient     *plaid.APIClient
	SandboxInstitution = "ins_109508"
	PaymentProcessor   = "dwolla"
)

type GetInstitutionReq struct {
	InstitutionId string   `json:"institution_id"`
	CountryCode   []string `json:"country_codes"`
}

func CreatePlaidConfig() {

	configuration := plaid.NewConfiguration()
	configuration.AddDefaultHeader("PLAID-CLIENT-ID", PlaidClientId)
	configuration.AddDefaultHeader("PLAID-SECRET", PlaidSecret)
	configuration.UseEnvironment(plaid.Sandbox)
	PlaidAPIClient = plaid.NewAPIClient(configuration)

}

func CreatePlaidLinkToken(plaidUser User) (string, error) {
	ctx := context.Background()
	user := plaid.LinkTokenCreateRequestUser{
		ClientUserId: plaidUser.UserId,
	}
	request := plaid.NewLinkTokenCreateRequest(
		plaidUser.Name,
		"en",
		[]plaid.CountryCode{plaid.COUNTRYCODE_US},
		user,
	)
	request.SetProducts([]plaid.Products{plaid.PRODUCTS_AUTH, plaid.PRODUCTS_TRANSACTIONS, plaid.PRODUCTS_IDENTITY})
	request.SetLinkCustomizationName("default")
	// request.SetWebhook("https://webhook-uri.com")
	// request.SetRedirectUri("https://domainname.com/oauth-page.html")
	request.SetAccountFilters(plaid.LinkTokenAccountFilters{
		Depository: &plaid.DepositoryFilter{
			AccountSubtypes: []plaid.AccountSubtype{plaid.ACCOUNTSUBTYPE_CHECKING, plaid.ACCOUNTSUBTYPE_SAVINGS},
		},
	})
	resp, _, err := PlaidAPIClient.PlaidApi.LinkTokenCreate(ctx).LinkTokenCreateRequest(*request).Execute()
	if err != nil {
		log.Println(err.Error())
		return "", fmt.Errorf("error while creating link token: %v", err.Error())
	}
	linkToken := resp.GetLinkToken()
	return linkToken, nil
}

func ExchangePublicTokenSandBoxMethod(plaidUser User) (string, error) {
	ctx := context.Background()
	testProducts := []plaid.Products{plaid.PRODUCTS_AUTH, plaid.PRODUCTS_TRANSACTIONS, plaid.PRODUCTS_IDENTITY}
	sandboxPublicTokenResp, _, err := PlaidAPIClient.PlaidApi.SandboxPublicTokenCreate(ctx).SandboxPublicTokenCreateRequest(
		*plaid.NewSandboxPublicTokenCreateRequest(
			SandboxInstitution,
			testProducts,
		),
	).Execute()
	if err != nil {
		log.Println(err.Error())
		return "", fmt.Errorf("error while creating public token: %v", err.Error())

	}
	publicToken := sandboxPublicTokenResp.GetPublicToken()

	return publicToken, nil
}

func ExchangePublicToken(publicToken string) (string, string, error) {
	ctx := context.Background()
	exchangePublicTokenResp, _, err := PlaidAPIClient.PlaidApi.ItemPublicTokenExchange(ctx).ItemPublicTokenExchangeRequest(
		*plaid.NewItemPublicTokenExchangeRequest(publicToken),
	).Execute()
	if err != nil {
		log.Println(err.Error())
		return "", "", fmt.Errorf("error while exhanging token: %v", err.Error())
	}
	accessToken := exchangePublicTokenResp.GetAccessToken()
	itemId := exchangePublicTokenResp.GetItemId()
	return accessToken, itemId, nil

}

func GetAccounts(accessToken string) (plaid.AccountBase, plaid.Item, error) {
	ctx := context.Background()
	accountsGetResp, _, err := PlaidAPIClient.PlaidApi.AccountsGet(ctx).AccountsGetRequest(
		*plaid.NewAccountsGetRequest(accessToken),
	).Execute()
	if err != nil {
		log.Println(err.Error())
		return plaid.AccountBase{}, plaid.Item{}, fmt.Errorf("error while getting account info: %v", err.Error())
	}
	accountData := accountsGetResp.GetAccounts()[0]
	accountItem := accountsGetResp.GetItem()
	// request := plaid.NewInstitutionsGetRequest(10, 0, []plaid.CountryCode{plaid.COUNTRYCODE_US})
	// resp, _, err := PlaidAPIClient.PlaidApi.InstitutionsGet(ctx).InstitutionsGetRequest(*request).Execute()
	// if err != nil {
	// 	log.Println("error occured while getting institution: ", err.Error())
	// }
	// log.Println("Response from Get Institution: ", resp.)
	return accountData, accountItem, nil
}

func CreataDwollaAccount(accessToken string, accountID string) (string, error) {
	ctx := context.Background()
	processorTokenCreateResp, _, err := PlaidAPIClient.PlaidApi.ProcessorTokenCreate(ctx).ProcessorTokenCreateRequest(
		*plaid.NewProcessorTokenCreateRequest(accessToken, accountID, PaymentProcessor),
	).Execute()
	if err != nil {
		log.Println(err.Error())
		return "", fmt.Errorf("error while creating Dwolla account: %v", err.Error())
	}
	processorToken := processorTokenCreateResp.ProcessorToken
	return processorToken, nil
}

func GetAccountInstituionId(institutionId string) (string, error) {
	ctx := context.Background()
	requestPayload := PlaidAPIClient.PlaidApi.InstitutionsGetById(ctx).InstitutionsGetByIdRequest(plaid.InstitutionsGetByIdRequest{InstitutionId: institutionId, CountryCodes: []plaid.CountryCode{plaid.COUNTRYCODE_US}})
	institutionResponse, _, err := PlaidAPIClient.PlaidApi.InstitutionsGetByIdExecute(requestPayload)
	if err != nil {
		return "", fmt.Errorf("error while getting institution: %v", err.Error())
	}
	instId := institutionResponse.Institution.InstitutionId
	return instId, nil
}
