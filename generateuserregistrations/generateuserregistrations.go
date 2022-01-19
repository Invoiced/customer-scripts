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

//This program will generate an excel file with email, first_name,last_name, registration_url

const sheet = "Sheet1"

func main() {
	sandBoxEnv := true
	key := flag.String("key", "", "api key in Settings > Developer")
	environment := flag.String("env", "", "your environment production or sandbox")
	fileLocation := flag.String("file", "", "specify your excel file")

	flag.Parse()

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("This program will generate user registrations")

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

	conn := invdapi.NewConnection(*key, sandBoxEnv)

	//Load a list of all users
	userConn := conn.NewUser()

	users, err := userConn.ListAll(nil,nil)

	if err != nil {
		fmt.Println("Got an error fetching users => ",err)
		return
	} else {
		fmt.Println("Fetched",len(users), "users")
	}

	f := excelize.NewFile()
	// Create a new sheet.
	index := f.NewSheet(sheet)
	// Set value of a cell.
	f.SetActiveSheet(index)
	// Save xlsx file by the given path.

	err = f.SetCellValue(sheet, "A1", "First Name")
	if err != nil {
		panic(err)
	}

	err = f.SetCellValue(sheet, "B1", "Last Name")
	if err != nil {
		panic(err)
	}

	err = f.SetCellValue(sheet, "C1", "Email")
	if err != nil {
		panic(err)
	}

	err = f.SetCellValue(sheet, "D1", "User Registration URL")
	if err != nil {
		panic(err)
	}

	j := 0
	for _, user := range users {
		if !user.User.Registered {
			err = f.SetCellValue("Sheet1", "A" + strconv.Itoa(j + 2), user.User.FirstName)
			if err != nil {
				panic(err)
			}
			err = f.SetCellValue("Sheet1", "B"+ strconv.Itoa(j + 2), user.User.LastName)
			if err != nil {
				panic(err)
			}
			err = f.SetCellValue("Sheet1", "C"+ strconv.Itoa(j + 2), user.User.Email)
			if err != nil {
				panic(err)
			}
			err = f.SetCellValue("Sheet1", "D"+ strconv.Itoa(j + 2), user.GenerateRegistrationURL())
			if err != nil {
				panic(err)
			}

			j +=1
		}
	}

	if err := f.SaveAs(*fileLocation); err != nil {
		fmt.Println("Error saving excel file -> ",err)
	}

	fmt.Println("Finished generating user registrations")


}

