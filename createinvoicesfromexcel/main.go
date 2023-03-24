package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/Invoiced/invoiced-go/v2"
	"github.com/Invoiced/invoiced-go/v2/api"
	"github.com/xuri/excelize/v2"
	"os"
	"strconv"
	"strings"
	"time"
)

//This program will create invoices Invoiced; one per row

const DATELAYOUT = "01/02/2006"

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

	conn := api.New(*key, sandBoxEnv)

	f, err := excelize.OpenFile(*fileLocation)

	if err != nil {
		panic(err)
	}

	fmt.Println("Read in excel file ", *fileLocation, ", successfully")

	typeIndex := 3             //column D
	dateIndex := 5             //column F
	memoIndex := 7             //column H
	invoiceNumberIndex := 9    //column J
	customerNameIndex := 11    //column L
	termsIndex := 13           //column N
	dueDateIndex := 15         //column P
	itemIndex := 17            //column R
	itemDescriptionIndex := 19 //column T
	quantityIndex := 21        //column V
	unitCostIndex := 25        //column Z

	rows, err := f.GetRows("Sheet1")

	if err != nil {
		panic("Error trying to get rows for the sheet" + err.Error())
	}

	if len(rows) == 0 {
		fmt.Println("No customer statements to send")
	}

	rows = rows[1:len(rows)]

	for k, row := range rows {

		typeParsed := strings.TrimSpace(row[typeIndex])
		dateParsed := strings.TrimSpace(row[dateIndex])
		memoParsed := strings.TrimSpace(row[memoIndex])
		invoiceNumberParsed := strings.TrimSpace(row[invoiceNumberIndex])
		customerNameParsed := strings.TrimSpace(row[customerNameIndex])
		termsParsed := strings.TrimSpace(row[termsIndex])
		dueDateParsed := strings.TrimSpace(row[dueDateIndex])
		itemParsed := strings.TrimSpace(row[itemIndex])
		itemDescriptionParsed := strings.TrimSpace(row[itemDescriptionIndex])
		quantityParsed := strings.TrimSpace(row[quantityIndex])
		unitCostParsed := strings.TrimSpace(row[unitCostIndex])

		// we don't need to error check these because they can be defaults (0) and it's fine
		quantity, _ := strconv.ParseFloat(quantityParsed, 64)
		unitCost, _ := strconv.ParseFloat(unitCostParsed, 64)

		if typeParsed != "invoice" {
			fmt.Println("Skipping row ", k, " because it is not an invoice")
			continue
		}

		//fetch customer by name
		customer, err := conn.Customer.ListCustomerByName(customerNameParsed)

		if err != nil {
			fmt.Println("Failed to fetch customer by name "+customerNameParsed+" in row => ", k, "; error: "+err.Error())
			continue
		}

		// create invoice
		invoice := new(invoiced.InvoiceRequest)

		invoice.Customer = invoiced.Int64(customer.Id)
		invoice.Number = invoiced.String(invoiceNumberParsed)
		invoice.PaymentTerms = invoiced.String(termsParsed)

		// date format 01/01/1970
		parsedDate, err := time.Parse(dateParsed, DATELAYOUT)

		if err != nil {
			fmt.Println("Failed to parse date in row => ", k, "; error: "+err.Error())
			continue
		}

		invoice.Date = invoiced.Int64(parsedDate.Unix())
		invoice.Notes = invoiced.String(memoParsed)

		// date format 01/01/1970
		parsedDueDate, err := time.Parse(dueDateParsed, DATELAYOUT)

		if err != nil {
			fmt.Println("Failed to parse due date in row => ", k, "; error: "+err.Error())
			continue
		}

		invoice.DueDate = invoiced.Int64(parsedDueDate.Unix())

		// create item and attach to invoice (matches embedded Invoice importer)
		singleItem := make([]*invoiced.LineItemRequest, 1)
		singleItem[0].Name = invoiced.String(itemParsed)
		singleItem[0].Description = invoiced.String(itemDescriptionParsed)
		singleItem[0].UnitCost = invoiced.Float64(unitCost)
		singleItem[0].Quantity = invoiced.Float64(quantity)
		invoice.Items = singleItem

		invResp, err := conn.Invoice.Create(invoice)

		if err != nil {
			fmt.Println("Failed to create invoice for contents of row " + strconv.Itoa(k) + "; error: " + err.Error())
			continue
		}

		fmt.Println("Invoice created successfully: " + invResp.Number)
	}

}
