package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/Invoiced/invoiced-go/invdendpoint"
	"github.com/Invoiced/invoiced-go"
	"os"
	"strconv"
	"strings"
	"time"
)

//This program will create credit memos on Invoiced; one per row
// rows must be of the form:
// [0]: {invoice_number}
// [1]: {amount_to_credit}

func main() {
	sandBoxEnv := true
	key := flag.String("key", "", "api key in Settings > Developer")
	environment := flag.String("env", "", "your environment production or sandbox")
	fileLocation := flag.String("file", "", "specify your excel file")

	flag.Parse()

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("This program will create credit memos to partially pay off invoices from the excel file")

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
	creditMemoQtyColumn := 6
	creditMemoUnitCostColumn := 7
	creditMemoLineItemNameColumn := 4
	creditMemoDescriptionColumn := 5
	creditMemoNumberColumn := 1
	creditMemoDateColumn := 2

	rows, err := f.GetRows("Sheet1")

	if err != nil {
		panic("Error trying to get rows for the sheet" + err.Error())
	}

	if len(rows) == 0 {
		fmt.Println("No invoice numbers in excel sheet to create credit memos")
	}

	rows = rows[1:len(rows)]

	creditNoteMap := make(map[string]*invdapi.CreditNote)

	for _, row := range rows {

		customerNumberParsed := strings.TrimSpace(row[customerNumberIndex])
		creditMemoQtyRaw := strings.TrimSpace(row[creditMemoQtyColumn])
		creditMemoUnitCostRaw := strings.TrimSpace(row[creditMemoUnitCostColumn])
		creditMemoNumber := strings.TrimSpace(row[creditMemoNumberColumn])
		creditMemoLineItemID := strings.TrimSpace(row[creditMemoLineItemNameColumn])
		creditMemoDescription := strings.TrimSpace(row[creditMemoDescriptionColumn])
		creditMemoDate := strings.TrimSpace(row[creditMemoDateColumn])

		if len(creditMemoDate) == 7 {
			creditMemoDate = "0" + creditMemoDate
		}



		creditNote, ok := creditNoteMap[creditMemoNumber]

		if !ok {
			customer, err := conn.NewCustomer().ListCustomerByNumber(customerNumberParsed)

			if err != nil {
				fmt.Println("Error retrieving customer for credit note;" +
					"credit note #" + creditMemoNumber + " does not exist")
				continue
			}

			if customer == nil {
				fmt.Println("Error retrieving customer for credit note;" +
					"credit note #" + creditMemoNumber + " does not exist")
				continue
			}

			creditNote = conn.NewCreditNote()
			creditNote.Customer = customer.Id
			if len(creditMemoNumber) > 0 {
				creditNote.Number = creditMemoNumber
			}

			tm, err := time.Parse("01-02-06",creditMemoDate)

			if err != nil {
				fmt.Println("Could note parse date" + err.Error())
				continue
			}

			creditNote.Date = tm.Unix()
		}

		fmt.Println("Adding line items for credit memo ",creditMemoNumber)

		//03/24/23
		creditMemoQty, err := strconv.ParseFloat(creditMemoQtyRaw,64)

		if err != nil {
			fmt.Println("Error parsing qty value," + creditMemoQtyRaw + ",for" +
				"credit memo #"+ customerNumberParsed)
			continue
		}

		creditMemoUnitCost, err := strconv.ParseFloat(creditMemoUnitCostRaw,64)

		if err != nil {
			fmt.Println("Error parsing unit cost value," + creditMemoUnitCostRaw + ",for" +
				"invoice #"+ customerNumberParsed)
			continue
		}

		creditMemoQty = 1
		// create simplified items to pass into credit note
		// if we don't do this, request will fail
		item := new(invdendpoint.LineItem)
		item.Name = creditMemoLineItemID
		item.Description = creditMemoDescription
		item.Quantity = creditMemoQty
		item.UnitCost = creditMemoUnitCost

		creditNote.Items = append(creditNote.Items, *item)

		if !ok {
			creditNoteMap[creditMemoNumber] = creditNote
		}


	}

   fmt.Println("Number of credit notes to create => ", len(creditNoteMap))
	for creditMemoNumber, creditNote := range creditNoteMap {

				_, err = creditNote.Create(creditNote)

				if err != nil {
					fmt.Println("Error creating credit note for invoice " + creditMemoNumber +
						" - error: " + err.Error())
					continue
				}

				fmt.Println("Successfully created & issued credit note for invoice " + creditMemoNumber)


	}


}
