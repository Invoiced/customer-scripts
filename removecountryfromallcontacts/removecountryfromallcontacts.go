package main

import (
	"bufio"
	"fmt"
	"github.com/Invoiced/invoiced-go/invdendpoint"
	"github.com/invoiced/invoiced-go"
	"os"
	"strings"
)

//This program will remove country from all contacts.
//Be careful using this in production

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Please enter your API Key: ")
	prodEnv := true
	key, _ := reader.ReadString('\n')

	for {

		fmt.Println("What is your environment, please enter P for production or S for sandbox: ")
		env, _ := reader.ReadString('\n')
		key = strings.TrimSpace(key)
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

	fmt.Println("Please confirm, this program is about to remove all countries from your contacts, please type in YES to continue: ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(confirm)

	if confirm != "YES" {
		fmt.Println("Halting program, confirm sequence not confirmed")
		return
	}

	conn := invdapi.NewConnection(key, prodEnv)

	customerConn := conn.NewCustomer()

	fmt.Println("Fetching all the customers")

	customers, err := customerConn.ListAll(nil, nil)

	if err != nil {
		panic("Could not fetch customers => " + err.Error())
	}

	fmt.Println("Number of customers to remove countries from the contact section", len(customers))

	for _, customer := range customers {
		contacts, err := customer.ListAllContacts()
		if err != nil {
			fmt.Println("Error getting contact for customer -> ", customer.Name)
		}

		for _, contact := range contacts {
			contactToUpdate := new(invdendpoint.Contact)
			contactToUpdate.Id = contact.Id
			contactToUpdate.Country = ""
			fmt.Println("Removing country from contact with email ", contact.Email)
			_, err := customer.UpdateContact(contactToUpdate)

			if err != nil {
				fmt.Println("Error updating contact with email ", contact.Email, "got error -> ", err)
			} else {
				fmt.Println("Successfully removed country from contact with email ", contact.Email)
			}

		}
	}

}
