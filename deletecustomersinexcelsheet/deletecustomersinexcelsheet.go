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

//This program will delete customers in an excel sheet

func main() {
	sandBoxEnv := true
	key := flag.String("key", "", "api key in Settings > Developer")
	environment := flag.String("env", "", "your environment production or sandbox")
	fileLocation := flag.String("file", "", "specify your excel file")

	flag.Parse()

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("This program will delete customers specified in the excel sheet")

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

	fmt.Println("Please confirm, this program is about to delete all of the customers specified in the excel sheet, please type in YES to continue: ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(confirm)

	if confirm != "YES" {
		fmt.Println("Halting program, confirm sequence not confirmed")
		return
	}

	conn := invdapi.NewConnection(*key, sandBoxEnv)

	f, err := excelize.OpenFile(*fileLocation)

	if err != nil {
		panic(err)
	}

	fmt.Println("Read in excel file ", *fileLocation, ", successfully")

	customerNumberIndex := 0

	rows, err := f.GetRows("Sheet1")

	if err != nil {
		panic("Error trying to get rows for the sheet" + err.Error())
	}

	if len(rows) == 0 {
		fmt.Println("No customers in excel sheet to delete")
	}

	rows = rows[1:len(rows)]

	for _, row := range rows {

		customerNumber := strings.TrimSpace(row[customerNumberIndex])

		customer, err := conn.NewCustomer().ListCustomerByNumber(customerNumber)

		if err != nil {
			fmt.Println("Error getting customer with number => ", customerNumber, ", error => ", err)
			continue
		} else if customer == nil {
			fmt.Println("Customer does not exist with number => ", customerNumber, ", error => ", err)
			continue
		}

		fmt.Println("Deleting customer with number =>", customerNumber)

		err = customer.Delete()

		if err != nil {
			fmt.Println("Error deleting customer with number =>", customerNumber, ", error => ", err)
			continue
		}

		fmt.Println("Deleted customer with number =>", customerNumber)

	}



}
