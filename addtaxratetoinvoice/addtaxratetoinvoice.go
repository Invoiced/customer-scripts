package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/Invoiced/invoiced-go/v2"
	"github.com/Invoiced/invoiced-go/v2/api"
	"github.com/xuri/excelize/v2"
	"os"
	"strings"
)

//This program adds a tax rate (column B) to the invoice in (column A)

func main() {
	sandBoxEnv := true
	key := flag.String("key", "", "api key in Settings > Developer")
	environment := flag.String("env", "", "your environment production or sandbox")
	fileLocation := flag.String("file", "", "specify your excel file")

	flag.Parse()

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("This program will add a tax code to the invoice")

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

	client := api.New(*key, sandBoxEnv)

	f, err := excelize.OpenFile(*fileLocation)

	if err != nil {
		panic(err)
	}

	fmt.Println("Read in excel file ", *fileLocation, ", successfully")

	invoiceNumberIndex := 0
	taxCodeIndex := 1

	rows, err := f.GetRows("Sheet1")

	if err != nil {
		panic("Error trying to get rows for the sheet" + err.Error())
	}

	if len(rows) == 0 {
		fmt.Println("No tax rate data")
	}

	rows = rows[1:len(rows)]

	taxRates, err := client.TaxRate.ListAll(nil, nil)

	if err != nil {
		fmt.Println("Could not fetch tax rates, err=>", err)
		return
	}

	taxRateMap := make(map[string]*invoiced.TaxRate)

	for _, taxRate := range taxRates {
		taxRateMap[taxRate.Id] = taxRate
	}

	for i, row := range rows {

		invoiceNumber := strings.TrimSpace(row[invoiceNumberIndex])
		taxCode := strings.TrimSpace(row[taxCodeIndex])

		invoice, err := client.Invoice.ListInvoiceByNumber(invoiceNumber)

		if err != nil {
			fmt.Println("Error getting invoice with number => ", invoiceNumber, ", error => ", err)
			continue
		} else if invoice == nil {
			fmt.Println("Invoice does not exist with number => ", invoice)
			continue
		}

		fmt.Println("Updating invoice for with number => ", invoiceNumber, "with tax code => ", taxCode)

		taxToAdd := new(invoiced.TaxRequest)

		taxRateToAdd, ok := taxRateMap[taxCode]

		if !ok {
			fmt.Println("Tax rate ", taxCode, ",not found")
			continue
		}

		taxToAdd.TaxRate = taxRateToAdd

		invToUpdateToRequest := new(invoiced.InvoiceRequest)
		invToUpdateToRequest.Taxes = append(invToUpdateToRequest.Taxes, taxToAdd)
		invToUpdateToRequest.Closed = invoiced.Bool(invoice.Closed)

		if invoice.Closed {
			invToUpdateToRequest.Closed = invoiced.Bool(false)
		}

		_, err = client.Invoice.Update(invoice.Id, invToUpdateToRequest)

		if err != nil {
			fmt.Println("Error adding tax to invoice with number => ", invoiceNumber, ", error => ", err)
			continue
		}

		fmt.Println("Successfully added tax")

	}

}
