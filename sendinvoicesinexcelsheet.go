package main


import (
	"bufio"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/invoiced/invoiced-go"
	"os"
	"strings"
	"flag"
)

//This program will send all invoices in the excel sheet

func main() {
	//declare and init command line flags
	prodEnv := true
	key := flag.String("key","","api key in Settings > Developer")
	environment := flag.String("env","","your environment production or sandbox")
	fileLocation := flag.String("file","","specify your excel file")
	flag.Parse()

	reader := bufio.NewReader(os.Stdin)

	if *key == "" {
		fmt.Print("Please enter your API Key: ")
		*key, _ = reader.ReadString('\n')
		*key = strings.TrimSpace(*key)
	}

	*environment = strings.ToUpper(strings.TrimSpace(*environment))


	if *environment == "P" || strings.Contains(*environment, "PRODUCTION") {
		prodEnv = false
		fmt.Println("Using Production for the environment")
	} else if *environment == "S" || strings.Contains(*environment, "SANDBOX") {
		fmt.Println("Using Sandbox for the environment")
	} else {
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

	}

	if *fileLocation == "" {
		fmt.Println("Please specify your excel file: ")
		*fileLocation, _ = reader.ReadString('\n')
		*fileLocation = strings.TrimSpace(*fileLocation)
	}

	fmt.Println("Opening Excel File => ",*fileLocation )
	f, err := excelize.OpenFile(*fileLocation)

	if err != nil {
		panic(err)
	}

	fmt.Println("Read in excel file ", *fileLocation, ", successfully")

	columnIndex := 0

	rows, err := f.GetRows("Sheet1")

	if err != nil {
		panic("Error trying to get rows for the sheet" + err.Error())
	}


	fmt.Println("Please confirm, this program is about to send out the all of those invoices specified in the excel file, through email, please type in YES to continue: ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(confirm)

	if confirm != "YES" {
		fmt.Println("Halting program, sequence not confirmed")
		return
	}

	conn := invdapi.NewConnection(*key, prodEnv)

	for _, row := range rows {

		invoiceNumber := strings.TrimSpace(row[columnIndex])

		fmt.Println("Getting invoice with number => ",invoiceNumber)

		inv, err := conn.NewInvoice().ListInvoiceByNumber(invoiceNumber)

		if err != nil {
			fmt.Println("Error getting invoice with number => ",invoiceNumber, ", error => ", err)
			continue
		}

		if inv == nil {
			fmt.Println("Invoice does not exist =>",invoiceNumber)
			continue
		}

		if inv.Status != "not_sent" {
			fmt.Println("Invoice is already sent, paid, voided, viewed , moving on to next invoice ...")
			continue
		}


		fmt.Println("Sending invoice with number => ", invoiceNumber)

		_, err = inv.SendEmail(nil)

		if err != nil {
			fmt.Println("Could not send out the invoice due the following error => ", err)
			continue
		}

		fmt.Println("Successfully queued invoice => ",  inv.Number  ,"for sending, it should be sent soon. ")


	}



}
