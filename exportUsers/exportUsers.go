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
	"time"
)

//This program generates a excel file that exports out the users

func main() {
	//declare and init command line flags
	sandboxEnv := true
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

	if *environment == "" {
		fmt.Println("Enter P for Production, S for Sandbox: ")
		*environment, _ = reader.ReadString('\n')
		*environment = strings.TrimSpace(*environment)
	}

	if *environment == "P" {
		sandboxEnv = false
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

	conn := invdapi.NewConnection(*key, sandboxEnv)

	fmt.Println("Getting users ...")

	roles, err := conn.NewRole().ListAll(nil,nil)

	if err != nil {
		fmt.Println("Got error fetching roles, err => ",err)
		return
	}

	roleDict := make(map[string]string)

	for _, role := range roles {
		roleDict[role.Id] = role.Name
	}


	users, err := conn.NewUser().ListAll(nil,nil)

	if err != nil {
		fmt.Println("Got error fetching users, err => ",err)
		return
	}

	f := excelize.NewFile()
	// Create a new sheet.
	index := f.NewSheet("Sheet1")
	// Set value of a cell.
	f.SetActiveSheet(index)
	// Save xlsx file by the given path.

	err = f.SetCellValue("Sheet1", "A1", "User Name")
	if err != nil {
		panic(err)
	}
	err = f.SetCellValue("Sheet1", "B1", "User Email")
	if err != nil {
		panic(err)
	}
	err = f.SetCellValue("Sheet1", "C1", "User Role")
	if err != nil {
		panic(err)
	}
	err = f.SetCellValue("Sheet1", "D1", "User Restrictions")
	if err != nil {
		panic(err)
	}
	err = f.SetCellValue("Sheet1", "E1", "User Last Sign In")
	if err != nil {
		panic(err)
	}

	for i, userdata := range users {
		err = f.SetCellValue("Sheet1", "A" + strconv.Itoa(i + 2), userdata.User.FirstName + userdata.User.LastName)
		if err != nil {
			fmt.Println(err)
		}
		err = f.SetCellValue("Sheet1", "B"+ strconv.Itoa(i + 2), userdata.User.Email)
		if err != nil {
			fmt.Println(err)
		}
		err = f.SetCellValue("Sheet1", "C"+ strconv.Itoa(i + 2), roleDict[userdata.Role])
		if err != nil {
			fmt.Println(err)
		}
		err = f.SetCellValue("Sheet1", "D"+ strconv.Itoa(i + 2), userdata.RestrictionMode)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println("lastsignedin",userdata.LastSignedIn)
		err = f.SetCellValue("Sheet1", "E"+ strconv.Itoa(i + 2), time.Unix(userdata.LastSignedIn,0).String())
		if err != nil {
			fmt.Println(err)
		}

	}

	fmt.Println("Saving excel file " + *fileLocation)

	if err := f.SaveAs(*fileLocation); err != nil {
		fmt.Println("Error saving excel file -> ",err)
	}
}

