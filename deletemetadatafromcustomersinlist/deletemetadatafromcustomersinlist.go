package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/Invoiced/invoiced-go"
	"os"
	"strconv"
	"strings"
)

//This program will import in customer contacts

const sheet = "Sheet1"

func main() {
	sandBoxEnv := true
	key := flag.String("key", "", "api key in Settings > Developer")
	environment := flag.String("env", "", "your environment production or sandbox")
	fileLocation := flag.String("file", "", "specify your excel file")
	metadataKeyToDelete := flag.String("metadatakey","", "specify you metadata key")

	flag.Parse()

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("This program deleted specified metadata from users in customer list")

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

	if *metadataKeyToDelete == "" {
		fmt.Println("Please specify the metadatakeytodelete: ")
		*metadataKeyToDelete, _ = reader.ReadString('\n')
		*metadataKeyToDelete = strings.TrimSpace(*metadataKeyToDelete)
	}

	if *fileLocation == "" {
		fmt.Println("Please specify your excel file: ")
		*fileLocation, _ = reader.ReadString('\n')
		*fileLocation = strings.TrimSpace(*fileLocation)
	}

	*fileLocation = strings.TrimSpace(*fileLocation)

	f, err := excelize.OpenFile(*fileLocation)

	if err != nil {
		panic(err)
	}

	fmt.Println("Read in excel file ", fileLocation, ", successfully")

	customerNumber:= "A"

	rows, err := f.GetRows(sheet)

	if err != nil {
		panic("Error trying to get rows for the sheet" + err.Error())
	}


	conn := invdapi.NewConnection(*key, sandBoxEnv)
	customerConn := conn.NewCustomer()


	for i, row := range rows {

		if i == 0 {
			fmt.Println("Skipping header row")
			fmt.Println(row)
			continue
		}

		customerNumber, err := f.GetCellValue(sheet,customerNumber + strconv.Itoa(i + 1))

		if err != nil {
			fmt.Println("Error getting customer number for row = ", i, ", error => ",err)
			continue
		}

		customer, err := customerConn.ListCustomerByNumber(customerNumber)


		if err != nil {
			fmt.Println("Error fetching customer " , customerNumber, ", error => ",err)
			continue
		}

		customerMetadata := customer.Metadata

		if customerMetadata == nil {
			fmt.Println("Customer " + customerNumber + "has no metadata to delete")
			continue
		}

		_, ok := customerMetadata[*metadataKeyToDelete]

		if !ok {
			fmt.Println("Customer " + customerNumber + "has no metadata to delete for " + *metadataKeyToDelete)
			continue
		}

		custToUpdate := conn.NewCustomer()
		custToUpdate.Id = customer.Id

		delete(customerMetadata,*metadataKeyToDelete)

        custToUpdate.Metadata = customerMetadata

        err =  custToUpdate.Save()

        if err != nil {
        	fmt.Println("Error saving customer " + customerNumber + ", error -> ",err)
        	continue
		} else {
			fmt.Println("Successfully removed the metadata ", *metadataKeyToDelete + ", for user ")
		}




	}

}

