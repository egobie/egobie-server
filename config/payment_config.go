package config

import (
	"fmt"
	"os"

	"github.com/lionelbarrow/braintree-go"
)

const (
	CREDIT_CARD = "CREDIT"
	DEBIT_CARD  = "DEBIT"
)

var BT *braintree.Braintree

func init() {
	os.Getenv("")

	merchantId := os.Getenv("EGOBIE_MERCHANT_ID")
	publicKey := os.Getenv("EGOBIE_PUBLIC_KEY")
	privateKey := os.Getenv("EGOBIE_PRIVATE_KEY")
	braintreeEnv := os.Getenv("EGOBIE_BRAINTREE_ENV")

	if merchantId == "" || publicKey == "" ||
		privateKey == "" || braintreeEnv == "" {
		fmt.Println("Merchant not Configured properly")
		os.Exit(0)
	}

	/*
		fmt.Println("merchantId - ", merchantId)
		fmt.Println("publicKey - ", publicKey)
		fmt.Println("privateKey - ", privateKey)
		fmt.Println("braintreeEnv - ", braintreeEnv)
	*/

	BT = braintree.New(
		braintree.Environment(braintreeEnv),
		merchantId,
		publicKey,
		privateKey,
	)

	fmt.Println(BT.Environment)
}
