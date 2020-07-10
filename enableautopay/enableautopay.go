package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/invoiced/invoiced-go"
	"os"
	"strings"
)

//This program add credits to customers in Invoiced

func main() {
	sandBoxEnv := true
	key := flag.String("key", "", "api key in Settings > Developer")
	environment := flag.String("env", "", "your environment production or sandbox")
	fileLocation := flag.String("file", "", "specify your excel file")

	flag.Parse()

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("This program will enable autopay for the invoices specified in the excel file")

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

	if *fileLocation == "" {
		fmt.Println("Please specify your excel file: ")
		*fileLocation, _ = reader.ReadString('\n')
		*fileLocation = strings.TrimSpace(*fileLocation)
	}

	conn := invdapi.NewConnection(*key, sandBoxEnv)

	f, err := excelize.OpenFile(*fileLocation)

	if err != nil {
		panic(err)
	}

	fmt.Println("Read in excel file", *fileLocation, ", successfully")

	invoiceNumberIndex := 0

	rows, err := f.GetRows("Sheet1")

	if err != nil {
		panic("Error trying to get rows for the sheet" + err.Error())
	}

	if len(rows) == 0 {
		fmt.Println("No customer credits to add")
	}

	rows = rows[1:len(rows)]

	for _, row := range rows {

		invoiceNumber := strings.TrimSpace(row[invoiceNumberIndex])

		fetchedInvoice, err := conn.NewInvoice().ListInvoiceByNumber(invoiceNumber)

		if err != nil {
			fmt.Println("Error getting invoice with number => ", invoiceNumber, ", error => ", err)
			continue
		} else if fetchedInvoice == nil {
			fmt.Println("Invoice does not exist with number => ", invoiceNumber, ", error => ", err)
			continue
		} else if fetchedInvoice.AutoPay {
			fmt.Println("Invoice " + invoiceNumber + " already has autopay enabled")
			fmt.Println("Skipping invoice " + invoiceNumber)
			continue

		}

		fetchedCustomer, err := conn.NewCustomer().Retrieve(fetchedInvoice.Customer)

		if err != nil {
			fmt.Println("Error getting customer with number => ", fetchedCustomer, ", error => ", err)
			continue
		} else if fetchedCustomer == nil {
			fmt.Println("Customer does not exist with number => ", fetchedCustomer, ", error => ", err)
			continue
		} else if !fetchedCustomer.AutoPay {
			fmt.Println("Customer " + fetchedCustomer.Name + " does not have autopay enabled")
			continue
		}


		invToUpdate := conn.NewInvoice()
		invToUpdate.Id = fetchedInvoice.Id
		invToUpdate.AutoPay = true

		if fetchedInvoice.Closed == false && fetchedInvoice.Paid == true {
			invToUpdate.Closed = true
		}


		err = invToUpdate.Save()

		if err != nil {
			fmt.Println("Error updating invoice " + invoiceNumber + " got the following error message => "+err.Error())
			continue
		}

		fmt.Println("Successfully enabled autopay for invoice " + invoiceNumber)

	}



}
