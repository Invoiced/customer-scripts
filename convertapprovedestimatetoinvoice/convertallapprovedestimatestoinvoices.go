package main

import (
	"bufio"
	"fmt"
	"github.com/Invoiced/invoiced-go/v2"
	"github.com/Invoiced/invoiced-go/v2/api"
	"os"
	"strings"
)

//This program will generate invoices for all approved estimates
func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Please enter your API Key: ")
	sandBoxEnv := true
	key, _ := reader.ReadString('\n')
	key = strings.TrimSpace(key)
	for {

		fmt.Println("What is your environment, please enter P for production or S for sandbox: ")
		env, _ := reader.ReadString('\n')
		env = strings.ToUpper(strings.TrimSpace(env))

		if env == "P" || strings.Contains(env, "PRODUCTION") {
			sandBoxEnv = false
			fmt.Println("Using Production for the environment")
			break
		} else if env == "S" || strings.Contains(env, "SANDBOX") {
			fmt.Println("Using Sandbox for the environment")
			break
		}
	}

	fmt.Println("This program will generate an invoice for all approved estimates")

	client := api.New(key, sandBoxEnv)

	filter := invoiced.NewFilter()
	filter.Set("status", "approved")

	estimates, err := client.Estimate.ListAll(filter,nil)

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

		inv, err := client.Estimate.GenerateInvoice(estimate.Id)
		if err != nil {
			fmt.Println("Error generating invoice for estimate #" + estimate.Number + ", got error => " + err.Error())
			continue
		}
		fmt.Println("Generated invoice # " + inv.Number + " for estimate #" + estimate.Number)
	}

	fmt.Println("Successfully generated invoices for estimates")

}
