package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/Invoiced/invoiced-go/v2"
	"github.com/Invoiced/invoiced-go/v2/api"
	"github.com/xuri/excelize/v2"
	"os"
	"strconv"
	"strings"
	"time"
)

//This program outputs the user activity for a user

const (
	CustomerActivity = "Customer Activity"
	TaskActivity = "Task Activity"
	NoteActivity = "Note Activity"
	InvoiceActivity = "Invoice Activity"
	PaymentActivity = "Payment Activity"
	SubscriptionActivity = "Subscription Activity"
	EstimateActivity = "Estimate Activity"
)

func main() {
	sandBoxEnv := true
	key := flag.String("key", "", "api key in Settings > Developer")
	environment := flag.String("env", "", "enter P for production or S for sandbox")
	fileLocation := flag.String("file", "", "specify your excel file name")
	dateToRun := flag.String("date","","date to run the report for in YYYY-MM-DD format")

	flag.Parse()

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("This program will generate a daily activity report for all the users")

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
		*fileLocation = "user_activity_report_"
	}

	if *dateToRun == "" {
		*dateToRun = time.Now().Format("2006-01-02")
	}

	client := api.New(*key, sandBoxEnv)

	members, err := client.Member.ListAll(nil,nil)

	fmt.Println(*dateToRun)

	if err != nil {
		panic(err)
	}

	f := excelize.NewFile()
	// Create a new sheet.
	index := f.NewSheet("Sheet1")
	// Set value of a cell.
	f.SetActiveSheet(index)
	// Save xlsx file by the given path.

	err = f.SetCellValue("Sheet1", "A1", "Report Date")
	if err != nil {
		panic(err)
	}

	err = f.SetCellValue("Sheet1", "B1", "User Name")
	if err != nil {
		panic(err)
	}
	err = f.SetCellValue("Sheet1", "C1", "User Email")
	if err != nil {
		panic(err)
	}
	err = f.SetCellValue("Sheet1", "D1", CustomerActivity)
	if err != nil {
		panic(err)
	}
	err = f.SetCellValue("Sheet1", "E1", TaskActivity)
	if err != nil {
		panic(err)
	}
	err = f.SetCellValue("Sheet1", "F1", NoteActivity)
	if err != nil {
		panic(err)
	}

	err = f.SetCellValue("Sheet1", "G1", InvoiceActivity)
	if err != nil {
		panic(err)
	}

	err = f.SetCellValue("Sheet1", "H1", PaymentActivity)
	if err != nil {
		panic(err)
	}


	err = f.SetCellValue("Sheet1", "I1", SubscriptionActivity)
	if err != nil {
		panic(err)
	}

	err = f.SetCellValue("Sheet1", "J1", EstimateActivity)
	if err != nil {
		panic(err)
	}

	for i, member := range members {
		userActivity := make(map[string]int)
		InitializeActivityMap(userActivity)

		memberID := member.User.Id

		fmt.Println(memberID)

		startDate,EndDate, err := StartEndTimestampsForDay(*dateToRun)

		if err != nil {
			panic("Error parsing date for "+*dateToRun)
		}

		events, err := client.Event.ListAllByDatesAndUser(nil,nil,startDate,EndDate,strconv.FormatInt(memberID,10),"",-1)

		if err != nil {
			fmt.Println("error retrieving event for user = ",member.User.Email)
		}

		fmt.Println("number of event ",len(events), "for user ",member.User.Email)

		ProcessEventActivity(events,userActivity)

		err = WriteEventsToExcel(f,userActivity,i,member.User.Email,member.User.FirstName + " " + member.User.LastName,*dateToRun)

		if err != nil {
			fmt.Println("Error writing ",member.User.Email, "activity to excel, err -> ",err)
		}

	}

	if err := f.SaveAs(*fileLocation+*dateToRun+".xlsx"); err != nil {
		fmt.Println("Error saving excel file -> ",err)
	}

}

