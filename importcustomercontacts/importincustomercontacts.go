package main

import (
	"bufio"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/Invoiced/invoiced-go/invdendpoint"
	"github.com/invoiced/invoiced-go"
	"os"
	"strconv"
	"strings"
)

//This program will import in customer contacts

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

	fmt.Println("Is this a Production connection? => ", !prodEnv)

	fmt.Println("Please specify your excel file: ")
	fileLocation, _ := reader.ReadString('\n')

	fileLocation = strings.TrimSpace(fileLocation)

	f, err := excelize.OpenFile(fileLocation)

	if err != nil {
		panic(err)
	}

	fmt.Println("Read in excel file ", fileLocation, ", successfully")

	customerNumberIndex := 0
	contactNameIndex := 1
	contactTitleIndex := 2
	contactEmailIndex := 3
	contactPhoneIndex := 4
	contactPrimaryIndex := 5
	contactSMSEnabledIndex := 6
	contactDepartmentIndex := 7
	contactAddress1Index := 8
	contactAddress2Index := 9
	contactCityIndex := 10
	contactStateIndex := 11
	contactPostalCodeIndex := 12
	contactCountryIndex := 13

	rows, err := f.GetRows("Sheet1")

	if err != nil {
		panic("Error trying to get rows for the sheet" + err.Error())
	}

	fmt.Println("Please confirm, this program is about import contacts, specified by the customers in the excel file, please type in YES to continue: ")
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

		customerNumber := strings.TrimSpace(row[customerNumberIndex])
		contactName := strings.TrimSpace(row[contactNameIndex])
		contactTitle := strings.TrimSpace(row[contactTitleIndex])
		contactEmail := strings.TrimSpace(row[contactEmailIndex])
		contactPhone := strings.TrimSpace(row[contactPhoneIndex])
		contactPrimaryStr := strings.TrimSpace(row[contactPrimaryIndex])
		contactSMSEnabledStr := strings.TrimSpace(row[contactSMSEnabledIndex])
		contactDepartment := strings.TrimSpace(row[contactDepartmentIndex])
		contactAddress1 := strings.TrimSpace(row[contactAddress1Index])
		contactAddress2 := strings.TrimSpace(row[contactAddress2Index])
		contactCity := strings.TrimSpace(row[contactCityIndex])
		contactState := strings.TrimSpace(row[contactStateIndex])
		contactPostalCode := strings.TrimSpace(row[contactPostalCodeIndex])
		contactCountry := strings.TrimSpace(row[contactCountryIndex])

		contactPrimary := false
		contactSMSEnabled := false

		fmt.Println(contactPrimaryStr,contactSMSEnabledStr)

		contactPrimary, _ = strconv.ParseBool(contactPrimaryStr)
		contactSMSEnabled, _ = strconv.ParseBool(contactSMSEnabledStr)

		fmt.Println("customerNumber=>", customerNumber)

		customer, err := conn.NewCustomer().ListCustomerByNumber(customerNumber)

		if err != nil {
			fmt.Println("Error getting customer with number -> ", customerNumber, ", skipping.  Error =>", err)
			continue
		}

		if customer == nil {
			fmt.Println("Could not retrieve customer with number -> ", customerNumber)
			continue
		}

		fetchedContacts, err := customer.ListAllContacts()

		if err != nil {
			fmt.Println("Error getting contacts for customer => ", customerNumber, "Error =>",err)
			continue
		}

		contactMap := make(map[string]invdendpoint.Contact)

		for _, fetchedContact := range fetchedContacts {
			contactMap[fetchedContact.Email] = fetchedContact
		}

		contactToAddUpdate, contactMatched := contactMap[contactEmail]

		if !contactMatched {
			contactToAddUpdate = invdendpoint.Contact{}
		}

		contactToAddUpdate.Name = contactName
		contactToAddUpdate.Email = contactEmail
		contactToAddUpdate.Title = contactTitle
		contactToAddUpdate.Phone = contactPhone
		contactToAddUpdate.Primary = contactPrimary
		contactToAddUpdate.SmsEnabled = contactSMSEnabled
		contactToAddUpdate.Department = contactDepartment
		contactToAddUpdate.Address1 = contactAddress1
		contactToAddUpdate.Address2 = contactAddress2
		contactToAddUpdate.City = contactCity
		contactToAddUpdate.State = contactState
		contactToAddUpdate.PostalCode = contactPostalCode
		contactToAddUpdate.Country = contactCountry

		if contactMatched {
			_, err := customer.UpdateContact(&contactToAddUpdate)
			if err != nil {
				fmt.Println("Error updating contact with email =>", contactEmail, ", for Customer =>", customerNumber, ", err =>",err)
				continue
			}
			fmt.Println("Successfully updated contact with email =>",contactEmail, ", for Customer =>",customerNumber)
		} else {
			_, err := customer.CreateContact(&contactToAddUpdate)
			if err != nil {
				fmt.Println("Error added contact with email =>", contactEmail, ", for Customer =>", customerNumber, ", err =>",err)
				continue
			}
			fmt.Println("Successfully added contact with email =>",contactEmail, ", for Customer =>",customerNumber)
		}

	}

	}
