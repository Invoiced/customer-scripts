package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/Invoiced/invoiced-go"
	"github.com/Invoiced/invoiced-go/invdendpoint"
	"io"
	"net/http"
	"os"
	"strings"
)

//This program exports out open invoice pdf belonging to customers with a certain metadata

func main() {
	//declare and init command line flags
	sandboxEnv := true
	key := flag.String("key", "", "api key in Settings > Developer")
	environment := flag.String("env", "", "your environment production or sandbox")
	customerMetadataKey := flag.String("customermetadatakey","","specify the customer metadata key here")
	customerMetadataValue := flag.String("customermetadatavalue","","specify the customer metadata value here")
	flag.Parse()

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("This program will filter your open invoices based on customer metadata value and download them as pdfs")

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
		sandboxEnv = false
		fmt.Println("Using Production for the environment")
	} else if *environment == "S" {
		fmt.Println("Using Sandbox for the environment")
	} else {
		fmt.Println("Unrecognized value ", *environment, ", enter P or S only")
		return
	}


	if *customerMetadataKey == "" {
		fmt.Println("Please specify your customer metadata key here: ")
		*customerMetadataKey, _ = reader.ReadString('\n')
		*customerMetadataKey = strings.TrimSpace(*customerMetadataKey)
	}

	if *customerMetadataValue == "" {
		fmt.Println("Please specify your customer metadata value here: ")
		*customerMetadataValue, _ = reader.ReadString('\n')
		*customerMetadataValue = strings.TrimSpace(*customerMetadataValue)
	}

	conn := invdapi.NewConnection(*key, sandboxEnv)


	filterCustomer := invdendpoint.NewMetadataFilter()
	filterCustomer.Set( *customerMetadataKey,*customerMetadataValue)

	customers, err := conn.NewCustomer().ListAll(filterCustomer,nil)

	if err != nil {
		fmt.Println("Got error fetching customers, err => ",err)
		return
	}


	for _,customer:= range customers {
		filterInvoice := invdendpoint.NewFilter()
		filterInvoice.Set("paid","0")
		filterInvoice.Set("closed","0")
		filterInvoice.Set("draft","0")
		filterInvoice.Set("customer",customer.Id)

		invoices, err := conn.NewInvoice().ListAll(filterInvoice,nil)

		if err != nil {
			fmt.Println("Got error fetching invoices, err => ",err)
			return
		}

		for _, invoice := range invoices {
			err := DownloadFile(invoice.Number + ".pdf",invoice.PdfUrl)
			if err != nil {
				fmt.Println("Got error downloading pdf ", invoice.PdfUrl,err)
			}
		}
	}

}

func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}