func WriteEventsToExcel(f *excelize.File, userActivity map[string]int,row int,userEmail string, userName string, reportDate string) error {

	err := f.SetCellValue("Sheet1", "A" +  strconv.Itoa(row + 2), reportDate)
	if err != nil {
		return err
	}
	err = f.SetCellValue("Sheet1", "B" +  strconv.Itoa(row + 2), userName)
	if err != nil {
		return err
	}
	err = f.SetCellValue("Sheet1", "C" +strconv.Itoa(row + 2), userEmail)
	if err != nil {
		return err
	}
	err = f.SetCellValue("Sheet1", "D" + strconv.Itoa(row + 2), userActivity[CustomerActivity])
	if err != nil {
		return err
	}
	err = f.SetCellValue("Sheet1", "E" + strconv.Itoa(row + 2), userActivity[TaskActivity])
	if err != nil {
		return err
	}
	err = f.SetCellValue("Sheet1", "F" + strconv.Itoa(row + 2), userActivity[NoteActivity])
	if err != nil {
		return err
	}

	err = f.SetCellValue("Sheet1", "G" + strconv.Itoa(row + 2), userActivity[InvoiceActivity])
	if err != nil {
		return err
	}

	err = f.SetCellValue("Sheet1", "H" + strconv.Itoa(row + 2), userActivity[PaymentActivity])
	if err != nil {
		return err
	}


	err = f.SetCellValue("Sheet1", "I" + strconv.Itoa(row + 2), userActivity[SubscriptionActivity])
	if err != nil {
		return err
	}

	err = f.SetCellValue("Sheet1", "J" + strconv.Itoa(row + 2), userActivity[EstimateActivity])
	if err != nil {
		return err
	}

	return nil

}

func ProcessEventActivity(events invoiced.Events, userActivity map[string]int) {

	for _, event := range events {

		if event.Type == "invoice.created" ||  event.Type == "invoice.updated" || event.Type == "invoice.deleted" || event.Type == "invoice.payment_expected" || event.Type == "invoice.paid" {
			userActivity[InvoiceActivity] += 1
		} else if event.Type == "task.created" ||  event.Type == "task.updated" || event.Type == "task.deleted" || event.Type == "task.completed" {
			userActivity[TaskActivity] += 1
		} else if event.Type == "customer.created" ||  event.Type == "customer.updated" || event.Type == "customer.deleted" || event.Type == "customer.merged" {
			userActivity[CustomerActivity] += 1
		} else if event.Type == "note.created" ||  event.Type == "note.updated" || event.Type == "note.deleted" {
			userActivity[NoteActivity] += 1
		} else if event.Type == "payment.created" ||  event.Type == "payment.updated" || event.Type == "payment.deleted" {
			userActivity[PaymentActivity] += 1
		}  else if event.Type == "subscription.created" ||  event.Type == "subscription.updated" || event.Type == "subscription.deleted" {
			userActivity[SubscriptionActivity] += 1
		} else if event.Type == "estimate.created" ||  event.Type == "estimate.updated" || event.Type == "estimate.deleted" || event.Type == "estimate.approved"{
			userActivity[SubscriptionActivity] += 1
		}
	}

}

func StartEndTimestampsForDay(day string)(int64,int64,error) {

	location := time.Now().Location()
	t, err := time.ParseInLocation("2006-01-02", day,location)

	if err != nil {
		return -1,-1,err
	}

	return t.Unix(),t.Unix()+86400-1, nil
}

func InitializeActivityMap(userActivity map[string]int) {
	userActivity[InvoiceActivity] = 0
	userActivity[SubscriptionActivity] = 0
	userActivity[PaymentActivity] = 0
	userActivity[CustomerActivity] = 0
	userActivity[EstimateActivity] = 0
	userActivity[TaskActivity] = 0
	userActivity[NoteActivity] = 0

}