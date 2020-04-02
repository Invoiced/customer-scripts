package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/Invoiced/invoiced-go/invdendpoint"
	"github.com/invoiced/invoiced-go"
	"os"
	"strings"
)

//This program will copy invoices & payments from customer to customer

func main() {
	sandBoxEnv := true
	key := flag.String("key", "", "api key in Settings > Developer")
	environment := flag.String("env", "", "your environment production or sandbox")
	fileLocation := flag.String("file", "", "specify your excel file")

	flag.Parse()

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("This program will copy invoices & payments from customer in column 1 to customer in column 2 in the excel sheet.")

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

	fmt.Println("Read in excel file",strings.TrimSpace(*fileLocation),",successfully")

	customerFromIndex := 0
	customerToIndex := 1

	rows, err := f.GetRows("Sheet1")

	if err != nil {
		panic("Error trying to get rows for the sheet" + err.Error())
	}

	if len(rows) == 0 {
		fmt.Println("No subscription data to update")
	}

	rows = rows[1:len(rows)]

	for _, row := range rows {

		customerFromNumber := row[customerFromIndex]
		customerToNumber := row[customerToIndex]

		customerFrom, err := conn.NewCustomer().ListCustomerByNumber(customerFromNumber)

		if err != nil {
			fmt.Println("Error in customer from =>" ,err)
			continue
		}

		if customerFrom == nil {
			fmt.Println("Could not fetch customer")
			continue
		}

		customerTo, err := conn.NewCustomer().ListCustomerByNumber(customerToNumber)

		if err != nil {
			fmt.Println("Error in customer to =>" ,err)
			continue
		}

		if customerTo == nil {
			fmt.Println("Could not fetch customer")
			continue
		}

		filter := invdendpoint.NewFilter()
		filter.Set("customer",customerFrom.Id)
		filter.Set("closed",false)
		filter.Set("draft",false)
		invoicesCustomerFrom, err := conn.NewInvoice().ListAll(filter,nil)

		if err != nil {
			fmt.Println("Error in customer to =>" ,*customerTo)
			continue
		}

		for _, invoice := range invoicesCustomerFrom {
			if invoice.Status != "voided" {
				//fetch invoice to make sure we don't recreate it
				fetchInv, err := conn.NewInvoice().ListInvoiceByNumber(invoice.Number + "CP")

				if err != nil {
					fmt.Println("Error fetching invoice")
					continue
				}

				if fetchInv != nil {
					fmt.Println("Found invoice", invoice.Number + "CP,","skipping transferring it")
					continue
				}

				if fetchInv == nil {
					//create invoice if it does not exist
					lineItems := make([]invdendpoint.LineItem, 0)
					for _, lineItem := range invoice.Items {
						lineItem.Id = 0
						lineItems = append(lineItems, lineItem)
					}
					invoice.Items = lineItems
					fmt.Println("customer to -> ", customerTo.Id)
					invToCreate := conn.NewInvoice()
					invToCreate.Invoice = invoice.Invoice
					invToCreate.Customer = customerTo.Id
					invToCreate.Number = invToCreate.Number + "CP"

					createdInv, err := conn.NewInvoice().Create(invToCreate)

					if err != nil {
						fmt.Println("error -> ", err)
					} else {

						fmt.Println("Successfully transferred invoice from customer", customerFromNumber, ", to customer ", customerToNumber)
						fmt.Println("Invoice transferred was", createdInv.Number)
					}

					filterTransaction := invdendpoint.NewFilter()
					filterTransaction.Set("invoice",invoice.Id)
					filterTransaction.Set("type","payment")
					filterTransaction.Set("status","succeeded")

					transactions, err := conn.NewTransaction().ListAll(filterTransaction,nil)

					if err != nil {
						fmt.Println("error getting transactions for invoice ",invoice.Number)
						continue
					}

					for _, transaction := range transactions {
						transaction.Invoice = createdInv.Id
						transaction.Customer = createdInv.Customer
						transaction.ParentTransaction = 0
						tranactionToCreate := conn.NewTransaction()
						tranactionToCreate.Transaction = transaction.Transaction
						createdTrans, err := conn.NewTransaction().Create(tranactionToCreate)

						if err != nil {
							fmt.Println("error creating transaction for invoice ",createdInv.Number)
							continue
						}

						fmt.Println("Created transaction successfully for ", createdTrans.Transaction)

					}



				}

			}
		}




	}

}
