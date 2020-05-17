package main

import (
"bufio"
"flag"
"fmt"
"github.com/360EntSecGroup-Skylar/excelize"
"github.com/invoiced/invoiced-go"
"os"
	"strconv"
	"strings"
)

//This program add credits to customers in Invoiced

func main() {
	sandBoxEnv := true
	key := flag.String("key", "", "api key in Settings > Developer")
	environment := flag.String("env", "", "your environment production or sandbox")
	fileLocation := flag.String("file", "", "specify your excel file")

	flag.Parse()

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("This program will add credits to customers in the excel file")

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

	fmt.Println("Read in excel file", *fileLocation, ", successfully")

	customerNumberIndex := 0
	creditAmtToAddIndex := 1

	rows, err := f.GetRows("Sheet1")

	if err != nil {
		panic("Error trying to get rows for the sheet" + err.Error())
	}

	if len(rows) == 0 {
		fmt.Println("No customer credits to add")
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

		creditAmtToAddVal := strings.TrimSpace(row[creditAmtToAddIndex])
		creditAmt, err := strconv.ParseFloat(creditAmtToAddVal,64)

		if err != nil {
			fmt.Println("Could not process the credit for",-1*creditAmt, ", for customer => ",customerNumber)
			continue
		}

		if creditAmt >= 0 {
			fmt.Println("Skipping adding credit cause the credit amount is positive or equals zero for customer => ",customerNumber)
			continue
		}

		fmt.Println("Going to add credit for customer with number", customerNumber,"for amount => ",1*creditAmt)

		transactionToCreate := conn.NewTransaction()
		transactionToCreate.Amount = creditAmt
		transactionToCreate.Type = "adjustment"
		transactionToCreate.Customer = customer.Id

		_, err = conn.NewTransaction().Create(transactionToCreate)

		if err != nil && !strings.Contains(err.Error(),"cannot unmarshal string into Go struct field Transaction.id of type int64") {
			fmt.Println("Error creating credit for customer with number => ", customerNumber, ", error => ", err)
			continue
		}

		fmt.Println("Successfully created credit for customer with number", customerNumber,"for amount => ",1*creditAmt)

	}



}
