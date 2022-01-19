package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/Invoiced/invoiced-go"
	"os"
	"strings"
)

//This program will create update metadata for invoices"; one per row

func main() {
	sandBoxEnv := true
	key := flag.String("key", "", "api key in Settings > Developer")
	environment := flag.String("env", "", "your environment production or sandbox")
	fileLocation := flag.String("file", "", "specify your excel file")

	flag.Parse()

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("This program will create update metadata for invoices")

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

	invoiceNumberIndex := 0
	maxCustomFieldIndex := 11

	rows, err := f.GetRows("Sheet1")

	if err != nil {
		panic("Error trying to get rows for the sheet" + err.Error())
	}

	if len(rows) == 0 {
		fmt.Println("No invoice numbers in excel sheet to create credit memos")
	}

	//calculate the number of columns to process
	rows = rows[0:len(rows)]

	//make a map of columns to custom field values
	customFieldMap := make(map[int]string)

	for i, row := range rows {
		if i == 0 {
			if maxCustomFieldIndex > len(row) {
				maxCustomFieldIndex = len(row)
			}

			for j := 1; j < maxCustomFieldIndex; j++ {
				//skip the first column, since it contains the invoice number
				metaDataFieldKey := strings.TrimSpace(row[j])
				fmt.Println(metaDataFieldKey)
				if len(metaDataFieldKey) > 0 {
					customFieldMap[j] = metaDataFieldKey
				}

			}
		} else {

			invoiceNumberParsed := strings.TrimSpace(row[invoiceNumberIndex])

			invoice, err := conn.NewInvoice().ListInvoiceByNumber(invoiceNumberParsed)

			if err != nil {
				fmt.Println("Error retrieving." +
					"invoice #" + invoiceNumberParsed + " does not exist")
				continue
			}

			if invoice == nil {
				fmt.Println("Error retrieving," +
					"invoice #" + invoiceNumberParsed + " does not exist")
				continue
			}

			invoiceMetaData := invoice.Metadata

			if invoiceMetaData == nil {
				invoiceMetaData = make(map[string]interface{})
			}

			for j := 1; j <= maxCustomFieldIndex; j++ {
				val, ok := customFieldMap[j]

				if ok {
					invoiceMetaData[val] = strings.TrimSpace(row[j])
				}
			}

			invToUpdate := conn.NewInvoice()
			invToUpdate.Id = invoice.Id
			invToUpdate.Metadata = invoiceMetaData

			err = invToUpdate.Save()

			if err != nil {
			} else {
				fmt.Println("Successfully updated Metadata for invoice " + invoiceNumberParsed)
			}

		}

	}

}
