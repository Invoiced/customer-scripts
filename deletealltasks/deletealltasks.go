package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/Invoiced/invoiced-go/v2/api"
	"os"
	"strings"
)

func main() {
	// Declare program options
	sandBoxEnv := true
	key := flag.String("key", "", "api key in Settings > Developer")
	environment := flag.String("env", "", "your environment production or sandbox")

	flag.Parse()

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("This program will delete all task")

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

	conn := api.New(*key, sandBoxEnv)

	tasks, err := conn.Task.ListAll(nil, nil)

	if err != nil {
		fmt.Println("Error fetching tasks => ", err)
		return
	}

	fmt.Println("Number of tasks => ", len(tasks))

	fmt.Println("Deleting all tasks")

	for _, task := range tasks {
		err := conn.Task.Delete(task.Id)
		if err != nil {
			fmt.Println("Error deleting task => ", err)
			continue
		}
		fmt.Println("Deleted task => ", task.Id)
	}

}
