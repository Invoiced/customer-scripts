package main

import (
	"bufio"
	"fmt"
	"github.com/Invoiced/invoiced-go/invdendpoint"
	"github.com/invoiced/invoiced-go"
	"os"
	"strings"
)

//This program will generate invoices for all approved estimates
func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Please enter your API Key: ")
	sandboxEnv := true
	key, _ := reader.ReadString('\n')
	key = strings.TrimSpace(key)
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

	fmt.Println("Is this a Production connection? => ", !sandboxEnv)

	fmt.Println("This program will generate an invoice for all approved estimates")

	conn := invdapi.NewConnection(key, sandboxEnv)

	filter := invdendpoint.NewFilter()
	filter.Set("status", "approved")

	estimates, err := conn.NewEstimate().ListAll(filter, nil)

	if err != nil {
		fmt.Println("Error getting estimates error => ", err)
		return
	}

	if estimates == nil || len(estimates) == 0 {
		fmt.Println("No estimates to process")
		return
	}

	fmt.Println("Successfully fetched estimates")

	for _, estimate := range estimates {

		fmt.Println("Generating invoice for estimate #" + estimate.Number)
		inv, err := estimate.GenerateInvoice()
		if err != nil {
			fmt.Println("Error generating invoice for estimate #" + estimate.Number + ", got error => " + err.Error())
			continue
		}
		fmt.Println("Generated invoice # " + inv.Number + " for estimate #" + estimate.Number)
	}

	fmt.Println("Successfully generated invoices for estimates")

}
