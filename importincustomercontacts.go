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

//This program will mark the customers in the excel sheet as paid.

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

	fmt.Println("Is this a Production connection? => ",prodEnv)

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

	for i, row := range rows {

		if i == 0 {
			fmt.Println("Skipping header row")
			continue
		}

		customerNumber := strings.TrimSpace(row[columnIndex])
		fmt.Println("customerNumber=>",customerNumber)

		filter := invdendpoint.NewFilter()
		filter.Set("number",customerNumber)

		customers, err := conn.NewCustomer().ListAll(filter,nil)

		if err != nil {
			fmt.Println("Error getting customer with number -> ",customerNumber, ", skipping.  Error => ",err)
			continue
		}

		if customers == nil || len(customers) == 0{
			fmt.Println("Could not retrieve customer with number -> ",customerNumber)
			continue
		}

		emails := strings.Split( strings.TrimSpace(row[3]),",")

		if len(emails) == 1 {
			fmt.Println("Adding contact with email = ",emails[0])
			contactToAdd := new(invdendpoint.Contact)
			contactToAdd.Name = emails[0]
			contactToAdd.Email = emails[0]
			contactToAdd.Phone = strings.TrimSpace(row[4])

			customer := customers[0]

			_, err = customer.CreateContact(contactToAdd)

			if err != nil {
				fmt.Println("Could not add contact, for customer => ",customer.Number,", got error =>  ",err)
				continue
			}


			fmt.Println("Successfully added contact ", contactToAdd.Name)
		} else if len(emails) > 1 {
			for _, email := range emails {
				fmt.Println("Adding contact with email = ",emails[0])
				contactToAdd := new(invdendpoint.Contact)
				contactToAdd.Name = email
				contactToAdd.Email = email
				contactToAdd.Phone = strings.TrimSpace(row[4])

				customer := customers[0]

				_, err = customer.CreateContact(contactToAdd)

				if err != nil {
					fmt.Println("Could not add contact, for customer => ",customer.Number,", got error =>  ",err)
					continue
				}


				fmt.Println("Successfully added contact ", contactToAdd.Name)


			}


		}




	}



}
