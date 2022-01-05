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

//This program will update customer numbers

func main() {
	sandBoxEnv := true
	key := flag.String("key", "", "api key in Settings > Developer")
	environment := flag.String("env", "", "your environment production or sandbox")
	fileLocation := flag.String("file", "", "specify your excel file")

	flag.Parse()

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("This program will update the customer numbers in the excel sheet")

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

	customerNumberOriginalIndex := 0
	customerNumberNewIndex := 1

	rows, err := f.GetRows("Sheet1")

	if err != nil {
		panic("Error trying to get rows for the sheet" + err.Error())
	}

	if len(rows) == 0 {
		fmt.Println("No customer numbers to update")
	}

	rows = rows[1:len(rows)]

	for i, row := range rows {

		if len(row) < 2 {
			fmt.Println("Skipping updating customer, because we require both the old and new customer number for customer with number for row = ",i)
			continue
		}



		customerNumberOriginal := strings.TrimSpace(row[customerNumberOriginalIndex])
		customerNumberNew := strings.TrimSpace(row[customerNumberNewIndex])

		if len(customerNumberNew) == 0 {
			fmt.Println("Skipping updating customer, because new customer number is blank")
			continue
		}

		customerOriginal, err := conn.NewCustomer().ListCustomerByNumber(customerNumberOriginal)

		if err != nil {
			fmt.Println("Error getting customer with number => ", customerNumberOriginal, ", error => ", err)
			continue
		} else if customerOriginal == nil {
			fmt.Println("Customer does not exist with number => ", customerOriginal, ", error => ", err)
			continue
		}

		fmt.Println("Updating customer number with number => ", customerNumberOriginal)

		customerNew := customerOriginal.NewCustomer()
		customerNew.Id = customerOriginal.Id
		customerNew.Number = customerNumberNew

		err = customerNew.Save()

		if err != nil {
			fmt.Println("Error updating customer number => ", customerNumberOriginal, ", error => ", err)
			continue
		}

		fmt.Println("Successfully updated customer number => ", customerNumberOriginal, ", set new number to ", customerNumberNew)

	}



}
