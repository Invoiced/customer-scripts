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
)

//This program will import in customer contacts

func main() {
	sandBoxEnv := true
	key := flag.String("key", "", "api key in Settings > Developer")
	environment := flag.String("env", "", "your environment production or sandbox")
	fileLocation := flag.String("file", "", "specify your excel file")

	flag.Parse()

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("This program will create invoices with metadata based on the excel file")

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

	*fileLocation = strings.TrimSpace(*fileLocation)

	f, err := excelize.OpenFile(*fileLocation)

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


	conn := invdapi.NewConnection(*key, sandBoxEnv)


	for i, row := range rows {

		if i == 0 {
			fmt.Println("Skipping header row")
			fmt.Println(row)
			continue
		}

		customerNumber := strings.TrimSpace(row[customerNumberIndex])
		if customerNumber == "NONE"  {
			customerNumber = ""
		}
		contactName := strings.TrimSpace(row[contactNameIndex])
		if contactName == "NONE"  {
			contactName = ""
		}

		contactTitle := strings.TrimSpace(row[contactTitleIndex])
		if contactTitle == "NONE"  {
			contactTitle = ""
		}
		contactEmail := strings.TrimSpace(row[contactEmailIndex])
		if contactEmail == "NONE"  {
			contactEmail = ""
		}
		contactPhone := strings.TrimSpace(row[contactPhoneIndex])
		if contactPhone == "NONE"  {
			contactPhone = ""
		}
		contactPrimaryStr := strings.TrimSpace(row[contactPrimaryIndex])
		contactSMSEnabledStr := strings.TrimSpace(row[contactSMSEnabledIndex])
		contactDepartment := strings.TrimSpace(row[contactDepartmentIndex])
		if contactDepartment == "NONE"  {
			contactDepartment = ""
		}
		contactAddress1 := strings.TrimSpace(row[contactAddress1Index])
		if contactAddress1 == "NONE"  {
			contactAddress1 = ""
		}
		contactAddress2 := strings.TrimSpace(row[contactAddress2Index])
		if contactAddress2 == "NONE"  {
			contactAddress2 = ""
		}
		contactCity := strings.TrimSpace(row[contactCityIndex])
		if contactCity == "NONE"  {
			contactCity = ""
		}
		contactState := strings.TrimSpace(row[contactStateIndex])
		if contactState == "NONE"  {
			contactState = ""
		}
		contactPostalCode := strings.TrimSpace(row[contactPostalCodeIndex])
		if contactPostalCode == "NONE"  {
			contactPostalCode = ""
		}
		contactCountry := strings.TrimSpace(row[contactCountryIndex])
		if contactCountry == "NONE"  {
			contactCountry = ""
		}

		if contactName == ""  {
			contactName = contactEmail
		}

		contactPrimary := false
		contactSMSEnabled := false

		//invoiced will make the uppercase email lowercase so to void duplicates, just make it lowercase
		contactEmail = strings.ToLower(contactEmail)

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
