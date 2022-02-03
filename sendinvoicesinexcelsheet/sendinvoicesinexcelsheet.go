package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/xuri/excelize/v2"
	"github.com/Invoiced/invoiced-go/v2/api"
	"os"
	"strings"
)

//This program will send all invoices in the excel sheet

func main() {
	//declare and init command line flags
	prodEnv := true
	key := flag.String("key", "", "api key in Settings > Developer")
	environment := flag.String("env", "", "your environment production or sandbox")
	fileLocation := flag.String("file", "", "specify your excel file")
	flag.Parse()

	reader := bufio.NewReader(os.Stdin)

	if *key == "" {
		fmt.Print("Please enter your API Key: ")
		*key, _ = reader.ReadString('\n')
		*key = strings.TrimSpace(*key)
	}

	*environment = strings.ToUpper(strings.TrimSpace(*environment))

	if *environment == "P" || strings.Contains(*environment, "PRODUCTION") {
		prodEnv = false
		fmt.Println("Using Production for the environment")
	} else if *environment == "S" || strings.Contains(*environment, "SANDBOX") {
		fmt.Println("Using Sandbox for the environment")
	} else {
		for {

			fmt.Println("What is your environment, please enter P for production or S for sandbox: ")
			env, _ := reader.ReadString('\n')
			env = strings.ToUpper(strings.TrimSpace(env))

			if env == "P" || strings.Contains(env, "PRODUCTION") {
				prodEnv = false
				fmt.Println("Using Production for the environment")
				break
			} else if env == "S" || strings.Contains(env, "SANDBOX") {
				fmt.Println("Using Sandbox for the environment")
				break
			}
		}

	}

	if *fileLocation == "" {
		fmt.Println("Please specify your excel file: ")
		*fileLocation, _ = reader.ReadString('\n')
		*fileLocation = strings.TrimSpace(*fileLocation)
	}

	fmt.Println("Opening Excel File => ", *fileLocation)
	f, err := excelize.OpenFile(*fileLocation)

	if err != nil {
		panic(err)
	}

	fmt.Println("Read in excel file ", *fileLocation, ", successfully")

	columnIndex := 0

	rows, err := f.GetRows("Sheet1")

	if err != nil {
		panic("Error trying to get rows for the sheet" + err.Error())
	}


	conn := api.New(*key, prodEnv)

	for i, row := range rows {

		if i == 0 {
			continue
		}

		invoiceNumber := strings.TrimSpace(row[columnIndex])

		fmt.Println("Getting invoice with number => ", invoiceNumber)

		inv, err := conn.Invoice.ListInvoiceByNumber(invoiceNumber)

		if err != nil {
			fmt.Println("Error getting invoice with number => ", invoiceNumber, ", error => ", err)
			continue
		}

		if inv == nil {
			fmt.Println("Invoice does not exist =>", invoiceNumber)
			continue
		}

		if inv.Status == "draft" || inv.Status == "pending" || inv.Status == "paid" || inv.Status == "voided" {
			fmt.Println("Invoice is already in draft, pending, paid, voided, skipping and moving on to next invoice ...")
			continue
		}

		fmt.Println("Sending invoice with number => ", invoiceNumber)

		err = conn.Invoice.SendEmail(inv.Id,nil)

		if err != nil {
			fmt.Println("Could not send out the invoice due the following error => ", err)
			continue
		}

		fmt.Println("Successfully queued invoice => ", inv.Number, "for sending, it should be sent soon. ")

	}

}
