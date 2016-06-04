package config

import (
	"os"
	"fmt"

	"github.com/lionelbarrow/braintree-go"
)

const (
	CREDIT_CARD = "CREDIT"
	DEBIT_CARD  = "DEBIT"
)

var BT *braintree.Braintree

func init() {
	os.Getenv("");

	merchantId := os.Getenv("EGOBIE_MERCHANT_ID")
	publicKey := os.Getenv("EGOBIE_PUBLIC_KEY")
	privateKey := os.Getenv("EGOBIE_PRIVATE_KEY")

	if merchantId == "" || publicKey == "" || privateKey == "" {
		fmt.Println("Merchant not Configured properly")
		os.Exit(1)
	}

	fmt.Println("merchantId - ", merchantId)
	fmt.Println("publicKey - ", publicKey)
	fmt.Println("privateKey - ", privateKey)

	BT = braintree.New(
		braintree.Production,
		merchantId,
		publicKey,
		privateKey,
	)
}
