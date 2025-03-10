package api

import (
	"context"
	"fmt"

	"github.com/kolanos/dwolla-v2-go"
)

func createAccount() error {
	ctx := context.Background()

	client := dwolla.New("<your dwolla key here>", "<your dwolla secret here>", dwolla.Sandbox)

	res, err := client.Account.Retrieve(ctx)

	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	fmt.Println("Account ID:", res.ID)
	fmt.Println("Account Name:", res.Name)
	return nil
}
