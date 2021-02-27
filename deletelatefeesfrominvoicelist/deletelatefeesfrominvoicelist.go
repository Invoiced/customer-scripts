package main

import (
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/Invoiced/invoiced-go"
	"github.com/Invoiced/invoiced-go/invdendpoint"
	"strconv"
	"strings"
	"flag"
	"bufio"
	"os"
)

const sheet = "Sheet1"

func main() {
	sandBoxEnv := true
	key := flag.String("key", "", "api key in Settings > Developer")
	environment := flag.String("env", "", "your environment production or sandbox")
	fileLocation := flag.String("file", "", "specify your excel file")

	flag.Parse()

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("This program will remove late fees from the specified invoices")

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

	*fileLocation = strings.TrimSpace(*fileLocation)

	f, err := excelize.OpenFile(*fileLocation)

	if err != nil {
		panic(err)
	}

	fmt.Println("Read in excel file ", fileLocation, ", successfully")

	invoiceNumber:= "A"

	rows, err := f.GetRows(sheet)

	if err != nil {
		panic("Error trying to get rows for the sheet" + err.Error())
	}

	conn := invdapi.NewConnection(*key, sandBoxEnv)
	invoiceConn := conn.NewInvoice()

	for i, row := range rows {

		if i == 0 {
			fmt.Println("Skipping header row =>",row)
			fmt.Println(row)
			continue
		}

		invoiceNumber, err := f.GetCellValue(sheet,invoiceNumber + strconv.Itoa(i + 1))

		if err != nil {
			fmt.Println("Error getting invoice number for row = ", i, ", error => ",err)
			continue
		}

		invoiceNumber = strings.TrimSpace(invoiceNumber)

		invoiceFetched, err := invoiceConn.ListInvoiceByNumber(invoiceNumber)

		if err != nil {
			fmt.Println("Error getting invoice from Invoiced for invoice number " + invoiceNumber)
			continue
		} else {
			fmt.Println("Successfully fetched invoice number# ",invoiceNumber)
		}

		lateFeePresent := false
		lineItems := invoiceFetched.Items

		lineItemWithOutLateFees := make([]invdendpoint.LineItem, 0)

		for _, item := range lineItems {
			if item.Type != "late_fee" {
				lineItemWithOutLateFees = append(lineItemWithOutLateFees, item)
			} else {
				lateFeePresent = true
			}
		}

		if lateFeePresent {
			invdInvoiceToUpdate := conn.NewInvoice()
			invdInvoiceToUpdate.Id = invoiceFetched.Id
			invdInvoiceToUpdate.Items = lineItemWithOutLateFees

			err := invdInvoiceToUpdate.Save()

			if err != nil {
				fmt.Println("Error updating invoice number ", invoiceFetched.Number, ",error => ", err)
			} else {
				fmt.Println("Successfully removed late fees for Invoice ")
			}
		} else {
			fmt.Println("No late fees for Invoice ", invoiceFetched.Number)
		}



	}


}
