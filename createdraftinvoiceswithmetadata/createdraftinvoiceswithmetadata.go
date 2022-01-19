package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/Invoiced/invoiced-go/invdendpoint"
	invdapi "github.com/Invoiced/invoiced-go"
)

//This program will create credit memos on Invoiced; one per row
// rows must be of the form:
// [0]: {customer_number}
// [1]: {invoice_number}
// [2]: {quantity}
// [3]: {unit_cost}
// [4]: {metadata} of the form field1=value1&field2=value2 etc.

func main() {
	sandBoxEnv := true
	key := flag.String("key", "", "api key in Settings > Developer")
	environment := flag.String("env", "", "your environment production or sandbox")
	fileLocation := flag.String("file", "", "specify your excel file")

	flag.Parse()

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("This program will create invoices with metadata based on the excel file")

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

	fmt.Println("Read in excel file ", *fileLocation, ", successfully")

	customerNumberIndex := 0
	invoiceNumberIndex := 1
	quantityIndex := 2
	unitCostIndex := 3
	metadataIndex := 4

	rows, err := f.GetRows("Sheet1")

	if err != nil {
		panic("Error trying to get rows for the sheet" + err.Error())
	}

	if len(rows) == 0 {
		fmt.Println("No customer statements to send")
	}

	rows = rows[1:len(rows)]

	for k, row := range rows {

		customerParsed := strings.TrimSpace(row[customerNumberIndex])
		invoiceParsed := strings.TrimSpace(row[invoiceNumberIndex])
		quantityParsed := strings.TrimSpace(row[quantityIndex])
		unitCostParsed := strings.TrimSpace(row[unitCostIndex])
		metadataParsed := strings.TrimSpace(row[metadataIndex])

		// we don't need to error check these because they can be defaults (0) and it's fine
		quantity, _ := strconv.ParseFloat(quantityParsed, 64)
		unitCost, _ := strconv.ParseFloat(unitCostParsed, 64)

		customerNumber, err := strconv.Atoi(customerParsed)

		if err != nil {
			fmt.Println("Error parsing customer number " + customerParsed + "; skipping")
			continue
		}

		// create prototype invoice in draft
		invoice := conn.NewInvoice()
		invoice.Draft = true
		invoice.Customer = int64(customerNumber)
		invoice.Number = invoiceParsed

		// create item and attach to invoice (matches embedded Invoice importer)
		singleItem := make([]invdendpoint.LineItem, 1)
		singleItem[0].UnitCost = unitCost
		singleItem[0].Quantity = quantity
		invoice.Items = singleItem

		// deal with metadata. we expect metadata in a _single_ cell of the form:
		// field1=value1&field2=value2 (like HTML query parameters; no leading `?`)
		metadataKeyValues := strings.Split(metadataParsed, "&")

		var metadataNestedSlices [][]string

		for _, v := range metadataKeyValues {
			metadataNestedSlices = append(metadataNestedSlices,
				strings.Split(v, "="))
		}

		finalMetadata := make(map[string]interface{})

		for _, pair := range metadataNestedSlices {
			finalMetadata[pair[0]] = pair[1]
		}

		invoice.Metadata = finalMetadata

		invoice, err = invoice.Create(invoice)

		if err != nil {
			fmt.Println("Failed to create invoice for contents of row " + strconv.Itoa(k) + "; error: " + err.Error())
			continue
		}

		fmt.Println("Invoice created successfully: " + invoice.Number)
	}

}
