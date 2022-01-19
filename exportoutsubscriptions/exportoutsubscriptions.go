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

//This program generates a excel file with a export of active subscriptions

func main() {
	//declare and init command line flags
	sandBoxEnv := true
	key := flag.String("key", "", "api key in Settings > Developer")
	environment := flag.String("env", "", "your environment production or sandbox")
	fileLocation := flag.String("file", "", "specify your excel file")
	flag.Parse()

	reader := bufio.NewReader(os.Stdin)

	if *key == "" {
		fmt.Print("Please enter your API Key: ")
		*key, _ = reader.ReadString('\n')
		*key = strings.TrimSpace(*key)
	}

	*environment = strings.ToUpper(strings.TrimSpace(*environment))

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

	filter := invdendpoint.NewFilter()
	filter.Set("canceled", false)
	filter.Set("finished", false)

	fmt.Println("This program generates a excel file with a export of active subscriptions")

	subscriptions, err := conn.NewSubscription().ListAll(filter, nil)

	if err != nil {
		panic(err)
	}

	f := excelize.NewFile()
	// Create a new sheet.
	index := f.NewSheet("Sheet1")
	// Set value of a cell.
	f.SetActiveSheet(index)
	// Save xlsx file by the given path.

	//set headers
	err = f.SetCellValue("Sheet1", "A1", "Subscription ID")

	if err != nil {
		panic(err)
	}
	err = f.SetCellValue("Sheet1", "B1", "Customer Name")

	if err != nil {
		panic(err)
	}
	err = f.SetCellValue("Sheet1", "C1", "CreatedAt")

	if err != nil {
		panic(err)
	}

	err = f.SetCellValue("Sheet1", "D1", "Paused")

	if err != nil {
		panic(err)
	}

	err = f.SetCellValue("Sheet1", "E1", "Plan")

	if err != nil {
		panic(err)
	}

	err = f.SetCellValue("Sheet1", "F1", "Recurring Total")

	if err != nil {
		panic(err)
	}

	err = f.SetCellValue("Sheet1", "G1", "Status")

	if err != nil {
		panic(err)
	}

	for i, subscription := range subscriptions {

		err = f.SetCellValue("Sheet1", "A"+strconv.Itoa(i+2), subscription.Id)

		if err != nil {
			panic(err)
		}

		customer, err := conn.NewCustomer().Retrieve(subscription.Customer)

		if err != nil {
			panic(err)
		}

		if customer == nil {
			panic("Customer with id = " + strconv.FormatInt(subscription.Customer, 10) + ", not found")
		}

		err = f.SetCellValue("Sheet1", "B"+strconv.Itoa(i+2), customer.Name)

		if err != nil {
			panic(err)
		}

		t := time.Unix(subscription.CreatedAt, 0)

		err = f.SetCellValue("Sheet1", "C"+strconv.Itoa(i+2), t.String())

		if err != nil {
			panic(err)
		}

		err = f.SetCellValue("Sheet1", "D"+strconv.Itoa(i+2), subscription.Paused)

		if err != nil {
			panic(err)
		}

		err = f.SetCellValue("Sheet1", "E"+strconv.Itoa(i+2), subscription.Plan)

		if err != nil {
			panic(err)
		}

		err = f.SetCellValue("Sheet1", "F"+strconv.Itoa(i+2), subscription.RecurringTotal)

		if err != nil {
			panic(err)
		}

		err = f.SetCellValue("Sheet1", "G"+strconv.Itoa(i+2), subscription.Status)

		if err != nil {
			panic(err)
		}
	}

	if err := f.SaveAs(*fileLocation); err != nil {
		panic(err)
	}

	fmt.Println("Subscriptions successfully saved to ", *fileLocation)
}
