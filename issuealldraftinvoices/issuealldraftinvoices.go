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

//This will issue all draft invoices

func main() {
	sandBoxEnv := true
	key := flag.String("key", "", "api key in Settings > Developer")
	environment := flag.String("env", "", "your environment production or sandbox")
	export := flag.Bool("export", false, "Export out issued invoices to excel")

	flag.Parse()

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("This program will issue all draft invoices")

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


	conn := api.New(*key, sandBoxEnv)

	fmt.Println("Fetching draft invoices")

	filter := invoiced.NewFilter()
	filter.Set("status","draft")

	invoices, err := conn.Invoice.ListAll(filter,nil)

	if err != nil {
		panic("could not fetch draft invoices")
	}

	fmt.Println("Fetched ", len(invoices), ", draft invoices to issue.")

	issuedList := make([]string,0)

	for _, invoice := range invoices {
		if invoice.Draft == true {
			fmt.Println("Issuing invoice ", invoice.Number)
			invToUpdate := new(invoiced.InvoiceRequest)

			draft := false
			invToUpdate.Draft = invoiced.Bool(draft)
			_, err := conn.Invoice.Update(invoice.Id,invToUpdate)

			if err != nil {
				fmt.Println("Error updating draft invoice ", invoice.Number)
			} else {
				issuedList = append(issuedList,invoice.Number)
			}
		}

	}

	if  *export && len(issuedList) > 0 {
		now := time.Now().Format(time.RFC822)
		nowParsed := strings.Replace(now, " ","-",-1)
		fileName := "issued"+ "-"+nowParsed + ".xlsx"
		f := excelize.NewFile()
		// Create a new sheet.
		index := f.NewSheet("Sheet1")
		// Set value of a cell.
		f.SetActiveSheet(index)
		// Save xlsx file by the given path.

		err = f.SetCellValue("Sheet1", "A1", "Invoice Number")
		if err != nil {
			panic(err)
		}

		for i, issuedInvoiceNumber := range issuedList {
			err = f.SetCellValue("Sheet1", "A" + strconv.Itoa(i + 2), issuedInvoiceNumber)
			if err != nil {
				panic(err)
			}

		}

		fmt.Println("Saving excel file " + fileName)

		if err := f.SaveAs(fileName); err != nil {
			fmt.Println("Error saving excel file -> ",err)
		}


	}

}
