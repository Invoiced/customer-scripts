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

//This program generates a excel file that exports out the events

func main() {
	//declare and init command line flags
	sandboxEnv := true
	key := flag.String("key", "", "api key in Settings > Developer")
	environment := flag.String("env", "", "your environment production or sandbox")
	startdate := flag.String("startdate", "", "Your start date for the event period in MMDDYYYY format")
	enddate := flag.String("enddate", "", "Your end date for the event period in MMDDYYYY format")
	eventType := flag.String("eventType", "", "What event are filtering on ie invoice.created, invoiced.updated")
	invoicedUser := flag.String("invoicedUser", "", "What event are filtering on ie invoice.created, invoiced.updated")

	fileLocation := flag.String("file", "", "specify your excel file")
	flag.Parse()

	reader := bufio.NewReader(os.Stdin)

	if *key == "" {
		fmt.Print("Please enter your API Key: ")
		*key, _ = reader.ReadString('\n')
		*key = strings.TrimSpace(*key)
	}

	*environment = strings.ToUpper(strings.TrimSpace(*environment))

	if *environment == "P" || strings.Contains(*environment, "PROD") {
		sandboxEnv = false
		fmt.Println("Using Production for the environment")
	} else if *environment == "S" || strings.Contains(*environment, "SAND") {
		fmt.Println("Using Sandbox for the environment")
	} else {
		for {

			fmt.Println("What is your environment, please enter P for production or S for sandbox: ")
			env, _ := reader.ReadString('\n')
			env = strings.ToUpper(strings.TrimSpace(env))

			if env == "P" || strings.Contains(env, "PRODUCTION") {
				sandboxEnv = false
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

	if *eventType == "" {
			fmt.Print("Please enter your event type: ")
			*eventType, _ = reader.ReadString('\n')
			*eventType = strings.TrimSpace(*eventType)
	}

	if *invoicedUser == "" {
			fmt.Print("Please enter your user type: ")
			*invoicedUser, _ = reader.ReadString('\n')
			*invoicedUser = strings.TrimSpace(*invoicedUser)
	}

	fmt.Println(beginTime.Unix())
	fmt.Println(endTime.Unix())
	fmt.Println(sandboxEnv)

	conn := invdapi.NewConnection(*key, sandboxEnv)

	filter := invdendpoint.NewFilter()
	filter.Set("type", *eventType)

	fmt.Println("Getting events ...")
	events, err := conn.NewEvent().ListAllByDatesAndUser(filter,nil,beginTime.Unix(),endTime.Unix(),*invoicedUser,"",-1)

	if err != nil {
		fmt.Println("Got error fetching events, err => ",err)
		return
	}

	f := excelize.NewFile()
	// Create a new sheet.
	index := f.NewSheet("Sheet1")
	// Set value of a cell.
	f.SetActiveSheet(index)
	// Save xlsx file by the given path.

	err = f.SetCellValue("Sheet1", "A1", "Event ID")
	if err != nil {
		panic(err)
	}
	err = f.SetCellValue("Sheet1", "B1", "Event Type")
	if err != nil {
		panic(err)
	}
	err = f.SetCellValue("Sheet1", "C1", "Event Date")
	if err != nil {
		panic(err)
	}
	err = f.SetCellValue("Sheet1", "D1", "Event Timestamp")
	if err != nil {
		panic(err)
	}

	for i, event := range events {
		err = f.SetCellValue("Sheet1", "A" + strconv.Itoa(i + 2), event.Id)
		if err != nil {
			panic(err)
		}
		err = f.SetCellValue("Sheet1", "B"+ strconv.Itoa(i + 2), event.Type)
		if err != nil {
			panic(err)
		}
		err = f.SetCellValue("Sheet1", "C"+ strconv.Itoa(i + 2), string(event.Data))
		if err != nil {
			panic(err)
		}
		err = f.SetCellValue("Sheet1", "D"+ strconv.Itoa(i + 2), event.Timestamp)
		if err != nil {
			panic(err)
		}


	}

	fmt.Println("Saving excel file " + *fileLocation)

	if err := f.SaveAs(*fileLocation); err != nil {
		fmt.Println("Error saving excel file -> ",err)
	}
}

