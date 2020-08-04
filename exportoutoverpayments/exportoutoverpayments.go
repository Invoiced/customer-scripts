package main


import (
"bufio"
"flag"
"fmt"
"github.com/360EntSecGroup-Skylar/excelize"
"github.com/Invoiced/invoiced-go/invdendpoint"
"github.com/invoiced/invoiced-go"
"os"
"strconv"
"strings"
"time"
)

//This program export out invoice payments

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
	filter.Set("type", "adjustment")
	filter.Set("method", "balance")

	fmt.Println("This program generates a excel file with a export of over payments")

	transactions, err := conn.NewTransaction().ListAll(filter, nil)

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
	err = f.SetCellValue("Sheet1", "A1", "Customer ID")

	if err != nil {
		panic(err)
	}
	err = f.SetCellValue("Sheet1", "B1", "Payment Date")

	if err != nil {
		panic(err)
	}

	err = f.SetCellValue("Sheet1", "C1", "Invoiced Amount")

	if err != nil {
		panic(err)
	}

	err = f.SetCellValue("Sheet1", "D1", "Check Number")

	if err != nil {
		panic(err)
	}

	err = f.SetCellValue("Sheet1", "E1", "Location ID")

	if err != nil {
		panic(err)
	}


	for i, transaction := range transactions {

		if transaction.Notes != "Overpayment" {
			continue
		}

		err = f.SetCellValue("Sheet1", "A"+strconv.Itoa(i+2), transaction.Customer)

		if err != nil {
			fmt.Println(err)
			continue
		}


		err = f.SetCellValue("Sheet1", "B"+strconv.Itoa(i+2), time.Unix(transaction.Date,0).String())

		if err != nil {
			fmt.Println(err)
			continue
		}

		err = f.SetCellValue("Sheet1", "C"+strconv.Itoa(i+2), transaction.Amount*-1)

		if err != nil {
			fmt.Println(err)
			continue
		}


		err = f.SetCellValue("Sheet1", "D"+strconv.Itoa(i+2), transaction.GatewayId)

		if err != nil {
			fmt.Println(err)
			continue
		}

	   if transaction.ParentTransaction > 0 {

	   	fetchedTransactionLevel1, err := conn.NewTransaction().Retrieve(transaction.ParentTransaction)

	   	if err != nil {
	   		fmt.Println("Error fetching parent transaction 1 => ", fetchedTransactionLevel1)
	   		continue
		} else if fetchedTransactionLevel1 == nil {
			fmt.Println("No data for fetchedTransactionLevel1")
			continue
		} else if fetchedTransactionLevel1.ParentTransaction == 0 {
			fmt.Println("No ParentTransactionLevel2")
			continue
		}

		fetchedTransactionLevel2, err := conn.NewTransaction().Retrieve(fetchedTransactionLevel1.ParentTransaction)

		   if err != nil {
			   fmt.Println("Error fetching parent transaction 2 => ", fetchedTransactionLevel1)
			   continue
		   } else if fetchedTransactionLevel2 == nil {
			   fmt.Println("No data for fetchedTransactionLevel2")
			   continue
		   } else if fetchedTransactionLevel2.Metadata == nil {
		   	   fmt.Println("No metadata for fetchedTransactionLevel2")
		   	   continue
		   }

		   val, ok := fetchedTransactionLevel2.Metadata["location"]

		   if ok {
			   err = f.SetCellValue("Sheet1", "E"+strconv.Itoa(i+2), val.(string))

			   if err != nil {
				   fmt.Println(err)
				   continue
			   }
		   }

	   }

	}

	if err := f.SaveAs(*fileLocation); err != nil {
		panic(err)
	}

	fmt.Println("Overpayments successfully saved to ", *fileLocation)
}

