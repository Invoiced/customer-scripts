package main

import (
	"bufio"
	"fmt"
	"github.com/Invoiced/invoiced-go/invdendpoint"
	"github.com/invoiced/invoiced-go"
	"os"
	"strings"
)

//This will issue all draft invoices

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

	fmt.Println("Is this a Production connection => ", prodEnv)

	fmt.Println("Please confirm, this program is about issue all draft invoices, please type in YES to continue: ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(confirm)

	if confirm != "YES" {
		fmt.Println("Halting program, sequence not confirmed")
		return
	}

	conn := invdapi.NewConnection(key, prodEnv)

	//filter := invdendpoint.NewFilter()
	//filter.Set("draft","false")

	fmt.Println("Fetching draft invoices")

	filter := invdendpoint.NewFilter()
	filter.Set("status","draft")

	invoices, err := conn.NewInvoice().ListAll(filter, nil)

	if err != nil {
		panic("could not fetch draft invoices")
	}

	fmt.Println("Fetched ", len(invoices), ", draft invoices to issue.")

	for _, invoice := range invoices {
		if invoice.Draft == true {
			fmt.Println("Issuing invoice ", invoice.Number)
			invToUpdate := conn.NewInvoice()
			invToUpdate.Id = invoice.Id
			draft := false
			invToUpdate.Draft = draft
			invToUpdate.Sent = true
			err := invToUpdate.Save()

			if err != nil {
				fmt.Println("Error updating draft invoice ", invoice.Number)
			}
		}

	}

}
