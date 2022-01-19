package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/Invoiced/invoiced-go/invdendpoint"
	"github.com/Invoiced/invoiced-go"
	"os"
	"strings"
	"time"
)

//This program generates a excel file that reports the total discounts between two dates inclusively

func main() {
	//declare and init command line flags
	prodEnv := true
	key := flag.String("key", "", "api key in Settings > Developer")
	environment := flag.String("env", "", "your environment production or sandbox")
	startdate := flag.String("startdate", "", "Your start date for the discount period in MMDDYYYY format")
	enddate := flag.String("enddate", "", "Your end date for the discount period in MMDDYYYY format")
	fileLocation := flag.String("file", "", "specify your excel file")
	flag.Parse()

	reader := bufio.NewReader(os.Stdin)

	if *key == "" {
		fmt.Print("Please enter your API Key: ")
		*key, _ = reader.ReadString('\n')
		*key = strings.TrimSpace(*key)
	}

	*environment = strings.ToUpper(strings.TrimSpace(*environment))

	if *environment == "P" || strings.Contains(*environment, "PRODUCTION") {
		prodEnv = false
		fmt.Println("Using Production for the environment")
	} else if *environment == "S" || strings.Contains(*environment, "SANDBOX") {
		fmt.Println("Using Sandbox for the environment")
	} else {
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

	}

	if *fileLocation == "" {
		fmt.Println("Please specify your excel file: ")
		*fileLocation, _ = reader.ReadString('\n')
		*fileLocation = strings.TrimSpace(*fileLocation)
	}

	beginTime := time.Time{}
	endTime := time.Time{}

	if *startdate == "" {
		for {
			fmt.Print("Please enter your start date in MMDDYYYY format: ")
			*startdate, _ = reader.ReadString('\n')
			*startdate = strings.TrimSpace(*startdate)

			loc := time.Now().Location()
			var err error

			beginTime, err = time.ParseInLocation("01022006", *startdate, loc)

			if err == nil {
				break
			}

		}
	}

	if *enddate == "" {
		for {
			fmt.Print("Please enter your end date in MMDDYYYY format: ")
			*enddate, _ = reader.ReadString('\n')
			*enddate = strings.TrimSpace(*enddate)

			loc := time.Now().Location()
			var err error
			endTime, err = time.ParseInLocation("01022006", *enddate, loc)

			if err == nil {
				break
			}
		}
	}

	fmt.Println(beginTime.Unix())
	fmt.Println(endTime.Unix())
	fmt.Println(prodEnv)

	conn := invdapi.NewConnection(*key, prodEnv)

	filter := invdendpoint.NewFilter()
	filter.Set("draft", 0)
	filter.Set("voided", 0)

	fmt.Println("Getting invoices ...")
	invoices, err := conn.NewInvoice().ListAllInvoicesStartEndDate(filter, nil, beginTime.Unix(), endTime.Unix())

	if err != nil {
		panic(err)
	}

	discountTotal := float64(0.0)

	for _, invoice := range invoices {

		items := invoice.Items

		for _, item := range items {
			for _, discount := range item.Discounts {
				discountTotal += discount.Amount
			}
		}

		for _, discount := range invoice.Discounts {
			discountTotal += discount.Amount
		}

	}

	fmt.Println("Total discount is ", discountTotal)
	fmt.Println("Saving data to", *fileLocation)

	f := excelize.NewFile()
	// Create a new sheet.
	index := f.NewSheet("Sheet1")
	// Set value of a cell.
	f.SetActiveSheet(index)
	// Save xlsx file by the given path.

	err = f.SetCellValue("Sheet1", "A1", "Discount Total")
	if err != nil {
		panic(err)
	}
	err = f.SetCellValue("Sheet1", "B1", "Start Date")
	if err != nil {
		panic(err)
	}
	err = f.SetCellValue("Sheet1", "C1", "End Date")
	if err != nil {
		panic(err)
	}

	err = f.SetCellValue("Sheet1", "A2", discountTotal)
	if err != nil {
		panic(err)
	}
	err = f.SetCellValue("Sheet1", "B2", *startdate)
	if err != nil {
		panic(err)
	}
	err = f.SetCellValue("Sheet1", "C2", *enddate)
	if err != nil {
		panic(err)
	}

	if err := f.SaveAs(*fileLocation); err != nil {
		panic(err)
	}
}
