package main

import (
	"bufio"
	"fmt"
	"github.com/invoiced/invoiced-go"
	"os"
	"strings"
)

//This program will delete all the plans in the account
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

	fmt.Println("Please confirm, this program is about to delete all of the plans, please type in YES to continue: ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(confirm)

	if confirm != "YES" {
		fmt.Println("Halting program, confirm sequence not confirmed")
		return
	}

	conn := invdapi.NewConnection(key, prodEnv)

	planConn := conn.NewPlan()

	fmt.Println("Fetching all the plans to delete")

	plans, err := planConn.ListAll(nil, nil)

	if err != nil {
		panic("Could not fetch customers => " + err.Error())
	}

	fmt.Println("Number of plans to delete", len(plans))

	for _, plan := range plans {
		fmt.Println("Deleting plan => ", plan.Name)
		err := plan.Delete()
		if err != nil {
			fmt.Println("Could not delete plan => ",  plan.Name, ", due to the following error => ", err)
			continue
		}

		fmt.Println("Deleted plan => ",  plan.Name)

	}

}
