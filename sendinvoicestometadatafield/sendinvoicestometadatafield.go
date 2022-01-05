package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/Invoiced/invoiced-go/invdendpoint"
	"github.com/Invoiced/invoiced-go"
	"os"
	"strings"
)

//This program will send invoices to email addresses in the metadata field

func main() {
	sandBoxEnv := true
	key := flag.String("key", "", "api key in Settings > Developer")
	environment := flag.String("env", "", "your environment production or sandbox")
	emailMetadataKey := flag.String("metadatakey", "", "your environment production or sandbox")
	flag.Parse()

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("This program will send customers statements for customers in the excel file")

	if *key == "" {
		fmt.Print("Please enter your API Key: ")
		*key, _ = reader.ReadString('\n')
		*key = strings.TrimSpace(*key)
	}

	*environment = strings.ToUpper(strings.TrimSpace(*environment))

	if *environment == "" {
		fmt.Println("Enter P for Production, S for Sandbox: ")
		*environment, _ = reader.ReadString('\n')
		*environment = strings.TrimSpace(*environment)
	}

	if *environment == "P" {
		sandBoxEnv = false
		fmt.Println("Using Production for the environment")
	} else if *environment == "S" {
		fmt.Println("Using Sandbox for the environment")
	} else {
		fmt.Println("Unrecognized value ", *environment, ", enter P or S only")
		return
	}

	if *emailMetadataKey == "" {
		fmt.Println("Enter metadata key for email address")
		*emailMetadataKey, _ = reader.ReadString('\n')
		*emailMetadataKey = strings.TrimSpace(*emailMetadataKey)
	}

	conn := invdapi.NewConnection(*key, sandBoxEnv)

	invoices, err := conn.NewInvoice().ListAll(nil,nil)

	if err != nil {
		fmt.Println(err)
		return
	}

	for _, invoice := range invoices {
		val, ok := invoice.Metadata[*emailMetadataKey]

		if !ok {
			fmt.Println("No email address in custom field to send to for invoice " + invoice.Number)

		}

		emailAddressToSend := val.(string)

		emailRequest := new(invdendpoint.EmailRequest)
		emailDetail := new(invdendpoint.EmailDetail)
		emailDetail.Email = emailAddressToSend
		emailDetail.Name = emailAddressToSend
		emailRequest.To = append(emailRequest.To, *emailDetail)

		_, err := invoice.SendEmail(emailRequest)

		if err != nil {
			fmt.Println("Error sending invoice " + invoice.Number + ", error => " + err.Error())
			continue
		}

		fmt.Println("Email successfully queued for invoice number" + invoice.Number)

	}



}
