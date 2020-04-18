package main

import (
	"bufio"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/Invoiced/invoiced-go/invdendpoint"
	"github.com/invoiced/invoiced-go"
	"os"
	"strings"
)

//This program marks all the invoices bad debt under the customer.  It looks up the customer by the customer name provided in the excel sheet.

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Please enter your API Key: ")
	prodEnv := true
	key, _ := reader.ReadString('\n')
	key = strings.TrimSpace(key)
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

	fmt.Println("Is this a Production connection? => ", prodEnv)

	fmt.Println("Please specify your excel file: ")
	fileLocation, _ := reader.ReadString('\n')

	fileLocation = strings.TrimSpace(fileLocation)

	f, err := excelize.OpenFile(fileLocation)

	if err != nil {
		panic(err)
	}

	fmt.Println("Read in excel file ", fileLocation, ", successfully")

	columnIndex := 0

	rows, err := f.GetRows("Sheet1")

	if err != nil {
		panic("Error trying to get rows for the sheet" + err.Error())
	}

	fmt.Println("Please confirm, this program is about mark the invoices as bad debt, specified by the customers in the excel file, please type in YES to continue: ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(confirm)

	if confirm != "YES" {
		fmt.Println("Halting program, sequence not confirmed")
		return
	}

	conn := invdapi.NewConnection(key, prodEnv)

	for _, row := range rows {

		customerName := strings.TrimSpace(row[columnIndex])

		fmt.Println("Getting customer with name => ", customerName)

		customerFilter := invdendpoint.NewFilter()
		customerFilter.Set("name", customerName)

		customers, err := conn.NewCustomer().ListAll(customerFilter, nil)

		if err != nil {
			fmt.Println("Error getting customer with name => ", customerName, ", error => ", err)
			continue
		}

		if customers == nil {
			fmt.Println("Customer does not exist =>", customerName)
			continue
		}

		if len(customers) == 0 {
			fmt.Println("Customer does not exist =>", customerName)
			continue
		}

		fmt.Println("Successfully got customer with name => ", customerName)

		fmt.Println("Now getting the associated invoices")

		filter := invdendpoint.NewFilter()
		filter.Set("customer", customers[0].Id)

		invoices, err := conn.NewInvoice().ListAll(filter, nil)

		if err != nil {
			fmt.Println("Error getting invoices ", err)
			continue
		}

		for _, invoice := range invoices {
			if !invoice.Closed {
				invToUpdate := conn.NewInvoice()
				invToUpdate.Id = invoice.Id
				invToUpdate.Closed = true
				err := invToUpdate.Save()

				if err != nil {
					fmt.Println("Error closing invoice => ", invoice.Number, ", error message => ", err)
				}

				fmt.Println("Successfully closed invoice ", invoice.Number)

			}

		}
	}

}
