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
)

//This program will import in customer contacts

const sheet = "Sheet1"

func main() {
	sandBoxEnv := true
	key := flag.String("key", "", "api key in Settings > Developer")
	environment := flag.String("env", "", "your environment production or sandbox")
	fileLocation := flag.String("file", "", "specify your excel file")

	flag.Parse()

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("This program import in users")

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

	userEmailIndex := "A"
	userFirstNameIndex := "B"
	userLastNameIndex := "C"
	userRoleIndex := "D"
	userRestrictByIndex := "E"
	userRestrictionCustomFieldIndex := "F"
	userRestrictionCustomFieldValueIndex := "G"


	rows, err := f.GetRows(sheet)

	if err != nil {
		panic("Error trying to get rows for the sheet" + err.Error())
	}


	conn := invdapi.NewConnection(*key, sandBoxEnv)
	userConn := conn.NewUser()


	for i, row := range rows {

		if i == 0 {
			fmt.Println("Skipping header row")
			fmt.Println(row)
			continue
		}

		userEmail, _ := f.GetCellValue(sheet,userEmailIndex + strconv.Itoa(i + 1))
		userFirstName,_ :=  f.GetCellValue(sheet,userFirstNameIndex + strconv.Itoa(i + 1))
		userLastName,_ :=  f.GetCellValue(sheet,userLastNameIndex + strconv.Itoa(i + 1))
		userRole,_ := f.GetCellValue(sheet,userRoleIndex + strconv.Itoa(i + 1))
		userRestrictBy,_ :=  f.GetCellValue(sheet,userRestrictByIndex + strconv.Itoa(i + 1))
		userRestrictionCustomField,_ :=  f.GetCellValue(sheet,userRestrictionCustomFieldIndex + strconv.Itoa(i + 1))
		userRestrictionCustomFieldValue, _:=  f.GetCellValue(sheet,userRestrictionCustomFieldValueIndex + strconv.Itoa(i + 1))

		fmt.Println("Add user with email address = ", userEmail)
		userReq := new(invdendpoint.UserRequest)
		userReq.Email = userEmail
		userReq.FirstName = userFirstName
		userReq.LastName = userLastName
		userReq.Role = userRole
		if len(userRestrictBy) > 0 {
			userReq.RestrictionMode = userRestrictBy

			if userRestrictBy == "custom_field" {
				userReq.Restrictions = make(map[string][]string)
				userReq.Restrictions[userRestrictionCustomField] = []string{userRestrictionCustomFieldValue}
			}
		}

		createdUser, err := userConn.Create(userReq)

		if err != nil {
			fmt.Println("Got an error while creating user ", userEmail, ", error -> ",err.Error())
		} else {
			fmt.Println("Successfully created user with email = ",createdUser.User.Email, "and id = ",createdUser.Id)
		}


	}

}
