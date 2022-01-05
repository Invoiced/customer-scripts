package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/Invoiced/invoiced-go"
	"os"
	"strings"
)

//This program will disable all of the credit card payments.

func main() {
	//declare and init command line flags
	prodEnv := true
	key := flag.String("key", "", "api key in Settings > Developer")
	environment := flag.String("env", "", "your environment production or sandbox")

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

	fmt.Println("Please confirm, this program is about disable credit card payments for all customers, please type in YES to continue: ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(confirm)

	if confirm != "YES" {
		fmt.Println("Halting program, sequence not confirmed")
		return
	}

	conn := invdapi.NewConnection(*key, prodEnv)

	customerConn := conn.NewCustomer()

	customers, err := customerConn.ListAll(nil, nil)

	if err != nil {
		panic("Error fetching customers => " + err.Error())
	}

	for _, customer := range customers {
		tmpCustToUpdate := conn.NewCustomer()
		tmpCustToUpdate.Id = customer.Id
		tmpCustToUpdate.DisabledPaymentMethods = append([]string{}, "credit_card")
		fmt.Println("Disabling credit card for customer # ", customer.Number)
		err := tmpCustToUpdate.Save()
		if err != nil {
			fmt.Println("Error saving customer => ", customer.Number)
		}
		fmt.Println("Successfully disabled credit card for customer # ", customer.Number)
	}

}
