package main

import (
	"bufio"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/invoiced/invoiced-go"
	"os"
	"strings"
)

//This program will mark the customers in the excel sheet as paid, by adding payments to close out the open invoices.

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

	fmt.Println("Please confirm, this program is about add payments to all of those invoices specified in the excel file, please type in YES to continue: ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(confirm)

	if confirm != "YES" {
		fmt.Println("Halting program, sequence not confirmed")
		return
	}

	conn := invdapi.NewConnection(key, prodEnv)

	for _, row := range rows {

		invoiceNumber := strings.TrimSpace(row[columnIndex])

		fmt.Println("Getting invoice with number => ", invoiceNumber)

		inv, err := conn.NewInvoice().ListInvoiceByNumber(invoiceNumber)

		if err != nil {
			fmt.Println("Error getting invoice with number => ", invoiceNumber, ", error => ", err)
			continue
		}

		if inv == nil {
			fmt.Println("Invoice does not exist =>", invoiceNumber)
			continue
		}

		fmt.Println("Successfully got invoice with number => ", invoiceNumber)

		balance := inv.Balance

		if inv.Closed {
			fmt.Println("Invoice ", invoiceNumber, " is closed, so we are skipping the creation of a payment")
			continue
		}

		if balance == 0 {
			fmt.Println("Creating a payment is not necessary since the balance is 0, so we are skipping the creation of a payment")
			continue
		}

		transactionToCreate := conn.NewTransaction()
		transactionToCreate.Invoice = inv.Id
		transactionToCreate.Amount = balance
		transactionToCreate.Type = "payment"
		transactionToCreate.Notes = "This payment is to close out the invoice"

		createdTransaction, err := conn.NewTransaction().Create(transactionToCreate)

		if err != nil {
			fmt.Println("Error creating payment with invoice with number => ", invoiceNumber)
			continue
		}

		fmt.Println("Successfully created payment with ID => ", createdTransaction.Id)

	}

}
